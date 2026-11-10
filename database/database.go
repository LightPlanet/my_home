// Database handlers

// This file contains:
// [1] 	 Helpers
// [1.1] Product
// [1.2] Recipe
//
// [2] 	 Methods
// [2.1] Open
// [2.2] Close
// [2.3] GetProducts
// [2.4] AddProduct
// [2.5] DeleteProduct
// [2.6] GetRecipes
// [2.7] AddRecipe

package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// [1] Helpers --------------------------------------------------------------------------

// [1.1] Product

type Product struct {
	ID        int
	Name      string
	Count     int
	ExpiresAt time.Time
	Status    string // Display only: "fresh"|"expiring"|"expired"
}

// Time until expiration date, when "expiring" status will be shown
const expiringWarningDuration = time.Duration(3 * 24 * time.Hour) // 3 days

// SQLite does not have time format
// Using Format("2006-01-02") workaround - with such formatting
// lexicographical order of 'date as time' and 'date as text' is the same
const sqlDateFormat = "2006-01-02"

// NOTE: This function is called after db-select (see GetProducts())
func (this *Product) CalcStatus() {
	remaining := time.Until(this.ExpiresAt)
	switch {
	case remaining <= 0:
		this.Status = "expired"
	case remaining <= expiringWarningDuration:
		this.Status = "expiring"
	default:
		this.Status = "fresh"
	}
}

// [1.2] Recipe

type RecipeIngredient struct {
	Name string
	Need int // By recipe
	Have int // Available
}

type Recipe struct {
	ID          int
	Name        string
	Description string
	Ingredients []RecipeIngredient
}

// [2] Database methods -----------------------------------------------------------------

var db *sql.DB

// [2.1] Open
func Open(file string) (err error) {
	const fnName = "Open"

	db, err = sql.Open("sqlite", file)
	if err != nil {
		return fmt.Errorf("%s: Open: %w", fnName, err)
	}

	// We need to delete ingrs automatically, if the recipe, which is using them
	// is deleted. We use "ON DELETE CASCADE" feature, which requires foreign keys support
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("%s: EnableForeignKeys: %w", fnName, err)
	}

	queries := []struct {
		Table string
		Query string
	}{
		{Table: "Products", Query: `CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			count INTEGER NOT NULL,
			expires_at TEXT NOT NULL
		);`},
		{Table: "Recipes", Query: `CREATE TABLE IF NOT EXISTS recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT
		);`},
		{Table: "Recipeingrs", Query: `CREATE TABLE IF NOT EXISTS recipe_ingredients (
			recipe_id INTEGER NOT NULL,
			ingr_name TEXT NOT NULL,
			required_count INTEGER NOT NULL CHECK (required_count > 0),
			PRIMARY KEY (recipe_id, ingr_name),
			FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE
		);`},
	}
	for _, q := range queries {
		if _, err := db.Exec(q.Query); err != nil {
			return fmt.Errorf("%s: CreateTable '%s': %w", fnName, q.Table, err)
		}
	}
	return nil
}

// [2.2] Close
func Close() error {
	return db.Close()
}

// [2.3] GetProducts

// Product helper. HTML [filter, search] -> SQL [query, queryArgs]
func transformFilterURLToSQL(urlFilter, urlSearch string) (string, []any) {
	nowTime := time.Now()
	now := nowTime.Format(sqlDateFormat)
	expiringWarningAt := nowTime.Add(expiringWarningDuration).Format(sqlDateFormat)

	var filter string
	var args []any

	switch urlFilter {
	case "fresh":
		filter = "WHERE expires_at > ?"
		args = []any{expiringWarningAt}
	case "expiring":
		filter = "WHERE expires_at >= ? AND expires_at <= ?"
		args = []any{now, expiringWarningAt}
	case "expired":
		filter = "WHERE expires_at < ?"
		args = []any{now}
	default: // "all"
		filter = ""
		args = nil
	}

	if urlSearch != "" {
		if filter == "" { // is empty - needs "WHERE"
			filter = "WHERE LOWER(name) LIKE ?"
		} else {
			filter += " AND LOWER(name) LIKE ?"
		}
		args = append(args, "%"+strings.ToLower(urlSearch)+"%")
	}

	return filter, args
}

func GetProducts(urlFilter, urlSearch string) ([]Product, error) {
	const fnName = "GetProducts"

	// SELECT matching
	filter, args := transformFilterURLToSQL(urlFilter, urlSearch)
	selectProducts :=
		`SELECT id, name, count, expires_at FROM products ` +
			filter + ` ORDER BY expires_at`
	rows, err := db.Query(selectProducts, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w\n", fnName, err)
	}
	defer rows.Close()

	// Parse products
	var products []Product
	for rows.Next() {
		var p Product
		var expiresAtStr string

		if err := rows.Scan(&p.ID, &p.Name, &p.Count, &expiresAtStr); err != nil {
			return nil, fmt.Errorf("%s: %w\n", fnName, err)
		}
		if p.ExpiresAt, err = time.Parse(sqlDateFormat, expiresAtStr); err != nil {
			return nil, fmt.Errorf("%s: %w\n", fnName, err)
		}
		p.CalcStatus()
		products = append(products, p)

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("%s: %w\n", fnName, err)
		}
	}
	return products, nil
}

// [2.4] AddProduct
func AddProduct(p Product) error {
	const fnName = "AddProduct"

	insert := `INSERT INTO products (name, count, expires_at) VALUES (?, ?, ?);`
	_, err := db.Exec(insert, p.Name, p.Count, p.ExpiresAt.Format(sqlDateFormat))
	if err != nil {
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}

// [2.5] DeleteProduct
func DeleteProduct(id int) error {
	const fnName = "DeleteProduct"

	_, err := db.Exec(`DELETE FROM products WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}

// [2.6] GetRecipes

// Recipe helper. HTML [search] -> SQL [query, queryArgs]
func transformSearchURLToSQL(urlSearch string) (string, []any) {
	// "FROM recipes r" is same as "using r = recipes"
	query := `
		SELECT r.id, r.name, r.description, ri.ingr_name, ri.required_count
		FROM recipes r
		JOIN recipe_ingredients ri ON r.id = ri.recipe_id
	`
	var args []any

	if urlSearch != "" {
		query += " WHERE r.name LIKE ?"
		args = append(args, "%"+urlSearch+"%")
	}

	query += " ORDER BY r.id, ri.ingr_name;"
	return query, args
}

func GetRecipes(urlSearch string) ([]Recipe, error) {
	const fnName = "GetRecipesWithIngredients"

	// 1. SELECT all available products
	select_ := `SELECT LOWER(name), SUM(count) FROM products GROUP BY LOWER(name);`
	productRows, err := db.Query(select_)
	if err != nil {
		return nil, fmt.Errorf("%s: products query: %w", fnName, err)
	}
	defer productRows.Close()

	myProducts := make(map[string]int)
	for productRows.Next() {
		var name string
		var totalCount int
		if err := productRows.Scan(&name, &totalCount); err != nil {
			return nil, fmt.Errorf("%s: productRows.Scan: %w", fnName, err)
		}
		myProducts[name] = totalCount
	}
	if err = productRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: productRows: %w", fnName, err)
	}

	// 2. SELECT all searched recipes
	select_, args := transformSearchURLToSQL(urlSearch)
	rows, err := db.Query(select_, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: recipes query: %w", fnName, err)
	}
	defer rows.Close()

	// 3. Group and calculate recipe missing products

	// The order must be calculated explicitly, because page reloading
	// may lead to item reordering, which is bad for UX
	var recipeOrder []int
	recipeMap := make(map[int]*Recipe)
	for rows.Next() {
		var rID, needCount int
		var rName, rDescr, ingrName string

		if err := rows.Scan(&rID, &rName, &rDescr, &ingrName, &needCount); err != nil {
			return nil, fmt.Errorf("%s: recipes scan: %w", fnName, err)
		}

		if _, exists := recipeMap[rID]; !exists {
			recipeMap[rID] = &Recipe{
				ID:          rID,
				Name:        rName,
				Description: rDescr,
				Ingredients: []RecipeIngredient{},
			}
			recipeOrder = append(recipeOrder, rID)
		}

		haveCount := myProducts[strings.ToLower(ingrName)]
		recipeMap[rID].Ingredients = append(recipeMap[rID].Ingredients, RecipeIngredient{
			Name: ingrName,
			Need: needCount,
			Have: haveCount,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fnName, err)
	}

	result := make([]Recipe, 0, len(recipeOrder))
	for _, id := range recipeOrder {
		result = append(result, *recipeMap[id])
	}
	return result, nil
}

// [2.7] AddRecipe
func AddRecipe(rName, rDescription string, ingrNames, ingrCounts []string) error {
	const fnName = "AddRecipe"

	// Open database transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %s\n", fnName, err)
	}
	// Allows safe fallback in case of panic (if no Commit() then Rollback())
	defer tx.Rollback()

	insert := `INSERT INTO recipes (name, description) VALUES (?, ?)`
	res, err := tx.Exec(insert, strings.TrimSpace(rName), strings.TrimSpace(rDescription))
	if err != nil {
		return fmt.Errorf("%s: %s\n", fnName, err)
	}

	recipeID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s: %s\n", fnName, err)
	}

	insert = `INSERT INTO recipe_ingredients
		(recipe_id, ingr_name, required_count) VALUES (?, ?, ?)`
	for i, ingName := range ingrNames {
		ingName = strings.TrimSpace(ingName)
		if ingName == "" {
			continue
		}

		count, err := strconv.Atoi(ingrCounts[i])
		if err != nil || count <= 0 {
			return fmt.Errorf("%s: %s\n", fnName, err)
		}

		_, err = tx.Exec(insert, recipeID, ingName, count)
		if err != nil {
			return fmt.Errorf("%s: %s\n", fnName, err)
		}
	}

	// Fix database changes
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %s\n", fnName, err)
	}
	return nil
}
