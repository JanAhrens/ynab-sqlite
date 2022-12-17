package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type YNAB struct {
	prefix   string
	apiKey   string
	budgetId string
}

func NewYNAB(prefix string, apiKey string, budgetId string) YNAB {
	return YNAB{prefix: prefix, apiKey: apiKey, budgetId: budgetId}
}

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
		Month struct {
			Month
			Categories []category
		}
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
			ID                    string  `json:"id"`
			Date                  string  `json:"date"`
			Amount                int     `json:"amount"`
			Memo                  string  `json:"memo"`
			Cleared               string  `json:"cleared"`
			Approved              bool    `json:"approved"`
			FlagColor             *string `json:"flag_color"`
			AccountID             string  `json:"account_id"`
			PayeeID               string  `json:"payee_id"`
			CategoryID            string  `json:"category_id"`
			TransferAccountID     string  `json:"transfer_account_id"`
			TransferTransactionID *string `json:"transfer_transaction_id"`
			MatchedTransactionID  *string `json:"matched_transaction_id"`
			ImportID              *string `json:"import_id"`
			Deleted               bool    `json:"deleted"`
			AccountName           string  `json:"account_name"`
			PayeeName             string  `json:"payee_name"`
			CategoryName          string  `json:"category_name"`
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
			ID                string  `json:"id"`
			Name              string  `json:"name"`
			TransferAccountID *string `json:"transfer_account_id"`
			Deleted           bool    `json:"deleted"`
		} `json:"payees"`
		ServerKnowledge int `json:"server_knowledge"`
	} `json:"data"`
}

func (ynab YNAB) request(url string) (*[]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ynab.apiKey))
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

	bytes, _ := io.ReadAll(res.Body)
	return &bytes, nil
}

func (ynab YNAB) LoadCategories(serverKnowledge int) Categories {
	url := fmt.Sprintf("%s/budgets/%s/categories?last_knowledge_of_server=%d", ynab.prefix, ynab.budgetId, serverKnowledge)
	bytes, err := ynab.request(url)
	if err != nil {
		log.Panic("failed to load categories list")
	}
	var categories Categories
	json.Unmarshal(*bytes, &categories)
	return categories
}

func (ynab YNAB) LoadMonths(serverKnowledge int) Months {
	bytes, err := ynab.request(
		fmt.Sprintf("%s/budgets/%s/months?last_knowledge_of_server=%d",
			ynab.prefix,
			ynab.budgetId,
			serverKnowledge,
		),
	)
	if err != nil {
		log.Panic("failed to load month list")
	}
	var months Months
	json.Unmarshal(*bytes, &months)
	return months
}

func (ynab YNAB) LoadCategoryMonths(monthID string) CategoryMonth {
	var categoryMonth CategoryMonth

	bytes, err := ynab.request(
		fmt.Sprintf("%s/budgets/%s/months/%s", ynab.prefix, ynab.budgetId, monthID),
	)
	if err != nil {
		log.Panicf("fail to load category month %s", monthID)
	}
	json.Unmarshal(*bytes, &categoryMonth)
	return categoryMonth
}

func (ynab YNAB) LoadAccounts(serverKnowledge int) Accounts {
	bytes, err := ynab.request(
		fmt.Sprintf("%s/budgets/%s/accounts?last_knowledge_of_server=%d",
			ynab.prefix,
			ynab.budgetId,
			serverKnowledge,
		),
	)
	if err != nil {
		log.Panic("Failed to load accounts list")
	}
	var accounts Accounts
	json.Unmarshal(*bytes, &accounts)
	return accounts
}

func (ynab YNAB) LoadTransactions(serverKnowledge int) Transactions {
	bytes, err := ynab.request(
		fmt.Sprintf("%s/budgets/%s/transactions?last_knowledge_of_server=%d",
			ynab.prefix,
			ynab.budgetId,
			serverKnowledge),
	)
	if err != nil {
		log.Panic("Failed to load transactions list")
	}
	var transactions Transactions
	json.Unmarshal(*bytes, &transactions)
	return transactions
}

func (ynab YNAB) LoadPayees(serverKnowledge int) Payees {
	bytes, err := ynab.request(
		fmt.Sprintf(
			"%s/budgets/%s/payees?last_knowledge_of_server=%d",
			ynab.prefix,
			ynab.budgetId,
			serverKnowledge,
		),
	)
	if err != nil {
		log.Panic("failed to load payees")
	}
	var payees Payees
	json.Unmarshal(*bytes, &payees)
	return payees
}
