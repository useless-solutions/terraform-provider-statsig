package statsig

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TagsAPIResponse struct {
	Message string `json:"message"`
	Data    []Tag  `json:"data"`
}

func (c *Client) GetTags(ctx context.Context) ([]Tag, error) {
	params := map[string]string{"page": "1", "limit": "100"}
	response, err := c.Get("GET", "tags", params)
	if err != nil {
		return nil, err
	}

	// Log the response body
	tflog.Debug(ctx, "Response Body: %s", map[string]interface{}{"response": string(response)})
	tagsResponse := TagsAPIResponse{}
	if err := json.Unmarshal(response, &tagsResponse); err != nil {
		return nil, err
	}

	return tagsResponse.Data, nil
}
