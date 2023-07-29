package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	BaseURL             string
	TimerForSSH         time.Duration
	HttpPort            string
	SshPort             string
	TokenSecretKey      string
	MongoDB             MongoDB
	Github              Github
	Google              Google
	AuthCookieName      string
	EncryptedPrivateKey string
	EncryptSecretKey    string
	LocationInfoKey     string
}

type Github struct {
	ClientID    string
	SecretKey   string
	RedirectURI string
}

type Google struct {
	Conf *oauth2.Config
}

type MongoDB struct {
	Url      string
	Username string
	Password string
}

type LimiterConfig struct {
	RPS   int
	Burst int
	TTL   time.Duration
}

func Load() Config {
	godotenv.Load()

	conf := viper.New()
	conf.AutomaticEnv()

	return Config{
		BaseURL:     conf.GetString("BASE_URL"),
		TimerForSSH: conf.GetDuration("TIMER_FOR_SSH"),
		HttpPort:    conf.GetString("HTTP_PORT"),
		SshPort:     conf.GetString("SSH_PORT"),
		MongoDB: MongoDB{
			Url:      conf.GetString("MONGODB_URL"),
			Username: conf.GetString("MONGODB_USERNAME"),
			Password: conf.GetString("MONGODB_PASSWORD"),
		},
		Github: Github{
			ClientID:    conf.GetString("GITHUB_CLIENT_ID"),
			SecretKey:   conf.GetString("GITHUB_SECRET_KEY"),
			RedirectURI: conf.GetString("GITHUB_REDIRECT_URI"),
		},
		Google: Google{
			Conf: &oauth2.Config{
				ClientID:     conf.GetString("GOOGLE_CLIENT_ID"),
				ClientSecret: conf.GetString("GOOGLE_SECRET_KEY"),
				RedirectURL:  conf.GetString("GOOGLE_REDIRECT_URI"),
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
				Endpoint: google.Endpoint,
			},
		},
		AuthCookieName:      conf.GetString("AUTH_COOKIE_NAME"),
		TokenSecretKey:      conf.GetString("TOKEN_SECRET_KEY"),
		EncryptedPrivateKey: conf.GetString("ENCRYPTED_PRIVATE_KEY"),
		EncryptSecretKey:    conf.GetString("ENCRYPT_SECRET_KEY"),
		LocationInfoKey:     conf.GetString("LOCATION_INFO_KEY"),
	}
}
