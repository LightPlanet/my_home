package main

import (
	"log"
	"my_home/cli"
	"my_home/database"
	"my_home/handlers"
	"net/http"
)

func main() {
	// Defaults
	dbFile := "database.db"
	serverAddr := ":8080"

	userWantsContinue := cli.Parse(&dbFile, &serverAddr)
	if !userWantsContinue {
		return
	}
	//fmt.Println("dbFile", dbFile)
	//fmt.Println("serverAddr", serverAddr)
	//return

	// Prepare DB
	if err := database.Open(dbFile); err != nil {
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

	log.Printf("Server starting at %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Println(err)
		return
	}
}
