// mockgen -source=correction.go -destination=internal/mocks/mock_correction_processor.go -package=mocks CorrectionProcessor
package workers_test

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
	DB = connectToTestDB(ctx)
})

var _ = BeforeEach(func() {
	db.Truncate(DB)
})

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
