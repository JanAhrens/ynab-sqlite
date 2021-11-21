package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type category struct {
	Id                      string `json:"id"`
	CategoryGroupId         string `json:"category_group_id"`
	Name                    string `json:"name"`
	Hidden                  bool   `json:"hidden"`
	OriginalCategoryGroupId string `json:"original_category_group_id"`
	Note                    string `json:"note"`
	Budgeted                int    `json:"budgeted"`
	Activity                int    `json:"activity"`
	Balance                 int    `json:"balance"`
	GoalType                string `json:"goal_type"`
	GoalCreationMonth       string `json:"goal_creation_month"`
	GoalTarget              int    `json:"goal_target"`
	GoalTargetMonth         string `json:"goal_target_month"`
	GoalPercentageComplete  int    `json:"goal_percentage_complete"`
	GoalMonthsToBudget      int    `json:"goal_months_to_budget"`
	GoalUnderFunded         int    `json:"goal_under_funded"`
	GoalOverallFunded       int    `json:"goal_overall_funded"`
	GoalOverallLeft         int    `json:"goal_overall_left"`
	Deleted                 bool   `json:"deleted"`
}

// GET /v1/budgets/:id/months/:month_id/categories/:category_id
type CategoryMonth struct {
	Data struct {
		Category category
	}
}

// GET /v1/budgets/:id/categories
type Categories struct {
	Data struct {
		CategoryGroups []struct {
			Id         string     `json:"id"`
			Name       string     `json:"name"`
			Hidden     bool       `json:"hidden"`
			Deleted    bool       `json:"deleted"`
			Categories []category `json:"categories"`
		} `json:"category_groups"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// GET /v1/budgets/:id/months
type Months struct {
	Data struct {
		Months []struct {
			Month        string `json:"month"`
			Note         string `json:"note"`
			Income       int    `json:"income"`
			Budgeted     int    `json:"budgeted"`
			Activity     int    `json:"activity"`
			ToBeBudgeted int    `json:"to_be_budgeted"`
			AgeOfMoney   int    `json:"age_of_money"`
			Deleted      bool   `json:"deleted"`
		} `json:"months"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// GET /v1/budgets/:id/accounts
type Accounts struct {
	Data struct {
		Accounts []struct {
			Id                  string `json:"id"`
			Name                string `json:"name"`
			Type                string `json:"type"`
			OnBudget            bool   `json:"on_budget"`
			Closed              bool   `json:"closed"`
			Note                string `json:"note"`
			ClearedBalance      int    `json:"cleared_balance"`
			UnclearedBalance    int    `json:"uncleared_balane"`
			TransferPayeeId     string `json:"transfer_payee_id"`
			DirectImportLinked  bool   `json:"direct_import_linked"`
			DirectImportInError bool   `json:"direct_import_in_error"`
			Deleted             bool   `json:"deleted"`
		} `json:"accounts"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// GET /v1/budgets/:id/transactions
type Transactions struct {
	Data struct {
		Transactions []struct {
			Id                    string `json:"id"`
			Date                  string `json:"date"`
			Amount                int    `json:"amount"`
			Memo                  string `json:"memo"`
			Cleared               string `json:"cleared"`
			Approved              bool   `json:"approved"`
			FlagColor             string `json:"flag_color"`
			AccountId             string `json:"account_id"`
			PayeeId               string `json:"payee_id"`
			CategoryId            string `json:"category_id"`
			TransferAccountId     string `json:"transfer_account_id"`
			TransferTransactionId string `json:"transfer_transaction_id"`
			MatchedTransactionId  string `json:"matched_transaction_id"`
			ImportId              string `json:"import_id"`
			Deleted               bool   `json:"deleted"`
			AccountName           string `json:"account_name"`
			PayeeName             string `json:"payee_name"`
			CategoryName          string `json:"category_name"`
			Subtransactions       []struct {
				Id                    string `json:"id"`
				TransactionId         string `json:"transaction_id"`
				Amount                int    `json:"amount"`
				Memo                  string `json:"memo"`
				PayeeId               string `json:"payee_id"`
				PayeeName             string `json:"payee_name"`
				CategoryId            string `json:"category_id"`
				CategoryName          string `json:"category_name"`
				TransferAccountId  string `json:"transfer_account_id"`
				TransferTransactionId string `json:"transfer_transaction_id"`
				Deleted               bool   `json:"deleted"`
			} `json:"subtransactions"`
		} `json:"transactions"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

func request(url string, apiKey string) *[]byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	res, _ := client.Do(req)
	defer res.Body.Close()
	bytes, _ := ioutil.ReadAll(res.Body)
	return &bytes
}

func loadCategories(budgetId string, apiKey string) Categories {
	bytes := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/categories", budgetId),
		apiKey)

	var categories Categories
	json.Unmarshal(*bytes, &categories)
	return categories
}

func loadMonths(budgetId string, apiKey string) Months {
	bytes := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/months", budgetId),
		apiKey)

	var months Months
	json.Unmarshal(*bytes, &months)
	return months
}

func loadCategoryMonths(budgetId string, apiKey, monthId string, categoryId string) CategoryMonth {
	url := "https://api.youneedabudget.com/v1/budgets/%s/months/%s/categories/%s"
	bytes := request(
		fmt.Sprintf(url, budgetId, monthId, categoryId),
		apiKey)

	var categoryMonth CategoryMonth
	json.Unmarshal(*bytes, &categoryMonth)
	return categoryMonth
}

func loadAccounts(budgetId string, apiKey string) Accounts {
	bytes := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/accounts", budgetId),
		apiKey)

	var accounts Accounts
	json.Unmarshal(*bytes, &accounts)
	return accounts
}

func loadTransactions(budgetId string, apiKey string) Transactions {
	bytes := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/transactions", budgetId),
		apiKey)

	var transactions Transactions
	json.Unmarshal(*bytes, &transactions)
	return transactions
}
