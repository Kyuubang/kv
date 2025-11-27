# Azure Key Vault Secret Version Browser

A CLI tool to browse Azure Key Vault secret versions with an interactive Terminal User Interface (TUI).

## Features

- 🔐 Browse secret versions in Azure Key Vault
- 🎨 Beautiful terminal UI powered by Charm's Bubble Tea
- ⌨️ Keyboard navigation (arrow keys)
- 📊 View detailed version information including timestamps, status, and tags
- 🚀 Fast and maintainable Go code

## Installation

```bash
go mod download
go build -o kv
```

## Prerequisites

- Go 1.21 or later
- Azure authentication configured (Azure CLI, managed identity, or environment variables)
- Access to an Azure Key Vault

## Authentication

This tool uses Azure's Default Credential Chain. Make sure you're authenticated via one of:
- `az login` (Azure CLI)
- Environment variables (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
- Managed Identity (when running on Azure)

## Usage

```bash
# Browse secret versions
./kv version https://your-vault.vault.azure.net/ your-secret-name
```

### Keyboard Controls

- `←` / `→` - Navigate between versions
- `h` / `l` - Alternative navigation (vim-style)
- `ESC` / `q` - Quit the application

## Project Structure

```
kv/
├── main.go                 # Application entry point
├── cmd/
│   ├── root.go            # Root command
│   └── version.go         # Version browsing command
├── pkg/
│   └── keyvault/
│       └── client.go      # Azure Key Vault client wrapper
└── internal/
    └── tui/
        └── model.go       # Bubble Tea TUI model
```

## Development

The project follows best practices:
- Clean separation of concerns (cmd, pkg, internal)
- Interface-based design for testability
- Uses Cobra for CLI structure (similar to GitHub CLI)
- Bubble Tea for reactive TUI
- Lipgloss for styling

## Example

```bash
# List versions of a secret named "database-password"
./kv version https://my-keyvault.vault.azure.net/ database-password
```

The TUI will show:
- Version ID
- Status (Enabled/Disabled)
- Creation and update timestamps
- Expiration date (if set)
- Tags (if any)
- Secret value

## License

MIT
