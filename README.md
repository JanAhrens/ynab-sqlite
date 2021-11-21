# ynab-sqlite

## Queries

> sqlite3 database.db "SELECT cg.name, c.name FROM category c LEFT JOIN category_group cg ON c.category_group_id = cg.id WHERE c.hidden <> 1 AND c.deleted <> 1 AND cg.hidden <> 1 AND cg.deleted <> 1;"
