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
	MailHost             string `mapstructure:"MAIL_HOST"`
	MailPort             string `mapstructure:"MAIL_PORT"`
	MailUsername         string `mapstructure:"MAIL_USERNAME"`
	MailPassword         string `mapstructure:"MAIL_PASSWORD"`
	MailEncryption       string `mapstructure:"MAIL_ENCRYPTION"`
	MailFromAddress      string `mapstructure:"MAIL_FROM_ADDRESS"`
	MailFromName         string `mapstructure:"MAIL_FROM_NAME"`
	ImageKitPublicKey    string `mapstructure:"IMAGEKIT_PUBLIC_KEY"`
	ImageKitPrivateKey   string `mapstructure:"IMAGEKIT_PRIVATE_KEY"`
	ImageKitID           string `mapstructure:"IMAGEKIT_ID"`
	ImageKitURLEndpoint  string `mapstructure:"IMAGEKIT_URL_ENDPOINT"`
	GoogleClientID       string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret   string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI    string `mapstructure:"GOOGLE_REDIRECT_URI"`
	RedisAddr            string `mapstructure:"REDIS_ADDR"`
	RedisPassword        string `mapstructure:"REDIS_PASSWORD"`
	RedisDB              int    `mapstructure:"REDIS_DB"`
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
