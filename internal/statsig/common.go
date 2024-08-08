package statsig

type APIResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type APIListResponse[T any] struct {
	Message string `json:"message"`
	Data    []T    `json:"data"`
}
