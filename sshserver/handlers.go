package sshserver

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/SaidovZohid/swiftsend.it/config"
	"github.com/SaidovZohid/swiftsend.it/storage/mongodb"
	"github.com/gliderlabs/ssh"
	"github.com/logrusorgru/aurora"
)

const (
	baseURI = "https://jtf.zohiddev.me"
)

// if io.Copy does not give and response in 5 seconds it returns error
func performCopyOperation(session ssh.Session, pipe *Tunnel) error {
	// Set timeout duration to 5 seconds
	timeout := 3 * time.Second

	// Create a channel to receive the result
	resultCh := make(chan error)

	// Start a goroutine to perform the copy operation
	go func() {
		// Perform the copy operation from `val.W` to the buffer
		fileSize, err := io.Copy(pipe.File.W, session)
		resultCh <- err
		if fileSize != 0 {
			pipe.File.FileSize = fileSize
		}
	}()

	// Wait for the operation to complete or the timeout to elapse
	select {
	case err := <-resultCh:
		// Operation completed before the timeout
		if err != nil {
			// Handle the error appropriately
			return fmt.Errorf("error handling: %v", err)
		}
		// Success
		return nil
	case <-time.After(timeout):
		// Timeout duration elapsed
		return fmt.Errorf("timeout occurred")
	}
}

// error handler writes to user session console!
func writeErrorAndHowToUse(s ssh.Session) {
	io.WriteString(s, "\n")
	io.WriteString(s, "\t"+aurora.Red("🔵❗ JTF Error").String()+"\n\n")

	io.WriteString(s, aurora.Blue("Uh-oh! It seems like something went wrong with your request. Please check the details below and try again:").String()+"\n\n")

	io.WriteString(s, aurora.Green("🌟 Quick Tips:").String()+"\n")

	io.WriteString(s, `	- Use the "from=" option to set a custom name for the download page.
	- Add a personalized "msg=" to include a special message along with the file.
	- Customize the "filename=" parameter to give your downloaded file a unique name.
	- Set "t=" option (0<sv<60) during file upload to control download time.
	`)

	io.WriteString(s, "\n"+aurora.Green("💡 Did you know?").String()+"\n")
	io.WriteString(s, "\t- You can even set multiple options together to create a highly customized experience. Feel free to explore the possibilities!\n")

	io.WriteString(s, aurora.Green("🚀 Example Command:").String()+"\n")
	io.WriteString(s, "To send your file with a personalized message and write your name, try the following command:\n")

	io.WriteString(s, aurora.Magenta(`ssh jtf.zohiddev.me -p 2222 from="Alex" msg="Hello, John! Here's your special file"  < myfile.txt`).String()+"\n\n")

	io.WriteString(s, aurora.Yellow("✨ Get creative and enjoy using JTF! If you need any further assistance, don't hesitate to reach out. mailto='support@zohiddev.me'✨").String()+"\n")
}

// func handleUsageExceededNotUser(s ssh.Session) {
// 	io.WriteString(s, "\n"+aurora.Red("\t❗ JTF 3 times usage exceeded ❗").String()+"\n\n")

// 	io.WriteString(s, aurora.Blue("⚠️  Oops! Usage limit exceeded. Link your laptop's SSH key to continue. ⚡️").String()+"\n\n")
// 	io.WriteString(s, "\t"+aurora.Green("🚀 New to JTF? Sign up at https://jtf.zohiddev.me/signup, get verified with a subdomain, and link your key for limitless access! 🔑✨").String()+"\n\n")
// 	io.WriteString(s, "\t"+aurora.Green("🔗 Already a JTF member? Link your SSH key now at https://jtf.zohiddev.me/s/settings/keys/add for seamless file transfers! 🚀🔒").String()+"\n\n")

// 	io.WriteString(s, aurora.Yellow("✨ Get creative and enjoy using JTF! If you need any further assistance, don't hesitate to reach out. mailto='support@zohiddev.me'✨").String()+"\n")
// }

// func handleUsageExceededNotSubdomain(s ssh.Session) {
// 	io.WriteString(s, "\n"+aurora.Red("\t❗ JTF 3 times usage exceeded ❗").String()+"\n\n")

// 	io.WriteString(s, aurora.Blue("⚠️  Oops! Usage limit exceeded. Get a subdomain to be verified user to continue. ⚡️").String()+"\n\n")
// 	io.WriteString(s, "\t"+aurora.Blue("🚀 New to JTF? Sign up at https://jtf.zohiddev.me/signup, get verified with a subdomain, and link your key for limitless access! 🔑✨").String()+"\n\n")
// 	io.WriteString(s, "\t🔗 Already a JTF member? Get a subdomain now at https://jtf.zohiddev.me/s/settings/account for seamless file transfers! 🚀🔒\n\n")

// 	yellow := color.New(color.FgYellow)
// 	yellow.Fprint(s, "✨ Get creative and enjoy using JTF! If you need any further assistance, don't hesitate to reach out. mailto='support@zohiddev.me'✨\n")
// }

func greatingHi(s ssh.Session) {
	io.WriteString(s, "\t"+aurora.Green("🌟✨ Welcome to JTF! ✨🌟").String()+"\n\n")
}

func handleFinished(timer *time.Timer, s ssh.Session, pipe Tunnel) {
	io.WriteString(s, aurora.Yellow("🛎  Exciting news! 📥 Your file downloaded. 🎉✨").String()+"\n")
	timer.Stop()
}

func handleDeleted(timer *time.Timer, s ssh.Session, pipe Tunnel) {
	io.WriteString(s, aurora.Red("🛎  Exciting news! 🗑  Your file deleted. ❌").String()+"\n")
	timer.Stop()
}

func handleNooneDownloaded(s ssh.Session) {
	io.WriteString(s, aurora.Yellow("⏳ Time's up! No downloaded 😭. Keep sharing the link! 🔥").String()+"\n")
}

func handleUserHas(s ssh.Session, user *mongodb.User, cfg *config.Config, link string, pipe Tunnel) {
	subdomainUrl := fmt.Sprintf(cfg.BaseURL+"/domain/%v/info", *user.Subdomain)
	if cfg.BaseURL == baseURI {
		subdomainUrl = fmt.Sprintf("https://%v."+cfg.BaseURL[8:], *user.Subdomain)
	}

	io.WriteString(s, fmt.Sprintf("%v %v 🔒🌟\n\n", aurora.Green("🌟🔒 Detected verified user domain").String(), aurora.Cyan(subdomainUrl).Underline().String()))

	handleLinkSent(s, cfg, link, user.Subdomain, pipe)
}

func handleUserNot(s ssh.Session, cfg *config.Config, link string, pipe Tunnel) {
	io.WriteString(s, aurora.Yellow("💫 Discover the Power of JTF! Get your Own Verified Link Today! 🌟").String()+"\n")

	io.WriteString(s, fmt.Sprintf("\nWant your own personal verified link %v?\n", aurora.Cyan("username.jtf.zohiddev.me").Underline().String()))

	io.WriteString(s, "\t"+aurora.Red("-> Visit https://jtf.zohiddev.me to get your verified subdomain, it's FREE and get UNLIMITED transfers").String()+"\n")

	handleLinkSent(s, cfg, link, nil, pipe)
}

func handleLinkSent(s ssh.Session, cfg *config.Config, link string, subdomain *string, pipe Tunnel) {
	// Download link to frontend page
	io.WriteString(s, "Download link:\n")
	downloadLink := cfg.BaseURL + "/download/unknown/" + link
	if subdomain != nil {
		downloadLink = cfg.BaseURL + "/download/" + *subdomain + "/" + link
	}
	if cfg.BaseURL == baseURI {
		if subdomain == nil {
			downloadLink = "https://" + "unknown." + cfg.BaseURL[8:] + "/" + link
		} else {
			downloadLink = "https://" + *subdomain + "." + cfg.BaseURL[8:] + "/" + link
		}
		// https://zohiddev.me
	}
	io.WriteString(s, "\t"+aurora.Yellow(downloadLink).String()+"\n")

	// Direct download link
	io.WriteString(s, "\nDirect download link:\n")
	directLink := cfg.BaseURL + "/direct/" + link
	if cfg.BaseURL == baseURI {
		directLink = "https://direct." + cfg.BaseURL[8:] + "/" + link
	}
	io.WriteString(s, "\t"+aurora.Yellow(directLink).String()+"\n")

	io.WriteString(s, "\nDelete file link:\n")
	deleteLink := cfg.BaseURL + "/delete/" + link
	if cfg.BaseURL == baseURI {
		deleteLink = "https://delete." + cfg.BaseURL[8:] + "/" + link
	}
	io.WriteString(s, "\t"+aurora.Red(deleteLink).String()+"\n")

	tm := fmt.Sprintf("%v minutes", 15)
	if pipe.User.Options != nil && pipe.User.Options.Save != nil {
		if *pipe.User.Options.Save > 1 {
			tm = fmt.Sprintf("%v minutes", *pipe.User.Options.Save)
		} else {
			tm = fmt.Sprintf("%v minute", *pipe.User.Options.Save)
		}
	}

	io.WriteString(s, "\n"+aurora.Cyan("⏳ Please hurry! Your link will expire in "+tm+". After that, the session will automatically close, and the link will become invalid. Let's patiently wait for the download to commence... 🕒").String()+"\n\n")
}

func parseUserInput(input []string, pipe *Tunnel) error {
	var (
		msg, filename, from, lastKey string
		save                         int
	)
	for _, v := range input {
		if strings.Contains(v, "=") {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return errors.New("not true option")
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(strings.TrimSpace(parts[1]))
			switch key {
			case "msg":
				lastKey = key
				if msg != "" {
					msg = msg + " " + value
					continue
				}
				msg = value
			case "from":
				lastKey = key
				if from != "" {
					from = from + " " + value
					continue
				}
				from = value
			case "filename":
				lastKey = key
				if filename != "" {
					filename = filename + " " + value
					continue
				}
				filename = value
			case "t":
				val, err := strconv.Atoi(value)
				if err != nil {
					return errors.New("not true option")
				}
				save = val
			default:
				return errors.New("not true option")
			}
		} else {
			str := strings.TrimSpace(v)
			switch lastKey {
			case "msg":
				if msg != "" {
					msg = msg + " " + str
					continue
				}
				msg = str
			case "from":
				if from != "" {
					from = from + " " + str
					continue
				}
				from = str
			case "filename":
				if filename != "" {
					filename = filename + " " + str
					continue
				}
				filename = str
			default:
				return errors.New("not true option")
			}
		}
	}

	if save > 0 && save <= 60 {
		pipe.User.Options.Save = &save
	} else {
		return errors.New("not true option")
	}
	if filename != "" {
		pipe.User.Options.Filename = &filename
	}
	if from != "" {
		pipe.User.Options.From = &from
	}
	if msg != "" {
		pipe.User.Options.Message = &msg
	}

	return nil
}
