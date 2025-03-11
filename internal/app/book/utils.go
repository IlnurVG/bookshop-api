package book

import (
	"strings"

	"github.com/bookshop/api/internal/domain/models"
)

// FormatBookTitle форматирует название книги
func FormatBookTitle(title string) string {
	// Удаляем лишние пробелы
	title = strings.TrimSpace(title)

	// Приводим первую букву к верхнему регистру
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

// FormatBookAuthor форматирует имя автора
func FormatBookAuthor(author string) string {
	// Удаляем лишние пробелы
	author = strings.TrimSpace(author)

	// Разделяем имя автора на части
	parts := strings.Split(author, " ")

	// Форматируем каждую часть
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	// Объединяем части обратно
	return strings.Join(parts, " ")
}

// CalculateTotalPages вычисляет общее количество страниц
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

// ValidatePageParams проверяет и корректирует параметры пагинации
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

// FilterBooksByCategory фильтрует книги по категории
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

// FilterBooksByPriceRange фильтрует книги по диапазону цен
func FilterBooksByPriceRange(books []models.Book, minPrice, maxPrice *float64) []models.Book {
	filtered := make([]models.Book, 0, len(books))

	for _, book := range books {
		// Проверяем минимальную цену
		if minPrice != nil && book.Price < *minPrice {
			continue
		}

		// Проверяем максимальную цену
		if maxPrice != nil && book.Price > *maxPrice {
			continue
		}

		filtered = append(filtered, book)
	}

	return filtered
}
