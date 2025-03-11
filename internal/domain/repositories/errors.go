package repositories

import "errors"

// Определение общих ошибок для репозиториев
var (
	// ErrNotFound возвращается, когда запись не найдена
	ErrNotFound = errors.New("запись не найдена")

	// ErrDuplicateKey возвращается при нарушении уникальности
	ErrDuplicateKey = errors.New("запись с таким ключом уже существует")
)
