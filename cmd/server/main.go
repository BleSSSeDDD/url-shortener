package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Тут будет сокращение ссылок"))
		})
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World"))
		})
		fmt.Println("Сервер запущен")
		http.ListenAndServe("localhost:8080", nil)
	}()

	<-stop
	fmt.Println("Программа завершена!")
}
