package keyvault

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// Client wraps the Azure Key Vault secrets client
type Client struct {
	client *azsecrets.Client
}

// SecretVersion represents a version of a secret
type SecretVersion struct {
	Version   string
	Value     string
	Enabled   bool
	CreatedOn *time.Time
	UpdatedOn *time.Time
	ExpiresOn *time.Time
	Tags      map[string]string
}

// NewClient creates a new Key Vault client
func NewClient(vaultURL string) (*Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{client: client}, nil
}

// ListSecretVersions lists all versions of a secret
func (c *Client) ListSecretVersions(ctx context.Context, secretName string) ([]SecretVersion, error) {
	pager := c.client.NewListSecretPropertiesVersionsPager(secretName, nil)

	var versions []SecretVersion
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get page: %w", err)
		}

		for _, props := range page.Value {
			if props.ID == nil {
				continue
			}

			version := props.ID.Version()
			if version == "" {
				continue
			}

			// Fetch the actual secret value for this version
			resp, err := c.client.GetSecret(ctx, secretName, version, nil)
			if err != nil {
				// If we can't get the secret value, still add it but without value
				versions = append(versions, SecretVersion{
					Version:   version,
					Value:     fmt.Sprintf("Error fetching value: %v", err),
					Enabled:   props.Attributes != nil && props.Attributes.Enabled != nil && *props.Attributes.Enabled,
					CreatedOn: props.Attributes.Created,
					UpdatedOn: props.Attributes.Updated,
					ExpiresOn: props.Attributes.Expires,
					Tags:      convertTags(props.Tags),
				})
				continue
			}

			value := ""
			if resp.Value != nil {
				value = *resp.Value
			}

			versions = append(versions, SecretVersion{
				Version:   version,
				Value:     value,
				Enabled:   props.Attributes != nil && props.Attributes.Enabled != nil && *props.Attributes.Enabled,
				CreatedOn: props.Attributes.Created,
				UpdatedOn: props.Attributes.Updated,
				ExpiresOn: props.Attributes.Expires,
				Tags:      convertTags(props.Tags),
			})
		}
	}

	// Sort versions by creation date (newest first)
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].CreatedOn == nil {
			return false
		}
		if versions[j].CreatedOn == nil {
			return true
		}
		return versions[i].CreatedOn.After(*versions[j].CreatedOn)
	})

	return versions, nil
}

// SetSecret sets a secret value in the Key Vault
func (c *Client) SetSecret(ctx context.Context, secretName, value string) error {
	_, err := c.client.SetSecret(ctx, secretName, azsecrets.SetSecretParameters{
		Value: &value,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}
	return nil
}

// convertTags converts Azure SDK tags (map[string]*string) to map[string]string
func convertTags(azureTags map[string]*string) map[string]string {
	if azureTags == nil {
		return nil
	}

	tags := make(map[string]string, len(azureTags))
	for key, value := range azureTags {
		if value != nil {
			tags[key] = *value
		}
	}
	return tags
}
