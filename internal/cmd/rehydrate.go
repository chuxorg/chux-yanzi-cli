package cmd

import (
	"errors"
	"fmt"
	"sort"

	yanzilibrary "github.com/chuxorg/chux-yanzi-library"
)

// RunRehydrate renders the latest checkpoint and artifacts since.
func RunRehydrate(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: yanzi rehydrate")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	payload, err := yanzilibrary.RehydrateProject(project)
	if err != nil {
		if errors.Is(err, yanzilibrary.ErrCheckpointNotFound) {
			return errors.New("no checkpoint found for active project")
		}
		return err
	}

	artifacts := payload.ArtifactsSinceCheckpoint
	sort.SliceStable(artifacts, func(i, j int) bool {
		if artifacts[i].CreatedAt == artifacts[j].CreatedAt {
			return artifacts[i].ID < artifacts[j].ID
		}
		return artifacts[i].CreatedAt < artifacts[j].CreatedAt
	})

	fmt.Printf("Project: %s\n", payload.Project)
	fmt.Println("Latest Checkpoint:")
	fmt.Printf("* CreatedAt: %s\n", payload.LatestCheckpoint.CreatedAt)
	fmt.Printf("* Summary: %s\n", payload.LatestCheckpoint.Summary)
	fmt.Println("Artifacts Since Checkpoint:")
	if len(artifacts) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	for i, artifact := range artifacts {
		fmt.Printf("%d. %s %s %s\n", i+1, artifact.ID, artifact.CreatedAt, artifact.Type)
	}
	return nil
}
