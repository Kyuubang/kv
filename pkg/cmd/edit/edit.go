package edit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bayhaqi/kv/internal/difftui"
	"github.com/bayhaqi/kv/pkg/cmd/root"
	"github.com/bayhaqi/kv/pkg/keyvault"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	editor string
)

var EditCmd = &cobra.Command{
	Use:   "edit <vault-name> <secret-name>",
	Short: "Edit a secret in Azure Key Vault",
	Long:  `Edit the latest version of a secret in Azure Key Vault using your preferred editor.`,
	Args:  cobra.ExactArgs(2),
	Run:   runEdit,
}

func init() {
	EditCmd.Flags().StringVarP(&editor, "editor", "e", "", "Editor to use (default: $EDITOR or vim)")
	root.RootCmd.AddCommand(EditCmd)
}

func runEdit(cmd *cobra.Command, args []string) {
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
		root.ExitWithError(fmt.Errorf("no versions found for secret: %s", secretName))
	}

	// Get the latest version (first in the list)
	latestVersion := versions[0]

	// Determine editor
	editorCmd := getEditor()

	// Create secure temporary file
	tempFile, err := createSecureTempFile(latestVersion.Value)
	if err != nil {
		root.ExitWithError(fmt.Errorf("failed to create temporary file: %w", err))
	}
	defer func() {
		// Securely delete the temporary file
		if err := secureDelete(tempFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to securely delete temp file: %v\n", err)
		}
	}()

	fmt.Printf("Editing secret '%s' (version: %s)\n", secretName, latestVersion.Version[:8])
	fmt.Printf("Opening editor: %s\n\n", editorCmd)

	// Open editor
	if err := openEditor(editorCmd, tempFile); err != nil {
		root.ExitWithError(fmt.Errorf("failed to open editor: %w", err))
	}

	// Read the edited content
	newValue, err := os.ReadFile(tempFile)
	if err != nil {
		root.ExitWithError(fmt.Errorf("failed to read edited file: %w", err))
	}

	newValueStr := string(newValue)

	// Check if content was changed
	if newValueStr == latestVersion.Value {
		fmt.Println("No changes detected. Secret not updated.")
		return
	}

	// Show diff in TUI for confirmation
	fmt.Println("\nReview changes...")
	diffModel := difftui.NewModel(latestVersion.Value, newValueStr, secretName)
	p := tea.NewProgram(diffModel, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		root.ExitWithError(fmt.Errorf("diff viewer error: %w", err))
	}

	diffResult := finalModel.(difftui.Model)
	if !diffResult.Confirmed() {
		fmt.Println("Changes discarded.")
		return
	}

	// Update the secret in Key Vault
	if err := client.SetSecret(ctx, secretName, newValueStr); err != nil {
		root.ExitWithError(fmt.Errorf("failed to update secret: %w", err))
	}

	fmt.Printf("✓ Secret '%s' updated successfully\n", secretName)
}

func getEditor() string {
	if editor != "" {
		return editor
	}

	if env := os.Getenv("EDITOR"); env != "" {
		return env
	}

	return "vim"
}

func createSecureTempFile(content string) (string, error) {
	// Generate random filename
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	filename := "kv-secret-" + hex.EncodeToString(randomBytes) + ".tmp"

	// Create temp file with restricted permissions (0600 - owner read/write only)
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, filename)

	file, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	return tempPath, nil
}

func openEditor(editorCmd, filePath string) error {
	cmd := exec.Command(editorCmd, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func secureDelete(filePath string) error {
	// Get file size
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Overwrite file with random data
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	size := info.Size()
	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err != nil {
		file.Close()
		return err
	}

	if _, err := file.Write(randomData); err != nil {
		file.Close()
		return err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}

	file.Close()

	// Delete the file
	return os.Remove(filePath)
}
