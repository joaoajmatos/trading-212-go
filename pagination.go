package trading212

import "context"

// Page is the raw decoded response from a paginated Trading 212 endpoint.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextPagePath"` // Trading 212 returns the full path
}

// Cursor is a stateful, lazy iterator over pages of T returned by a paginated
// endpoint. Use it like a scanner:
//
//	cur := client.HistoryOrders(ctx, trading212.HistoryOrdersParams{})
//	for cur.Next(ctx) {
//	    item := cur.Item()
//	    _ = item
//	}
//	if err := cur.Err(); err != nil {
//	    log.Fatal(err)
//	}
//
// Cursor is not safe for concurrent use.
type Cursor[T any] struct {
	// fetch is called with the cursor token to retrieve the next page.
	// An empty token fetches the first page.
	fetch  func(ctx context.Context, cursor string) (Page[T], error)
	cursor string
	buf    []T
	pos    int
	done   bool
	err    error
}

// newCursor creates a new Cursor. fetch must not be nil.
func newCursor[T any](fetch func(ctx context.Context, cursor string) (Page[T], error)) *Cursor[T] {
	return &Cursor[T]{fetch: fetch}
}

// Next advances to the next item, fetching the next page from the API if
// needed. Returns false when all items have been consumed or an error occurs.
// Call [Cursor.Err] after Next returns false to distinguish exhaustion from
// errors.
func (c *Cursor[T]) Next(ctx context.Context) bool {
	if c.done || c.err != nil {
		return false
	}

	// Still have buffered items from the current page.
	if c.pos < len(c.buf) {
		c.pos++
		return true
	}

	// Need to fetch the next page.
	page, err := c.fetch(ctx, c.cursor)
	if err != nil {
		c.err = err
		return false
	}

	if len(page.Items) == 0 {
		c.done = true
		return false
	}

	c.buf = page.Items
	c.pos = 1 // point to first item (index 0)

	if page.NextCursor == "" {
		c.done = true
	} else {
		c.cursor = page.NextCursor
	}

	return true
}

// Item returns the current item. It is valid only after [Cursor.Next] has
// returned true.
func (c *Cursor[T]) Item() T {
	return c.buf[c.pos-1]
}

// Err returns the first error encountered during iteration, if any.
func (c *Cursor[T]) Err() error {
	return c.err
}
