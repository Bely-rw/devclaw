// Package main é o ponto de entrada do CLI do DevClaw.
package main

import (
	"fmt"
	"os"

	"github.com/jholhewres/devclaw/cmd/devclaw/commands"
)

// version é injetado em build time via ldflags.
var version = "dev"

func main() {
	rootCmd := commands.NewRootCmd(version)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
