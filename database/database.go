package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Product struct {
	ID        int
	Name      string
	Count     int
	ExpiresAt time.Time
	Status    string // Display only: "fresh"|"expiring"|"expired"
}

// Time until expiration date, when "expiring" status will be shown
const expiringWarningDuration = time.Duration(3 * 24 * time.Hour) // 3 days

// NOTE: This function is called after db-select
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

func Open() (*sql.DB, error) {
	return sql.Open("sqlite", "database.db")
}

// Helper, converts HTML ProductFilterQ to [sqlQuery, sqlQueryArgs]
func transformURLQueryToSQL(urlQuery string) (string, []any) {
	// SQLite does not have time format
	// Using Format("2006-01-02") workaround (text and date ordering are the same)
	const sqlDateFormat = "2006-01-02"

	nowTime := time.Now()
	now := nowTime.Format(sqlDateFormat)
	expiringWarningAt := nowTime.Add(expiringWarningDuration).Format(sqlDateFormat)

	switch urlQuery {
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

func GetProducts(db *sql.DB, productsFilter string) ([]Product, error) {
	const fnName = "GetProducts"

	createTable := `CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name STRING NOT NULL
		count INTEGER NOT NULL
		expires_at STRING NOT NULL
	)`
	if _, err := db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("%s: %w", fnName, err)
	}

	// SELECT
	query, args := transformURLQueryToSQL(productsFilter)
	selectProducts :=
		`SELECT id, name, count, expires_at FROM products ` +
			query + ` ORDER BY expires_at`
	rows, err := db.Query(selectProducts, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse products
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Count, &p.ExpiresAt); err != nil {
			return nil, err
		}
		p.CalcStatus()
		products = append(products, p)

		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return products, nil
}

func AddProduct(db *sql.DB, p Product) error {

	return nil
}
