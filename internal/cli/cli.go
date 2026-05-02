package cli

import (
	"flag"
	"fmt"
	"io"
)

const version = "0.0.0-dev"

// Run executes the osf CLI and returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("osf", flag.ContinueOnError)
	flags.SetOutput(stderr)

	showVersion := flags.Bool("version", false, "print version")
	showHelp := flags.Bool("help", false, "print help")
	flags.BoolVar(showHelp, "h", false, "print help")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	if *showHelp || flags.NArg() == 0 {
		printHelp(stdout)
		return 0
	}

	fmt.Fprintf(stderr, "unknown command: %s\n\n", flags.Arg(0))
	printHelp(stderr)
	return 2
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `osf is a command-line client for the Open Science Framework.

Usage:
  osf [--version]
  osf <command> [options]

Planned commands:
  auth       Manage OSF personal access tokens
  projects   List and inspect OSF projects and components
  files      List, download, and upload OSF Storage files
  export     Create portable project exports

Use "osf <command> --help" for command-specific help once implemented.`)
}
