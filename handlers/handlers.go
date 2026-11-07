package handlers

import (
	"fmt"
	"html/template"
	"log"
	"my_home/database"
	"net/http"
	"strconv"
	"time"
)

// Template cache -----------------------------------------------------------------------

var templates = make(map[string](*template.Template))

// Helper to minimize copy-pasting
func parseTemplate(handlerFn string, filenames ...string) (err error) {
	const fnName = "parseTemplate"

	templates[handlerFn], err = template.ParseFiles(filenames...)
	if err != nil {
		return fmt.Errorf("%s: %s: %w", fnName, handlerFn, err)
	}
	return nil
}

func ParseTemplates() error {
	tmplInfos := []struct {
		HandlerFn string
		Filenames []string
	}{
		{HandlerFn: "ProductsHandler", Filenames: []string{
			"templates/layout.html",
			"templates/products.html",
		}},
		{HandlerFn: "HomeHandler", Filenames: []string{
			"templates/layout.html",
			"templates/home.html",
		}},
	}
	for _, info := range tmplInfos {
		if err := parseTemplate(info.HandlerFn, info.Filenames...); err != nil {
			return err
		}
	}
	return nil
}

// Content pages ------------------------------------------------------------------------

// Possible states
// ActiveMenu          : "home" | "products" | "recipes"
// ProductsFilterQuery : "all" | "fresh" | "expiring" | "expired"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	const fnName = "HomeHandler"

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet { // Разрешаем только GET-запросы на главную
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	expiringProducts, err := database.GetProducts("expiring", "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	expiredProducts, err := database.GetProducts("expired", "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	expiringProducts = append(expiringProducts, expiredProducts...)

	// "layout" is a primary template
	err = templates[fnName].ExecuteTemplate(w, "layout", map[string]any{
		"Title":            "Сводка по продуктам",
		"ActiveMenu":       "home",
		"ExpiringProducts": expiringProducts,
	})
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func productsAdd(w http.ResponseWriter, r *http.Request) {
	const fnName = "productsAdd"

	var product database.Product
	var err error

	product.Name = r.FormValue("name")
	countStr := r.FormValue("count")
	expiresAtStr := r.FormValue("expires_at") // Приходит в формате "YYYY-MM-MM" (например, "2026-11-07")

	product.Count, err = strconv.Atoi(countStr)
	if err != nil || product.Count < 1 {
		http.Error(w, "Некорректное количество товара", http.StatusBadRequest)
		return
	}
	product.ExpiresAt, err = time.Parse("2006-01-02", expiresAtStr)
	if err != nil {
		// Will never happen?
		http.Error(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	err = database.AddProduct(product)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	const fnName = "ProductsHandler"

	switch r.Method {
	case http.MethodPost:
		productsAdd(w, r)
		return

	case http.MethodGet:
		filter := r.URL.Query().Get("filter")
		search := r.URL.Query().Get("search")
		products, err := database.GetProducts(filter, search)
		if err != nil {
			log.Printf("%s: %s\n", fnName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = templates[fnName].ExecuteTemplate(w, "layout", map[string]any{
			"Title":               "Список продуктов",
			"ActiveMenu":          "products",
			"ProductsFilterQuery": filter,
			"SearchQuery":         search,
			"Products":            products,
		})
		if err != nil {
			log.Printf("%s: %s\n", fnName, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	default:
		log.Printf("%s unsupporetd http.Method: %v\n", fnName, r.Method)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func ProductsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const fnName = "ProductsDeleteHandler"

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		return
	}

	err = database.DeleteProduct(id)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
