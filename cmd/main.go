package main

import (
	"encoding/base64"

	"github.com/SaidovZohid/swiftsend.it/api"
	"github.com/SaidovZohid/swiftsend.it/api/handlers"
	"github.com/SaidovZohid/swiftsend.it/config"
	"github.com/SaidovZohid/swiftsend.it/pkg/logger"
	"github.com/SaidovZohid/swiftsend.it/pkg/mongodb"
	"github.com/SaidovZohid/swiftsend.it/sshserver"
	"github.com/SaidovZohid/swiftsend.it/storage"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	logger.Init()
	log := logger.GetLogger()
	log.Info("logger initialized")

	cfg := config.Load()
	log.Info("config initialized")

	database, err := mongodb.NewClient(cfg.MongoDB.Url, cfg.MongoDB.Username, cfg.MongoDB.Password)
	if err != nil {
		log.Fatal("error while connecting to mongodb:", err)
	}
	log.Info("database initialized")

	// for tunneling throw ssh and http
	pipes := make(map[string]sshserver.Tunnel)

	strg := storage.NewStorage(database)

	app := api.New(&api.RoutetOptions{
		Cfg:   &cfg,
		Log:   log,
		Strg:  strg,
		Pipes: pipes,
	})

	// run htpp port in goroutine
	go func() {
		if err = app.Listen(cfg.HttpPort); err != nil {
			log.Fatal("error while listening http port:", err)
		}
	}()

	// Decrypt the data
	decoded, err := base64.StdEncoding.DecodeString(cfg.EncryptedPrivateKey)
	if err != nil {
		log.Fatal("Decoding error:", err)
	}

	// decrypt the encrypted private key
	b, err := handlers.Decrypt(decoded, []byte(cfg.EncryptSecretKey))
	if err != nil {
		log.Error(err)
		return
	}

	privateKey, err := gossh.ParsePrivateKey(b)
	if err != nil {
		log.Fatal("Faild to parse private key:", err)
	}

	// listen and serve ssh
	log.Fatal(sshserver.ListenAndServe(privateKey, &cfg, pipes, strg))
}
