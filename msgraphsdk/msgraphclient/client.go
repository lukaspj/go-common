package msgraphclient

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/patrickmn/go-cache"

	log "github.com/sirupsen/logrus"

	"github.com/webdevops/go-common/azuresdk/azidentity"
	"github.com/webdevops/go-common/azuresdk/cloudconfig"
	"github.com/webdevops/go-common/azuresdk/prometheus/tracing"
)

type (
	MsGraphClient struct {
		cloud    cloudconfig.CloudEnvironment
		tenantID string

		logger *log.Logger

		cache              *cache.Cache
		cacheTtl           time.Duration
		cacheAuthorizerTtl time.Duration

		userAgent string
	}
)

// NewMsGraphClient creates new MS Graph client
func NewMsGraphClient(cloudConfig cloudconfig.CloudEnvironment, tenantID string, logger *log.Logger) *MsGraphClient {
	client := &MsGraphClient{}
	client.cloud = cloudConfig
	client.tenantID = tenantID

	client.cacheTtl = 1 * time.Hour
	client.cache = cache.New(60*time.Minute, 60*time.Second)

	client.cacheAuthorizerTtl = 15 * time.Minute

	client.logger = logger
	client.userAgent = "go-common/unknown"

	return client
}

// NewMsGraphClientWithCloudName creates new MS Graph client with environment name as string
func NewMsGraphClientWithCloudName(cloudName string, tenantID string, logger *log.Logger) (*MsGraphClient, error) {
	cloudConfig, err := cloudconfig.NewCloudConfig(cloudName)
	if err != nil {
		logger.Panic(err.Error())
	}
	return NewMsGraphClient(cloudConfig, tenantID, logger), nil
}

// ServiceClient returns msgraph service client
func (c *MsGraphClient) ServiceClient() *msgraphsdk.GraphServiceClient {
	return msgraphsdk.NewGraphServiceClient(c.RequestAdapter())
}

// RequestAdapter returns msgraph request adapter
func (c *MsGraphClient) RequestAdapter() *msgraphsdk.GraphRequestAdapter {
	cacheKey := "adapter"
	if v, ok := c.cache.Get(cacheKey); ok {
		if adapter, ok := v.(*msgraphsdk.GraphRequestAdapter); ok {
			return adapter
		}
	}

	adapter := c.createRequestAdapter()

	c.cache.Set(cacheKey, adapter, c.cacheAuthorizerTtl)

	return adapter
}

// createRequestAdapter returns new msgraph request adapter
func (c *MsGraphClient) createRequestAdapter() (adapter *msgraphsdk.GraphRequestAdapter) {
	cred, err := azidentity.NewAzCredential(c.NewAzCoreClientOptions())
	if err != nil {
		c.logger.Panic(err)
	}

	auth, err := a.NewAzureIdentityAuthenticationProvider(cred)
	if err != nil {
		c.logger.Panic(err)
	}

	adapter, err = msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		c.logger.Panic(err)
	}

	// set endpoint from cloudconfig
	if c.cloud.Services != nil {
		if serviceConfig, exists := c.cloud.Services[cloudconfig.ServiceNameMicrosoftGraph]; exists {
			adapter.SetBaseUrl(serviceConfig.Endpoint + "/v1.0")
		}
	}

	return
}

// NewAzCoreClientOptions returns new client options for all arm clients
func (c *MsGraphClient) NewAzCoreClientOptions() *azcore.ClientOptions {
	clientOptions := azcore.ClientOptions{
		Cloud:            c.cloud.Configuration,
		PerCallPolicies:  []policy.Policy{},
		PerRetryPolicies: nil,
	}

	// azure prometheus tracing
	if tracing.TracingIsEnabled() {
		clientOptions.PerRetryPolicies = append(
			clientOptions.PerRetryPolicies,
			tracing.NewTracingPolicy(),
		)
	}

	return &clientOptions
}

// SetUserAgent set user agent for all API calls
func (c *MsGraphClient) SetUserAgent(useragent string) {
	c.userAgent = useragent
}

// SetCacheTtl set TTL for service discovery cache
func (c *MsGraphClient) SetCacheTtl(ttl time.Duration) {
	c.cacheTtl = ttl
}

// GetTenantID returns the current set TenantID
func (c *MsGraphClient) GetTenantID() string {
	return c.tenantID
}
