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
	localMigrationsDir = "migrations"
	localMigrationName = "0001_init.sql"
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

func ensureLocalMigrations(stateDir string) error {
	path := filepath.Join(stateDir, localMigrationsDir)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}
	file := filepath.Join(path, localMigrationName)
	if _, err := os.Stat(file); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat migration: %w", err)
	}
	if err := os.WriteFile(file, []byte(localMigrationSQL), 0o644); err != nil {
		return fmt.Errorf("write migration: %w", err)
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
