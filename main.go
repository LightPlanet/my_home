package main

import (
	"flag"
	"log"
	"my_home/database"
	"my_home/handlers"
	"net/http"
)

func main() {
	dbFile := flag.String("db", "database.db", "Путь к файлу базы данных")
	serverAddr := flag.String("addr", ":8080", "Адрес сервера")
	flag.Parse()

	// Prepare DB
	if err := database.Open(*dbFile); err != nil {
		log.Println(err)
		return
	}
	defer database.Close()

	// Prepare files & callbacks
	if err := handlers.ParseTemplates(); err != nil {
		log.Println(err)
		return
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/products", handlers.ProductHandler)
	http.HandleFunc("/products/{id}/delete", handlers.ProductDeleteHandler)
	http.HandleFunc("/recipes", handlers.RecipeHandler)

	log.Printf("Server starting at %s", *serverAddr)
	if err := http.ListenAndServe(*serverAddr, nil); err != nil {
		log.Println(err)
		return
	}
}
