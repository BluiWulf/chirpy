package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{}

	server.Handler = mux
	server.Addr = ":8080"

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error starting server: %v", err)
	}
}
