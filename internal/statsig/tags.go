package statsig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type GetTagsAPIResponse struct {
	Message string          `json:"message"`
	Data    []TagAPIRequest `json:"data"`
}

type TagAPIRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsCore      bool   `json:"isCore"`
}

type CreateTagAPIResponse struct {
	Message string        `json:"message"`
	Data    TagAPIRequest `json:"data"`
}

func (c *Client) GetTags(ctx context.Context) ([]TagAPIRequest, error) {
	params := map[string]string{"page": "1", "limit": "100"}
	response, err := c.Get("tags", params)
	if err != nil {
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, "Response Body: %s", map[string]interface{}{"response": string(response)})
	tagsResponse := GetTagsAPIResponse{}
	if err := json.Unmarshal(response, &tagsResponse); err != nil {
		return nil, err
	}

	return tagsResponse.Data, nil
}

func (c *Client) CreateTag(ctx context.Context, tag TagAPIRequest) (*TagAPIRequest, error) {
	response, err := c.Post("tags", tag)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating tag: %s", err))
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, fmt.Sprintf("Create tag response: %s", response))
	createdTag := CreateTagAPIResponse{}
	if err := json.Unmarshal(response, &createdTag); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error unmarshalling tag: %s", err))
		return nil, err
	}

	tflog.Debug(ctx, fmt.Sprintf("Returning created tag: %+v", createdTag.Data))
	tflog.Debug(ctx, fmt.Sprintf("Full response: %+v", createdTag))
	tflog.Trace(ctx, fmt.Sprintf("Tag created with ID: %s", createdTag.Data.ID))

	return &createdTag.Data, nil
}
