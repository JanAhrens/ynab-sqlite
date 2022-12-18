CREATE TABLE IF NOT EXISTS server_knowledge (
    "endpoint" TEXT NOT NULL PRIMARY KEY,
    "value"    INTEGER
);

-- initialize endpoints with 0 unless they are already initialized
INSERT INTO server_knowledge(endpoint,value) VALUES
    ('categories',   0),
    ('accounts',     0),
    ('transactions', 0),
    ('payees',       0),
    ('months',       0)
ON CONFLICT(endpoint) DO NOTHING;

CREATE TABLE IF NOT EXISTS category_group (
    id      TEXT NOT NULL PRIMARY KEY,
    name    TEXT NOT NULL,
    hidden  INTEGER,
    deleted INTEGER
);

CREATE TABLE IF NOT EXISTS category (
    id                  TEXT NOT NULL PRIMARY KEY,
    category_group_id   TEXT NOT NULL,
    name                TEXT NOT NULL,
    note				TEXT,
    hidden              INTEGER,
    deleted             INTEGER,
    goal_type           TEXT,
    goal_creation_month TEXT,
    goal_target         TEXT,
    goal_target_month   TEXT
);

CREATE TABLE IF NOT EXISTS month (
    id             TEXT PRIMARY KEY NOT NULL,
    note           TEXT,
    income         INTEGER,
    budgeted       INTEGER,
    activity       INTEGER,
    to_be_budgeted INTEGER,
    age_of_money   INTEGER,
    deleted        INTEGER
);

CREATE TABLE IF NOT EXISTS category_month (
    month_id    TEXT,
    category_id TEXT,
    budgeted    INTEGER,
    activity    INTEGER,
    balance     INTEGER,
    PRIMARY KEY (month_id, category_id)
);

CREATE TABLE IF NOT EXISTS "transaction" (
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
);

CREATE TABLE IF NOT EXISTS subtransaction (
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
);

CREATE TABLE IF NOT EXISTS account (
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
);

CREATE TABLE IF NOT EXISTS payee (
    id 					TEXT NOT NULL PRIMARY KEY,
    name 				TEXT NOT NULL,
    transfer_account_id INTEGER,
    deleted				INTEGER
);