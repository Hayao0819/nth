package main

import (
	"os"

	"github.com/Hayao0819/nth/internal/cli"
	"github.com/Hayao0819/nth/internal/version"
)

func main() { os.Exit(cli.Execute(version.Version)) }
