package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/chuxorg/chux-yanzi-cli/internal/client"
	"github.com/chuxorg/chux-yanzi-cli/internal/config"
)

// RunList lists intent records.
func RunList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	author := fs.String("author", "", "author filter")
	source := fs.String("source", "", "source filter")
	limit := fs.Int("limit", 20, "max records to return")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var intents []client.IntentRecord
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		resp, err := cli.ListIntents(context.Background(), *author, *source, *limit)
		if err != nil {
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		intents = resp.Intents
	case config.ModeLocal:
		ctx := context.Background()
		store, err := openLocalStore(ctx, cfg)
		if err != nil {
			return err
		}
		defer store.Close()

		localIntents, err := listLocalIntents(ctx, store, *author, *source, *limit)
		if err != nil {
			return err
		}
		intents = localIntents
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	if len(intents) == 0 {
		return fmt.Errorf("No records found")
	}

	fmt.Println("ID\tCreated_At\tAuthor\tSource\tTitle")
	for _, intent := range intents {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", intent.ID, intent.CreatedAt, intent.Author, intent.SourceType, intent.Title)
	}

	return nil
}
