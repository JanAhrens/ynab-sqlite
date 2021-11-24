package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	budgetId := "last-used"
	apiKey, ok := os.LookupEnv("YNAB_API_KEY")
	if !ok {
		log.Fatal("YNAB_API_KEY not set")
	}

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	createTables(db)

	serverKnowledge := loadServerKnowledge(db)

	categories := loadCategories(budgetId, apiKey, serverKnowledge["categories"])
	updateCategories(categories, db)

	transactions := loadTransactions(budgetId, apiKey, serverKnowledge["transactions"])
	updateTransactions(transactions, db)

	accounts := loadAccounts(budgetId, apiKey, serverKnowledge["accounts"])
	updateAccounts(accounts, db)

	os.Exit(0)
	months := loadMonths(budgetId, apiKey)

	for _, month := range months.Data.Months {
		categoryMonth := loadCategoryMonths(budgetId, apiKey, month.Month, categories.Data.CategoryGroups[0].Categories[0].Id)
		fmt.Printf("%s %d\n", month.Month, categoryMonth.Data.Category.Budgeted)
	}
}
