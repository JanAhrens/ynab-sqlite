package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

	content, err := ioutil.ReadFile(file)
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

	data := loadCategories(ts.URL, "last-user", "token", 0).Data
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

	categoryMonth, _ := loadCategoryMonths(ts.URL, "last-used", "token", "month-id")
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

	months := loadMonths(ts.URL, "last-user", "token", 0)
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

	accounts := loadAccounts(ts.URL, "last-used", "token", 0)
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
