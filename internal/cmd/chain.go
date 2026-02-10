package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/chuxorg/chux-yanzi-cli/internal/client"
	"github.com/chuxorg/chux-yanzi-cli/internal/config"
)

// RunChain prints the intent chain from oldest to newest.
func RunChain(args []string) error {
	fs := flag.NewFlagSet("chain", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi chain <intent-id>")
	}

	id := fs.Arg(0)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cli := client.New(cfg.BaseURL)

	resp, err := cli.ChainIntent(context.Background(), id)
	if err != nil {
		return err
	}

	fmt.Printf("chain head: %s\n", resp.HeadID)
	for i, intent := range resp.Intents {
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", i+1, intent.CreatedAt, intent.Title, intent.Author, intent.Hash)
	}
	if len(resp.MissingLinks) > 0 {
		fmt.Printf("missing_links: %s\n", joinComma(resp.MissingLinks))
	}

	return nil
}

func joinComma(values []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for i := 1; i < len(values); i++ {
		out += "," + values[i]
	}
	return out
}
