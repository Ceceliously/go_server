package main

import (
	"net/http"
	"server/service"
	
)

func main() {



	mux := http.NewServeMux()

	srv, err := service.NewDB()
	if err != nil {
		fmt.Println(err)
	}

	mux.HandleFunc("/register", srv.Register)
	mux.HandleFunc("/read", srv.Read)

	http.ListenAndServe(":8000", mux)
}