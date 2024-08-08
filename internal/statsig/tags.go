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

func (c *Client) GetTags(ctx context.Context) ([]TagAPIRequest, error) {
	params := map[string]string{"page": "1", "limit": "100"}
	response, err := c.Get("tags", params)
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
	// Get all tags and find the one with the matching ID
	tags, err := c.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	var tag TagAPIRequest
	for _, t := range tags {
		if t.Name == tagName {
			tag = t
			tflog.Trace(ctx, fmt.Sprintf("Tag retrieved with Name: %s; and ID: %s", tag.Name, tag.ID))
			break
		}
	}

	// Log an error if the the tag is null
	if tag == (TagAPIRequest{}) {
		tflog.Error(ctx, fmt.Sprintf("Tag with Name %s not found.", tagName))
		return nil, fmt.Errorf("Tag with Name '%s' not found.", tagName)
	}

	return &tag, nil
}

func (c *Client) CreateTag(ctx context.Context, tag TagAPIRequest) (*TagAPIRequest, error) {
	response, err := c.Post("tags", tag)
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
