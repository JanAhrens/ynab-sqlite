package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type category struct {
	Id                string `json:"id"`
	CategoryGroupId   string `json:"category_group_id"`
	Name              string `json:"name"`
	Hidden            bool   `json:"hidden"`
	Deleted           bool   `json:"deleted"`
	Note              string `json:"note"`
	Budgeted          int    `json:"budgeted"`
	Activity          int    `json:"activity"`
	Balance           int    `json:"balance"`
	GoalType          string `json:"goal_type"`
	GoalCreationMonth string `json:"goal_creation_month"`
	GoalTarget        int    `json:"goal_target"`
	GoalTargetMonth   string `json:"goal_target_month"`
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
			Id      string `json:"id"`
			Date    string `json:"date"`
			Amount  int    `json:"amount"`
			Memo    string `json:"memo"`
			Cleared string `json:"cleared"`
		} `json:"transactions"`
	} `json:"data"`
}

func createTables(db *sql.DB) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS server_knowledge (
      "endpoint" TEXT NOT NULL PRIMARY KEY,
      "value"    INTEGER
    );`,
		`CREATE TABLE IF NOT EXISTS category_group (
      id      TEXT NOT NULL PRIMARY KEY,
      name    TEXT NOT NULL,
      hidden  INTEGER,
      deleted INTEGER
    );`,
		`CREATE TABLE IF NOT EXISTS category (
      id                  TEXT NOT NULL PRIMARY KEY,
      category_group_id   TEXT NOT NULL,
      name                TEXT NOT NULL,
      hidden              INTEGER,
      deleted             INTEGER,
      goal_type           TEXT,
      goal_creation_month TEXT,
      goal_target         TEXT,
      goal_target_month   TEXT
    );`,
	}
	for _, sql := range tables {
		statement, err := db.Prepare(sql)
		if err != nil {
			log.Fatal(err.Error())
		}
		statement.Exec()
	}
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

func updateCategories(categories Categories, db *sql.DB) {
	insertCategoryGroupSql := `
    INSERT INTO category_group (
      id, name, hidden, deleted
    ) VALUES(?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      name=excluded.name, hidden=excluded.hidden, deleted=excluded.hidden;
  `
	insertCategorySql := `
    INSERT INTO category (
      id, category_group_id, name, hidden, deleted, goal_type,
      goal_creation_month, goal_target, goal_target_month
    ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      name=excluded.name, category_group_id=excluded.category_group_id,
      hidden=excluded.name, deleted=excluded.deleted,
      goal_type=excluded.goal_type,
      goal_creation_month=excluded.goal_creation_month,
      goal_target=excluded.goal_target,
      goal_target_month=excluded.goal_target_month;
  `
	for _, group := range categories.Data.CategoryGroups {
		statement, err := db.Prepare(insertCategoryGroupSql)
		if err != nil {
			log.Fatal(err.Error())
		}
		_, err = statement.Exec(group.Id, group.Name, group.Hidden, group.Deleted)

		for _, category := range group.Categories {
			statement, err := db.Prepare(insertCategorySql)
			if err != nil {
				log.Fatal(err.Error())
			}
			_, err = statement.Exec(category.Id, category.CategoryGroupId, category.Name, category.Hidden, category.Deleted, category.GoalType, category.GoalCreationMonth, category.GoalTarget, category.GoalTargetMonth)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
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

	createTables(db)

	categories := loadCategories(budgetId, apiKey)
	updateCategories(categories, db)

	os.Exit(0)
	months := loadMonths(budgetId, apiKey)
	accounts := loadAccounts(budgetId, apiKey)

	for _, month := range months.Data.Months {
		categoryMonth := loadCategoryMonths(budgetId, apiKey, month.Month, categories.Data.CategoryGroups[0].Categories[0].Id)
		fmt.Printf("%s %d\n", month.Month, categoryMonth.Data.Category.Budgeted)
	}

	os.Exit(0)

	for _, account := range accounts.Data.Accounts {
		if !account.OnBudget || account.Deleted || account.Closed {
			continue
		}
		fmt.Printf("%v\n", account)
	}

	for _, month := range months.Data.Months {
		fmt.Println(month.Month)
	}

	for _, i := range categories.Data.CategoryGroups {
		if i.Deleted || i.Hidden {
			continue
		}
		fmt.Println(i.Name)
		for _, s := range i.Categories {
			if i.Deleted || i.Hidden {
				continue
			}
			fmt.Printf("  %s\n", s.Name)
		}
	}
}
