package book

// Constants for pagination
const (
	// DefaultPage - default page number
	DefaultPage = 1

	// DefaultPageSize - default page size
	DefaultPageSize = 10

	// MaxPageSize - maximum page size
	MaxPageSize = 100
)

// Constants for sorting
const (
	// SortByTitle - sort by title
	SortByTitle = "title"

	// SortByAuthor - sort by author
	SortByAuthor = "author"

	// SortByPrice - sort by price
	SortByPrice = "price"

	// SortByPublished - sort by publication year
	SortByPublished = "year_published"

	// SortByCreated - sort by creation date
	SortByCreated = "created_at"

	// SortAsc - ascending sort order
	SortAsc = "asc"

	// SortDesc - descending sort order
	SortDesc = "desc"
)
