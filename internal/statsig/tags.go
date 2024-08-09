package statsig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TagAPIRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsCore      bool   `json:"isCore"`
}

const tagsEndpoint = "tags"

func (c *Client) GetTags(ctx context.Context) ([]TagAPIRequest, error) {
	params := QueryParams{"page": "1", "limit": "100"}
	response, err := c.Get(tagsEndpoint, params)
	if err != nil {
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Response Body: %s", map[string]interface{}{"response": string(response)}))
	tagsResponse := APIListResponse[TagAPIRequest]{}
	if err := json.Unmarshal(response, &tagsResponse); err != nil {
		return nil, err
	}

	return tagsResponse.Data, nil
}

// GetTag retrieves a tag by its name from the Statsig API.
//
// The API does not use IDs for identifying unique objects, so we must retrieve the tag by its Name.
func (c *Client) GetTag(ctx context.Context, tagName string) (*TagAPIRequest, error) {
	response, err := c.Get(createEndpointPath(tagName), nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error getting tag: %s", err))
		return nil, err
	}

	tag := APIResponse[TagAPIRequest]{}
	if err := json.Unmarshal(response, &tag); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling tag: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag retrieved with Name: %s; and ID: %s", tag.Data.Name, tag.Data.ID))
	return &tag.Data, nil
}

func (c *Client) CreateTag(ctx context.Context, tag TagAPIRequest) (*TagAPIRequest, error) {
	response, err := c.Post(tagsEndpoint, tag)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating tag: %s", err))
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Create tag response: %s", response))
	createdTag := APIResponse[TagAPIRequest]{}
	if err := json.Unmarshal(response, &createdTag); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling tag: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag created with ID: %s", createdTag.Data.ID))

	return &createdTag.Data, nil
}

func (c *Client) UpdateTag(ctx context.Context, tagName string, planTag TagAPIRequest) (*TagAPIRequest, error) {
	response, err := c.Patch(createEndpointPath(tagName), planTag)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating tag '%s': %s", tagName, err))
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Update tag response: %s", response))
	updatedTag := APIResponse[TagAPIRequest]{}
	if err := json.Unmarshal(response, &updatedTag); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling tag: %s", err))
		return nil, err
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag updated with ID: %s", updatedTag.Data.ID))

	return &updatedTag.Data, nil
}

func (c *Client) DeleteTag(ctx context.Context, tagName string) error {
	_, err := c.Delete(createEndpointPath(tagName), nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting tag: %s", err))
		return err
	}

	tflog.Trace(ctx, fmt.Sprintf("Tag deleted with Name: %s", tagName))

	return nil
}

func createEndpointPath(path string) string {
	return fmt.Sprintf("%s/%s", tagsEndpoint, path)
}
