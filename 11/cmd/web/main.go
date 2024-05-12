package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":4000", "Сетевой адрес веб-сервера")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet", showSnippet)
	mux.HandleFunc("/snippet/create", createSnippet)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Инициализируем новую структуру http.Server. Мы устанавливаем поля Addr и Handler, так
	// что сервер использует тот же сетевой адрес и маршруты, что и раньше, и назначаем
	// поле ErrorLog, чтобы сервер использовал наш логгер
	// при возникновении проблем.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	// Вызываем метод ListenAndServe() от нашей новой структуры http.Server
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
