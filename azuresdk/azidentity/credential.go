package azidentity

import (
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	azureAuthorityHost      = "AZURE_AUTHORITY_HOST"
	azureClientID           = "AZURE_CLIENT_ID"
	azureFederatedTokenFile = "AZURE_FEDERATED_TOKEN_FILE"
	azureTenantID           = "AZURE_TENANT_ID"
)

func NewAzDefaultCredential(clientOptions *azcore.ClientOptions) (azcore.TokenCredential, error) {
	// azure authorizer
	switch strings.ToLower(os.Getenv("AZURE_AUTH")) {
	case "az", "cli", "azcli":
		// azurecli authentication
		return NewAzCliCredential()
	case "wi", "workload", "workloadidentity", "federation":
		var tokenFile, tenantID, clientID string
		var ok bool

		if _, ok = os.LookupEnv(azureAuthorityHost); !ok {
			panic(fmt.Sprintf(`missing environment variable "%s" for workload identity. Check webhook and pod configuration`, azureAuthorityHost))
		}

		if tokenFile, ok = os.LookupEnv(azureFederatedTokenFile); !ok {
			panic(fmt.Sprintf(`missing environment variable "%s" for workload identity. Check webhook and pod configuration`, azureFederatedTokenFile))
		}

		if tenantID, ok = os.LookupEnv(azureTenantID); !ok {
			panic(fmt.Sprintf(`missing environment variable "%s" for workload identity. Check webhook and pod configuration`, azureTenantID))
		}

		if clientID, ok = os.LookupEnv(azureClientID); !ok {
			panic(fmt.Sprintf(`missing environment variable "%s" for workload identity. Check webhook and pod configuration`, azureClientID))
		}

		opts := azidentity.WorkloadIdentityCredentialOptions{}
		if clientOptions != nil {
			opts.ClientOptions = *clientOptions
		}

		return azidentity.NewWorkloadIdentityCredential(tenantID, clientID, tokenFile, &opts)
	default:
		// general azure authentication (env vars, service principal, msi, ...)
		opts := azidentity.DefaultAzureCredentialOptions{}
		if clientOptions != nil {
			opts.ClientOptions = *clientOptions
		}

		return azidentity.NewDefaultAzureCredential(&opts)
	}
}

func NewAzCliCredential() (azcore.TokenCredential, error) {
	opts := azidentity.AzureCLICredentialOptions{}
	return azidentity.NewAzureCLICredential(&opts)
}
