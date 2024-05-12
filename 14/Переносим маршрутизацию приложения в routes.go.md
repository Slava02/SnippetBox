# Переносим маршрутизацию приложения в routes.go
---
## Изоляция маршрутизации приложения в отдельный файл

Пока мы проводим **рефакторинг кода**, стоит внести еще одно изменение. Наша главная функция `main()` становится немного переполненной.Чтобы она стала более компактной и легко читаемой, можно переместить [объявления маршрутов](https://golangs.org/routing-servemux "роутинг в golang") для приложения в отдельный файл `routes.go`. Например:

```shell
$ cd $HOME/code/snippetbox
$ touch cmd/web/routes.go
```

```go
package main

import "net/http"

func (app *application) routes() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/", app.home)
    mux.HandleFunc("/snippet", app.showSnippet)
    mux.HandleFunc("/snippet/create", app.createSnippet)

    fileServer := http.FileServer(http.Dir("./ui/static/"))
    mux.Handle("/static/", http.StripPrefix("/static", fileServer))

    return mux
}
```
Теперь мы можем обновить файл `main.go` следующим образом:

```go
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	addr := flag.String("addr", ":4000", "Сетевой адрес веб-сервера")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(), // Вызов нового метода app.routes()
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
```
Такой вариант выглядит аккуратнее. Маршруты для нашего приложения теперь изолированы и инкапсулированы в методе `app.routes()`, а обязанности функции `main()` сводятся к следующему:

-   Парсинг настроек конфигурации среды выполнения для приложения;
-   Установление зависимостей для обработчиков;
-   Запуск HTTP-сервера.

### Исходный код веб-приложения

**Скачать**: [snippetbox-14.zip](https://golangs.org/wp-content/uploads/2021/01/snippetbox-14.zip)

