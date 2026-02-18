package cmd

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	"github.com/chuxorg/chux-yanzi-core/hash"
	"github.com/chuxorg/chux-yanzi-core/model"
	"github.com/chuxorg/chux-yanzi-core/store"
)

const (
	localMigrationsDir  = "migrations"
	localMigrationName  = "0001_init.sql"
	localProjectName    = "0002_projects.sql"
	localCheckpointName = "0003_checkpoints.sql"
)

const localMigrationSQL = `CREATE TABLE IF NOT EXISTS intents (
	id TEXT PRIMARY KEY,
	created_at TEXT NOT NULL,
	author TEXT NOT NULL,
	source_type TEXT NOT NULL,
	title TEXT,
	prompt TEXT NOT NULL,
	response TEXT NOT NULL,
	meta TEXT,
	prev_hash TEXT,
	hash TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intents_hash ON intents(hash);
CREATE INDEX IF NOT EXISTS idx_intents_created_at ON intents(created_at);
`

const localProjectMigrationSQL = `CREATE TABLE IF NOT EXISTS projects (
	name TEXT PRIMARY KEY,
	description TEXT,
	created_at TEXT NOT NULL,
	prev_hash TEXT,
	hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_projects_prev_hash ON projects (prev_hash);
`

const localCheckpointMigrationSQL = `CREATE TABLE IF NOT EXISTS checkpoints (
	hash TEXT PRIMARY KEY,
	project TEXT NOT NULL,
	summary TEXT NOT NULL,
	created_at TEXT NOT NULL,
	artifact_ids TEXT NOT NULL,
	previous_checkpoint_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_project_created_at ON checkpoints (project, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_checkpoints_previous_id ON checkpoints (previous_checkpoint_id);
`

func openLocalStore(ctx context.Context, cfg config.Config) (*store.Store, error) {
	if cfg.DBPath == "" {
		return nil, errors.New("db_path is required when mode=local")
	}
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	if err := ensureLocalMigrations(dir); err != nil {
		return nil, err
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	if err := withDir(dir, func() error {
		return st.Migrate(ctx)
	}); err != nil {
		_ = st.Close()
		return nil, err
	}

	return st, nil
}

func openSQLiteDB(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ensureLocalMigrations(stateDir string) error {
	path := filepath.Join(stateDir, localMigrationsDir)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}
	migrations := map[string]string{
		localMigrationName:  localMigrationSQL,
		localProjectName:    localProjectMigrationSQL,
		localCheckpointName: localCheckpointMigrationSQL,
	}
	for name, contents := range migrations {
		file := filepath.Join(path, name)
		if _, err := os.Stat(file); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat migration: %w", err)
		}
		if err := os.WriteFile(file, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("write migration: %w", err)
		}
	}
	return nil
}

func withDir(dir string, fn func() error) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("enter %s: %w", dir, err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	return fn()
}

func buildLocalIntent(req createIntentInput) (model.IntentRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id, err := newIntentID()
	if err != nil {
		return model.IntentRecord{}, err
	}

	record := model.IntentRecord{
		ID:         id,
		CreatedAt:  now,
		Author:     req.Author,
		SourceType: req.SourceType,
		Title:      req.Title,
		Prompt:     req.Prompt,
		Response:   req.Response,
		PrevHash:   req.PrevHash,
		Meta:       req.Meta,
	}
	sum, err := hash.HashIntent(record)
	if err != nil {
		return model.IntentRecord{}, err
	}
	record.Hash = sum
	return record, nil
}

func newIntentID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func verifyLocalIntent(ctx context.Context, st *store.Store, id string) (verifyResult, error) {
	record, err := st.GetIntent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return verifyResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return verifyResult{}, err
	}

	computed, err := hash.HashIntent(record)
	result := verifyResult{
		ID:           record.ID,
		StoredHash:   record.Hash,
		ComputedHash: computed,
		PrevHash:     record.PrevHash,
		Valid:        err == nil && computed == record.Hash,
	}
	if err != nil {
		msg := err.Error()
		result.Error = &msg
	}
	return result, nil
}

func chainLocalIntent(ctx context.Context, st *store.Store, id string) (chainResult, error) {
	head, err := st.GetIntent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chainResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return chainResult{}, err
	}

	intents := []model.IntentRecord{head}
	current := head
	var missing []string
	for current.PrevHash != "" {
		prev, err := st.GetIntentByHash(ctx, current.PrevHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				missing = append(missing, current.PrevHash)
				break
			}
			return chainResult{}, err
		}
		intents = append(intents, prev)
		current = prev
	}

	for i, j := 0, len(intents)-1; i < j; i, j = i+1, j-1 {
		intents[i], intents[j] = intents[j], intents[i]
	}

	return chainResult{
		HeadID:       head.ID,
		Length:       len(intents),
		Intents:      intents,
		MissingLinks: missing,
	}, nil
}

func listLocalIntents(ctx context.Context, st *store.Store, author, source string, limit int, metaFilters map[string]string) ([]model.IntentRecord, error) {
	fetchLimit := limit
	if fetchLimit <= 0 {
		fetchLimit = 20
	}
	if author != "" || source != "" || len(metaFilters) > 0 {
		fetchLimit = fetchLimit * 5
		if fetchLimit < 100 {
			fetchLimit = 100
		}
	}

	intents, err := st.ListIntents(ctx, fetchLimit)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.IntentRecord, 0, len(intents))
	for _, intent := range intents {
		if author != "" && intent.Author != author {
			continue
		}
		if source != "" && intent.SourceType != source {
			continue
		}
		filtered = append(filtered, intent)
	}

	if len(metaFilters) > 0 {
		filtered, err = store.FilterIntentsByMeta(filtered, metaFilters)
		if err != nil {
			return nil, err
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func getLocalIntent(ctx context.Context, st *store.Store, id string) (model.IntentRecord, error) {
	record, err := st.GetIntent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentRecord{}, fmt.Errorf("intent not found for ID %s", id)
		}
		return model.IntentRecord{}, err
	}
	return record, nil
}

type createIntentInput struct {
	Author     string
	SourceType string
	Title      string
	Prompt     string
	Response   string
	Meta       []byte
	PrevHash   string
}

type verifyResult struct {
	ID           string
	Valid        bool
	StoredHash   string
	ComputedHash string
	PrevHash     string
	Error        *string
}

type chainResult struct {
	HeadID       string
	Length       int
	Intents      []model.IntentRecord
	MissingLinks []string
}
