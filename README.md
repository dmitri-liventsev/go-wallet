# Wallet API Service

This project provides a service for managing user balances, allowing users to retrieve, update, and manage their balance information.

## Features

- Create new transactions
- Retrieve current balance
- Manage balance information

## Manual Testing Instructions

When performing manual testing of the application, please keep in mind the following:

- The record in the `balances` table must be unique and must have the ID `0f31adad-bfb6-41d1-aeff-c110ca13cbfa`. Any other records will be ignored.
- The record in the `corrections` table must also be unique and must have the ID `3d8e7990-7a74-4613-9ed4-154dbba1d3b5`. Any other records will be ignored.

### Lost transaction example
```bash
curl -X POST http://localhost:8081/transaction \
     -H "Content-Type: application/json" \
     -H "Source-Type: game" \
     -d '{
           "state": "lost",
           "amount": "-50.00",
           "transactionId": "your-transaction-id"
         }'

```

### Win transaction example

```bash
curl -X POST http://localhost:8082/transaction \
     -H "Content-Type: application/json" \
     -H "Source-Type: game" \
     -d '{
           "state": "win",
           "amount": "100.00",
           "transactionId": "your-transaction-id"
         }'
```

## Prerequisites

- Docker
- Docker Compose

## Getting Started

### Clone the Repository

```bash
git clone git@github.com:dmitri-liventsev/go-wallet.git
cd go-wallet
```

## Start the Services
Use Docker Compose to start the services. This will start the Wallet API service listening on **localhost:80** and a PostgreSQL database.

```bash
docker-compose up -d
```

## API Documentation
The API follows the OpenAPI 3.0.3 specification. The OpenAPI yaml file can be found in the **gen/http** directory.

## API Endpoints
### Create Transaction
* **Endpoint: /transaction**
* **Method: POST**
* **Headers:**
  * Source-Type: Source type header (required, example: game, enum: game, server, payment)
* **Request Body:**
  * amount: Amount of the transaction (string, example: 10.15)
  * state: State of the transaction (string, enum: win, lost, example: win)
  * transactionId: Transaction ID (string, example: some generated identificator)
* **Responses:**
  * 202 Accepted: Transaction accepted
  * 400 Bad Request: Invalid input
  * 500 Internal Server Error: Internal server error

Example request body:

```json
{
  "amount": "10.15",
  "state": "win",
  "transactionId": "some generated identificator"
}
```

## Database Access
The current state of the balance can be viewed by connecting to the PostgreSQL database using the following credentials:

* **Host: localhost**
* **Port: 5432**
* **Username: postgres**
* **Password: password**
* **Database: txdb**

## Stopping the Services
To stop the services, use:

```bash
docker-compose down
```

## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Contact
For any questions or feedback, please write a letter to Santa Claus.
