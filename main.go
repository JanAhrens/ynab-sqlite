package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func execTransaction(ctx context.Context, tx *sql.Tx, apiKey string, budgetID string) error {
	prefix := "https://api.youneedabudget.com/v1"
	serverKnowledge, err := loadServerKnowledge(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to load server knowledge from databse: %s", err)
	}

	categories := loadCategories(prefix, budgetID, apiKey, serverKnowledge["categories"])
	if err = updateCategories(ctx, categories, tx); err != nil {
		return fmt.Errorf("couldn't update categories: %s", err)
	}

	months := loadMonths(prefix, budgetID, apiKey, serverKnowledge["months"])
	if err = updateMonthServerKnowledge(ctx, months, tx); err != nil {
		return fmt.Errorf("could not update month server knowledge: %s", err)
	}

	accounts := loadAccounts(prefix, budgetID, apiKey, serverKnowledge["accounts"])
	if err = updateAccounts(ctx, accounts, tx); err != nil {
		log.Panicf("could not update accounts: %s", err)
	}

	transactions := loadTransactions(prefix, budgetID, apiKey, serverKnowledge["transactions"])
	if err = updateTransactions(ctx, transactions, tx); err != nil {
		log.Panicf("could not update transactions: %s", err)
	}

	for _, month := range months.Data.Months {

		log.Printf("Loading month %s", month.Month)
		if err = updateMonth(ctx, month, tx); err != nil {
			log.Panicf("could not update months: %s", err)
		}

		categoryMonth, err := loadCategoryMonths(prefix, budgetID, apiKey, month.Month)
		if err != nil {
			log.Printf("skipping month %s", month.Month)
		}
		if err = updateCategoryMonth(ctx, month.Month, categoryMonth, tx); err != nil {
			log.Panicf("could not update category month %s", err)
		}

	}

	payees := loadPayees(prefix, budgetID, apiKey, serverKnowledge["payees"])
	return updatePayees(ctx, payees, tx)
}

func main() {
	budgetID := "last-used"
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
	err = execTransaction(ctx, tx, apiKey, budgetID)
	if err != nil {
		tx.Rollback()
		log.Fatalf("transaction failed: %s", err)
	} else {
		tx.Commit()
	}
	db.Close()
}
