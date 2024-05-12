# Модель для работы с базой данных в Golang
---
## Проектирование модели в Go

В этом уроке мы **спроектируем модель для работы с базой данных** нашего проекта.

Если вам не нравится термин *модель*, вы можете использовать название *уровень обслуживания* или *уровень доступа к данным*. Как данный паттерн не назови, идея состоит в том, что мы инкапсулируем код для работы с [MySQL](https://golangs.org/mysql-web-site "Golang MySQL база данных") в отдельный пакет которого можно использовать в остальной части нашего **веб-приложения**.


Сейчас мы создадим скелет модели [базы данных](https://golangs.org/go-sql-driver-mysql "база данных в golang"), которая возвращает тестовые данные. Прежде чем мы перейдем к мельчайшим деталям **SQL-запросов**, давайте разберем данный паттерн поподробнее.

Звучит неплохо? Давайте создадим пару новых папок и файлов в каталоге `pkg`:

```shell
$ cd $HOME/code/snippetbox
$ mkdir -p pkg/models/mysql
$ touch pkg/models/models.go
$ touch pkg/models/mysql/snippets.go
```
![Модель MySQL базы данных](https://golangs.org/wp-content/uploads/2021/04/mysql-model-golang.png)

> **Запомните**: Папка `pkg` используется для хранения вспомогательного кода, не зависящего от приложения, который может использоваться повторно. Модель базы данных, которая может использоваться другими приложениями в будущем (например, приложение с интерфейсом командной строки), соответствует всем требованиям.

Мы начнем с использования файла `pkg/models/models.go`, чтобы определить **типы данных верхнего уровня**, которые модель базы данных будет использовать и возвращать. Откройте данный файл и добавьте следующий код:

```sql
package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: подходящей записи не найдено")

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}
```
Обратите внимание, как названия полей [структуры](https://golangs.org/struct "структуры в golang") `Snippet` соответствуют полям в MySQL таблице `snippets`?

Перейдем к файлу `pkg/models/mysql/snippets.go`, который будет содержать код, специально предназначенный для работы с заметками в нашей **MySQL базе данных**. В этом файле мы определим новый тип `SnippetModel` и реализуем в нем некоторые [методы](https://golangs.org/methods "методы в golang") для доступа и управления базой данных. К примеру:

pkg/models/mysql/snippets.go

```go
package mysql

import (
	"database/sql"

	// Импортируем только что созданный пакет models. Нам нужно добавить к нему префикс
	// независимо от того, какой путь к модулю вы установили
	// чтобы оператор импорта выглядел как
	// "{ваш-путь-модуля}/pkg/models"..
	"golangify.com/snippetbox/pkg/models"
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
```
Важно отметить оператор импорта для `golangify.com/snippetbox/pkg/models`. Обратите внимание, что путь импорта для внутреннего пакета указан с префиксом в виде выбранного пути модуля.

## Создание модели базы данных в Go

Чтобы использовать эту модель в наших обработчиках, требуется вызвать новую структуру `SnippetModel` в [функции](https://golangs.org/func "создание функции в golang") `main()`, а затем внедрить ее как зависимость через [структуру application](https://golangs.org/dependency-injection "Dependency Injection") — точно так же, как мы это делали с другими зависимостями.

Это делается следующим образом:

```go
package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"golangify.com/snippetbox/pkg/models/mysql" // Новый импорт

	_ "github.com/go-sql-driver/mysql"
)

// Добавляем поле snippets в структуру application. Это позволит
// сделать объект SnippetModel доступным для наших обработчиков.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *mysql.SnippetModel
}

func main() {
	addr := flag.String("addr", ":4000", "Сетевой адрес веб-сервера")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "Название MySQL источника данных")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	
	defer db.Close()

	// Инициализируем экземпляр mysql.SnippetModel и добавляем его в зависимостях.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &mysql.SnippetModel{DB: db},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
```
## Преимущества модели базы данных

Такая настройка моделей может показаться немного сложной и запутанной, особенно если вы [новичок в Go](https://golangs.org/go/kurs-izucheniya-golang-dlya-nachinayuschih "курс для чайников голанг"). Однако по мере того, как наше **веб-приложение** продолжает расти, должно стать яснее, почему мы выбрали именно такую структуру.

Если вы вспомните все проделанное, то сможете осознать несколько преимуществ:

-   Присутствует четкое разделение проблем. Логика базы данных не привязана к нашим обработчикам, а это означает, что обязанности обработчика ограничены HTTP элементами (то есть проверкой запросов и возвращение ответов). Это упростит написание тестов в будущем;
-   Создав настраиваемый тип `SnippetModel` и реализовав для него методы, мы смогли сделать модель одним аккуратно инкапсулированным объектом, который можно легко инициализировать и затем передать обработчикам в качестве зависимости. Это упрощает поддержку и тестирование кода;
-   Так как действия модели определены как методы объекта — в данном случае, у `SnippetModel` — есть возможность создать [интерфейс](https://golangs.org/interface "interface golang") и смоделировать его для тестирования;
-   Мы полностью контролируем, какая база данных используется во время работы, используя [флаг из командной строки](https://golangs.org/managing-configuration-settings);
-   В конечном итоге структура каталогов хорошо масштабируется, если у проекта есть несколько бэкэндов. К примеру, если некоторые данные хранятся на **Redis**, можно поместить все их модели в пакет `pkg/models/redis`.

### Скачать исходный код

**Скачать**: [snippetbox-18.zip](https://golangs.org/wp-content/uploads/2021/04/snippetbox-18.zip)


