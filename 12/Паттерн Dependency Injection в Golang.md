# Паттерн Dependency Injection в Golang

---
Есть еще одна проблема [логирования](https://golangs.org/logging "лог в golang"), с которой нам нужно разобраться. При просмотре файла `handlers.go` вы заметите, что функция обработчик `home()` по-прежнему записывает сообщения об ошибках с использованием **стандартного логгера от Go**, а не созданный нами логгер `errorLog`, которого нам нужно использовать.

```go
func home(w http.ResponseWriter, r *http.Request) {
    ...

    ts, err := template.ParseFiles(files...)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, "Internal Server Error", 500)
        return
    }

    err = ts.Execute(w, nil)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, "Internal Server Error", 500)
    }
}
```
Возникает вопрос: как сделать наш новый логгер `errorLog` доступным для функции `home` из `main()` ?

В текущем виде кода, логгеры **infoLog** и **errorLog** не видны из за [области видимости](https://golangs.org/variable-scope "область видимости в golang") функции `main()`.

У большинства веб-приложений будет несколько зависимостей, к которым их обработчики должны обращаться. Например пул соединений с базой данных, централизованные обработчики ошибок и кэши [шаблонов](https://golangs.org/template-engine "шаблонизатор в golang"). Однако, ответ на вопрос который нам действительно нужен, это: *как сделать любую зависимость доступной нашим обработчикам*?

Для этого есть несколько способов. Самый простой — просто поместить зависимости в [глобальные переменные](https://golangs.org/variable-scope#global-scope) которых видно везде. В целом рекомендуется **внедрять зависимости в обработчики**. Это делает код более явным, менее подверженным ошибкам и более простым для модульного тестирования, чем в случае **использования глобальных переменных**.

Для приложений, в которых все обработчики находятся в одном пакете (как наше приложение), отличным способом внедрения зависимостей будет их размещение в [структуру](https://golangs.org/struct "структуры в golang") `application`.

Ниже мы продемонстрируем всё на примере.

Откройте файл `main.go` и **создайте новую структуру** `application` следующим образом:

```go
package main

import (
    "flag"
    "log"
    "net/http"
    "os"
)

// Создаем структуру `application` для хранения зависимостей всего веб-приложения.
// Пока, что мы добавим поля только для двух логгеров, но
// мы будем расширять данную структуру по мере усложнения приложения.
type application struct {
    errorLog *log.Logger
    infoLog  *log.Logger
}


func main() {
    ...
}
```
Затем в файле `handlers.go` обновим обработчик `home()` так, чтобы ему стали доступны поля из структуры `application`, для этого он должен стать частью этой структуры.

Если данная тема кажется запутанной, то советую ознакомиться [методами в golang](https://golangs.org/methods) из нашего [курса по изучению go](https://golangs.org/go/kurs-izucheniya-golang-dlya-nachinayuschih).

```go
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// Меняем сигнатуры обработчика home, чтобы он определялся как метод
// структуры *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		// Поскольку обработчик home теперь является методом структуры application
		// он может получить доступ к логгерам из структуры.
		// Используем их вместо стандартного логгера от Go.
		app.errorLog.Println(err.Error())
		http.Error(w, "Внутренняя ошибка сервера", 500)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		// Обновляем код для использования логгера-ошибок
		// из структуры application.
		app.errorLog.Println(err.Error())
		http.Error(w, "Внутренняя ошибка сервера", 500)
	}
}

// Меняем сигнатуру обработчика showSnippet, чтобы он был определен как метод
// структуры *application
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Отображение выбранной заметки с ID %d...", id)
}

// Меняем сигнатуру обработчика createSnippet, чтобы он определялся как метод
// структуры *application.
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", 405)
		return
	}

	w.Write([]byte("Создание новой заметки..."))
}
```
Обновляем файл `main.go` в котором мы определяем структуру **application**:

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

	// Инициализируем новую структуру с зависимостями приложения.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	// Используем методы из структуры в качестве обработчиков маршрутов.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/snippet/create", app.createSnippet)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
```
Такой подход может показаться немного сложным и запутанным, особенно когда альтернативой является простое создание **глобальных переменных** логгеров `infoLog` и `errorLog`. По мере того, как приложение будет расти, наши обработчики будут требовать больше зависимостей. Тогда мы и увидим всю прелесть **паттерна Dependency Injection**.

## Тестирование логгеров в Golang

Давайте попробуем специально **создать ошибку** в наше веб-приложение.

Откройте терминал и переименуйте `ui/html/home.page.tmpl` на `ui/html/home.page.bak`.Теперь при запуске веб-приложения и создании HTTP-запроса к домашней страницы, результатом будет ошибка, потому что файл `ui/html/home.page.tmpl` больше не существует.

Попробуйте сделать изменение:

```shell
$ cd $HOME/code/snippetbox
$ mv ui/html/home.page.tmpl ui/html/home.page.tmpl
```
Затем, запустите приложение и выполните запрос на [http://127.0.0.1:4000](http://127.0.0.1:4000/). В браузере вы должны получить HTTP ответ `Внутренняя ошибка сервера`, а также соответствующее сообщение об ошибке в терминале, наподобие следующего:

![Логирование ошибок в Golang](https://golangs.org/wp-content/uploads/2021/01/logging-error-golang.jpg)

Обратите внимание, что у сообщения об ошибке теперь есть префикс `ERROR` она была сгенерирована на 29-й строки файла `handlers.go`. Это наглядно демонстрирует, что созданный нами логгер `errorLog` передается обработчику `home()` в качестве зависимости и работает как и ожидалось.

> Оставьте пока **преднамеренную ошибку** с несуществующим файлом шаблона `home.page.tmpl`  — она нам понадобится в следующем уроке.

### Замыкания для внедрения зависимостей

Паттерн, который мы используем для **внедрения зависимостей**, не будет работать, если обработчики распределены по нескольким пакетам. В этом случае, альтернативным подходом будет создание пакета `config`, который экспортирует структуру `Application`. К примеру:

```go
func main() {
    app := &config.Application{
        ErrorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
    }

    mux.Handle("/", handlers.Home(app))
}
```
```go
func Home(app *config.Application) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ...
        ts, err := template.ParseFiles(files...)
        if err != nil {
            app.ErrorLog.Println(err.Error())
            http.Error(w, "Внутренняя ошибка сервера", 500)
            return
        }
        ...
    }
}
```
Вы можете найти более подробный пример использования паттерна замыкания ознакомившись со статьей: [Замыкания и анонимные функции](https://golangs.org/closures-anonymous-first-class-func#closures-and-anonymous-func)

[крутое видео по данной теме](https://www.youtube.com/watch?v=w2xl-GIPK7Q)

### Скачать исходный код

**Скачать**: [snippetbox-12.zip](https://golangs.org/wp-content/uploads/2021/01/snippetbox-12.zip)

