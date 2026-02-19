package yanzilibrary

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	chuxlib "github.com/chuxorg/chux-yanzi-library"
)

// Project types and helpers are re-exported from chux-yanzi-library.
type Project = chuxlib.Project
type ProjectValidationError = chuxlib.ProjectValidationError
type DuplicateProjectNameError = chuxlib.DuplicateProjectNameError

// IsDuplicateProjectName returns true when err indicates an existing project name.
func IsDuplicateProjectName(err error) bool {
	return chuxlib.IsDuplicateProjectName(err)
}

// CreateProject creates a new project artifact.
func CreateProject(ctx context.Context, db *sql.DB, name, description string) (Project, error) {
	return chuxlib.CreateProject(ctx, db, name, description)
}

// ListProjects returns projects ordered by creation time, newest first.
func ListProjects(ctx context.Context, db *sql.DB) ([]Project, error) {
	return chuxlib.ListProjects(ctx, db)
}

// HashProject computes a deterministic hash for a project record.
func HashProject(project Project) (string, error) {
	return chuxlib.HashProject(project)
}

// Checkpoint represents a saved checkpoint for a project.
type Checkpoint struct {
	Hash                 string
	Project              string
	Summary              string
	CreatedAt            string
	ArtifactIDs          []string
	PreviousCheckpointID string
}

// Artifact represents an artifact created since a checkpoint.
type Artifact struct {
	ID        string
	CreatedAt string
	Type      string
}

// RehydratePayload returns the latest checkpoint and artifacts since.
type RehydratePayload struct {
	Project                  string
	LatestCheckpoint         Checkpoint
	ArtifactsSinceCheckpoint []Artifact
}

// CheckpointNotFoundError indicates no checkpoints for a project.
type CheckpointNotFoundError struct {
	Project string
}

func (e CheckpointNotFoundError) Error() string {
	if strings.TrimSpace(e.Project) == "" {
		return "checkpoint not found"
	}
	return "checkpoint not found for project: " + e.Project
}

// CreateCheckpoint creates and stores a new checkpoint.
func CreateCheckpoint(ctx context.Context, db *sql.DB, project, summary string, artifactIDs []string) (Checkpoint, error) {
	if db == nil {
		return Checkpoint{}, errors.New("checkpoint db is not initialized")
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return Checkpoint{}, errors.New("project is required")
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return Checkpoint{}, errors.New("summary is required")
	}
	if artifactIDs == nil {
		artifactIDs = []string{}
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	prevID, err := latestCheckpointID(ctx, db, project)
	if err != nil {
		return Checkpoint{}, err
	}

	checkpoint := Checkpoint{
		Project:              project,
		Summary:              summary,
		CreatedAt:            createdAt,
		ArtifactIDs:          artifactIDs,
		PreviousCheckpointID: prevID,
	}
	checkpoint.Hash, err = HashCheckpoint(checkpoint)
	if err != nil {
		return Checkpoint{}, err
	}

	artifactPayload, err := json.Marshal(artifactIDs)
	if err != nil {
		return Checkpoint{}, err
	}

	var prev any
	if prevID != "" {
		prev = prevID
	}

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		checkpoint.Hash,
		checkpoint.Project,
		checkpoint.Summary,
		checkpoint.CreatedAt,
		string(artifactPayload),
		prev,
	)
	if err != nil {
		return Checkpoint{}, err
	}

	return checkpoint, nil
}

// ListCheckpoints lists checkpoints for a project, newest first.
func ListCheckpoints(ctx context.Context, db *sql.DB, project string) ([]Checkpoint, error) {
	if db == nil {
		return nil, errors.New("checkpoint db is not initialized")
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}

	rows, err := db.QueryContext(
		ctx,
		`SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id
		FROM checkpoints
		WHERE project = ?
		ORDER BY created_at DESC`,
		project,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var checkpoint Checkpoint
		var artifactIDs string
		var prevID sql.NullString
		if err := rows.Scan(
			&checkpoint.Hash,
			&checkpoint.Project,
			&checkpoint.Summary,
			&checkpoint.CreatedAt,
			&artifactIDs,
			&prevID,
		); err != nil {
			return nil, err
		}
		if prevID.Valid {
			checkpoint.PreviousCheckpointID = prevID.String
		}
		checkpoint.ArtifactIDs = []string{}
		if strings.TrimSpace(artifactIDs) != "" {
			if err := json.Unmarshal([]byte(artifactIDs), &checkpoint.ArtifactIDs); err != nil {
				return nil, err
			}
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return checkpoints, nil
}

// HashCheckpoint computes a deterministic hash for a checkpoint.
func HashCheckpoint(checkpoint Checkpoint) (string, error) {
	payload := struct {
		Project              string   `json:"project"`
		Summary              string   `json:"summary"`
		CreatedAt            string   `json:"created_at"`
		ArtifactIDs          []string `json:"artifact_ids"`
		PreviousCheckpointID string   `json:"previous_checkpoint_id"`
	}{
		Project:              strings.TrimSpace(checkpoint.Project),
		Summary:              strings.TrimSpace(checkpoint.Summary),
		CreatedAt:            strings.TrimSpace(checkpoint.CreatedAt),
		ArtifactIDs:          checkpoint.ArtifactIDs,
		PreviousCheckpointID: strings.TrimSpace(checkpoint.PreviousCheckpointID),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// RehydrateProject loads the latest checkpoint and artifacts since.
func RehydrateProject(project string) (RehydratePayload, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return RehydratePayload{}, errors.New("project is required")
	}

	path := strings.TrimSpace(os.Getenv("YANZI_LIBRARY_DB"))
	if path == "" {
		path = filepath.Join(".", "yanzi-library.db")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return RehydratePayload{}, err
	}
	defer db.Close()

	ctx := context.Background()
	checkpoint, err := latestCheckpoint(ctx, db, project)
	if err != nil {
		return RehydratePayload{}, err
	}
	artifacts, err := listArtifactsSince(ctx, db, project, checkpoint.CreatedAt)
	if err != nil {
		return RehydratePayload{}, err
	}

	return RehydratePayload{
		Project:                  project,
		LatestCheckpoint:         checkpoint,
		ArtifactsSinceCheckpoint: artifacts,
	}, nil
}

func latestCheckpointID(ctx context.Context, db *sql.DB, project string) (string, error) {
	var id string
	row := db.QueryRowContext(
		ctx,
		`SELECT hash FROM checkpoints WHERE project = ? ORDER BY created_at DESC LIMIT 1`,
		project,
	)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return id, nil
}

func latestCheckpoint(ctx context.Context, db *sql.DB, project string) (Checkpoint, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id
		FROM checkpoints
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT 1`,
		project,
	)

	var checkpoint Checkpoint
	var artifactIDs string
	var prevID sql.NullString
	if err := row.Scan(
		&checkpoint.Hash,
		&checkpoint.Project,
		&checkpoint.Summary,
		&checkpoint.CreatedAt,
		&artifactIDs,
		&prevID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Checkpoint{}, CheckpointNotFoundError{Project: project}
		}
		return Checkpoint{}, err
	}
	if prevID.Valid {
		checkpoint.PreviousCheckpointID = prevID.String
	}
	checkpoint.ArtifactIDs = []string{}
	if strings.TrimSpace(artifactIDs) != "" {
		if err := json.Unmarshal([]byte(artifactIDs), &checkpoint.ArtifactIDs); err != nil {
			return Checkpoint{}, err
		}
	}
	return checkpoint, nil
}

func listArtifactsSince(ctx context.Context, db *sql.DB, project, createdAfter string) ([]Artifact, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, meta FROM intents WHERE created_at > ? ORDER BY created_at ASC, id ASC`,
		createdAfter,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []Artifact
	for rows.Next() {
		var id string
		var createdAt string
		var metaText sql.NullString
		if err := rows.Scan(&id, &createdAt, &metaText); err != nil {
			return nil, err
		}
		if !metaText.Valid {
			continue
		}
		var meta map[string]string
		if err := json.Unmarshal([]byte(metaText.String), &meta); err != nil {
			continue
		}
		if strings.TrimSpace(meta["project"]) != project {
			continue
		}
		artifacts = append(artifacts, Artifact{
			ID:        id,
			CreatedAt: createdAt,
			Type:      "intent",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artifacts, nil
}
