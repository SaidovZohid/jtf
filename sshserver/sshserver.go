package sshserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/SaidovZohid/swiftsend.it/config"
	"github.com/SaidovZohid/swiftsend.it/pkg/utils"
	"github.com/SaidovZohid/swiftsend.it/storage"
	"github.com/SaidovZohid/swiftsend.it/storage/mongodb"
	"github.com/gliderlabs/ssh"
	"go.mongodb.org/mongo-driver/mongo"
	gossh "golang.org/x/crypto/ssh"
)

type Tunnel struct {
	File       File
	DoneChan   chan struct{}
	DeleteChan chan struct{}
	SentAt     time.Time
	ExpiresAt  time.Time
	User       *User
}

type File struct {
	W        io.Writer
	FileSize int64
}

type User struct {
	Subdomain string
	Options   *UserOption
}

type UserOption struct {
	From     *string
	Filename *string
	Message  *string
	Save     *int
}

// ListenAndServer configures ssh key with private key of server and start ssh server
func ListenAndServe(privateKey gossh.Signer, cfg *config.Config, pipes map[string]Tunnel, strg storage.StorageI) error {
	var tunnel Tunnel
	// Configure the SSH server
	server := ssh.Server{
		Addr: cfg.SshPort,
		Handler: func(s ssh.Session) {
			tunnel.HandleSSH(s, cfg, pipes, strg)
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Add your logic here to validate the client's public key
			// and authorize the connection

			// For simplicity, this example allows any client to connect
			return key.Type() != "dsdsd"
		},
	}

	// Set the server private key
	server.AddHostKey(privateKey)
	log.Println("SSH server listening on port:", cfg.SshPort)

	return server.ListenAndServe()
}

func (p *Tunnel) HandleSSH(session ssh.Session, cfg *config.Config, pipes map[string]Tunnel, strg storage.StorageI) {
	// Extracting the IP address from the connection
	userIP, _, _ := net.SplitHostPort(session.RemoteAddr().String())

	// usage, err := strg.Usage().GetUsage(context.Background(), userIP)
	// if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
	// 	log.Println(err)
	// 	writeErrorAndHowToUse(session)
	// 	return
	// }

	marshaledPublicKey := session.PublicKey().Marshal()
	// Parse the SSH authorized key
	pubKey, err := ssh.ParsePublicKey(marshaledPublicKey)
	if err != nil {
		log.Println(err)
		writeErrorAndHowToUse(session)
		return
	}

	// Calculate the fingerprint
	fingerprint := gossh.FingerprintSHA256(pubKey)[7:]
	user, err := strg.User().GetUserInfoByHashSSH(context.Background(), fingerprint)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Println(err)
		writeErrorAndHowToUse(session)
		return
	}

	// if usage != nil {
	// 	if usage.Usage >= 3 {
	// 		if user == nil {
	// 			handleUsageExceededNotUser(session)
	// 			return
	// 		} else if user.Subdomain == nil {
	// 			handleUsageExceededNotSubdomain(session)
	// 			return
	// 		}
	// 	}
	// }

	// Create a fixed time zone for GMT+5 (Asia/Tashkent)
	timezone := time.FixedZone("GMT+5", 5*60*60) // 5 hours ahead of UTC

	var link string
	for {
		link = utils.GenerateRandomLink(7)
		if _, ok := pipes[link]; !ok {
			break
		}
	}

	timeNow := time.Now().In(timezone)

	pipes[link] = Tunnel{
		File: File{
			W:        &bytes.Buffer{}, // Use a bytes.Buffer as the io.Writer
			FileSize: 0,
		},
		DoneChan:   make(chan struct{}),
		DeleteChan: make(chan struct{}),
		SentAt:     timeNow,
		ExpiresAt:  timeNow.Add(time.Minute * 15),
		User: &User{
			Subdomain: "",
			Options:   &UserOption{},
		},
	}
	pipe := pipes[link]

	// Copy the data from val.W to the buffer
	err = performCopyOperation(session, &pipe)
	if err != nil {
		writeErrorAndHowToUse(session)
		delete(pipes, link)
		return
	}

	// [from=Alex msg=Hello, John! Heres your special file filename=main.txt]
	if session.Command() != nil {
		parts := strings.Fields(session.RawCommand())
		err = parseUserInput(parts, &pipe)
		if err != nil {
			writeErrorAndHowToUse(session)
			return
		}
	} else {
		pipe.User.Options = nil
	}

	// greeting
	greatingHi(session)

	if user != nil && user.Subdomain != nil {
		pipe.User.Subdomain = *user.Subdomain
		handleUserHas(session, user, cfg, link, pipe)
	} else {
		handleUserNot(session, cfg, link, pipe)
	}

	// Calculate the time to wait for 15 minutes
	var (
		waitTime time.Time
	)
	if pipe.User.Options != nil {
		if pipe.User.Options.Save != nil {
			waitTime = timeNow.Add(time.Duration(*pipe.User.Options.Save * int(time.Minute)))
		} else {
			waitTime = timeNow.Add(cfg.TimerForSSH)
		}
	} else {
		waitTime = timeNow.Add(cfg.TimerForSSH)
	}

	// Start a timer to wait for 15 minutes or user option from 1 minute to 60 minute acceptable
	timer := time.NewTimer(waitTime.Sub(timeNow))

	_, err = strg.Usage().CreateUsage(context.Background(), &mongodb.Usage{
		IPAddress: userIP,
		Usage:     1,
	})
	if err != nil {
		log.Println(err)
		writeErrorAndHowToUse(session)
		delete(pipes, link)
		return
	}

	// Wait for either the timer to expire or the DoneChan to be closed
	select {
	case <-timer.C:
		// Timer expired, close the DoneChan
		close(pipe.DoneChan)
		delete(pipes, link)
		handleNooneDownloaded(session)
		return
	case <-pipe.DoneChan:
		handleFinished(timer, session, pipe)
		return
	case <-pipe.DeleteChan:
		handleDeleted(timer, session, pipe)
		return
	}
}
