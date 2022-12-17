package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func prepareDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("failed to open in-memory db")
	}
	if err := createTables(db); err != nil {
		t.Fatalf("createTables err = %s, want nil", err)
	}
	return db
}

func prepareDBTx(t *testing.T) (*sql.DB, context.Context, *sql.Tx) {
	t.Helper()
	db := prepareDB(t)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal("failed to start transaction")
	}
	return db, ctx, tx
}

func queryString(ctx context.Context, tx *sql.Tx, query string, t *testing.T) string {
	t.Helper()
	res := tx.QueryRowContext(ctx, query)
	var got string
	err := res.Scan(&got)
	if err != nil {
		t.Fatalf("failed to query db: %s", err)
	}
	return got
}

func loadFixture(path string, typ interface{}, t *testing.T) {
	t.Helper()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load fixure file %s", err)
	}
	json.Unmarshal(content, typ)
}

func TestCreateTables(t *testing.T) {
	db := prepareDB(t)
	defer db.Close()
	res, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table' ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query database %s", err)
	}
	defer res.Close()

	var tableName string
	var tables []string
	for res.Next() {
		if err := res.Scan(&tableName); err != nil {
			t.Fatalf("failed to query database %s", err)
		}
		tables = append(tables, tableName)
	}
	if err := res.Err(); err != nil {
		t.Fatalf("failed to query database %s", err)
	}
	want := []string{"account", "category", "category_group", "category_month",
		"month", "payee", "server_knowledge", "subtransaction", "transaction"}
	if !reflect.DeepEqual(want, tables) {
		t.Fatalf("%v != %v", want, tables)
	}
}

func TestLoadServerKnowledge(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()
	res, err := loadServerKnowledge(ctx, tx)
	if err != nil {
		t.Fatalf("loadServerKnowledge err = %s, want nil", err)
	}
	want := map[string]int{
		"accounts":     0,
		"categories":   0,
		"months":       0,
		"payees":       0,
		"transactions": 0,
	}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("loadServerKnowledge = %v, want %v", res, want)
	}
}

func TestUpdateServerKnowledge(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()
	want := 42
	if err := updateServerKnowledge(ctx, tx, "accounts", want); err != nil {
		t.Fatalf("updateServerKnowledge err = %s, want nil", err)
	}
	res, err := loadServerKnowledge(ctx, tx)
	if err != nil {
		t.Fatalf("loadServerKnowledge err = %s, want nil", err)
	}
	if got := res["accounts"]; got != want {
		t.Fatalf(`res["accounts"] = %d, want %d`, got, want)
	}
}
func TestUpdateCategories(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var categories Categories
	loadFixture("./fixtures/categories.json", &categories, t)

	err := updateCategories(ctx, categories, tx)
	if err != nil {
		t.Fatalf("updateCategories err = %s, want nil", err)
	}
	got := queryString(ctx, tx, "SELECT name FROM category_group WHERE id = 'e8e9fa0e-0667-4b8f-afb8-8f0c0a151a1d'", t)
	if want := "Internal Master Category"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdateTransactions(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var transactions Transactions
	loadFixture("./fixtures/transactions.json", &transactions, t)

	if err := updateTransactions(ctx, transactions, tx); err != nil {
		t.Fatalf("updateTransactions err = %s, want nil", err)
	}
	got := queryString(ctx, tx, `SELECT date FROM "transaction" WHERE id = "295c1843-14dd-46ed-bed5-3d02c17a82db"`, t)
	if want := "2021-11-24"; got != want {
		t.Fatalf("%q != %q", want, got)
	}

	// check if updating a transaction record works

	transactions.Data.Transactions[0].Date = "1999-12-24"
	if err := updateTransactions(ctx, transactions, tx); err != nil {
		t.Fatalf("updateTransactions err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT date FROM "transaction" WHERE id = "295c1843-14dd-46ed-bed5-3d02c17a82db"`, t)
	if want := "1999-12-24"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdateTransactionsSubtransactions(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var transactions Transactions
	loadFixture("./fixtures/transactions.json", &transactions, t)

	if err := updateTransactions(ctx, transactions, tx); err != nil {
		t.Fatalf("updateTransactions err = %s, want nil", err)
	}
	got := queryString(ctx, tx, `SELECT transaction_id FROM "subtransaction" WHERE id = "9e53be73-3f80-4047-aacf-ca975a1b430e"`, t)
	if want := "dcc9865c-dd45-468b-93c3-fa6b327db3fe_2021-11-25"; got != want {
		t.Fatalf("%q != %q", want, got)
	}

	// update subtransaction

	for _, transaction := range transactions.Data.Transactions {
		if transaction.ID == "dcc9865c-dd45-468b-93c3-fa6b327db3fe_2021-11-25" {
			transaction.Subtransactions[0].CategoryName = "Whatever"
		}
	}
	if err := updateTransactions(ctx, transactions, tx); err != nil {
		t.Fatalf("updateTransactions err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT category_name FROM "subtransaction" WHERE id = "9e53be73-3f80-4047-aacf-ca975a1b430e"`, t)
	if want := "Whatever"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdateAccounts(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var accounts Accounts
	loadFixture("./fixtures/accounts.json", &accounts, t)

	if err := updateAccounts(ctx, accounts, tx); err != nil {
		t.Fatalf("updateAccounts err = %s, want nil", err)
	}
	got := queryString(ctx, tx, `SELECT type FROM account WHERE id = "9a329f5e-1eca-40c6-8ba1-a19b0d8cadd1"`, t)
	if want := "checking"; got != want {
		t.Fatalf("%q != %q", want, got)
	}

	// test updating existing account

	accounts.Data.Accounts[0].Type = "savings"
	if err := updateAccounts(ctx, accounts, tx); err != nil {
		t.Fatalf("updateAccounts err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT type FROM account WHERE id = "9a329f5e-1eca-40c6-8ba1-a19b0d8cadd1"`, t)
	if want := "savings"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdateCategoryMonth(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var categoryMonth CategoryMonth
	loadFixture("./fixtures/category-month.json", &categoryMonth, t)

	if err := updateCategoryMonth(ctx, "2021-11-01", categoryMonth, tx); err != nil {
		t.Fatalf("updateCategoryMonth err = %s, want nil", err)
	}
	got := queryString(ctx, tx, `SELECT budgeted FROM category_month WHERE month_id = "2022-12-01" and category_id = "94b9ac05-6a55-4e33-8f52-65931515da96"`, t)
	if want := "2000000"; got != want {
		t.Fatalf("%q != %q", want, got)
	}

	// update category_month
	categoryMonth.Data.Month.Categories[0].Budgeted = 4200

	if err := updateCategoryMonth(ctx, "2022-12-01", categoryMonth, tx); err != nil {
		t.Fatalf("updateCategoryMonth err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT budgeted FROM category_month WHERE month_id = "2022-12-01" and category_id = "94b9ac05-6a55-4e33-8f52-65931515da96"`, t)
	if want := "4200"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdateMonth(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var months Months
	loadFixture("./fixtures/month.json", &months, t)

	for _, month := range months.Data.Months {
		if err := updateMonth(ctx, month, tx); err != nil {
			t.Fatalf("updateMonth err = %s, want nil", err)
		}
	}
	got := queryString(ctx, tx, `SELECT income FROM month WHERE id = "2021-11-01"`, t)
	if want := "120000"; got != want {
		t.Fatalf("%q != %q", want, got)
	}

	// test update month

	months.Data.Months[0].Income = 230000
	if err := updateMonth(ctx, months.Data.Months[0], tx); err != nil {
		t.Fatalf("updateMonth err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT income FROM month WHERE id = "2021-11-01"`, t)
	if want := "230000"; got != want {
		t.Fatalf("%q != %q", want, got)
	}
}

func TestUpdatePayees(t *testing.T) {
	db, ctx, tx := prepareDBTx(t)
	defer db.Close()

	var payees Payees
	loadFixture("./fixtures/payees.json", &payees, t)
	if err := updatePayees(ctx, payees, tx); err != nil {
		t.Fatalf("updatePayees err = %s, want nil", err)
	}
	got := queryString(ctx, tx, `SELECT name FROM payee WHERE id = "306c522d-93c1-436d-8667-b9a32661322e"`, t)
	if want := "Hugo"; got != want {
		t.Fatalf("%v != %v", got, want)
	}

	payees.Data.Payees[len(payees.Data.Payees)-1].Name = "John"
	if err := updatePayees(ctx, payees, tx); err != nil {
		t.Fatalf("updatePayees err = %s, want nil", err)
	}
	got = queryString(ctx, tx, `SELECT name FROM payee WHERE id = "306c522d-93c1-436d-8667-b9a32661322e"`, t)
	if want := "John"; got != want {
		t.Fatalf("%v != %v", got, want)
	}
}
