package db

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"goa.design/clue/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"wallet/transaction/internal/domain/entities"
)

// DbConnection holds configuration details for connecting to a database.
type DbConnection struct {
	Username string
	Password string
	DbName   string
	Host     string
	Port     string
}

// Connect establishes a connection to the PostgreSQL database using the config,
// performs necessary setup, and returns a *gorm.DB instance.
func (dbConn DbConnection) Connect(ctx context.Context) *gorm.DB {
	db := dbConn.doConnection(ctx)
	err := RunAutoMigrations(db)
	if err != nil {
		log.Fatal(ctx, err)
	}

	return db
}

func (dbConn DbConnection) doConnection(ctx context.Context) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		dbConn.Host,
		dbConn.Port,
		dbConn.Username,
		dbConn.DbName,
		dbConn.Password,
	)

	log.Printf(ctx, "Connecting to database with DSN: %s", dsn)

	// Connect to PostgreSQL database using GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Fatal(ctx, err)
	}

	return db
}

// ConnectToSchema establishes a database connection, switches to the specified schema,
// and runs auto-migrations, returning the *gorm.DB instance.
func (dbConn DbConnection) ConnectToSchema(ctx context.Context, schemaName string) (*gorm.DB, error) {
	db := dbConn.doConnection(ctx)
	err := SwitchSchema(db, schemaName)
	if err != nil {
		return nil, err
	}
	err = RunAutoMigrations(db)

	if err != nil {
		return nil, err
	}

	return db, nil
}

// RunAutoMigrations performs auto-migrations for the Transaction, Correction, and Balance entities.
func RunAutoMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(&entities.Transaction{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entities.Correction{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entities.Balance{})
	if err != nil {
		return err
	}

	return nil
}

// SwitchSchema creates the specified schema if it doesn't exist and sets it as the search path for the database connection.
func SwitchSchema(db *gorm.DB, schemaName string) error {
	err := CreateSchemaIfNotExists(db, schemaName)
	if err != nil {
		return err
	}
	query := "SET search_path TO " + schemaName

	return db.Exec(query).Error

}

// CreateSchemaIfNotExists creates the specified schema if it does not already exist, ignoring duplicate schema errors.
func CreateSchemaIfNotExists(db *gorm.DB, schemaName string) error {
	query := "CREATE SCHEMA IF NOT EXISTS " + schemaName
	err := db.Exec(query).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}

// NewConnection returns new connections instance.
func NewConnection() DbConnection {
	return DbConnection{
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
		DbName:   viper.GetString("db.dbname"),
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
	}
}
