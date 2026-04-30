
# Concurrency Test Utility

A small stress-test tool that fires concurrent transactions at the wallet application and then verifies that every balance is mathematically consistent.

## Overview

1. **Balance initialisation** — for each of the three test users (IDs 1, 2, 3) the tool ensures a balance row exists in the database. If no row is found it inserts one with a starting value of 1 000.00.
2. **Transaction flood** — `numOfTransactions` jobs are distributed across `numWorkers` goroutines. Each job picks a random user and a random amount between -10.00 and +10.00 and posts it to a random server.
3. **Wait for processing** — polls `transactions` every 500 ms until no rows remain in `new` or `locked` state (60-second deadline).
4. **Validation** — for each user computes `initial_balance + SUM(amount WHERE status = 'done')` and compares it to the current value in `balances`. Cancelled transactions are counted and shown but do not affect the expected balance.

## Prerequisites

Start the application with Docker before running the tool:

```bash
docker-compose up
```

## Running

```bash
go run main.go \
  --numOfTransactions=1000 \
  --numWorkers=20 \
  --connStr="user=postgres password=password dbname=txdb host=localhost port=5432 sslmode=disable" \
  --servers="localhost:8081,localhost:8082"
```

| Flag | Default | Description |
|---|---|---|
| `--numOfTransactions` | 1000 | Total number of transactions to send |
| `--numWorkers` | 20 | Number of concurrent goroutines |
| `--connStr` | *(local postgres)* | PostgreSQL connection string for validation queries |
| `--servers` | `localhost:8081,localhost:8082` | Comma-separated list of application instances |

## Example Output

```
waiting: 312 transactions still pending...
waiting: 58 transactions still pending...
User 1: balance=983.41 | computed=983.41 | cancelled=14 | OK
User 2: balance=1047.20 | computed=1047.20 | cancelled=9  | OK
User 3: balance=999.75 | computed=999.75 | cancelled=21 | OK
All balances match.
```

`balance` is the raw value from the `balances` table; `computed` is derived purely from `done` transactions. A `MISMATCH` line means a concurrency or atomicity bug was detected.