# Wallet API Service

This project provides a service for managing user balances, allowing users to retrieve, update, and manage their balance information.

## Features

- Create new transactions
- Retrieve current balance
- Manage balance information

## Manual Testing

### Win transaction example

```bash
curl -X POST http://localhost:8081/user/1/transaction \
     -H "Content-Type: application/json" \
     -H "Source-Type: game" \
     -d '{
           "state": "win",
           "amount": "100.00",
           "transactionId": "your-transaction-id"
         }'
```

### Lose transaction example

```bash
curl -X POST http://localhost:8081/user/1/transaction \
     -H "Content-Type: application/json" \
     -H "Source-Type: game" \
     -d '{
           "state": "lose",
           "amount": "50.00",
           "transactionId": "your-transaction-id"
         }'
```

> Amounts are always positive. For `lose` transactions the server deducts the amount from the user's balance.

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

* **Endpoint:** `/user/{userId}/transaction`
* **Method:** `POST`

* **Headers:**
  * `Source-Type` *(required)*  
    Allowed values: `game`, `server`, `payment`

* **Path Parameters:**
  * `userId` *(uint64, required)*  
    Example: `1`

* **Request Body:**
  * `amount` *(string)*: Amount of the transaction (example: `10.15`)
  * `state` *(string)*: State of the transaction  
    Allowed values: `win`, `lose` (example: `win`)
  * `transactionId` *(string)*: Transaction ID (example: `some generated identifier`)

* **Responses:**
  * `200 OK`: Transaction successfully processed
  * `400 Bad Request`: Invalid input
  * `409 Conflict`: Duplicate transactionId (already processed)
  * `500 Internal Server Error`: Internal server error

Example request body:

```json
{
  "amount": "10.15",
  "state": "win",
  "transactionId": "some generated identificator"
}
```

### Get User Balance

* **Endpoint:** `/user/{userId}/balance`
* **Method:** `GET`

* **Path Parameters:**
  * `userId` *(uint64, required)*  
    Example: `1`

* **Responses:**
  * `200 OK`: Returns current user balance
  * `404 Not Found`: User or balance not found
  * `500 Internal Server Error`: Internal server error

* **Response Body:**
```json
{
  "userId": 1,
  "balance": "9.25"
}
````

## Database Access
The current state of the balance can be viewed by connecting to the PostgreSQL database using the following credentials:

* **Host: localhost**
* **Port: 5432**
* **Username: postgres**
* **Password: password**
* **Database: txdb**


## Concurrency Stress Test

A concurrency stress-test utility is available in the `concurrency` folder. It can be used to validate correct balance calculation and idempotent transaction processing under concurrent load. Please refer to the README inside that folder for usage instructions.

## Stopping the Services
To stop the services, use:

```bash
docker-compose down
```

## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Contact
For any questions or feedback, please write a letter to Santa Claus.
