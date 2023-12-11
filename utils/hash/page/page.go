package gxpage

// Page is the default implementation of Page interface
type Page struct {
	requestOffset int
	pageSize      int
	totalSize     int
	data          []interface{}
	totalPages    int
	hasNext       bool
}

// GetOffSet will return the offset
func (d *Page) GetOffset() int {
	return d.requestOffset
}

// GetPageSize will return the page size
func (d *Page) GetPageSize() int {
	return d.pageSize
}

// GetTotalPages will return the number of total pages
func (d *Page) GetTotalPages() int {
	return d.totalPages
}

// GetData will return the data
func (d *Page) GetData() []interface{} {
	return d.data
}

// GetDataSize will return the size of data.
// it's len(GetData())
func (d *Page) GetDataSize() int {
	return len(d.GetData())
}

// HasNext will return whether has next page
func (d *Page) HasNext() bool {
	return d.hasNext
}

// HasData will return whether this page has data.
func (d *Page) HasData() bool {
	return d.GetDataSize() > 0
}

// NewPage will create an instance
func NewPage(requestOffset int, pageSize int,
	data []interface{}, totalSize int) *Page {

	remain := totalSize % pageSize
	totalPages := totalSize / pageSize
	if remain > 0 {
		totalPages++
	}

	hasNext := totalSize-requestOffset-pageSize > 0

	return &Page{
		requestOffset: requestOffset,
		pageSize:      pageSize,
		data:          data,
		totalSize:     totalSize,
		totalPages:    totalPages,
		hasNext:       hasNext,
	}
}