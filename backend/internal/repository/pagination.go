package repository

// PageParams holds cursor-based pagination query parameters.
type PageParams struct {
	Limit  int
	Cursor string
}

// PageResult holds a page of items and an optional next cursor.
type PageResult[T any] struct {
	Items      []T
	NextCursor *string
}
