package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kardianos/osext"
	"github.com/nitrous-io/gopass"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/domain"
	"github.com/nitrous-io/nitrous-solo/pkg/le"
)

var (
	exeName = "nitrousadmin"
	emailRe = regexp.MustCompile(`\A[^@\s]+@([^@\s]+\.)+[^@\s]+\z`)
)

func main() {
	dataDirPathFlag := flag.String("d", "", "path to data directory")
	flag.Parse()

	u, err := user.Current()
	if err != nil {
		logrus.Fatal("could not determine current user")
	}

	dataDirPath := *dataDirPathFlag
	if dataDirPath == "" {
		dataDirPath = filepath.Join(u.HomeDir, ".nitrous-solo")
	}

	data.CurrentUser = u.Username
	data.DataDirPath = dataDirPath
	if err := data.LoadUserData(); err != nil {
		logrus.Fatal(err)
	}

	args := flag.Args()

	exePath, err := osext.Executable()
	if err != nil {
		logrus.Fatal("could not determine executable file path")
	}

	exeName = filepath.Base(exePath)

	if len(args) > 0 {
		params := args[1:]
		subcommand := strings.ToLower(args[0])
		switch subcommand {
		case "adduser":
			addUser(params)
		case "deluser":
			delUser(params)
		case "sethost":
			setHost(params)
		case "renewcert":
			renewCert(params)
		default:
			logrus.Fatalf("Unrecognized subcommand \"%s\"", subcommand)
		}
		os.Exit(0)
	}

	fmt.Printf(strings.TrimSpace(`
Usage
=====
%s subcommand [arguments...]

Sub-Commands
============
adduser SYSTEM_USERNAME IDE_USERNAME [IDE_PASSWORD]
deluser USERNAME
sethost HOSTNAME EMAIL
renewcert
	`)+"\n", exeName)
}

func addUser(params []string) {
	if len(params) < 2 {
		logrus.Fatalf("Usage: %s adduser SYSTEM_USERNAME IDE_USERNAME [IDE_PASSWORD]", exeName)
	}

	systemUser := strings.ToLower(params[0])
	_, err := user.Lookup(systemUser)
	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			logrus.Fatalf(`could not find user "%s" on this system`, systemUser)
		}
		logrus.Fatal(err)
	}

	var (
		login    = strings.ToLower(params[1])
		password string
	)

	if len(params) < 3 {
		for {
			var (
				pw, pw2 string
				err     error
			)

			for {
				pw, err = gopass.GetPass(fmt.Sprintf(`Password for "%s" (entered text will be hidden): `, login))
				if err != nil {
					logrus.Fatal(err)
				}

				if pw != "" && len(pw) >= 6 {
					break
				}
				fmt.Println("Password must be at least 6 characters long, try again.")
			}

			pw2, err = gopass.GetPass(fmt.Sprintf(`Confirm Password for "%s": `, login))
			if err != nil {
				logrus.Fatal(err)
			}

			if pw == pw2 {
				password = pw
				break
			}
			fmt.Println("Passwords do not match, try again.")
		}
	} else {
		password = params[2]
	}

	ud := &data.UserData{
		Login: login,
		User:  systemUser,
	}
	ud.SetPassword(password)

	update := false
	if _, ok := data.D.UserData[login]; ok {
		fmt.Printf("User \"%s\" already exists, updating instead of creating...\n", login)
		update = true
	}

	data.D.UserData[login] = ud
	if err := data.SaveUserData(); err != nil {
		logrus.Fatal(err)
	}

	if update {
		fmt.Printf("Successfully updated user \"%s\".\n", login)
	} else {
		fmt.Printf("Successfully added user \"%s\".\n", login)
	}
}

func delUser(params []string) {
	if len(params) < 1 {
		logrus.Fatalf("Usage: %s deluser IDE_USERNAME", exeName)
	}

	login := strings.ToLower(params[0])

	if _, ok := data.D.UserData[login]; !ok {
		logrus.Fatalf(`User "%s" not found`, login)
	}
	delete(data.D.UserData, login)
	if err := data.SaveUserData(); err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("Successfully deleted user \"%s\".\n", login)
}

func setHost(params []string) {
	le.CertDataPath = filepath.Join(data.DataDirPath, "global")

	if len(params) < 2 {
		logrus.Fatalf("Usage: %s hostname HOSTNAME EMAIL", exeName)
	}

	var err error
	hostname, err := domain.Sanitize(params[0])
	if err != nil {
		logrus.Fatalf(`could not add hostname "%s" on this system: %v`, params[0], err)
	}

	email := params[1]
	if len(email) < 5 || !emailRe.MatchString(email) {
		logrus.Fatalf(`could not add email "%s" on this system: the email address is invalid`, email)
	}

	certPath, keyPath, err := le.GetCertPathes()
	if err != nil {
		if err != le.ErrUserNotFound {
			logrus.Fatalf(`could not read existing cert files: %v`, email, err)
		}
	}

	// If valid cert files exist, it does not generate again
	if certPath != "" && keyPath != "" {
		if certPath == le.CertificatePath(hostname) && keyPath == le.PrivateKeyPath(hostname) {
			if ok, err := le.VerifyExistingCert(hostname); err != nil || !ok {
				if err != nil {
					logrus.Warnf(`failed to load existing cert files: %v`, err)
				} else {
					logrus.Infof(`cert files will be expired soon in less 2 days`)
				}
			} else {
				logrus.Fatalln(`Valid certs already exists`)
			}
		}
	}

	// Generate new cert with new email and hostname
	if err := le.GenerateCert(email, hostname); err != nil {
		logrus.Fatalf(`could not generate certs: %v`, err)
	}

	fmt.Println("Successfully enable https with lets encrypt\n")
}

func renewCert(params []string) {
	le.CertDataPath = filepath.Join(data.DataDirPath, "global")

	if err := le.RenewCert(); err != nil {
		logrus.Fatalf(`could not renew certs: %v`, err)
	}

	fmt.Println("Successfully renew cert\n")
}
