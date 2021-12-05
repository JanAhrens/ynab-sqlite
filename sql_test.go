package main

import (
	"context"
	"database/sql"
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

	categories := Categories{}
	categories.Data.ServerKnowledge = 42
	// TODO initialize category_groups

	err := updateCategories(ctx, categories, tx)
	if err != nil {
		t.Fatalf("updateCategories err = %s, want nil", err)
	}
}
