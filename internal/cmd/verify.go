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

// RunVerify verifies the stored hash for a given intent id.
func RunVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi verify <intent-id>")
	}

	id := fs.Arg(0)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cli := client.New(cfg.BaseURL)

	resp, err := cli.VerifyIntent(context.Background(), id)
	if err != nil {
		return err
	}

	status := "✖ INVALID"
	if resp.Valid {
		status = "✔ VALID"
	}
	fmt.Println(status)
	fmt.Printf("stored_hash: %s\n", resp.StoredHash)
	fmt.Printf("computed_hash: %s\n", resp.ComputedHash)
	if resp.Error != nil {
		fmt.Printf("error: %s\n", *resp.Error)
	}

	return nil
}
