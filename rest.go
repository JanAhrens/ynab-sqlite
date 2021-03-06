package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type category struct {
	ID                      string  `json:"id"`
	CategoryGroupID         string  `json:"category_group_id"`
	Name                    string  `json:"name"`
	Hidden                  bool    `json:"hidden"`
	OriginalCategoryGroupID *string `json:"original_category_group_id"`
	Note                    *string `json:"note"`
	Budgeted                int     `json:"budgeted"`
	Activity                int     `json:"activity"`
	Balance                 int     `json:"balance"`
	GoalType                *string `json:"goal_type"`
	GoalCreationMonth       *string `json:"goal_creation_month"`
	GoalTarget              int     `json:"goal_target"`
	GoalTargetMonth         *string `json:"goal_target_month"`
	GoalPercentageComplete  *int    `json:"goal_percentage_complete"`
	GoalMonthsToBudget      *int    `json:"goal_months_to_budget"`
	GoalUnderFunded         *int    `json:"goal_under_funded"`
	GoalOverallFunded       *int    `json:"goal_overall_funded"`
	GoalOverallLeft         *int    `json:"goal_overall_left"`
	Deleted                 bool    `json:"deleted"`
}

// CategoryMonth GET /v1/budgets/:id/months/:month_id/categories/:category_id
type CategoryMonth struct {
	Data struct {
		Category category
	}
}

// Categories GET /v1/budgets/:id/categories
type Categories struct {
	Data struct {
		CategoryGroups []struct {
			ID         string     `json:"id"`
			Name       string     `json:"name"`
			Hidden     bool       `json:"hidden"`
			Deleted    bool       `json:"deleted"`
			Categories []category `json:"categories"`
		} `json:"category_groups"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// Month is part of GET /v1/budgets/:id/months
type Month struct {
	Month        string `json:"month"`
	Note         string `json:"note"`
	Income       int    `json:"income"`
	Budgeted     int    `json:"budgeted"`
	Activity     int    `json:"activity"`
	ToBeBudgeted int    `json:"to_be_budgeted"`
	AgeOfMoney   int    `json:"age_of_money"`
	Deleted      bool   `json:"deleted"`
}

// Months GET /v1/budgets/:id/months
type Months struct {
	Data struct {
		Months          []Month `json:"months"`
		ServerKnowledge int     `json:"server_knowledge"`
	} `json:"data"`
}

// Accounts GET /v1/budgets/:id/accounts
type Accounts struct {
	Data struct {
		Accounts []struct {
			ID                  string  `json:"id"`
			Name                string  `json:"name"`
			Type                string  `json:"type"`
			OnBudget            bool    `json:"on_budget"`
			Closed              bool    `json:"closed"`
			Note                *string `json:"note"`
			Balance             int     `json:"balance"`
			ClearedBalance      int     `json:"cleared_balance"`
			UnclearedBalance    int     `json:"uncleared_balance"`
			TransferPayeeID     string  `json:"transfer_payee_id"`
			DirectImportLinked  bool    `json:"direct_import_linked"`
			DirectImportInError bool    `json:"direct_import_in_error"`
			Deleted             bool    `json:"deleted"`
		} `json:"accounts"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// Transactions GET /v1/budgets/:id/transactions
type Transactions struct {
	Data struct {
		Transactions []struct {
			ID                    string `json:"id"`
			Date                  string `json:"date"`
			Amount                int    `json:"amount"`
			Memo                  string `json:"memo"`
			Cleared               string `json:"cleared"`
			Approved              bool   `json:"approved"`
			FlagColor             string `json:"flag_color"`
			AccountID             string `json:"account_id"`
			PayeeID               string `json:"payee_id"`
			CategoryID            string `json:"category_id"`
			TransferAccountID     string `json:"transfer_account_id"`
			TransferTransactionID string `json:"transfer_transaction_id"`
			MatchedTransactionID  string `json:"matched_transaction_id"`
			ImportID              string `json:"import_id"`
			Deleted               bool   `json:"deleted"`
			AccountName           string `json:"account_name"`
			PayeeName             string `json:"payee_name"`
			CategoryName          string `json:"category_name"`
			Subtransactions       []struct {
				ID                    string `json:"id"`
				TransactionID         string `json:"transaction_id"`
				Amount                int    `json:"amount"`
				Memo                  string `json:"memo"`
				PayeeID               string `json:"payee_id"`
				PayeeName             string `json:"payee_name"`
				CategoryID            string `json:"category_id"`
				CategoryName          string `json:"category_name"`
				TransferAccountID     string `json:"transfer_account_id"`
				TransferTransactionID string `json:"transfer_transaction_id"`
				Deleted               bool   `json:"deleted"`
			} `json:"subtransactions"`
		} `json:"transactions"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

// Payees GET /v1/budgets/:budget_id/payees
type Payees struct {
	Data struct {
		Payees []struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			TransferAccountID string `json:"transfer_account_id"`
			Deleted           bool   `json:"deleted"`
		} `json:"payees"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

func request(url string, apiKey string) (*[]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer res.Body.Close()

	// https://api.youneedabudget.com/#rate-limiting
	// every access token can generate 200 requests per hour
	log.Printf("%s %v %s\n", url, res.Status, res.Header.Get("X-Rate-Limit"))

	// check if response outside of 2xx or 3xx code
	if !(res.StatusCode >= 200 && res.StatusCode <= 399) {
		return nil, fmt.Errorf("failed request with status code %d", res.StatusCode)
	}

	bytes, _ := ioutil.ReadAll(res.Body)
	return &bytes, nil
}

func loadCategories(prefix string, budgetID string, apiKey string, serverKnowledge int) Categories {
	url := fmt.Sprintf("%s/budgets/%s/categories?last_knowledge_of_server=%d", prefix, budgetID, serverKnowledge)
	bytes, err := request(
		url,
		apiKey)
	if err != nil {
		log.Panic("failed to load categories list")
	}
	var categories Categories
	json.Unmarshal(*bytes, &categories)
	return categories
}

func loadMonths(prefix string, budgetID string, apiKey string, serverKnowledge int) Months {
	bytes, err := request(
		fmt.Sprintf("%s/budgets/%s/months?last_knowledge_of_server=%d", prefix, budgetID, serverKnowledge),
		apiKey)
	if err != nil {
		log.Panic("failed to load month list")
	}
	var months Months
	json.Unmarshal(*bytes, &months)
	return months
}

func loadCategoryMonths(prefix string, budgetID string, apiKey, monthID string, categoryID string) (CategoryMonth, error) {
	var categoryMonth CategoryMonth

	url := "%s/budgets/%s/months/%s/categories/%s"
	bytes, err := request(
		fmt.Sprintf(url, prefix, budgetID, monthID, categoryID),
		apiKey)
	if err != nil {
		return categoryMonth, err
	}
	json.Unmarshal(*bytes, &categoryMonth)
	return categoryMonth, nil
}

func loadAccounts(prefix string, budgetID string, apiKey string, serverKnowledge int) Accounts {
	bytes, err := request(
		fmt.Sprintf("%s/budgets/%s/accounts?last_knowledge_of_server=%d", prefix, budgetID, serverKnowledge),
		apiKey)
	if err != nil {
		log.Panic("Failed to load accounts list")
	}
	var accounts Accounts
	json.Unmarshal(*bytes, &accounts)
	return accounts
}

func loadTransactions(budgetID string, apiKey string, serverKnowledge int) Transactions {
	bytes, err := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/transactions?last_knowledge_of_server=%d", budgetID, serverKnowledge),
		apiKey)
	if err != nil {
		log.Panic("Failed to load transactions list")
	}
	var transactions Transactions
	json.Unmarshal(*bytes, &transactions)
	return transactions
}

func loadPayees(budgetID string, apiKey string, serverKnowledge int) Payees {
	bytes, err := request(
		fmt.Sprintf("https://api.youneedabudget.com/v1/budgets/%s/payees?last_knowledge_of_server=%d", budgetID, serverKnowledge),
		apiKey)
	if err != nil {
		log.Panic("failed to load payees")
	}
	var payees Payees
	json.Unmarshal(*bytes, &payees)
	return payees
}
