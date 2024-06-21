package client

import (
	"context"
	"net/http"
	"time"
)

type Client struct {
	ctx      context.Context
	api      string
	apiKey   string
	metadata statsigMetadata
	client   *http.Client
}

func NewClient(_ context.Context, apiKey string) (*Client, error) {
	return &Client{
		api:      "https://api.statsig.com/console/v1",
		apiKey:   apiKey,
		metadata: getStatsigMetadata(),
		client:   &http.Client{Timeout: time.Second * 10},
	}, nil
}

// All API calls must include 'STATSIG-API-KEY' in the header. This is the apiKey value
