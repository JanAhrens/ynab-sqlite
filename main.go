package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Responses struct {
	categories    Categories
	months        Months
	accounts      Accounts
	transactions  Transactions
	payees        Payees
	categoryMonth []CategoryMonth
}

func updateDatabase(ctx context.Context, tx *sql.Tx, responses Responses) error {
	if err := updateCategories(ctx, responses.categories, tx); err != nil {
		return fmt.Errorf("couldn't update categories: %s", err)
	}

	if err := updateMonthServerKnowledge(ctx, responses.months, tx); err != nil {
		return fmt.Errorf("could not update month server knowledge: %s", err)
	}

	if err := updateAccounts(ctx, responses.accounts, tx); err != nil {
		log.Panicf("could not update accounts: %s", err)
	}

	if err := updateTransactions(ctx, responses.transactions, tx); err != nil {
		log.Panicf("could not update transactions: %s", err)
	}

	for _, month := range responses.months.Data.Months {
		if err := updateMonth(ctx, month, tx); err != nil {
			log.Panicf("could not update months: %s", err)
		}
	}

	for _, categoryMonth := range responses.categoryMonth {
		if err := updateCategoryMonth(ctx, categoryMonth, tx); err != nil {
			log.Panicf("could not update category month %s", err)
		}
	}

	return updatePayees(ctx, responses.payees, tx)
}

func main() {
	apiKey, ok := os.LookupEnv("YNAB_API_KEY")
	if !ok {
		log.Fatal("YNAB_API_KEY not set")
	}

	ynab := NewYNAB(
		"https://api.youneedabudget.com/v1",
		apiKey,
		"last-used",
	)

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	sqlite := NewSqliteService(db)

	if err = sqlite.CreateTables(); err != nil {
		log.Fatal("failed to create database tables")
	}

	err = sqlite.Transaction(func(ctx context.Context, tx *sql.Tx) error {
		serverKnowledge, err := loadServerKnowledge(ctx, tx)
		if err != nil {
			return err
		}

		responses := Responses{
			categories:    ynab.LoadCategories(serverKnowledge["categories"]),
			months:        ynab.LoadMonths(serverKnowledge["months"]),
			accounts:      ynab.LoadAccounts(serverKnowledge["accounts"]),
			transactions:  ynab.LoadTransactions(serverKnowledge["transactions"]),
			payees:        ynab.LoadPayees(serverKnowledge["payees"]),
			categoryMonth: nil, // wait until months are loaded and only load required monthly budgets
		}

		for _, month := range responses.months.Data.Months {
			responses.categoryMonth = append(responses.categoryMonth, ynab.LoadCategoryMonths(month.Month))
		}

		return updateDatabase(ctx, tx, responses)
	})
	if err != nil {
		log.Fatal("failure in database transaction")
	}
}
