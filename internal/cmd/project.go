package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/chuxorg/chux-yanzi-library"
)

// RunProject handles project subcommands.
func RunProject(args []string) error {
	if len(args) == 0 {
		return projectUsageError()
	}

	switch args[0] {
	case "use":
		return runProjectUse(args[1:])
	case "current":
		return runProjectCurrent(args[1:])
	default:
		return projectUsageError()
	}
}

func runProjectUse(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: yanzi project use <name>")
	}
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		ctx := context.Background()
		db, closeFn, err := openLocalProjectDB(ctx, cfg)
		if err != nil {
			return err
		}
		defer func() {
			_ = closeFn()
		}()

		projects, err := yanzilibrary.ListProjects(ctx, db)
		if err != nil {
			return err
		}

		found := false
		for _, project := range projects {
			if project.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("project not found: %s", name)
		}

		if err := saveActiveProject(name); err != nil {
			return err
		}

		fmt.Printf("Active project set to %s.\n", name)
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectCurrent(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: yanzi project current")
	}

	active, err := loadActiveProject()
	if err != nil {
		return err
	}
	if active == "" {
		fmt.Println("No active project")
		return nil
	}

	fmt.Printf("Active project: %s\n", active)
	return nil
}

func projectUsageError() error {
	return errors.New("usage: yanzi project <use|current>")
}

func openLocalProjectDB(ctx context.Context, cfg config.Config) (*sql.DB, func() error, error) {
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
