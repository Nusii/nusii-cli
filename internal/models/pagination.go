package models

// PaginationMeta holds pagination info from JSON:API responses.
type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	NextPage    *int `json:"next_page"`
	PrevPage    *int `json:"prev_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
}

// Response wraps a JSON:API response with data and optional pagination meta.
type Response[T any] struct {
	Data T               `json:"data"`
	Meta *PaginationMeta `json:"meta,omitempty"`
}

// Resource represents a JSON:API resource object.
type Resource[T any] struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes T      `json:"attributes"`
}

// ListResponse is a convenience alias for a paginated list of resources.
type ListResponse[T any] struct {
	Data []Resource[T]  `json:"data"`
	Meta *PaginationMeta `json:"meta,omitempty"`
}

// SingleResponse is a convenience alias for a single resource response.
type SingleResponse[T any] struct {
	Data Resource[T] `json:"data"`
}

