# Централизованная обработка ошибок в Golang
---
## Создание методов-помощников для обработки ошибок

Давайте немного настроим наше приложение, переместив часть кода **обработки ошибок** во вспомогательные методы. Это поможет разделить проблемы и избавиться от повторения кода по мере развития программы.

Добавьте новый файл `helpers.go` в папку `cmd/web`

И добавьте в него следующий код:

```go
package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Помощник serverError записывает сообщение об ошибке в errorLog и
// затем отправляет пользователю ответ 500 "Внутренняя ошибка сервера".
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Помощник clientError отправляет определенный код состояния и соответствующее описание
// пользователю. Мы будем использовать это в следующий уроках, чтобы отправлять ответы вроде 400 "Bad
// Request", когда есть проблема с пользовательским запросом.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// Мы также реализуем помощник notFound. Это просто
// удобная оболочка вокруг clientError, которая отправляет пользователю ответ "404 Страница не найдена".
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
```
Здесь используется совсем немного кода, но вместе с тем вводится несколько новых элементов, которые стоит обсудить:

-   В помощнике `serverError()` мы используем функцию [debug.Stack()](https://golang.org/pkg/runtime/debug/#Stack), чтобы получить трассировку стека для текущей горутины и добавить ее в логгер. Возможность видеть полный путь к приложению через трассировку стека может быть полезна при отладке возникнувших ошибок;
-   В помощнике `clientError()` мы используем функцию [http.StatusText()](https://golang.org/pkg/net/http/#StatusText) для автоматической генерации понятного человеку текстового представления для [кода состояния HTTP](https://golangs.org/customizing-http-headers#http-status-code). К примеру, `http.StatusText(400)` вернет [строку](https://golangs.org/string "строки в golang") `"Bad Request"`;
-   Мы начали использовать специальные [константы](https://golangs.org/osnovy-golang#const-var "константы golang") из пакета `net/http` для кодов состояния HTTP вместо целых чисел. В помощнике `serverError()` мы использовали константу `http.StatusInternalServerError` вместо `500`, а в помощнике `notFound()` — константу `http.StatusNotFound` вместо записи `404`.

Использование констант для кодов состояния HTTP является приятной особенностью, которая помогает сделать код понятным и самодокументируемым — особенно при работе с редко используемыми кодами состояния HTTP. Вы можете найти **полный список констант кодов состояния** [по ссылке](https://golang.org/pkg/net/http/#pkg-constants).

Как только все это будет сделано, вернитесь к файлу `handlers.go` и обновите его, для применения **функций-помощников** из `helpers.go`:

```go
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w) // Использование помощника notFound()
		return
	}

	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err) // Использование помощника serverError()
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		app.serverError(w, err) // Использование помощника serverError()
	}
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w) // Использование помощника notFound()
		return
	}

	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed) // Используем помощник clientError()
		return
	}

	w.Write([]byte("Создание новой заметки..."))
}
```

После обновления перезапустите веб-приложение и сделайте HTTP-запрос к [http://127.0.0.1:4000](http://127.0.0.1:4000/) в браузере.

Это все должно привести к возникновению нашей [специально оставленной ошибки](https://golangs.org/dependency-injection#add-error) из прошлого урока, и вы должны увидеть соответствующее **сообщение об ошибке** и **трассировку стека** в терминале:

```shell
$ go run ./cmd/web

INFO	2021/01/25 18:04:36 Запуск сервера на :4000

ERROR	2021/01/25 18:04:39 helpers.go:13: open ./ui/html/home.page.tmpl: no such file or directory
goroutine 6 [running]:
runtime/debug.Stack(0xc00007f200, 0xc00001e2c0, 0x38)
	/snap/go/6745/src/runtime/debug/stack.go:24 +0x9f
main.(*application).serverError(0xc000012c70, 0x827e60, 0xc00014e0e0, 0x8228e0, 0xc00007f200)
	/home/pc/code/snippetbox/cmd/web/helpers.go:12 +0x65
main.(*application).home(0xc000012c70, 0x827e60, 0xc00014e0e0, 0xc00017e000)
	/home/pc/code/snippetbox/cmd/web/handlers.go:24 +0x21c
net/http.HandlerFunc.ServeHTTP(0xc000012c80, 0x827e60, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2042 +0x44
net/http.(*ServeMux).ServeHTTP(0xc000072340, 0x827e60, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2417 +0x1ad
net/http.serverHandler.ServeHTTP(0xc00014e000, 0x827e60, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2843 +0xa3
net/http.(*conn).serve(0xc000117040, 0x8287e0, 0xc000072400)
	/snap/go/6745/src/net/http/server.go:1925 +0x8ad
created by net/http.(*Server).Serve
	/snap/go/6745/src/net/http/server.go:2969 +0x36c
```
Если вы внимательно посмотрите, то заметите небольшую проблему: название файла и номер строки, отображаемые в логе, как `helpers.go:13` — это то место, откуда записывается сообщение об ошибке.

> Но, зачем нам знать где находится помощник? Мы хотим знать где именно в работе программы возникла ошибка, а не место нахождения помощника который данную ошибку выводит!

В своих логах, мы хотим увидеть только название проблематичного файла и номер строки из трассировки стека, это даст более четкое представление о том, где возникла ошибка.

Это можно сделать, изменив немного код помощника `serverError()` чтобы он использовал логгер [Output()](https://golang.org/pkg/log/#Logger.Output) и установив глубину вызова на 2 (она и так по умолчанию на 2). Снова откройте файл `helpers.go` и обновите его следующим образом:

```go
package main

...

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

...
```
Перезапустите веб-приложение из терминала. При повторной попытке перехода на [http://127.0.0.1:4000](http://127.0.0.1:4000/) вы обнаружите, что нам сообщается соответствующее название файла и номер строки `(handlers.go:24)`:

```shell
$ go run ./cmd/web
INFO	2021/01/25 18:17:06 Запуск сервера на :4000

ERROR	2021/01/25 18:21:01 handlers.go:24: open ./ui/html/home.page.tmpl: no such file or directory
goroutine 6 [running]:
runtime/debug.Stack(0xc00007f200, 0xc00001e2c0, 0x38)
	/snap/go/6745/src/runtime/debug/stack.go:24 +0x9f
main.(*application).serverError(0xc000012c70, 0x827e40, 0xc00014e0e0, 0x822900, 0xc00007f200)
	/home/pc/code/snippetbox/cmd/web/helpers.go:12 +0x65
main.(*application).home(0xc000012c70, 0x827e40, 0xc00014e0e0, 0xc00017e000)
	/home/pc/code/snippetbox/cmd/web/handlers.go:24 +0x21c
net/http.HandlerFunc.ServeHTTP(0xc000012c80, 0x827e40, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2042 +0x44
net/http.(*ServeMux).ServeHTTP(0xc000072340, 0x827e40, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2417 +0x1ad
net/http.serverHandler.ServeHTTP(0xc00014e000, 0x827e40, 0xc00014e0e0, 0xc00017e000)
	/snap/go/6745/src/net/http/server.go:2843 +0xa3
net/http.(*conn).serve(0xc000116fa0, 0x8287c0, 0xc000072400)
	/snap/go/6745/src/net/http/server.go:1925 +0x8ad
created by net/http.(*Server).Serve
	/snap/go/6745/src/net/http/server.go:2969 +0x36c
```
### Исправляем ошибку из прошлого урока

На данный момент нам больше не нужна **искусственная ошибка**. Избавиться от нее можно следующий образом:

```shell
$ cd $HOME/code/snippetbox
$ mv ui/html/home.page.tmpl ui/html/home.page.tmpl
```
Если через терминал у вас не получается, то просто зайдите в папку `html`, найдите там файл `home.page.bak` и замените расширение `.bak` на `.tmpl` и всё!

### Скачать готовый код

**Скачать**: [snippetbox-13.zip](https://golangs.org/wp-content/uploads/2021/01/snippetbox-13.zip)

