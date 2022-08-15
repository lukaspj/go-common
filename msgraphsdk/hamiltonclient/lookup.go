package hamiltonclient

import (
	"context"
	"errors"

	"github.com/manicminer/hamilton/msgraph"

	"github.com/webdevops/go-common/utils/to"
)

type (
	DirectoryObject struct {
		Type                 string
		ServicePrincipalType string
		ManagedIdentity      string
		DisplayName          string
		ObjectID             string
		ApplicationID        string
	}
)

// LookupPrincipalID returns information about AzureAD directory object by objectid
func (c *MsGraphClient) LookupPrincipalID(ctx context.Context, princpalIds ...string) (map[string]*DirectoryObject, error) {
	ret := map[string]*DirectoryObject{}

	// inject cached entries
	for _, objectId := range princpalIds {
		cacheKey := "object:" + objectId
		if val, ok := c.cache.Get(cacheKey); ok {
			if directoryObjectInfo, ok := val.(*DirectoryObject); ok {
				ret[objectId] = directoryObjectInfo
			}
		}
	}

	// build list of not cached entries
	lookupPrincipalObjectIDList := []string{}
	for _, princpalId := range princpalIds {
		if _, exists := ret[princpalId]; !exists {
			lookupPrincipalObjectIDList = append(lookupPrincipalObjectIDList, princpalId)
		}
	}

	// azure limits objects ids
	chunkSize := 999
	for i := 0; i < len(lookupPrincipalObjectIDList); i += chunkSize {
		end := i + chunkSize
		if end > len(lookupPrincipalObjectIDList) {
			end = len(lookupPrincipalObjectIDList)
		}

		principalObjectIDChunkList := lookupPrincipalObjectIDList[i:end]

		client := msgraph.NewDirectoryObjectsClient(c.tenantID)
		client.BaseClient.Authorizer = c.Authorizer()

		result, _, err := client.GetByIds(ctx, principalObjectIDChunkList, []string{})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, errors.New(`bad MSGraph API response, nil results received`)
		}

		for _, row := range *result {
			objectInfo := &DirectoryObject{
				ObjectID: to.String(row.ID),
				Type:     "unknown",
			}

			// TODO

			ret[objectInfo.ObjectID] = objectInfo

			// store in cache
			cacheKey := "object:" + objectInfo.ObjectID
			c.cache.Set(cacheKey, objectInfo, c.cacheTtl)
		}
	}

	return ret, nil
}
