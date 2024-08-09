package statsig

type APIResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type APIListResponse[T any] struct {
	Message string `json:"message"`
	Data    []T    `json:"data"`
}

type QueryParams map[string]string // QueryParams is a map of query parameters
