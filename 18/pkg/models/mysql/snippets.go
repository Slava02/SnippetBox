package mysql

import (
	"database/sql"

	// Импортируем только что созданный пакет models. Нам нужно добавить к нему префикс
	// независимо от того, какой путь к модулю вы установили
	// чтобы оператор импорта выглядел как
	// "{ваш-путь-модуля}/pkg/models"..
	"golangs.org/snippetbox/pkg/models"
)

// SnippetModel - Определяем тип который обертывает пул подключения sql.DB
type SnippetModel struct {
	DB *sql.DB
}

// Insert - Метод для создания новой заметки в базе дынных.
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	return 0, nil
}

// Get - Метод для возвращения данных заметки по её идентификатору ID.
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	return nil, nil
}

// Latest - Метод возвращает 10 наиболее часто используемые заметки.
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return nil, nil
}
