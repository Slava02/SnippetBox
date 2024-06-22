# Кэширование сайта на Golang

---
Перед добавлением дополнительных возможностей для нашего шаблона мы внесем некоторые изменения в коде веб-приложения. На данный момент у нас есть две проблемы:

1.  Каждый раз при отображении веб-страницы наше приложение обрабатывает файлы шаблонов с помощью функции `template.ParseFiles()`. Мы можем избежать этой повторяющейся работы, обработав файлы один раз во время запуска приложения — и сохранив обработанные шаблоны в кэше в памяти;
2.  В обработчиках `home` и `showSnippet` есть код который повторяется. Повторяющийся код мы поместим в функцию-помощник.

Сначала займемся первым пунктом и создадим [карту](https://golangify.com/map) типа `map[string]*template.Template` для **кэширования обработанных шаблонов**. Откройте файл `cmd/web/templates.go` и добавьте туда следующий код:

```go
package main

import (
    "github.com/Slava02/SnippetBox/25/pkg/models"
    "html/template" // новый импорт
    "path/filepath" // новый импорт
)

// Эту структуру не трогаем, она тут из прошлого урока...
type templateData struct {
    Snippet  *models.Snippet
    Snippets []*models.Snippet
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
    // Инициализируем новую карту, которая будет хранить кэш.
    cache := map[string]*template.Template{}

    // Используем функцию filepath.Glob, чтобы получить срез всех файловых путей с
    // расширением '.page.tmpl'. По сути, мы получим список всех файлов шаблонов для страниц
    // нашего веб-приложения.
    pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
    if err != nil {
        return nil, err
    }

    // Перебираем файл шаблона от каждой страницы.
    for _, page := range pages {
        // Извлечение конечное названия файла (например, 'home.page.tmpl') из полного пути к файлу
        // и присваивание его переменной name.
        name := filepath.Base(page)

        // Обрабатываем итерируемый файл шаблона.
        ts, err := template.ParseFiles(page)
        if err != nil {
            return nil, err
        }

        // Используем метод ParseGlob для добавления всех каркасных шаблонов.
        // В нашем случае это только файл base.layout.tmpl (основная структура шаблона).
        ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
        if err != nil {
            return nil, err
        }

        // Используем метод ParseGlob для добавления всех вспомогательных шаблонов.
        // В нашем случае это footer.partial.tmpl "подвал" нашего шаблона.
        ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
        if err != nil {
            return nil, err
        }

        // Добавляем полученный набор шаблонов в кэш, используя название страницы
        // (например, home.page.tmpl) в качестве ключа для нашей карты.
        cache[name] = ts
    }

    // Возвращаем полученную карту.
    return cache, nil
}
```
Следующим шагом будет инициализация данного кэша в функцию `main()`. Также надо сделать, чтобы кэш был доступен для наших обработчиков как зависимость через [структуру application](https://golangify.com/dependency-injection).

Реализуем это следующим образом:

```go
package main

import (
    "database/sql"
    "flag"
    "html/template" // Новый импорт
    "log"
    "net/http"
    "os"

    _ "github.com/go-sql-driver/mysql"
    "github.com/Slava02/SnippetBox/25/pkg/models/mysql"
)

// Добавляем поле templateCache в структуру зависимостей. Это позволит
// получить доступ к кэшу во всех обработчиках.
type application struct {
    errorLog      *log.Logger
    infoLog       *log.Logger
    snippets      *mysql.SnippetModel
    templateCache map[string]*template.Template
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

    // Инициализируем новый кэш шаблона...
    templateCache, err := newTemplateCache("./ui/html/")
    if err != nil {
        errorLog.Fatal(err)
    }

    // И добавляем его в зависимостях нашего
    // веб-приложения.
    app := &application{
        errorLog:      errorLog,
        infoLog:       infoLog,
        snippets:      &mysql.SnippetModel{DB: db},
        templateCache: templateCache,
    }

    srv := &http.Server{
        Addr:     *addr,
        ErrorLog: errorLog,
        Handler:  app.routes(),
    }

    infoLog.Printf("Запуск сервера на http://127.0.0.1%s", *addr)
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
На данный момент в памяти есть кэш с соответствующим набором шаблонов для каждой из страниц, и у обработчиков есть доступ к этому кэшу из структуры зависимостей `application`. Теперь займемся второй проблемой — **дублирующимся кодом в обработчиках** и создадим **метод-помощник**, чтобы можно было легко отображать шаблоны из кэша не повторяя один и тот же код каждый раз во всех обработчиках.

Откройте файл `cmd/web/helpers.go` и [создаем метод](https://golangify.com/methods) `render`:

```go
package main

...

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
    // Извлекаем соответствующий набор шаблонов из кэша в зависимости от названия страницы
    // (например, 'home.page.tmpl'). Если в кэше нет записи запрашиваемого шаблона, то
    // вызывается вспомогательный метод serverError(), который мы создали ранее.
    ts, ok := app.templateCache[name]
    if !ok {
        app.serverError(w, fmt.Errorf("Шаблон %s не существует!", name))
        return
    }

    // Рендерим файлы шаблона, передавая динамические данные из переменной `td`.
    err := ts.Execute(w, td)
    if err != nil {
        app.serverError(w, err)
    }
}
```
Возможно вы гадаете, почему метод `render()` имеет параметр `*http.Request` — даже если он нигде не используется. Это нужно для защиты сигнатуры метода в будущем, потому что в следующих уроках мы расширим возможности метода `render` и нам понадобится этот параметр.

Теперь у нас всё готово, осталось обновить код в обработчиках `home()` и `showSnippet()` следующим образом:

```go
package main

import (
    "errors"
    "fmt"
    "github.com/Slava02/SnippetBox/25/pkg/models"
    "net/http"
    "strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        app.notFound(w)
        return
    }

    s, err := app.snippets.Latest()
    if err != nil {
        app.serverError(w, err)
        return
    }

    // Используем помощника render() для отображения шаблона.
    app.render(w, r, "home.page.tmpl", &templateData{
        Snippets: s,
    })
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil || id < 1 {
        app.notFound(w)
        return
    }

    s, err := app.snippets.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFound(w)
        } else {
            app.serverError(w, err)
        }
        return
    }

    // Используем помощника render() для отображения шаблона.
    app.render(w, r, "show.page.tmpl", &templateData{
        Snippet: s,
    })
}

...
```
## Скачать исходный код

**Скачать**: [snippetbox-25](https://golangify.com/wp-content/uploads/2021/10/snippetbox-25.zip)
