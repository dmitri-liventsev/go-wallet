
# Concurrency Test Utility

This is a small test utility designed to verify the concurrency handling of a parent application. It simulates multiple transactions being processed simultaneously to ensure that the application manages concurrent requests properly.

## Overview

The utility performs the following steps:
1. **Initial Balance Check**: Fetches the current balance from the application.
2. **Transaction Generation**: Generates a list of transactions to be sent to the application.
3. **Expected Balance Calculation**: Calculates the expected balance after all transactions, assuming no transaction cancellation due to a negative balance.
4. **Execution**: Sends the generated transactions to the application using multiple concurrent workers.
5. **Comparison**: Compares the expected balance with the actual balance after all transactions have been processed, displaying both results.

## Prerequisites

Before using this utility, ensure that the application is running on the host machine using Docker. You need to start the application containers with `docker-compose up`. Once the application is running, you can build and run this utility from the host machine.

## Configuration

You can configure the utility by modifying the following constants in the source code:

- **`numWorkers`**: The number of concurrent workers that will send requests to the servers.
- **`numOfTransactions`**: The number of transactions that will be sent to the servers.
- **`servers`**: A list of host instances where the application is running.

## Important Notes

- **Correction Process**: The application has a built-in correction mechanism that runs every 10 minutes. This utility does not track these corrections, so the expected and actual balances may differ if a correction occurs during the test run.

- **Transaction Anomalies**: If there is a discrepancy between the expected and actual balances, verify the transactions. Look for transactions with `source_type = Internal` or transactions marked with `status = cancelled`, as these could indicate where the correction process has intervened.

## Example Output

After running the utility, you might see output like this:

```json5
Expected Balance: 157924.00
Final Balance in Database: 157924.00
```

This output indicates a slight difference between the expected and actual balances, possibly due to the application's correction process.

## Conclusion

This utility is a helpful tool for testing the concurrency capabilities of your application. By simulating multiple simultaneous transactions, it helps ensure that your application can handle concurrent requests without issues.
