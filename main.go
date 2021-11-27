package main

import (
	"database/sql"
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

	months := loadMonths(budgetId, apiKey, serverKnowledge["months"])
	updateMonthServerKnowledge(months, db)

	for _, month := range months.Data.Months {
		updateMonth(month, db)
		for _, categoryGroup := range categories.Data.CategoryGroups {
			for _, category := range categoryGroup.Categories {
				categoryMonth := loadCategoryMonths(budgetId, apiKey, month.Month, category.Id)
				updateCategoryMonth(month.Month, categoryMonth, db)
			}
		}
	}

	payees := loadPayees(budgetId, apiKey, serverKnowledge["payees"])
	updatePayees(payees, db)
}
