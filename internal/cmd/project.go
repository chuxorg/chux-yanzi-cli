package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/yourusername/yanzi-library"
)

// RunProject handles project subcommands.
func RunProject(args []string) error {
	if len(args) == 0 {
		return projectUsageError()
	}

	switch args[0] {
	case "create":
		return runProjectCreate(args[1:])
	case "list":
		return runProjectList(args[1:])
	case "current":
		return runProjectCurrent(args[1:])
	default:
		return projectUsageError()
	}
}

func runProjectCreate(args []string) error {
	var (
		name        string
		description string
	)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--description":
			if i+1 >= len(args) {
				return errors.New("usage: yanzi project create <name> [--description \"...\"]")
			}
			description = args[i+1]
			i++
		case strings.HasPrefix(arg, "--description="):
			description = strings.TrimPrefix(arg, "--description=")
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown flag: %s", arg)
		default:
			if name != "" {
				return errors.New("usage: yanzi project create <name> [--description \"...\"]")
			}
			name = arg
		}
	}
	if name == "" {
		return errors.New("usage: yanzi project create <name> [--description \"...\"]")
	}

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

		project, err := yanzilibrary.CreateProject(ctx, db, name, description)
		if err != nil {
			return err
		}

		fmt.Printf("hash: %s\n", project.Hash)
		fmt.Println("Project created.")
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectList(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: yanzi project list")
	}

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

		fmt.Println("Name\tCreatedAt\tDescription")
		for _, project := range projects {
			fmt.Printf("%s\t%s\t%s\n", project.Name, project.CreatedAt, project.Description)
		}
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
	fmt.Println("No active project")
	return nil
}

func projectUsageError() error {
	return errors.New("usage: yanzi project <create|list|current>")
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
