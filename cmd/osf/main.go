package main

import (
	"os"

	"github.com/edithatogo/osf-cli-go/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
