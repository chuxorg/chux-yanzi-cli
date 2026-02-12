package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chuxorg/chux-yanzi-cli/internal/client"
	"github.com/chuxorg/chux-yanzi-cli/internal/config"
)

// RunCapture posts a new intent record to the library API.
func RunCapture(args []string) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		title        = fs.String("title", "", "optional title")
		author       = fs.String("author", "", "required author")
		source       = fs.String("source", "cli", "source type")
		promptFile   = fs.String("prompt-file", "", "prompt file path")
		responseFile = fs.String("response-file", "", "response file path")
		prevHash     = fs.String("prev-hash", "", "previous hash")
		useEditor    = fs.Bool("edit", false, "edit prompt in $EDITOR")
		metaPairs    = &kvPairs{}
	)
	fs.Var(metaPairs, "meta", "meta key=value (repeatable)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *author == "" {
		return errors.New("--author is required")
	}
	if *responseFile == "" {
		return errors.New("--response-file is required")
	}

	hasStdin, err := stdinHasData()
	if err != nil {
		return err
	}

	if *useEditor {
		if *promptFile != "" {
			return errors.New("--edit cannot be used with --prompt-file")
		}
		if hasStdin {
			return errors.New("--edit cannot be used with stdin")
		}
	}

	var prompt []byte
	switch {
	case *useEditor:
		content, err := readPromptFromEditor()
		if err != nil {
			return err
		}
		prompt = content
	case *promptFile != "":
		content, err := os.ReadFile(*promptFile)
		if err != nil {
			return fmt.Errorf("read prompt file: %w", err)
		}
		prompt = content
	case hasStdin:
		content, err := readPromptFromStdin()
		if err != nil {
			return err
		}
		prompt = content
	default:
		return errors.New("prompt must be provided via --prompt-file, stdin, or --edit")
	}

	response, err := os.ReadFile(*responseFile)
	if err != nil {
		return fmt.Errorf("read response file: %w", err)
	}

	meta, err := metaPairs.ToJSON()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	input := createIntentInput{
		Author:     *author,
		SourceType: *source,
		Title:      *title,
		Prompt:     string(prompt),
		Response:   string(response),
		PrevHash:   *prevHash,
		Meta:       meta,
	}

	var intent client.IntentRecord
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		req := client.CreateIntentRequest{
			Author:     input.Author,
			SourceType: input.SourceType,
			Title:      input.Title,
			Prompt:     input.Prompt,
			Response:   input.Response,
			PrevHash:   input.PrevHash,
		}
		if input.Meta != nil {
			req.Meta = input.Meta
		}
		intent, err = cli.CreateIntent(context.Background(), req)
		if err != nil {
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
	case config.ModeLocal:
		ctx := context.Background()
		store, err := openLocalStore(ctx, cfg)
		if err != nil {
			return err
		}
		defer store.Close()

		record, err := buildLocalIntent(input)
		if err != nil {
			return err
		}
		if err := store.CreateIntent(ctx, record); err != nil {
			return err
		}
		intent = record
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	fmt.Printf("id: %s\n", intent.ID)
	fmt.Printf("hash: %s\n", intent.Hash)

	if err := saveLastHash(intent.Hash); err != nil {
		return err
	}

	return nil
}

// kvPairs collects repeated key=value flags.
type kvPairs []string

func (k *kvPairs) String() string {
	return strings.Join(*k, ",")
}

func (k *kvPairs) Set(value string) error {
	*k = append(*k, value)
	return nil
}

func (k *kvPairs) ToJSON() (json.RawMessage, error) {
	if len(*k) == 0 {
		return nil, nil
	}
	obj := make(map[string]string, len(*k))
	for _, pair := range *k {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid meta value: %s", pair)
		}
		obj[parts[0]] = parts[1]
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("encode meta: %w", err)
	}
	return json.RawMessage(b), nil
}

func saveLastHash(hash string) error {
	dir, err := config.StateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	path := filepath.Join(dir, "last_hash")
	if err := os.WriteFile(path, []byte(hash+"\n"), 0o600); err != nil {
		return fmt.Errorf("write last hash: %w", err)
	}
	return nil
}

// stdinHasData reports whether stdin is connected to a non-terminal input.
func stdinHasData() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("stat stdin: %w", err)
	}
	return info.Mode()&os.ModeCharDevice == 0, nil
}

// readPromptFromStdin reads stdin and trims trailing whitespace only.
func readPromptFromStdin() ([]byte, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	trimmed := strings.TrimRightFunc(string(data), unicode.IsSpace)
	return []byte(trimmed), nil
}

// readPromptFromEditor opens $EDITOR to capture the prompt text.
func readPromptFromEditor() ([]byte, error) {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		return nil, errors.New("$EDITOR is not set")
	}

	tmp, err := os.CreateTemp("", "yanzi-prompt-*.txt")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}
	defer os.Remove(path)

	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return nil, errors.New("invalid $EDITOR value")
	}
	cmd := exec.Command(fields[0], append(fields[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor failed: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read temp file: %w", err)
	}
	return content, nil
}
