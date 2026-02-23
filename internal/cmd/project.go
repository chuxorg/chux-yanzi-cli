package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/chuxorg/chux-yanzi-library"
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
	case "use":
		return runProjectUse(args[1:])
	case "current":
		return runProjectCurrent(args[1:])
	default:
		return projectUsageError()
	}
}

func runProjectCreate(args []string) error {
	description := ""
	name := ""
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
		case strings.HasPrefix(arg, "--"):
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
		if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
			return fmt.Errorf("set YANZI_DB_PATH: %w", err)
		}

		project, err := yanzilibrary.CreateProject(name, description)
		if err != nil {
			return err
		}

		fmt.Printf("created_at: %s\n", project.CreatedAt.Format(time.RFC3339Nano))
		fmt.Println("Project created.")
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectList(args []string) error {
	fs := flag.NewFlagSet("project list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi project list")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
			return fmt.Errorf("set YANZI_DB_PATH: %w", err)
		}

		projects, err := yanzilibrary.ListProjects()
		if err != nil {
			return err
		}

		fmt.Println("Name\tCreatedAt\tDescription")
		for _, project := range projects {
			fmt.Printf("%s\t%s\t%s\n", project.Name, project.CreatedAt.Format(time.RFC3339Nano), project.Description)
		}
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
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
		if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
			return fmt.Errorf("set YANZI_DB_PATH: %w", err)
		}

		projects, err := yanzilibrary.ListProjects()
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
	return errors.New("usage: yanzi project <create|list|use|current>")
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
