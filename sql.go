package main

import (
	"database/sql"
	"log"
)

var tables = []string{
	`CREATE TABLE IF NOT EXISTS server_knowledge (
		"endpoint" TEXT NOT NULL PRIMARY KEY,
		"value"    INTEGER
	);`,

	// initialize endpoints with null values unless they are already initialized
	`INSERT INTO server_knowledge(endpoint,value) VALUES
		('categories',   0),
		('accounts',     0),
		('transactions', 0)
		ON CONFLICT(endpoint) DO NOTHING;
	`,

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

	`CREATE TABLE IF NOT EXISTS "transaction" (
		id                      TEXT NOT NULL PRIMARY KEY,
		date                    TEXT,
		amount                  INTEGER,
		memo                    TEXT,
		cleared                 TEXT,
		approved                INTEGER,
		flag_color              TEXT,
		account_id              TEXT,
		payee_id                TEXT,
		category_id             TEXT,
		transfer_account_id     TEXT,
		transfer_transaction_id TEXT,
		matched_transaction_id  TEXT,
		import_id               TEXT,
		deleted                 INTEGER,
		account_name            TEXT,
		payee_name              TEXT,
		category_name           TEXT
	);`,

	`CREATE TABLE IF NOT EXISTS subtransaction (
		id                      TEXT NOT NULL PRIMARY KEY,
		transaction_id          TEXT,
		amount                  INTEGER,
		memo                    TEXT,
		payee_id                TEXT,
		payee_name              TEXT,
		category_id             TEXT,
		category_name           TEXT,
		transfer_account_id     TEXT,
		transfer_transaction_id TEXT,
		deleted                 INTEGER
	);`,

	`CREATE TABLE IF NOT EXISTS account (
		id                     TEXT NOT NULL PRIMARY KEY,
		name                   TEXT,
		type                   TEXT,
		on_budget              INTEGER,
		closed                 INTEGER,
		note                   TEXT,
		cleared_balance        INTEGER,
		uncleared_balane       INTEGER,
		transfer_payee_id      TEXT,
		direct_import_linked   INTEGER,
		direct_import_in_error INTEGER,
		deleted                INTEGER
	);`,
}

func createTables(db *sql.DB) {
	for _, sql := range tables {
		statement, err := db.Prepare(sql)
		if err != nil {
			log.Fatal(err.Error())
		}
		statement.Exec()
	}
}

func loadServerKnowledge(db *sql.DB) map[string]int {
	var serverKnowledge = make(map[string]int)
	res, err := db.Query("SELECT endpoint, value FROM server_knowledge")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer res.Close()
	var (
		endpoint string
		value    int
	)
	for res.Next() {
		err := res.Scan(&endpoint, &value)
		if err != nil {
			log.Fatal(err.Error())
		}
		serverKnowledge[endpoint] = value
	}
	if err := res.Err(); err != nil {
		log.Fatal(err.Error())
	}
	return serverKnowledge
}

func updateServerKnowledge(db *sql.DB, sql string, value int) {
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(value)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func updateCategories(categories Categories, db *sql.DB) {
	serverKnowledgeSql := `INSERT INTO server_knowledge(endpoint, value) VALUES('categories', ?) ON CONFLICT(endpoint) DO UPDATE SET value=excluded.value;`
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
		if err != nil {
			log.Fatal(err.Error())
		}

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

	updateServerKnowledge(db, serverKnowledgeSql, categories.Data.ServerKnowledge)
}

func updateTransactions(transactions Transactions, db *sql.DB) {
	insertTransactionSql := `
    INSERT INTO "transaction" (
		id, date, amount, memo, cleared, approved,
		flag_color, account_id, payee_id, category_id,
		transfer_account_id, transfer_transaction_id,
		matched_transaction_id, import_id, deleted,
		account_name, payee_name, category_name
    ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
		date=excluded.date, amount=excluded.amount, memo=excluded.memo,
		cleared=excluded.cleared, approved=excluded.approved,
		flag_color=excluded.flag_color, account_id=excluded.account_id,
		payee_id=excluded.payee_id, category_id=excluded.category_id,
		transfer_account_id=excluded.transfer_account_id,
		transfer_transaction_id=excluded.transfer_transaction_id,
		matched_transaction_id=excluded.matched_transaction_id,
		import_id=excluded.import_id, deleted=excluded.deleted,
		account_name=excluded.account_name, payee_name=excluded.payee_name,
		category_name=excluded.category_name;`
	insertSubtransactionSql := `
    INSERT INTO "subtransaction" (
		id, transaction_id, amount, memo, payee_id, category_id,
		category_name, transfer_account_id, transfer_transaction_id,
		deleted
    ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
		transaction_id=excluded.transaction_id, amount=excluded.amount, memo=excluded.memo,
		payee_id=excluded.payee_id, category_id=excluded.category_id,
		category_name=excluded.category_name, transfer_account_id=excluded.transfer_account_id,
		transfer_transaction_id=excluded.transfer_transaction_id, deleted=excluded.deleted;
	`
	serverKnowledgeSql := `INSERT INTO server_knowledge(endpoint, value) VALUES('transactions', ?) ON CONFLICT(endpoint) DO UPDATE SET value=excluded.value;`

	for _, t := range transactions.Data.Transactions {
		statement, err := db.Prepare(insertTransactionSql)
		if err != nil {
			log.Fatal(err.Error())
		}
		_, err = statement.Exec(t.Id, t.Date, t.Amount, t.Memo, t.Cleared, t.Approved,
			t.FlagColor, t.AccountId, t.PayeeId, t.CategoryId,
			t.TransferAccountId, t.TransferTransactionId,
			t.MatchedTransactionId, t.ImportId, t.Deleted,
			t.AccountName, t.PayeeName, t.CategoryName)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, st := range t.Subtransactions {
			statement, err := db.Prepare(insertSubtransactionSql)
			if err != nil {
				log.Fatal(err.Error())
			}
			_, err = statement.Exec(st.Id, st.TransactionId, st.Amount, st.Memo,
				st.PayeeId, st.CategoryId, st.CategoryName, st.TransferAccountId,
				st.TransferTransactionId, st.Deleted)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

	updateServerKnowledge(db, serverKnowledgeSql, transactions.Data.ServerKnowledge)
}

func updateAccounts(accounts Accounts, db *sql.DB) {
	serverKnowledgeSql := `INSERT INTO server_knowledge(endpoint, value) VALUES('accounts', ?) ON CONFLICT(endpoint) DO UPDATE SET value=excluded.value;`
	insertAccountSql := `
		INSERT INTO account (
			id, name, type, on_budget, closed, note, cleared_balance,
			uncleared_balane, transfer_payee_id, direct_import_linked,
			direct_import_in_error, deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			type=excluded.type,
			on_budget=excluded.on_budget,
			closed=excluded.closed,
			note=excluded.note,
			cleared_balance=excluded.cleared_balance,
			uncleared_balane=excluded.uncleared_balane,
			transfer_payee_id=excluded.transfer_payee_id,
			direct_import_linked=excluded.direct_import_linked,
			direct_import_in_error=excluded.direct_import_in_error,
			deleted=excluded.deleted;
	`

	for _, account := range accounts.Data.Accounts {
		statement, err := db.Prepare(insertAccountSql)
		if err != nil {
			log.Fatal(err.Error())
		}
		_, err = statement.Exec(
			account.Id,
			account.Name,
			account.Type,
			account.OnBudget,
			account.Closed,
			account.Note,
			account.ClearedBalance,
			account.UnclearedBalance,
			account.TransferPayeeId,
			account.DirectImportLinked,
			account.DirectImportInError,
			account.Deleted)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	updateServerKnowledge(db, serverKnowledgeSql, accounts.Data.ServerKnowledge)
}
