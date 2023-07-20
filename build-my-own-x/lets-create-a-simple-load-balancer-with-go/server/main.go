package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Args[1]
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(port))
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
