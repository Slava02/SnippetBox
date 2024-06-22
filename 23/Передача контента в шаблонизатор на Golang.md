# Передача контента в шаблонизатор на Golang

---
В данной статье мы разберем принцип отображения динамических данных из [MySQL базы данных](https://golangs.org/mysql-web-site "MySQL Golang") на некоторых HTML страницах нашего **веб-приложения на Golang**.

## Вывод заметок в HTML шаблоне

На данный момент обработчик `showSnippet` извлекает заметку [models.Snippet](https://golangs.org/designing-a-database-model "создаем модель в golang") из **базы данных**, а затем показывает его содержимое в виде обычного **текстового HTTP ответа**. Сейчас мы обновим текущий HTML шаблона, чтобы наши данные отображались на веб-странице следующим образом:

![шаблоны html](https://golangs.org/wp-content/uploads/2020/12/shablony-html-1024x655.png)

Начнем с обработчика `showSnippet` и добавим код для **рендеринга файла шаблона** `show.page.tmpl` (которого мы создадим через минуту). Следующий код должен показаться знакомым по предыдущим урокам.

```go
package main

import (
    "errors"
    "fmt"
    "golangify.com/snippetbox/pkg/models"
    "html/template"
    "net/http"
    "strconv"
)

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

    // Инициализируем срез, содержащий путь к файлу show.page.tmpl
    // Добавив еще базовый шаблон и часть футера, который мы сделали ранее.
    files := []string{
        "./ui/html/show.page.tmpl",
        "./ui/html/base.layout.tmpl",
        "./ui/html/footer.partial.tmpl",
    }

    // Парсинг файлов шаблонов...
    ts, err := template.ParseFiles(files...)
    if err != nil {
        app.serverError(w, err)
        return
    }

    // А затем выполняем их. Обратите внимание на передачу заметки с данными
    // (структура models.Snippet) в качестве последнего параметра.
    err = ts.Execute(w, s)
    if err != nil {
        app.serverError(w, err)
    }
}
```
Затем нам нужно создать файл `show.page.tmpl`, содержащий HTML разметку для страницы отображения заметок. Но перед этим давайте немного поговорим о теории…

В ваших **HTML шаблонах** любые передаваемые в него динамические данные представлены символом точки `.` в начале.

В данном конкретном случае базовым типом точки будет структура `models.Snippet`. Если базовым типом точки является структура, вы можете отобразить (или получить) значение любого поля из неё, добавив точку в начале. Поскольку у структуры `models.Snippet` есть поле `Title`, можно получить заголовок заметки, указав `{{.Title}}` в наших шаблонах.

Создайте новый файл в `ui/html/show.page.tmpl` и добавьте следующую разметку:

```html
{{template "base" .}}

{{define "title"}}Заметка #{{.ID}}{{end}}

{{define "main"}}
<div class='snippet'>
    <div class='metadata'>
        <strong>{{.Title}}</strong>
        <span>#{{.ID}}</span>
    </div>
    <pre><code>{{.Content}}</code></pre>
    <div class='metadata'>
        <time>Создан: {{.Created}}</time>
        <time>Срок: {{.Expires}}</time>
    </div>
</div>
{{end}}
```
Если вы [перезапустите веб-приложение](https://golangify.com/managing-configuration-settings) и перейдете на страницу [http://127.0.0.1:4000/snippet?id=1](http://127.0.0.1:4000/snippet?id=1) в браузере, вы увидите, что нужная заметка была получена из базы данных, передана шаблону, и её содержимое красиво отображено на экране.

Содержимое страницы с заметкой:

![Создание сайта на Golang](https://golangify.com/wp-content/uploads/2021/09/show-snippet-go.jpg)

## Go-структура с данными для передачи в шаблон

Важно понять, что пакет `html/template` позволяет передавать **только один** источник динамических данных при рендеринге шаблона. Однако, в реальном приложении зачастую есть несколько источников данных, которых нужно отобразить на одной странице.

Легкий способ достичь этого — обернуть динамические данные в [новую структуру](https://golangify.com/struct "структура в golang"), которая будет действовать как единая «структура хранения» для ваших данных.

Для этого создадим новый файл `cmd/web/templates.go`, содержащий структуру `templateData`.

```go
package main

import "golangify.com/snippetbox/pkg/models"

// Создаем тип templateData, который будет действовать как хранилище для
// любых динамических данных, которые нужно передать HTML-шаблонам.
// На данный момент он содержит только одно поле, но мы добавим в него другие
// по мере развития нашего приложения.
type templateData struct {
    Snippet *models.Snippet
}
```
Затем обновляем обработчик `showSnippet` для использования данной структуры при отображении наших шаблонов:

```go
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

    // Создаем экземпляр структуры templateData, содержащей данные заметки.
    data := &templateData{Snippet: s}

    files := []string{
        "./ui/html/show.page.tmpl",
        "./ui/html/base.layout.tmpl",
        "./ui/html/footer.partial.tmpl",
    }

    ts, err := template.ParseFiles(files...)
    if err != nil {
        app.serverError(w, err)
        return
    }

    // Передаем структуру templateData в качестве данных для шаблона.
    err = ts.Execute(w, data)
    if err != nil {
        app.serverError(w, err)
    }
}
```
Теперь данные нашей заметки из структуры `models.Snippet` переданы внутри структуры `templateData`. Чтобы получить данные из неё, нужно вызвать соответствующие названия полей следующим образом (файл `ui/html/show.page.tmpl`):

```xhtml
{{template "base" .}}

{{define "title"}}Snippet #{{.Snippet.ID}}{{end}}

{{define "main"}}
<div class='snippet'>
    <div class='metadata'>
        <strong>{{.Snippet.Title}}</strong>
        <span>#{{.Snippet.ID}}</span>
    </div>
    <pre><code>{{.Snippet.Content}}</code></pre>
    <div class='metadata'>
        <time>Created: {{.Snippet.Created}}</time>
        <time>Expires: {{.Snippet.Expires}}</time>
    </div>
</div>
{{end}}
```
Попробуйте перезапустить приложение и обновите страницу [http://127.0.0.1:4000/snippet?id=1](http://127.0.0.1:4000/snippet?id=1). В браузере вы должны увидеть рендеринг той же страницы, что и раньше.

## Защита от XSS-атак в Golang — Экранирование данных

Пакет `html/template` автоматически **экранирует любые данные**, находящиеся между тегами `{{}}`. Такое поведение помогает избежать атак с использованием межсайтовых сценариев (XSS) и поэтому мы используем именно пакет `html/template` вместо более простого пакета `text/template`.

Допустим кто-то взломал наше приложение и вместо обычного текста для заметки, он добавил туда вредоносный JavaScript код, что произойдет далее?

```xhtml
<span>{{"<script>alert('xss attack')</script>"}}</span>
```
Пакет `html\template` обработает данный контент и выведет на экран пользователя следующее:

```xhtml
<span>&lt;script&gt;alert(&#39;xss attack&#39;)&lt;/script&gt;</span>
```
Пакет `html/template` также достаточно умен, чтобы выполнять экранирование в зависимости от обстоятельств. Он будет использовать соответствующие экранированные последовательности в зависимости от того, отображаются ли данные в части страницы, содержащей HTML, CSS или Javascript.

### Шаблоны с вложенными шаблонами

Очень важно отметить, что при вызове одного шаблона из другого, точка `.` должна быть явно передана в вызываемый шаблон. Это делается через добавление данной точки в конец каждого вызова `{{template}}` или `{{block}}`.

Например:

```go
{{template "base" .}}
{{template "main" .}}
{{template "footer" .}}

{{block "sidebar" .}}
Любой контент
{{end}}
```
Мы советуем выработать привычку использования точки всякий раз, когда вы вызываете шаблон с помощью `{{template}}` или `{{block}}`.

### Вызов методов в шаблоне

Если у переданного в шаблон объекта есть методы, вы можете вызывать их (при условии, что они экспортируются и возвращают только одно значение — или одно значение и [ошибку](https://golangify.com/errors "Ошибки в golang")).

Например, если поле `.Snippet.Created` является типом `time.Time`, вы можете **отобразить в шаблоне** название дня недели, вызвав его метод [Weekday()](https://golang.org/pkg/time/#Time.Unix) следующим образом:

```xhtml
<span>{{.Snippet.Created.Weekday}}</span>
```
Также, методу можно передать параметры. К примеру, вы можете использовать метод [AddDate()](https://golang.org/pkg/time/#Time.AddDate) чтобы увеличить текущее время на шесть месяцев:

```xhtml
<span>{{.Snippet.Created.AddDate 0 6 0}}</span>
```
Обратите внимание, что данный синтаксис отличается от синтаксиса вызова [функций в Go](https://golangify.com/func "создание функций в golang") — параметры не заключены в круглые скобки и разделены через пробел, а не запятой.

### HTML комментарии в шаблоне

Пакет `html/template` всегда удаляет любые HTML комментарии, которые вы оставляете в шаблонах, включая любые [условные комментарии](https://ru.wikipedia.org/wiki/%D0%A3%D1%81%D0%BB%D0%BE%D0%B2%D0%BD%D1%8B%D0%B9_%D0%BA%D0%BE%D0%BC%D0%BC%D0%B5%D0%BD%D1%82%D0%B0%D1%80%D0%B8%D0%B9) которые часто применяются фронтенд разработчиками.

Это необходимо во избежания XSS-атак при **отображении динамического контента**. Разрешение условных комментариев означало бы, что Go не всегда может предугадать, как браузер интерпретирует разметку на странице. Следовательно, не факт, что он сможет избежать все опасные сценарии. Для решения этой проблемы Go просто **удаляет все HTML комментарии**.

### Скачать исходный код

**Скачать**: [snippetbox-23](https://golangify.com/wp-content/uploads/2021/09/snippetbox-23.zip)
