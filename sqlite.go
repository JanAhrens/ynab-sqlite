package main

import (
	"context"
	"database/sql"
	_ "embed"
)

func NewSqliteService(db *sql.DB) sqliteService {
	return sqliteService{db: db}
}

type sqliteService struct {
	db *sql.DB
}

func (sql *sqliteService) Transaction(transactionFunction func(context.Context, *sql.Tx) error) error {
	ctx := context.Background()
	tx, err := sql.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = transactionFunction(ctx, tx)

	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return err
}

//go:embed schema.sql
var tables string

func (service sqliteService) CreateTables() error {
	_, err := service.db.Exec(tables)
	return err
}

func loadServerKnowledge(ctx context.Context, tx *sql.Tx) (map[string]int, error) {
	var serverKnowledge = make(map[string]int)
	res, err := tx.QueryContext(ctx, "SELECT endpoint, value FROM server_knowledge")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var (
		endpoint string
		value    int
	)
	for res.Next() {
		if err := res.Scan(&endpoint, &value); err != nil {
			return nil, err
		}
		serverKnowledge[endpoint] = value
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return serverKnowledge, nil
}

func updateServerKnowledge(ctx context.Context, tx *sql.Tx, endpoint string, value int) error {
	updateSQL := `
	INSERT INTO server_knowledge(endpoint, value)
	VALUES(?, ?)
	ON CONFLICT(endpoint)
	DO UPDATE SET value=excluded.value;
	`
	statement, err := tx.Prepare(updateSQL)
	if err != nil {
		return err
	}
	if _, err := statement.ExecContext(ctx, endpoint, value); err != nil {
		return err
	}
	return nil
}

func updateCategories(ctx context.Context, categories Categories, tx *sql.Tx) error {
	insertCategoryGroupSQL := `
    INSERT INTO category_group (
      id, name, hidden, deleted
    ) VALUES(?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      name=excluded.name, hidden=excluded.hidden, deleted=excluded.hidden;
  `
	insertCategorySQL := `
    INSERT INTO category (
      id, name, note, category_group_id, hidden, deleted, goal_type,
      goal_creation_month, goal_target, goal_target_month
    ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      name=excluded.name, note=excluded.note, category_group_id=excluded.category_group_id,
      hidden=excluded.name, deleted=excluded.deleted,
      goal_type=excluded.goal_type,
      goal_creation_month=excluded.goal_creation_month,
      goal_target=excluded.goal_target,
      goal_target_month=excluded.goal_target_month;
  `
	for _, group := range categories.Data.CategoryGroups {
		statement, err := tx.Prepare(insertCategoryGroupSQL)
		if err != nil {
			return err
		}
		_, err = statement.ExecContext(ctx, group.ID, group.Name, group.Hidden, group.Deleted)
		if err != nil {
			return err
		}

		for _, category := range group.Categories {
			statement, err := tx.Prepare(insertCategorySQL)
			if err != nil {
				return err
			}
			_, err = statement.ExecContext(ctx, category.ID, category.Name, category.Note, category.CategoryGroupID, category.Hidden, category.Deleted, category.GoalType, category.GoalCreationMonth, category.GoalTarget, category.GoalTargetMonth)
			if err != nil {
				return err
			}
		}
	}

	return updateServerKnowledge(ctx, tx, "categories", categories.Data.ServerKnowledge)
}

func updateTransactions(ctx context.Context, transactions Transactions, tx *sql.Tx) error {
	insertTransactionSQL := `
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
	insertSubtransactionSQL := `
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

	for _, t := range transactions.Data.Transactions {
		statement, err := tx.Prepare(insertTransactionSQL)
		if err != nil {
			return err
		}
		_, err = statement.ExecContext(ctx, t.ID, t.Date, t.Amount, t.Memo, t.Cleared, t.Approved,
			t.FlagColor, t.AccountID, t.PayeeID, t.CategoryID,
			t.TransferAccountID, t.TransferTransactionID,
			t.MatchedTransactionID, t.ImportID, t.Deleted,
			t.AccountName, t.PayeeName, t.CategoryName)
		if err != nil {
			return err
		}
		for _, st := range t.Subtransactions {
			statement, err := tx.Prepare(insertSubtransactionSQL)
			if err != nil {
				return nil
			}
			_, err = statement.ExecContext(ctx, st.ID, st.TransactionID, st.Amount, st.Memo,
				st.PayeeID, st.CategoryID, st.CategoryName, st.TransferAccountID,
				st.TransferTransactionID, st.Deleted)
			if err != nil {
				return nil
			}
		}
	}

	return updateServerKnowledge(ctx, tx, "transactions", transactions.Data.ServerKnowledge)
}

func updateAccounts(ctx context.Context, accounts Accounts, tx *sql.Tx) error {
	insertAccountSQL := `
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
		statement, err := tx.Prepare(insertAccountSQL)
		if err != nil {
			return err
		}
		_, err = statement.ExecContext(ctx,
			account.ID,
			account.Name,
			account.Type,
			account.OnBudget,
			account.Closed,
			account.Note,
			account.ClearedBalance,
			account.UnclearedBalance,
			account.TransferPayeeID,
			account.DirectImportLinked,
			account.DirectImportInError,
			account.Deleted)
		if err != nil {
			return err
		}
	}

	return updateServerKnowledge(ctx, tx, "accounts", accounts.Data.ServerKnowledge)
}

func updateCategoryMonth(ctx context.Context, categoryMonth CategoryMonth, tx *sql.Tx) error {
	insertCategortMonthSQL := `
		INSERT INTO category_month (
			month_id, category_id, budgeted, activity, balance
		) VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(month_id, category_id) DO UPDATE SET
			budgeted=excluded.budgeted,
			activity=excluded.activity,
			balance=excluded.balance;
	`
	statement, err := tx.Prepare(insertCategortMonthSQL)
	if err != nil {
		return err
	}

	for _, category := range categoryMonth.Data.Month.Categories {
		_, err = statement.ExecContext(ctx,
			categoryMonth.Data.Month.Month.Month,
			category.ID,
			category.Budgeted,
			category.Activity,
			category.Balance,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateMonthServerKnowledge(ctx context.Context, months Months, tx *sql.Tx) error {
	return updateServerKnowledge(ctx, tx, "months", months.Data.ServerKnowledge)
}

func updateMonth(ctx context.Context, month Month, tx *sql.Tx) error {
	insertMonthSQL := `
		INSERT INTO month (
			id, note, income, budgeted, activity, to_be_budgeted, age_of_money, deleted
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			note=excluded.note,
			income=excluded.income,
			budgeted=excluded.budgeted,
			activity=excluded.activity,
			to_be_budgeted=excluded.to_be_budgeted,
			age_of_money=excluded.age_of_money,
			deleted=excluded.deleted
	;`
	statement, err := tx.Prepare(insertMonthSQL)
	if err != nil {
		return err
	}
	_, err = statement.ExecContext(ctx,
		month.Month,
		month.Note,
		month.Income,
		month.Budgeted,
		month.Activity,
		month.ToBeBudgeted,
		month.AgeOfMoney,
		month.Deleted)
	return err
}

func updatePayees(ctx context.Context, payees Payees, tx *sql.Tx) error {
	insertPayeeSQL := `INSERT INTO payee (
		id, name, transfer_account_id, deleted
	) VALUES(?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name=excluded.name,
		transfer_account_id=excluded.transfer_account_id,
		deleted=excluded.deleted
	;`

	for _, payee := range payees.Data.Payees {
		statement, err := tx.Prepare(insertPayeeSQL)
		if err != nil {
			return err
		}
		_, err = statement.ExecContext(ctx, payee.ID, payee.Name, payee.TransferAccountID, payee.Deleted)
		if err != nil {
			return err
		}
	}

	return updateServerKnowledge(ctx, tx, "payees", payees.Data.ServerKnowledge)
}
