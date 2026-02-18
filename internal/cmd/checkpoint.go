package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi-library"
)

// RunCheckpoint handles checkpoint subcommands.
func RunCheckpoint(args []string) error {
	if len(args) == 0 {
		return checkpointUsageError()
	}

	switch args[0] {
	case "create":
		return runCheckpointCreate(args[1:])
	case "list":
		return runCheckpointList(args[1:])
	default:
		return checkpointUsageError()
	}
}

func runCheckpointCreate(args []string) error {
	fs := flag.NewFlagSet("checkpoint create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	summary := fs.String("summary", "", "checkpoint summary")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi checkpoint create --summary \"...\"")
	}
	if *summary == "" {
		return errors.New("summary is required")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		ctx := context.Background()
		db, closeFn, err := openLocalCheckpointDB(ctx, cfg)
		if err != nil {
			return err
		}
		defer func() {
			_ = closeFn()
		}()

		checkpoint, err := yanzilibrary.CreateCheckpoint(ctx, db, project, *summary, []string{})
		if err != nil {
			return err
		}

		fmt.Printf("id: %s\n", checkpoint.Hash)
		fmt.Printf("summary: %s\n", checkpoint.Summary)
		return nil
	case config.ModeHTTP:
		return errors.New("checkpoint commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runCheckpointList(args []string) error {
	fs := flag.NewFlagSet("checkpoint list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi checkpoint list")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		ctx := context.Background()
		db, closeFn, err := openLocalCheckpointDB(ctx, cfg)
		if err != nil {
			return err
		}
		defer func() {
			_ = closeFn()
		}()

		checkpoints, err := yanzilibrary.ListCheckpoints(ctx, db, project)
		if err != nil {
			return err
		}

		fmt.Println("Index\tCreatedAt\tSummary")
		for i, checkpoint := range checkpoints {
			fmt.Printf("%d\t%s\t%s\n", i+1, checkpoint.CreatedAt, checkpoint.Summary)
		}
		return nil
	case config.ModeHTTP:
		return errors.New("checkpoint commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func checkpointUsageError() error {
	return errors.New("usage: yanzi checkpoint <create|list>")
}

func openLocalCheckpointDB(ctx context.Context, cfg config.Config) (*sql.DB, func() error, error) {
	st, err := openLocalStore(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := st.Close(); err != nil {
		return nil, nil, err
	}

	db, err := openSQLiteDB(cfg.DBPath)
	if err != nil {
		return nil, nil, err
	}
	return db, db.Close, nil
}
