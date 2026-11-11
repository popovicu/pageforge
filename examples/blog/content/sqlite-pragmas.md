---
title: "Three SQLite pragmas to set on every new database"
date: 2026-04-08
description: "WAL, foreign keys, and busy_timeout. Set them, walk away."
tags: [sqlite, databases]
author: "Jules"
reading_minutes: 3
---

Every SQLite database I open in earnest gets the same three pragmas in the
first second of its life:

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
```

## `journal_mode = WAL`

Write-ahead logging lets readers and writers proceed concurrently. The default
rollback journal blocks readers during a write, which is fine for a CLI tool
and a disaster for anything serving requests.

WAL is a per-database setting that survives reopen — set it once at creation
and forget it.

## `foreign_keys = ON`

This one always surprises people. SQLite parses foreign-key constraints since
forever, but it does not *enforce* them unless you turn the pragma on.

The default is off for backwards compatibility with databases that predate
enforcement (2010!). If you are starting today, set it.

## `busy_timeout = 5000`

When a write transaction is holding the database, other writers get
`SQLITE_BUSY` immediately. With a busy timeout, they wait up to N milliseconds
for the lock instead of failing.

Five seconds is usually plenty. The number is forgettable; the habit of
setting *some* timeout is not.

## Why a function

I keep these three lines in a `prepare(db)` helper that runs on every open:

```go
func prepare(db *sql.DB) error {
    pragmas := []string{
        "PRAGMA journal_mode = WAL",
        "PRAGMA foreign_keys = ON",
        "PRAGMA busy_timeout = 5000",
    }
    for _, p := range pragmas {
        if _, err := db.Exec(p); err != nil {
            return fmt.Errorf("%s: %w", p, err)
        }
    }
    return nil
}
```

The function is dumb on purpose. It is also the difference between a database
that behaves and one that does not.
