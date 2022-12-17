package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func assertValue(t *testing.T, fieldName string, got interface{}, want interface{}) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", fieldName, got, want)
	}
}

func assertInt(t *testing.T, fieldName string, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %d, want %d", fieldName, got, want)
	}
}

func assertNil(t *testing.T, fieldName string, got *string) {
	t.Helper()
	var nilString *string = nil
	if got != nilString {
		t.Fatalf("%s = %d, want nil", fieldName, got)
	}
}

func assertNilInt(t *testing.T, fieldName string, got *int) {
	t.Helper()
	var nilInt *int = nil
	if got != nilInt {
		t.Fatalf("%s = %d, want nil", fieldName, got)
	}
}

func fixtureGET(t *testing.T, file string) *httptest.Server {
	t.Helper()

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read fixture file %s", file)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if got, want := r.Method, "GET"; got != want {
			t.Errorf("r.Method = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token"; got != want {
			t.Errorf(`r.Header.Get("Authorization") = %q, want %q`, got, want)
		}
		w.Write(content)
	}))
}

func TestLoadCategories(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/categories.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-user")
	data := ynab.LoadCategories(0).Data
	if got, want := data.ServerKnowledge, 98; got != want {
		t.Fatalf("data.ServerKnowledge = %d, want %d", got, want)
	}

	if got, want := len(data.CategoryGroups), 8; got != want {
		t.Fatalf("len(data.CategoryGroups) = %d, want %d", got, want)
	}

	if got, want := len(data.CategoryGroups[0].Categories), 2; got != want {
		t.Fatalf("len(data.CategoryGroups[0].Categories) = %d, want %d", got, want)
	}
}

func TestLoadCategoryMonths(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/category-month.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-used")
	categoryMonth := ynab.LoadCategoryMonths("month-id")
	if got, want := categoryMonth.Data.Month.Categories[0].Name, "Electric 213"; got != want {
		t.Fatalf("categoryMonth.Data.Category.Name = %q, want %q", got, want)
	}
	category := categoryMonth.Data.Month.Categories[0]
	assertValue(t, "ID", category.ID, "94b9ac05-6a55-4e33-8f52-65931515da96")
	assertValue(t, "CategoryGroupID", category.CategoryGroupID, "5423a142-b27a-4a54-b6a6-adfdb31a41bc")
	assertValue(t, "Name", category.Name, "Electric 213")
	assertValue(t, "Hidden", category.Hidden, false)
	assertNil(t, "OriginalCategoryGroupID", category.OriginalCategoryGroupID)
	assertNil(t, "Note", category.Note)
	assertInt(t, "Budgeted", category.Budgeted, 2000000)
	assertInt(t, "Activity", category.Activity, -2000)
	assertInt(t, "Balance", category.Balance, 2001000)
	assertNil(t, "GoalType", category.GoalType)
	assertNil(t, "GoalCreationMonth", category.GoalCreationMonth)
	assertInt(t, "GoalTarget", category.GoalTarget, 0)
	assertNil(t, "GoalTargetMonth", category.GoalTargetMonth)
	assertNilInt(t, "GoalPercentageComplete", category.GoalPercentageComplete)
	assertNilInt(t, "GoalMonthsToBudget", category.GoalMonthsToBudget)
	assertNilInt(t, "GoalUnderFunded", category.GoalUnderFunded)
	assertNilInt(t, "GoalOverallFunded", category.GoalOverallFunded)
	assertNilInt(t, "GoalOverallLeft", category.GoalOverallLeft)
	assertValue(t, "Deleted", category.Deleted, false)
}

func TestLoadMonth(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/month.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-user")
	months := ynab.LoadMonths(0)
	if got, want := len(months.Data.Months), 2; got != want {
		t.Fatalf("len(months.Data.Months) = %d, want %d", got, want)
	}
	if got, want := months.Data.ServerKnowledge, 98; got != want {
		t.Fatalf("months.Data.ServerKnowledge = %d, want %d", got, want)
	}
}

func TestAccounts(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/accounts.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-used")
	accounts := ynab.LoadAccounts(0)
	if got, want := len(accounts.Data.Accounts), 2; got != want {
		t.Fatalf("len(accounts.Data.Accounts) = %d, want %d", got, want)
	}
	first := accounts.Data.Accounts[0]
	assertValue(t, "ID", first.ID, "9a329f5e-1eca-40c6-8ba1-a19b0d8cadd1")
	assertValue(t, "Name", first.Name, "Checker")
	assertValue(t, "Type", first.Type, "checking")
	assertValue(t, "OnBudget", first.OnBudget, true)
	assertValue(t, "Closed", first.Closed, false)
	assertNil(t, "Note", first.Note)
	assertInt(t, "Balance", first.Balance, 95000)
	assertInt(t, "ClearedBalance", first.ClearedBalance, 118000)
	assertInt(t, "UnclearedBalance", first.UnclearedBalance, -23000)
	assertValue(t, "TransferPayeeId", first.TransferPayeeID, "db6deeec-b0ba-4b1e-a09f-1338822ec9d0")
	assertValue(t, "DirectImportLinked", first.DirectImportLinked, false)
	assertValue(t, "DirectImportInError", first.DirectImportInError, false)
	assertValue(t, "Deleted", first.Deleted, false)
}

func TestTransactions(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/transactions.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-used")
	transactions := ynab.LoadTransactions(0)
	if got, want := len(transactions.Data.Transactions), 4; got != want {
		t.Fatalf("len(transactions.Data.Transactions) = %d, want %d", got, want)
	}

	first := transactions.Data.Transactions[0]
	assertValue(t, "ID", first.ID, "295c1843-14dd-46ed-bed5-3d02c17a82db")
	assertValue(t, "Date", first.Date, "2021-11-24")
	assertValue(t, "Amount", first.Amount, -23000)
	assertValue(t, "Memo", first.Memo, "")
	assertValue(t, "Cleared", first.Cleared, "uncleared")
	assertValue(t, "Approved", first.Approved, true)
	assertNil(t, "FlagColor", first.FlagColor)
	assertValue(t, "AccountId", first.AccountID, "9a329f5e-1eca-40c6-8ba1-a19b0d8cadd1")
	assertValue(t, "AccountName", first.AccountName, "Checker")
	assertValue(t, "PayeeId", first.PayeeID, "306c522d-93c1-436d-8667-b9a32661322e")
	assertValue(t, "PayeeName", first.PayeeName, "Hugo")
	assertValue(t, "CategoryId", first.CategoryID, "7d3b19a3-a347-4a10-befc-b966f278aa3e")
	assertValue(t, "CategoryName", first.CategoryName, "Water")
	assertNil(t, "TransferTransactionId", first.TransferTransactionID)
	assertNil(t, "MatchedTransactionId", first.MatchedTransactionID)
	assertNil(t, "ImportId", first.ImportID)
	assertValue(t, "Deleted", first.Deleted, false)
	assertValue(t, "len(Subtransactions)=0", len(first.Subtransactions), 0)
}

func TestPayees(t *testing.T) {
	ts := fixtureGET(t, "./fixtures/payees.json")
	defer ts.Close()

	ynab := NewYNAB(ts.URL, "token", "last-used")
	payees := ynab.LoadPayees(0)
	if got, want := len(payees.Data.Payees), 7; got != want {
		t.Fatalf("len(payees.Data.Payees) = %d, want %d", got, want)
	}

	first := payees.Data.Payees[0]
	assertValue(t, "ID", first.ID, "8a8fbcd5-2eda-478d-a977-c8c1122f6e3a")
	assertValue(t, "Name", first.Name, "Starting Balance")
	assertNil(t, "TransferAccountId", first.TransferAccountID)
	assertValue(t, "Deleted", first.Deleted, false)
}
