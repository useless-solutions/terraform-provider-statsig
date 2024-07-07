package statsig

import (
	"context"
	"encoding/json"

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
