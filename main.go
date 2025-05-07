package main

import (
	"net/http"
	"server/service"
)

func main() {
	mux := http.NewServeMux()

	srv := service.New()

	mux.HandleFunc("/register", srv.Register)
	mux.HandleFunc("/read", srv.Read)

	http.ListenAndServe(":8000", mux)
}