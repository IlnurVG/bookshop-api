package book

// Константы для пагинации
const (
	// DefaultPage - страница по умолчанию
	DefaultPage = 1

	// DefaultPageSize - размер страницы по умолчанию
	DefaultPageSize = 10

	// MaxPageSize - максимальный размер страницы
	MaxPageSize = 100
)

// Константы для сортировки
const (
	// SortByTitle - сортировка по названию
	SortByTitle = "title"

	// SortByAuthor - сортировка по автору
	SortByAuthor = "author"

	// SortByPrice - сортировка по цене
	SortByPrice = "price"

	// SortByPublished - сортировка по году публикации
	SortByPublished = "year_published"

	// SortByCreated - сортировка по дате создания
	SortByCreated = "created_at"

	// SortAsc - сортировка по возрастанию
	SortAsc = "asc"

	// SortDesc - сортировка по убыванию
	SortDesc = "desc"
)
