// This file has been generated by `make gen`
// Do not edit this file, edit generate/collection.got and re-run make.

package srapi

// SeriesCollection is list of Series structs. It possible represents
// a slice of the entire dataset and has links to navigate through the pages.
type SeriesCollection struct {
	Data       []Series
	Pagination Pagination
	limit      int
}

// SeriesWalkerFunc is a function that can be used in Walk(). If it returns
// true, walking continues, else the walk stops.
type SeriesWalkerFunc func(s *Series) bool

// Limit returns a copy of the collection that is limited to a maximum amount
// of items in it. This is useful because the Cursor type does *not* affect
// how many items are in a collection, but only how many are fetched per
// request.
func (c *SeriesCollection) Limit(limit int) *SeriesCollection {
	return &SeriesCollection{
		Data:       c.Data,
		Pagination: c.Pagination,
		limit:      limit,
	}
}

// ManySeries returns a list of pointers to the structs; used for cases where
// there is no pagination and the caller wants to return a flat slice of items
// instead of a collection (which would be misleading, as collections imply
// pagination).
func (c *SeriesCollection) ManySeries() []*Series {
	var result []*Series

	c.Walk(func(item *Series) bool {
		result = append(result, item)
		return true
	})

	return result
}

// Walk applies a function to all items in the collection, in order. If the
// function returns false, iterating will be stopped.
func (c *SeriesCollection) Walk(f SeriesWalkerFunc) {
	it := c.Iterator()

	for item := range it.Output() {
		if !f(item) {
			it.Stop()
		}
	}
}

// Size returns the number of elements in the collection; returns -1 if the total
// number cannot be determined without iterating over additional pages (which
// requires network roundtrips) and fetchAllPages is set to false
func (c *SeriesCollection) Size(fetchAllPages bool) int {
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

	c.Walk(func(item *Series) bool {
		count++
		return true
	})

	return count
}

// Get returns the n-th element (the first one has idx 0) and nil if there is
// no such index.
func (c *SeriesCollection) Get(idx int) *Series {
	cur := 0
	it := c.Iterator()

	for item := range it.Output() {
		if cur == idx {
			return item
		}

		cur++
	}

	return nil
}

// First returns the first element, if any, otherwise nil.
func (c *SeriesCollection) First() *Series {
	if len(c.Data) == 0 {
		return nil
	}

	return &c.Data[0]
}

// ScanForID searches through the collection and looks for an item with the given ID.
func (c *SeriesCollection) ScanForID(id string) *Series {
	it := c.Iterator()

	for item := range it.Output() {
		if item.ID == id {
			return item
		}
	}

	return nil
}

// Iterator returns an interator for a SeriesCollection. There can be many
// independent iterators starting from the same collection.
func (c *SeriesCollection) Iterator() SeriesIterator {
	it := SeriesIterator{
		output:     make(chan *Series),
		killSwitch: make(chan struct{}),
		origin:     c,
		limit:      c.limit,
	}

	go it.work()

	return it
}

// SeriesIterator represents a list of manySeries.
type SeriesIterator struct {
	output     chan *Series
	killSwitch chan struct{}
	origin     *SeriesCollection
	limit      int
}

// Output returns a channel that can be used to read all manySeries
// from the iterator.
func (i *SeriesIterator) Output() <-chan *Series {
	return i.output
}

// Stop interrupts the iterator and cancels all further pending action. After
// calling this, the iterator returns no more manySeries and becomes
// unusable.
func (i *SeriesIterator) Stop() {
	close(i.killSwitch)

	// drain the remaining element(s)
	for _ = range i.output {
	}
}

// work is the goroutine that reads items from the current page and
// fetches new pages until all pages are fetched or the iteration is stopped.
func (i *SeriesIterator) work() {
	page := i.origin
	first := true
	remaining := i.limit

	defer close(i.output)

	for {
		select {
		case <-i.killSwitch:
			return

		default:
			// if this is not the first iteration, fetch the next page to work on
			if !first {
				// is there another one?
				nextLink := firstLink(&page.Pagination, "next")
				if nextLink == nil {
					return
				}

				// fetch the next page
				p, err := fetchManySeries(nextLink.request(nil, nil, NoEmbeds))
				if err != nil {
					return
				}

				// is this page empty?
				if len(p.Data) == 0 {
					return
				}

				// use this page from now on
				page = p
			}

			for idx := 0; idx < len(page.Data); idx++ {
				select {
				case <-i.killSwitch:
					return

				default:
					i.output <- &page.Data[idx]
					remaining--
				}

				// stop we we exhausted all allowed elements
				if i.limit > 0 && remaining <= 0 {
					return
				}
			}

			first = false
		}
	}
}
