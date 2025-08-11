package main

import (
	"log"

	"github.com/spf13/cobra"
)

var Version string

func main() {
	if err := RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func RootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "runesmith-enchanter",
		Short:   "Runesmith job unit for enchantments",
		RunE:    func(cmd *cobra.Command, args []string) error { return run() },
		Version: Version,
	}
	return rootCmd
}
