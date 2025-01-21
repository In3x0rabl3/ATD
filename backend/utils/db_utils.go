package utils

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func InitializeDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS sensitive_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS uploaded_images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL,
		content BLOB NOT NULL
	);`

	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully.")
	return db
}

func PopulateDB(db *sql.DB) {
	sampleData := map[string]string{
		"admin":         "P9#mNz4L2RxYv@Fk8T&dWq3V",
		"servadm":       "8Fv@Lk93zWq&XnP!d2R4YmTc",
		"database":      "5g&fxvTU3d5Gts3#56",
		"api_key":       "sk-TN8DF5GZ39R0sMkpJhQUV67wX4zyabCefgHiKlmnopq32TuvWxYZA7d12E",
		"root_password": "R00t_s3cr3t_p4$$wrd",
		"flag":          "ME{Da74_L3aK4ge}",
	}

	for key, value := range sampleData {
		_, err := db.Exec("INSERT OR IGNORE INTO sensitive_data (key, value) VALUES (?, ?)", key, value)
		if err != nil {
			log.Printf("Failed to insert data (%s: %s): %v", key, value, err)
		}
	}

	log.Println("Sample data populated successfully.")
}

func ExecuteQuery(db *sql.DB, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	if len(query) >= 6 && strings.ToUpper(query[:6]) == "SELECT" {
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Error executing SELECT query: %v", err)
			return "", fmt.Errorf("error executing SELECT query: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			log.Printf("Error fetching column names: %v", err)
			return "", fmt.Errorf("error fetching column names: %v", err)
		}

		results := []map[string]interface{}{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				log.Printf("Error scanning row: %v", err)
				return "", fmt.Errorf("error scanning row: %v", err)
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				row[col] = fmt.Sprintf("%v", values[i])
			}

			results = append(results, row)
		}

		log.Printf("Query results: %+v", results)
		return fmt.Sprintf("Query executed successfully.\nResults:\n%+v", results), nil
	}

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return "", fmt.Errorf("error executing query: %v", err)
	}

	log.Printf("Query executed successfully: %s", query)
	return "Query executed successfully.", nil
}
