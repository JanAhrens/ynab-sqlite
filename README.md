# ynab-sqlite

Exports all data from a [YNAB](https://youneedadbudget.com) budget and stores it in a local [SQLite](https://sqlite.org/) database.
When the program gets executed multiple times, only the changed data will be downloaded.

## Getting started

Prerequisites: You need a [YNAB](https://youneedadbudget.com) account. A trial account also works.

1. Create a Personal Access Token in the [Developer Settings](https://app.youneedabudget.com/settings/developer)

2. Copy the access token and set the YNAB_API_KEY environment variable

	export YNAB_API_KEY=722XXXXXXXXXXbbe4436302XXXXXXdc734XX35bd21cXXXXX2d4b5fafb3c06dXX

3. Run the program

     go run .

4. Explore the data using the sqlite3 cli (see queries section)


## Queries

```
$ sqlite3 --header --column database.db
```

```sql
SELECT
	cg.name, c.name
FROM category c LEFT JOIN category_group cg ON c.category_group_id = cg.id
WHERE c.hidden <> 1 AND c.deleted <> 1 AND cg.hidden <> 1 AND cg.deleted <> 1;
```

### Net worth

```sql
SELECT
	strftime("%Y-%m-01", "date"),
	CAST(SUM(amount) AS REAL)/1000
FROM "transaction"
GROUP BY strftime("%Y-%m-01", "date")
ORDER BY "date";
```