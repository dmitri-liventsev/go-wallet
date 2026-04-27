package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("Wallet", func() {
	Title("Wallet API")
	Description("Service for processing user balance transactions")

	Server("wallet", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("wallet", func() {
	Description("Wallet service")

	// ---------------------------
	// Healthcheck
	// ---------------------------
	Method("healthcheck", func() {
		Description("Health check")

		Result(func() {
			Attribute("status", String)
			Required("status")
		})

		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})
	})

	// ---------------------------
	// Create Transaction
	// ---------------------------
	Method("createTransaction", func() {
		Description("Process transaction for user")

		Payload(func() {
			Attribute("userId", UInt64, "User ID", func() {
				Minimum(1)
				Example(1)
			})

			Attribute("state", String, func() {
				Enum("win", "lose")
				Example("win")
			})

			Attribute("amount", String, func() {
				Example("10.15")
			})

			Attribute("transactionId", String, func() {
				Example("tx-123")
			})

			Attribute("sourceType", String, func() {
				Enum("game", "server", "payment")
				Example("game")
			})

			Required("userId", "state", "amount", "transactionId", "sourceType")
		})

		// 🔥 важно: пустой response
		Result(Empty)

		HTTP(func() {
			POST("/user/{userId}/transaction")

			Param("userId")

			Header("sourceType:Source-Type")

			Response(StatusOK)
			Response(StatusBadRequest)
			Response(StatusConflict)
			Response(StatusInternalServerError)
		})
	})

	// ---------------------------
	// Get Balance
	// ---------------------------
	Method("getBalance", func() {
		Description("Get user balance")

		Payload(func() {
			Attribute("userId", UInt64, "User ID", func() {
				Minimum(1)
				Example(1)
			})
			Required("userId")
		})

		Result(func() {
			Attribute("userId", UInt64)
			Attribute("balance", String, func() {
				Example("9.25")
			})

			Required("userId", "balance")
		})

		HTTP(func() {
			GET("/user/{userId}/balance")

			Param("userId")

			Response(StatusOK)
			Response(StatusNotFound)
			Response(StatusInternalServerError)
		})
	})
})
