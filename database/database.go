package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Database helpers ---------------------------------------------------------------------

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

// Database methods ---------------------------------------------------------------------

var db *sql.DB

func Open(file string) (err error) {
	const fnName = "Open"

	db, err = sql.Open("sqlite", file)
	if err != nil {
		return fmt.Errorf("%s: Open: %w", fnName, err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		count INTEGER NOT NULL,
		expires_at TEXT NOT NULL
	)`
	if _, err := db.Exec(createTable); err != nil {
		return fmt.Errorf("%s: CreateTable: %w", fnName, err)
	}
	return nil
}

func Close() error {
	return db.Close()
}

// HTML ProductFilterQuery -> [sqlQuery, sqlQueryArgs]
func transformFilterURLToSQL(urlFilter string) (string, []any) {
	nowTime := time.Now()
	now := nowTime.Format(sqlDateFormat)
	expiringWarningAt := nowTime.Add(expiringWarningDuration).Format(sqlDateFormat)

	switch urlFilter {
	case "fresh":
		return "WHERE expires_at > ?", []any{expiringWarningAt}
	case "expiring":
		return "WHERE expires_at >= ? AND expires_at <= ?", []any{now, expiringWarningAt}
	case "expired":
		return "WHERE expires_at < ?", []any{now}
	default: // "all"
		return "", nil
	}
}

func GetProducts(urlFilter, urlSearch string) ([]Product, error) {
	const fnName = "GetProducts"

	// SELECT matching
	filter, args := transformFilterURLToSQL(urlFilter)
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

func AddProduct(p Product) error {
	const fnName = "AddProduct"

	insert := `INSERT INTO products (name, count, expires_at) VALUES (?, ?, ?);`
	_, err := db.Exec(insert, p.Name, p.Count, p.ExpiresAt.Format(sqlDateFormat))
	if err != nil {
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}

func DeleteProduct(id int) error {
	const fnName = "DeleteProduct"

	_, err := db.Exec(`DELETE FROM products WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}
