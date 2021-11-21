# ynab-sqlite

## Queries

> sqlite3 database.db "SELECT cg.name, c.name FROM category c LEFT JOIN category_group cg ON c.category_group_id = cg.id WHERE c.hidden <> 1 AND c.deleted <> 1 AND cg.hidden <> 1 AND cg.deleted <> 1;"

### Net worth

```sql
SELECT
	strftime("%Y-%m-01", "date"),
	SUM(CAST(amount AS REAL)/1000)
FROM "transaction"
GROUP BY strftime("%Y-%m-01", "date")
ORDER BY "date";
```
