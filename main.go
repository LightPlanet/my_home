package main

import (
	"fmt"
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
	fmt.Println("dbFile", dbFile)
	fmt.Println("serverAddr", serverAddr)
	return

	db, err := database.Open(dbFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ProductListHandler(db, w, r)
	})
	log.Printf("Server starting at %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Println(err)
		return
	}
}
