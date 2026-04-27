package tests

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"testing"
	"wallet/config"
	"wallet/gen/wallet"
	"wallet/transaction/interfaces"
	"wallet/transaction/internal/infrastructure/db"
)

var client *wallet.Client
var DB *gorm.DB
var _ = BeforeSuite(func(ctx context.Context) {
	DB = connectToTestDB(ctx)
	client = createTestClient()
})

var _ = BeforeEach(func() {
	db.Truncate(DB)
})

func createTestClient() *wallet.Client {
	txEndpoint := wallet.NewCreateTransactionEndpoint(interfaces.NewTxController(DB))
	balanceEndpoint := wallet.NewGetBalanceEndpoint(interfaces.NewTxController(DB))
	healthcheck := wallet.NewHealthcheckEndpoint(interfaces.NewTxController(DB))
	return wallet.NewClient(healthcheck, txEndpoint, balanceEndpoint)
}

func connectToTestDB(ctx context.Context) *gorm.DB {
	config.Load()

	//Initialize db
	dbConnection := db.NewConnection()
	gormdb, err := dbConnection.ConnectToSchema(ctx, "testing")

	if err != nil {
		panic("failed to switch schema: " + err.Error())
	}

	DeferCleanup(func() {
		sqlDB, _ := gormdb.DB()
		_ = sqlDB.Close()
	})

	return gormdb
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "wallet/tests")
}
