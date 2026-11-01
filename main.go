package main

import (
	"log"
	"my_home/database"
	"my_home/handlers"
	"net/http"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ProductListHandler(db, w, r)
	})
	log.Println("Server starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
		return
	}
}
