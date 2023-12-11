package gxpage


// Pager is the abstraction for pagination usage.
type Pager interface {

	// GetOffset will return the offset
	GetOffset() int

	// GetPageSize will return the page size
	GetPageSize() int

	// GetTotalPages will return the number of total pages
	GetTotalPages() int

	// GetData will return the data
	GetData() []interface{}

	// GetDataSize will return the size of data.
	// Usually it's len(GetData())
	GetDataSize() int

	// HasNext will return whether has next page
	HasNext() bool

	// HasData will return whether this page has data.
	HasData() bool
}
