package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"my_home/database"
	"net/http"
)

func ProductListHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	const fnName = "ProductListHandler"

	filter := r.URL.Query().Get("ProductsFilterQ")
	products, err := database.GetProducts(db, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s: %s", fnName, err)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("%s: %s", fnName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{
		"Title":           "Продукты",
		"Active":          "products",
		"ProductsFilterQ": filter,
		"Products":        products,
	})
	if err != nil {
		log.Printf("%s: %s", fnName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
