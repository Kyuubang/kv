package show

import (
	"context"
	"fmt"

	"github.com/bayhaqi/kv/internal/tui"
	"github.com/bayhaqi/kv/pkg/cmd/root"
	"github.com/bayhaqi/kv/pkg/keyvault"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show <vault-name> <secret-name>",
	Short: "Browse secret versions in Azure Key Vault",
	Long:  `Browse different versions of a secret in Azure Key Vault using an interactive TUI.`,
	Args:  cobra.ExactArgs(2),
	Run:   runShow,
}

func init() {
	root.RootCmd.AddCommand(ShowCmd)
}

func runShow(cmd *cobra.Command, args []string) {
	vaultName := args[0]
	secretName := args[1]

	// Build vault URL from vault name
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)

	// Fetch secret versions
	ctx := context.Background()
	client, err := keyvault.NewClient(vaultURL)
	if err != nil {
		root.ExitWithError(fmt.Errorf("failed to create Key Vault client: %w", err))
	}

	versions, err := client.ListSecretVersions(ctx, secretName)
	if err != nil {
		root.ExitWithError(fmt.Errorf("failed to list secret versions: %w", err))
	}

	if len(versions) == 0 {
		fmt.Println("No versions found for this secret.")
		return
	}

	// Start TUI
	model := tui.NewModel(versions, secretName)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		root.ExitWithError(fmt.Errorf("TUI error: %w", err))
	}
}
