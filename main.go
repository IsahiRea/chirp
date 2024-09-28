package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/assets/logo", http.FileServer(http.Dir("logo.png")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
