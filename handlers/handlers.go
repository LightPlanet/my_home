// HTTP handlers

// This file contains:
// [1] Template cache
// [2] Content pages
// [2.1] Home page
// [2.2] Products page
// [2.3] Recipes page

package handlers

import (
	"fmt"
	"html/template"
	"log"
	"my_home/database"
	"net/http"
	"strconv"
	"strings"
)

// [1] Template cache -------------------------------------------------------------------

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
		{HandlerFn: "ProductHandler", Filenames: []string{
			"templates/layout.html",
			"templates/products.html",
		}},
		{HandlerFn: "HomeHandler", Filenames: []string{
			"templates/layout.html",
			"templates/home.html",
		}},
		{HandlerFn: "RecipeHandler", Filenames: []string{
			"templates/layout.html",
			"templates/recipes.html",
		}},
	}

	for _, info := range tmplInfos {
		if err := parseTemplate(info.HandlerFn, info.Filenames...); err != nil {
			return err
		}
	}
	return nil
}

// [2] Content pages --------------------------------------------------------------------

// Possible states
// ActiveMenu          : "home" | "products" | "recipes"
// ProductsFilterQuery : "all" | "fresh" | "expiring" | "expired"

// [2.1] Home page
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

// [2.2] Products page

func productAdd(w http.ResponseWriter, r *http.Request) {
	const fnName = "productsAdd"

	var product database.Product
	var err error

	// Using TrimSpace before any DB insertion to avoid spaces, which
	// may lead to string comparing errors
	product.Name = strings.TrimSpace(r.FormValue("name"))
	countStr := r.FormValue("count")
	expiresAtStr := r.FormValue("expires_at") // Always "YYYY-MM-MM" (browser standard?)

	product.Count, err = strconv.Atoi(countStr)
	if err != nil || product.Count < 1 {
		http.Error(w, "Некорректное количество товара", http.StatusBadRequest)
		return
	}

	// Passing Product struct without .ExpiresAt - not an error
	// No need to parse string -> time and then format time -> string just
	// to make an SQL query
	err = database.AddProduct(product, expiresAtStr)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func productShow(w http.ResponseWriter, r *http.Request) {
	const fnName = "productsShow"

	filter := r.URL.Query().Get("filter")
	search := r.URL.Query().Get("search")
	products, err := database.GetProducts(filter, search)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templates["ProductHandler"].ExecuteTemplate(w, "layout", map[string]any{
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
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	const fnName = "ProductHandler"

	switch r.Method {
	case http.MethodPost:
		productAdd(w, r)
	case http.MethodGet:
		productShow(w, r)
	default:
		log.Printf("%s unsupporetd http.Method: %v\n", fnName, r.Method)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func ProductDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

// [2.3] Recipes

func recipeAdd(w http.ResponseWriter, r *http.Request) {
	const fnName = "recipesAdd"

	if err := r.ParseForm(); err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}
	rName := strings.TrimSpace(r.FormValue("name"))
	rDesc := strings.TrimSpace(r.FormValue("description"))
	ingrNames := r.Form["ingr_names[]"]
	ingrCounts := r.Form["ingr_counts[]"]

	if rName == "" || len(ingrNames) == 0 {
		log.Printf("%s: %s\n", fnName, "recipeName == '' || len(ingrNames) == 0")
		http.Error(w, "Название рецепта и ингредиенты обязательны!", http.StatusBadRequest)
		return
	}

	err := database.AddRecipe(rName, rDesc, ingrNames, ingrCounts)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func recipeShow(w http.ResponseWriter, r *http.Request) {
	const fnName = "recipesShow"

	search := r.URL.Query().Get("search")
	recipes, err := database.GetRecipes(search)
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templates["RecipeHandler"].ExecuteTemplate(w, "layout", map[string]any{
		"Title":       "Рецепты",
		"ActiveMenu":  "recipes",
		"SearchQuery": search,
		"Recipes":     recipes,
	})
	if err != nil {
		log.Printf("%s: %s\n", fnName, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func RecipeHandler(w http.ResponseWriter, r *http.Request) {
	const fnName = "RecipesHandler"

	switch r.Method {
	case http.MethodPost:
		recipeAdd(w, r)
	case http.MethodGet:
		recipeShow(w, r)
	default:
		log.Printf("%s: %s\n", fnName, "Invalid HTTP method")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
