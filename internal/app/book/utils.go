package book

import (
	"strings"

	"github.com/bookshop/api/internal/domain/models"
)

// FormatBookTitle formats the book title
func FormatBookTitle(title string) string {
	// Remove extra spaces
	title = strings.TrimSpace(title)

	// Capitalize the first letter
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

// FormatBookAuthor formats the author's name
func FormatBookAuthor(author string) string {
	// Remove extra spaces
	author = strings.TrimSpace(author)

	// Split the author's name into parts
	parts := strings.Split(author, " ")

	// Format each part
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	// Join the parts back together
	return strings.Join(parts, " ")
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(totalCount, pageSize int) int {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	totalPages := totalCount / pageSize
	if totalCount%pageSize > 0 {
		totalPages++
	}

	return totalPages
}

// ValidatePageParams checks and corrects pagination parameters
func ValidatePageParams(page, pageSize int) (int, int) {
	if page <= 0 {
		page = DefaultPage
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return page, pageSize
}

// FilterBooksByCategory filters books by category
func FilterBooksByCategory(books []models.Book, categoryID int) []models.Book {
	if categoryID <= 0 {
		return books
	}

	filtered := make([]models.Book, 0, len(books))
	for _, book := range books {
		if book.CategoryID == categoryID {
			filtered = append(filtered, book)
		}
	}

	return filtered
}

// FilterBooksByPriceRange filters books by price range
func FilterBooksByPriceRange(books []models.Book, minPrice, maxPrice *float64) []models.Book {
	filtered := make([]models.Book, 0, len(books))

	for _, book := range books {
		// Check minimum price
		if minPrice != nil && book.Price < *minPrice {
			continue
		}

		// Check maximum price
		if maxPrice != nil && book.Price > *maxPrice {
			continue
		}

		filtered = append(filtered, book)
	}

	return filtered
}
