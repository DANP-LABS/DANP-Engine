package mcpclient

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func ListResources(ctx context.Context, client *Client) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	resourcesRequest := mcp.ListResourcesRequest{}
	resourcesResult, err := client.GetRawClient().ListResources(ctx, resourcesRequest)
	if err != nil {
		return fmt.Errorf("failed to list resources: %v", err)
	}

	log.Printf("Server has %d resources available\n", len(resourcesResult.Resources))
	for i, resource := range resourcesResult.Resources {
		log.Printf("  %d. %s - %s\n", i+1, resource.URI, resource.Name)
	}
	return nil
}
