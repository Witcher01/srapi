// This file has been generated by `make gen`
// Do not edit this file, edit generate/collection.got and re-run make.

package srapi

// RunCollection is list of Run structs. It possible represents
// a slice of the entire dataset and has links to navigate through the pages.
type RunCollection struct {
	Data       []Run
	Pagination Pagination
	limit      int
}

// RunWalkerFunc is a function that can be used in Walk(). If it returns
// true, walking continues, else the walk stops.
type RunWalkerFunc func(r *Run) bool

// Limit returns a copy of the collection that is limited to a maximum amount
// of items in it. This is useful because the Cursor type does *not* affect
// how many items are in a collection, but only how many are fetched per
// request.
func (c *RunCollection) Limit(limit int) *RunCollection {
	return &RunCollection{
		Data:       c.Data,
		Pagination: c.Pagination,
		limit:      limit,
	}
}

// Runs returns a list of pointers to the structs; used for cases where
// there is no pagination and the caller wants to return a flat slice of items
// instead of a collection (which would be misleading, as collections imply
// pagination).
func (c *RunCollection) Runs() []*Run {
	var result []*Run

	c.Walk(func(item *Run) bool {
		result = append(result, item)
		return true
	})

	return result
}

// Walk applies a function to all items in the collection, in order. If the
// function returns false, iterating will be stopped.
func (c *RunCollection) Walk(f RunWalkerFunc) {
	it := c.Iterator()

	for item := it.Start(); item != nil; item = it.Next() {
		if !f(item) {
			break
		}
	}
}

// Size returns the number of elements in the collection; returns -1 if the total
// number cannot be determined without iterating over additional pages (which
// requires network roundtrips) and fetchAllPages is set to false
func (c *RunCollection) Size(fetchAllPages bool) int {
	length := len(c.Data)
	if c.limit > 0 && length > c.limit {
		length = c.limit
	}

	// we have a simple collection if no pagination information is set
	if len(c.Pagination.Links) == 0 && c.Pagination.Max == 0 {
		return length
	}

	// we have only one page
	if c.Pagination.Size < c.Pagination.Max {
		return length
	}

	if !fetchAllPages {
		return -1
	}

	count := 0

	c.Walk(func(item *Run) bool {
		count++
		return true
	})

	return count
}

// Get returns the n-th element (the first one has idx 0) and nil if there is
// no such index.
func (c *RunCollection) Get(idx int) *Run {
	cur := 0
	it := c.Iterator()

	for item := it.Start(); item != nil; item = it.Next() {
		if cur == idx {
			return item
		}

		cur++
	}

	return nil
}

// First returns the first element, if any, otherwise nil.
func (c *RunCollection) First() *Run {
	if len(c.Data) == 0 {
		return nil
	}

	return &c.Data[0]
}

// ScanForID searches through the collection and looks for an item with the given ID.
func (c *RunCollection) ScanForID(id string) *Run {
	it := c.Iterator()

	for item := it.Start(); item != nil; item = it.Next() {
		if item.ID == id {
			return item
		}
	}

	return nil
}

// Iterator returns an interator for a RunCollection. There can be many
// independent iterators starting from the same collection.
func (c *RunCollection) Iterator() RunIterator {
	return RunIterator{
		origin:    c,
		cursor:    0,
		limit:     c.limit,
		remaining: c.limit,
	}
}

// RunIterator represents a list of runs.
type RunIterator struct {
	origin    *RunCollection
	page      *RunCollection
	cursor    int
	limit     int
	remaining int
}

// Start returns the iterator to the start of the original collection page
// and returns the first element if it exists.
func (i *RunIterator) Start() *Run {
	i.cursor = 0
	i.page = i.origin
	i.remaining = i.limit

	return i.fetch()
}

// Next advances to the next item. If there is no further item, nil is
// returned. All further calls to Next would return nil as well.
func (i *RunIterator) Next() *Run {
	i.cursor++

	return i.fetch()
}

// fetch tries to return the current item. If it doesn't exist, it attempts
// to fetch the next page and return its first item.
func (i *RunIterator) fetch() *Run {
	// handle item limit
	if i.limit > 0 {
		if i.remaining <= 0 {
			return nil
		}

		i.remaining--
	}

	// easy, just get the next item on the current page
	if i.cursor < len(i.page.Data) {
		return &i.page.Data[i.cursor]
	}

	// we reached the end of the current page; is there another one?
	nextLink := firstLink(&i.page.Pagination, "next")
	if nextLink == nil {
		return nil
	}

	// fetch the next page
	page, err := fetchRuns(nextLink.request(nil, nil, NoEmbeds))
	if err != nil {
		return nil
	}

	i.page = page
	i.cursor = 0

	if i.cursor < len(i.page.Data) {
		return &i.page.Data[i.cursor]
	}

	return nil
}
