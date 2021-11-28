package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func execTransaction(tx *sql.Tx, ctx context.Context, apiKey string, budgetId string) error {
	serverKnowledge, err := loadServerKnowledge(tx, ctx)
	if err != nil {
		return fmt.Errorf("failed to load server knowledge from databse: %s", err)
	}

	categories := loadCategories(budgetId, apiKey, serverKnowledge["categories"])
	if err = updateCategories(categories, tx, ctx); err != nil {
		return fmt.Errorf("couldn't update categories: %s", err)
	}

	transactions := loadTransactions(budgetId, apiKey, serverKnowledge["transactions"])
	if err = updateTransactions(transactions, tx, ctx); err != nil {
		return fmt.Errorf("could not update transactions: %s", err)
	}

	accounts := loadAccounts(budgetId, apiKey, serverKnowledge["accounts"])
	if err = updateAccounts(accounts, tx, ctx); err != nil {
		return fmt.Errorf("could not update accounts: %s", err)
	}

	months := loadMonths(budgetId, apiKey, serverKnowledge["months"])
	if err = updateMonthServerKnowledge(months, tx, ctx); err != nil {
		return fmt.Errorf("could not update month server knowledge: %s", err)
	}

	for _, month := range months.Data.Months {
		if err = updateMonth(month, tx, ctx); err != nil {
			return fmt.Errorf("could not update months: %s", err)
		}
		for _, categoryGroup := range categories.Data.CategoryGroups {
			for _, category := range categoryGroup.Categories {
				categoryMonth := loadCategoryMonths(budgetId, apiKey, month.Month, category.Id)
				if err = updateCategoryMonth(month.Month, categoryMonth, tx, ctx); err != nil {
					return fmt.Errorf("could not update category month", err)
				}
			}
		}
	}

	payees := loadPayees(budgetId, apiKey, serverKnowledge["payees"])
	return updatePayees(payees, tx, ctx)
}

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

	if err = createTables(db); err != nil {
		log.Fatal("failed to create database tables")
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal("could start database transaction")
	}
	err = execTransaction(tx, ctx, apiKey, budgetId)
	if err != nil {
		tx.Rollback()
		log.Fatalf("transaction failed: %s", err)
	} else {
		tx.Commit()
	}
}
