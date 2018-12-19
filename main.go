package main

import (
	"log"
	"net/http"
	"path"
)

func main() {
	// Static file hosting
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join("views", "index.html"))
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}
