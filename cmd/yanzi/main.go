package main

import (
	"fmt"
	"os"

	"github.com/chuxorg/chux-yanzi-cli/internal/cmd"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	if isHelpArg(os.Args[1]) {
		usage()
		return
	}

	var err error
	switch os.Args[1] {
	case "capture":
		err = cmd.RunCapture(os.Args[2:])
	case "verify":
		err = cmd.RunVerify(os.Args[2:])
	case "chain":
		err = cmd.RunChain(os.Args[2:])
	case "list":
		err = cmd.RunList(os.Args[2:])
	case "show":
		err = cmd.RunShow(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  yanzi <command> [args]

commands:
  capture  Create a new intent record via the library API.
  verify   Verify an intent by id.
  chain    Print an intent chain by id.
  list     List intent records.
  show     Show intent details by id.

capture args:
  --author <name>         Required author name.
  --response-file <path>  Required response file path.
  --prompt-file <path>    Read prompt from file.
  --edit                  Open $EDITOR to edit the prompt.
  (or)                    Provide prompt via stdin.
  --title <title>         Optional title.
  --source <source>       Optional source type (default "cli").
  --prev-hash <hash>      Optional previous hash.
  --meta k=v              Optional metadata (repeatable).

verify args:
  <intent-id>             Intent id to verify.

chain args:
  <intent-id>             Intent id to chain.

list args:
  --author <name>         Optional author filter.
  --source <source>       Optional source filter.
  --limit <n>             Max records to return (default 20).

show args:
  <intent-id>             Intent id to show.

examples:
  yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt --meta lang=go
  yanzi verify 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi chain 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi list --limit 10
  yanzi show 01HZX9Q4X8N9JZ1K2G9N8M4V3P`)
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "?"
}
