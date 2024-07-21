package service_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"testing"
	"wallet/config"
	"wallet/transaction/internal/infrastructure/db"
)

var DB *gorm.DB
var _ = BeforeSuite(func(ctx context.Context) {
	config.Load()

	//Initialize db
	dbConnection := db.NewConnection()
	var err error
	DB, err = dbConnection.ConnectToSchema(ctx, "testing")
	if err != nil {
		panic("failed to switch schema: " + err.Error())
	}

	DeferCleanup(func() {
		sqlDB, _ := DB.DB()
		_ = sqlDB.Close()
	})
})

var _ = BeforeEach(func() {
	db.Truncate(DB)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Balance Service")
}
