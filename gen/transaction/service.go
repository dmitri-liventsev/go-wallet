// Code generated by goa v3.17.2, DO NOT EDIT.
//
// transaction service
//
// Command:
// $ goa gen wallet/design

package transaction

import (
	"context"
)

// The transaction service
type Service interface {
	// Check if the service is running
	Healthcheck(context.Context) (res *HealthcheckResult, err error)
	// Create a new transaction
	Create(context.Context, *CreatePayload) (err error)
}

// APIName is the name of the API as defined in the design.
const APIName = "Wallet"

// APIVersion is the version of the API as defined in the design.
const APIVersion = "0.0.1"

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "transaction"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"healthcheck", "create"}

// CreatePayload is the payload type of the transaction service create method.
type CreatePayload struct {
	// State of the transaction
	State string
	// Amount of the transaction
	Amount string
	// Transaction ID
	TransactionID string
	// Source type header
	SourceType string
}

// HealthcheckResult is the result type of the transaction service healthcheck
// method.
type HealthcheckResult struct {
	// Service status
	Status string
}
