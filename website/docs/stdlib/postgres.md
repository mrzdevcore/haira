---
title: "Postgres"
description: "PostgreSQL database client for queries, transactions, and connection management."
---

# Postgres

## Connection

```haira
import "postgres"

db, err = postgres.connect(env("DATABASE_URL"))
if err != nil {
    io.println("Connection failed: ${err}")
}
```

## Queries

```haira
// Select
rows, err = postgres.query("SELECT name, email FROM users WHERE active = $1", true)
for row in rows {
    io.println("${row["name"]}: ${row["email"]}")
}

// Insert
_, err = postgres.exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Alice", "alice@example.com")

// Update
_, err = postgres.exec("UPDATE users SET active = $1 WHERE id = $2", false, 42)
```

## In Tools

A common pattern — tools that query databases:

```haira
tool lookup_user(email: string) -> string {
    """Look up a user by their email address in the database"""
    rows, err = postgres.query(
        "SELECT name, plan, created_at FROM users WHERE email = $1",
        email
    )
    if err != nil { return "Database error." }
    if len(rows) == 0 { return "User not found." }

    user = rows[0]
    return "Name: ${user["name"]}, Plan: ${user["plan"]}"
}
```

## Store Backend

Postgres can serve as the session/state store backend for workflows, replacing the default SQLite:

```haira
import "postgres"

// When postgres is imported, it registers as a store backend
// Sessions and workflow state persist in Postgres
```

## Extended Operations

```haira
// Upsert
postgres.upsert("users", {
    "email": "alice@example.com",
    "name": "Alice",
    "updated_at": time.now()
}, "email")

// Transaction
postgres.transaction(fn(tx) {
    tx.exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, from_id)
    tx.exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, to_id)
})
```
