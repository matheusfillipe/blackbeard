// Database

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	blb "github.com/matheusfillipe/blackbeard/blb"
)

var _ = blb.Breakpoint

type CacheDb struct {
	profile  string
	filename string
}

var cacheDb CacheDb

type CacheValue struct {
	Provider    string
	Query       string
	Description string
}

func createDb() *sql.DB{
	createCacheDir(cacheDb.profile)
	sqliteDatabase, err := sql.Open("sqlite3", cacheDb.filename)
  if err != nil {
    log.Fatal(err)
  }
	return sqliteDatabase
}

// Check if the cache tables don't exist first then creates them if needed
func createSearchTable(db *sql.DB) bool {
	createSerchTable := "CREATE TABLE IF NOT EXISTS search (id INTEGER PRIMARY KEY, provider TEXT, query TEXT)"

	var createQueries = []string{createSerchTable}
	for _, query := range createQueries {
		statement, err := db.Prepare(query) // Prepare SQL Statement
		if err != nil {
			log.Println(err.Error())
			println("Unable to create cache table!")
			return false
		}
		_, err = statement.Exec()
		if err != nil {
			log.Println(err.Error())
			println("Unable to create cache table!")
			return false
		}
	}
	return true
}


// Inserts into search cache table
// Avoids duplication
func insertSearch(db *sql.DB, provider string, query string) {
	// Check if entry exists
	showSQL := `SELECT * FROM search WHERE provider=? and query=?`
	statement, err := db.Prepare(showSQL)
	if err != nil {
		fmt.Println(err.Error())
		return

	}
	row, err := statement.Query(provider, query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for row.Next() {
		return
	}

	// Insert new
	insertSQL := `INSERT INTO search(provider, query) VALUES (?, ?)`
	statement, err = db.Prepare(insertSQL)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = statement.Exec(provider, query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

// Ensures the existence the cache directory and database file
func createCacheDir(profile string) {
	cacheDb.profile = profile
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	cacheDir := filepath.FromSlash(usr.HomeDir + "/.cache/blackbeard/" + profile + "/")
	os.MkdirAll(cacheDir, 0755)
	cacheFile := filepath.FromSlash(cacheDir + "/cache.db")

	// Create db file if doesn't exist, check access
	_, err = os.Stat(cacheFile)
	if os.IsNotExist(err) {
		file, err := os.Create(cacheFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
	}
	cacheDb.filename = cacheFile
}

func getSearchCache(provider string) ([]CacheValue, bool) {
	db := createDb()
	defer db.Close()
	createSearchTable(db)

	showSQL := `SELECT query FROM search WHERE provider=?`
	statement, err := db.Prepare(showSQL)
	if err != nil {
		fmt.Println(err.Error())
		return []CacheValue{}, false
	}
	row, err := statement.Query(provider)
	if err != nil {
		fmt.Println(err.Error())
		return []CacheValue{}, false
	}
	res := []CacheValue{}
	for row.Next() {
		r := CacheValue{Provider: provider}
		row.Scan(&r.Query)
		res = append(res, r)
	}
	return res, true
}

// Will attempt to write into cache
func writeSearchCache(provider string, s string) {
	db := createDb()
	defer db.Close()

	if !createSearchTable(db) {
		return
	}
	insertSearch(db, provider, s)
}
