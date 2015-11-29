// This file has been generated by `make gen`
// Do not edit this file, edit generate/collection.got and re-run make.

package srapi

// LevelCollection is list of Level structs. It possible represents
// a slice of the entire dataset and has links to navigate through the pages.
type LevelCollection struct {
	Data       []Level
	Pagination Pagination
}

// LevelWalkerFunc is a function that can be used in Walk(). If it returns
// true, walking continues, else the walk stops.
type LevelWalkerFunc func(l *Level) bool

// Levels returns a list of pointers to the structs; used for cases where
// there is no pagination and the caller wants to return a flat slice of items
// instead of a collection (which would be misleading, as collections imply
// pagination).
func (c *LevelCollection) Levels() []*Level {
	var result []*Level

	c.Walk(func(item *Level) bool {
		result = append(result, item)
		return true
	})

	return result
}

// Walk applies a function to all items in the collection, in order. If the
// function returns false, iterating will be stopped.
func (c *LevelCollection) Walk(f LevelWalkerFunc) {
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
func (c *LevelCollection) Size(fetchAllPages bool) int {
	// we have a simple collection if no pagination information is set
	if len(c.Pagination.Links) == 0 && c.Pagination.Max == 0 {
		return len(c.Data)
	}

	// we have only one page
	if c.Pagination.Size < c.Pagination.Max {
		return len(c.Data)
	}

	if !fetchAllPages {
		return -1
	}

	count := 0

	c.Walk(func(item *Level) bool {
		count++
		return true
	})

	return count
}

// Get returns the n-th element (the first one has idx 0) and nil if there is
// no such index.
func (c *LevelCollection) Get(idx int) *Level {
	// easy, the idx is on this page
	if idx < len(c.Data) {
		return &c.Data[idx]
	}

	// if there is no pagination information, we're out of luck
	if len(c.Pagination.Links) == 0 {
		return nil
	}

	// iterate through the data until we hit the idx we want
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
func (c *LevelCollection) First() *Level {
	if len(c.Data) == 0 {
		return nil
	}

	return &c.Data[0]
}

// ScanForID searches through the collection and looks for an item with the given ID.
func (c *LevelCollection) ScanForID(id string) *Level {
	it := c.Iterator()

	for item := it.Start(); item != nil; item = it.Next() {
		if item.ID == id {
			return item
		}
	}

	return nil
}

// Iterator returns an interator for a LevelCollection. There can be many
// independent iterators starting from the same collection.
func (c *LevelCollection) Iterator() LevelIterator {
	return LevelIterator{
		origin: c,
		cursor: 0,
	}
}

// LevelIterator represents a list of levels.
type LevelIterator struct {
	origin *LevelCollection
	page   *LevelCollection
	cursor int
}

// Start returns the iterator to the start of the original collection page
// and returns the first element if it exists.
func (i *LevelIterator) Start() *Level {
	i.cursor = 0
	i.page = i.origin

	return i.fetch()
}

// Next advances to the next item. If there is no further item, nil is
// returned. All further calls to Next would return nil as well.
func (i *LevelIterator) Next() *Level {
	i.cursor++

	return i.fetch()
}

// fetch tries to return the current item. If it doesn't exist, it attempts
// to fetch the next page and return its first item.
func (i *LevelIterator) fetch() *Level {
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
	page, err := fetchLevels(nextLink.request(nil, nil, NoEmbeds))
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
