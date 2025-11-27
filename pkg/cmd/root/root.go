package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Azure Key Vault CLI tool",
	Long:  `A CLI tool to browse and manage Azure Key Vault secrets with a beautiful TUI.`,
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}

func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
