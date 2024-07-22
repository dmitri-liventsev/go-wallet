# Wallet API Service

This project provides a service for managing user balances, allowing users to retrieve, update, and manage their balance information.

## Features

- Create new transactions
- Retrieve current balance
- Manage balance information

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
