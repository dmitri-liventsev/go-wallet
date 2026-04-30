
# Concurrency Test Utility

A small stress-test tool that fires concurrent transactions at the wallet application and then verifies that every balance is mathematically consistent.

## Overview

1. **Fetch initial balances** — calls `GET /user/{id}/balance` for each of the three test users (IDs 1, 2, 3) and records the starting balance.
2. **Transaction flood** — `numOfTransactions` jobs are distributed across `numWorkers` goroutines. Each job picks a random user and a random amount between -10.00 and +10.00 and posts it to a random server.
3. **Wait for processing** — sleeps for `waitSeconds` (default 60 s) to let the application finish processing all queued transactions.
4. **Validation** — calls `GET /user/{id}/balance` again for each user and prints the initial balance, final balance, and net difference.

## Prerequisites

Start the application with Docker before running the tool:

```bash
docker-compose up
```

## Known Limitation: Transaction Order

Because transactions are dispatched concurrently across multiple goroutines and servers, **the order in which the server receives and processes them is non-deterministic**. The `computed` balance is calculated locally in generation order, which will differ from the server's actual processing order.

As a result, when a user's balance is close to zero, we cannot reliably predict which transactions the server will cancel — the server may cancel a different set than the ones we skipped locally, causing a `MISMATCH` that is **not a concurrency bug**.

> **Recommendation:** set the initial balance high enough that no transaction can drive it negative during the test run. A value of **10 000.00** is sufficient for the default 1 000 transactions with amounts up to ±10.00.

## Running

```bash
go run main.go \
  --numOfTransactions=1000 \
  --numWorkers=20 \
  --servers="localhost:8081,localhost:8082"
```

| Flag | Default | Description |
|---|---|---|
| `--numOfTransactions` | 1000 | Total number of transactions to send |
| `--numWorkers` | 20 | Number of concurrent goroutines |
| `--servers` | `localhost:8081,localhost:8082` | Comma-separated list of application instances |

The wait time before reading final balances is derived automatically: **1 second per 20 transactions** (`numOfTransactions / 20`).

## Example Output

```
waiting 60 seconds for transaction processing...
User 1: initial=1000.00 | expected=983.41 | final=983.41 | OK
User 2: initial=1000.00 | expected=1047.20 | final=1047.20 | OK
User 3: initial=1000.00 | expected=999.75 | final=999.75 | OK
All balances match.
```

`expected` is computed locally as transactions are generated: losses that would drive the balance below zero are skipped, mirroring the server-side cancellation logic. Both `expected` and `final` are compared at the end — a `MISMATCH` means a concurrency or atomicity bug was detected.