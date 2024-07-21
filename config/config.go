package config

import (
	"github.com/spf13/viper"
	"os"
)

// Load initializes Viper to automatically read environment variables and sets default database configuration values.
func Load() {
	viper.AutomaticEnv()

	viper.Set("db.username", getEnv("DB_USERNAME", "postgres"))
	viper.Set("db.password", getEnv("DB_PASSWORD", "password"))
	viper.Set("db.dbname", getEnv("DB_NAME", "txdb"))
	viper.Set("db.host", getEnv("DB_HOST", "db"))
	viper.Set("db.port", getEnv("DB_PORT", "5432"))
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
