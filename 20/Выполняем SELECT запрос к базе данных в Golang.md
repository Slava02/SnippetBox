# Выполняем SELECT запрос к базе данных в Golang

---
Запрос для получения одной записи из базы данных через `SELECT` может показаться немного сложным. Давайте разберемся, как это можно выполнить используя обновленный [метод](https://golangs.org/methods "методы в golang") `SnippetModel.Get()`, чтобы он возвращал определенную заметку на основе её ID.

```sql
SELECT id, title, content, created, expires FROM snippets
WHERE expires > UTC_TIMESTAMP() AND id = ?
```
Поскольку таблица `snippets` использует столбец `id` в качестве первичного ключа, этот запрос всегда будет возвращать только одну запись из базы данных (или вообще ни одной). Запрос также включает проверку **срока годности заметки**, чтобы не возвращались заметки с истекшим сроком.

Обратите внимание, что мы снова используем [плейсхолдер](https://golangs.org/executing-sql-statements#placeholder-parameter "плейсхолдер golang") для `id`, это сделано **в целях безопасности**. Ведь, в будущем ID заметки мы получим от пользователя из URL. Это не безопасный источник для получения данных, потому мы передаём данные через плейсхолдер при **выполнении SQL запроса**.

Откройте файл `pkg/models/mysql/snippets.go` и добавьте следующий код:

```go
package mysql

import (
    "database/sql"
    "errors" // новый импорт
    "golangify.com/snippetbox/pkg/models"
)

type SnippetModel struct {
    DB *sql.DB
}

...

// Get - Метод для возвращения данных заметки по её идентификатору ID.
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	// SQL запрос для получения данных одной записи.
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// Используем метод QueryRow() для выполнения SQL запроса, 
	// передавая ненадежную переменную id в качестве значения для плейсхолдера
	// Возвращается указатель на объект sql.Row, который содержит данные записи.
	row := m.DB.QueryRow(stmt, id)

	// Инициализируем указатель на новую структуру Snippet.
	s := &models.Snippet{}

	// Используйте row.Scan(), чтобы скопировать значения из каждого поля от sql.Row в 
	// соответствующее поле в структуре Snippet. Обратите внимание, что аргументы 
	// для row.Scan - это указатели на место, куда требуется скопировать данные
	// и количество аргументов должно быть точно таким же, как количество 
	// столбцов в таблице базы данных.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// Специально для этого случая, мы проверим при помощи функции errors.Is()
		// если запрос был выполнен с ошибкой. Если ошибка обнаружена, то
		// возвращаем нашу ошибку из модели models.ErrNoRecord.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	// Если все хорошо, возвращается объект Snippet.
	return s, nil
}

...
```
> **На заметку**: Вам может быть интересно, почему мы возвращаем ошибку `models.ErrNoRecord` вместо `sql.ErrNoRows`. Причина в том, что требуется полностью инкапсулировать модель, чтобы приложение не было связано с базовым хранилищем данных или зависело от ошибок базы данных.

## Конвертирование типов из MySQL в Go

Под капотом метода `rows.Scan()`, драйвер базы данных автоматически преобразует MySQL типы в типы языка программирования Go:

-   `CHAR`, `VARCHAR` и `TEXT` соответствуют типу [string](https://golangs.org/string "строки в Golang");
-   `BOOLEAN` соответствует [bool](https://golangs.org/go-for-if-else-switch#true-false "boolean golang");
-   `INT` соответствует [int](https://golangs.org/integer "integer golang");
-   `BIGINT` соответствует [int64](https://golangs.org/big-numbers "int64 golang");
-   `DECIMAL` и `NUMERIC` соответствуют [float](https://golangs.org/float "float голанг");
-   `TIME`, `DATE` и `TIMESTAMP` соответствуют [time.Time](https://golangs.org/go/time "дата и время в golang").

Особенность нашего [MySQL драйвера](https://golangs.org/go-sql-driver-mysql "MySQL Голанг") заключается в том, нам нужно использовать параметр `parseTime=true` в нашей [строке подключения к MySQL](https://golangs.org/mysql-connection-pool-go "DSN MySQL"), чтобы заставить его преобразовывать поля `TIME` и `DATE` в `time.Time`. В противном случае, он вернёт их как объекты `[]byte`. Это один из многих [предлагаемых параметров](https://github.com/go-sql-driver/mysql#parameters) для предварительной настройки драйвера базы данных при подключении к нему.

## Модели базы данных в обработчиках

Отлично, давайте воспользуемся методом `SnippetModel.Get()` на практике.

Откройте файл `cmd/web/handlers.go` и обновите обработчик `showSnippet`, чтобы он возвращал данные для определенной записи:

```go
package main

import (
    "errors" // новый импорт
    "fmt"
    "html/template"
    "net/http"
    "strconv"

    "golangify.com/snippetbox/pkg/models" // новый импорт
)

...

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w) // Страница не найдена.
		return
	}

	// Вызываем метода Get из модели Snipping для извлечения данных для
	// конкретной записи на основе её ID. Если подходящей записи не найдено,
	// то возвращается ответ 404 Not Found (Страница не найдена).
	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Отображаем весь вывод на странице.
	fmt.Fprintf(w, "%v", s)
}

...
```
Запускаем наше веб-приложение из терминала:

```shell
go run ./cmd/web
```

Перейдите в браузере на страницу [http://127.0.0.1:4000/snippet?id=1](http://127.0.0.1:4000/snippet?id=1). Вы должны увидеть HTTP ответ, похожий на следующий:

![SELECT запрос в Golang](https://golangs.org/wp-content/uploads/2021/04/golang-select-one-snippet.png)

Также, можете попробовать сделать несколько запросов для других заметок, срок действия которых истек или которых **вообще не существуют** (например, `id=99`), чтобы убедиться, что они возвращают ответ `404 Not Found`:

![MySQL запись не найдена](https://golangs.org/wp-content/uploads/2021/04/snippet-page-not-found-id.png)

## Обработка ошибок используя errors.Is()

В коде выше мы использовали [функцию](https://golangs.org/func) [errors.Is()](https://tip.golang.org/pkg/errors/#Is), которая была введена в **Go 1.13**, чтобы проверить, нужная ли нам ошибка выскочила (в нашем случае проверяется, если нам попалась именно ошибка `sql.ErrNoRows`).

Обсудим данный вопрос немного подробнее.

Во-первых, [sql.ErrNoRows](https://golang.org/pkg/database/sql/#pkg-variables) является примером так называемых **предопределённых ошибок** (по аналогии с Python, это будут «**исключения**» exceptions), которых мы можем приблизительно определить как объект `error`, хранящийся в [глобальной переменной](https://golangs.org/variable-scope "область видимости"). Обычно они создаются с помощью функции `errors.New()`. Несколько примеров  ошибок из стандартной библиотеки — [io.ErrUnexpectedEOF](https://golang.org/pkg/io/#pkg-variables) и [bytes.ErrTooLarge](https://golang.org/pkg/bytes/#pkg-variables).

Только что созданная нами ошибка `models.ErrNoRecord` является примером **предопределённой ошибки**.

> У нас есть её «имя» и мы можем «ловить» её в случае если она возникнет при работе программы. Всё работает по принципу `try - catch` в Java, PHP или [try — except в Python](https://python-scripts.com/try-except-finally "обработка ошибок в Python").

В старых версиях Go (до 1.13) лучший способ проверить, если ошибка совпадает с **предопределённой ошибкой**, выглядел бы так:

```go
if err == sql.ErrNoRows {
    // Делает что-то
} else {
    // Делает что-то другое
}
```
Однако после Go 1.13 **лучше использовать функцию** `errors.Is()`:

```go
if errors.Is(err, sql.ErrNoRows) {
    // Делает что-то
} else {
    // Делает что-то другое
}
```
Причина в том, что в Go 1.13 появилась возможность **оборачивать ошибки** для добавления дополнительной информации. По сути, мы создаём новую ошибку на базе уже существующей ошибки.

Для строго стиля проверки на ошибку — предопределённая ошибка и созданная (обёрнутая) на основе неё другая новая ошибка — не будут совпадать, так как **обернутая ошибка** не равна оригинальной предопределённой ошибке.

Функция `errors.Is()` наоборот работает путем распаковки ошибок — при необходимости — перед проверкой на совпадения.

Если у вас Go 1.13 или новее, лучше использовать `error.Is()`. Это хороший способ защитить код в будущем и предотвратить проблемы, вызванные вами — или любыми пакетами, которые ваш код импортирует.

Существует также другая функция, [errors.As()](https://tip.golang.org/pkg/errors/#As), которую вы можете использовать, чтобы проверить, есть ли (у потенциально обернутой — *wrapped*) ошибки определенный тип. Мы будем использовать эту функцию в будущих уроках.

Для получения дополнительной информации об изменениях в [способах обработки ошибок](https://golangs.org/errors "ошибки в Golang") в Go 1.13 можете прочитать [официальный FAQ](https://github.com/golang/go/wiki/ErrorValueFAQ).

## Метод для получения данных записи

Мы специально сделали код в `SnippetModel.Get()` немного длиннее, чтобы было проще понять, что именно происходит за кулисами вашего кода.

На практике, можно сократить код (или хотя бы количество строк), воспользовавшись тем, что появление ошибки из `DB.QueryRow()` откладываются до вызова `Scan()`. Функциональной разницы нет, так что если вы хотите, можете переписать код следующим образом:

```go
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
    s := &models.Snippet{}
    err := m.DB.QueryRow("SELECT ...", id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, models.ErrNoRecord
        } else {
             return nil, err
        }
    }

    return s, nil
}
```
## Скачать исходный код урока

**Скачать**: [snippetbox-20.zip](https://golangs.org/wp-content/uploads/2021/04/snippetbox-20.zip)


