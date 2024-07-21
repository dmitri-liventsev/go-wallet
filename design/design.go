package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("Wallet", func() {
	Title("Wallet API")
	Description("Service for managing user balances, allowing users to retrieve, update, and manage their balance information.")

	Server("Wallet", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("transaction", func() {
	Description("The transaction service")

	HTTP(func() {
		Path("/transaction")
	})

	Method("create", func() {
		Description("Create a new transaction")

		Payload(func() {
			Attribute("state", String, "State of the transaction", func() {
				Enum("win", "lost")
				Example("win")
			})
			Attribute("amount", String, "Amount of the transaction", func() {
				Example("10.15")
			})
			Attribute("transactionId", String, "Transaction ID", func() {
				Example("some generated identificator")
			})
			Attribute("sourceType", String, "Source type header", func() {
				Enum("game", "server", "payment")
				Example("game")
			})
			Required("state", "amount", "transactionId", "sourceType")
		})

		Result(Empty)

		HTTP(func() {
			POST("/")
			Header("sourceType:Source-Type")
			Response(StatusAccepted)
			Response(StatusBadRequest, func() {
				Description("Invalid input")
			})
			Response(StatusInternalServerError, func() {
				Description("Internal server error")
			})
		})
	})
})
