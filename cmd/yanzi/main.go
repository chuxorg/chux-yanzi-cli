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
	fmt.Fprintln(os.Stderr, "usage: yanzi <capture|verify|chain> [args]")
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "?"
}
