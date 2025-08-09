package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost               string `mapstructure:"DB_HOST"`
	DBPort               string `mapstructure:"DB_PORT"`
	DBUsername           string `mapstructure:"DB_USERNAME"`
	DBPassword           string `mapstructure:"DB_PASSWORD"`
	DBDatabase           string `mapstructure:"DB_DATABASE"`
	DBSSLMode            string `mapstructure:"DB_SSLMODE"`
	AppURL               string `mapstructure:"APP_URL"`
	JWTSecret            string `mapstructure:"JWT_SECRET"`
	MidtransServerKey    string `mapstructure:"MIDTRANS_SERVER_KEY"`
	MidtransIsProduction bool   `mapstructure:"MIDTRANS_IS_PRODUCTION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
