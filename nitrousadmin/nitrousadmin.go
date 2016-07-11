package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kardianos/osext"
	"github.com/nitrous-io/gopass"
	"github.com/nitrous-io/nitrous-solo/data"
)

var exeName = "nitrousadmin"

func main() {
	dataDirPathFlag := flag.String("d", "", "path to data directory")
	flag.Parse()

	dataDirPath := *dataDirPathFlag
	if dataDirPath == "" {
		u, err := user.Current()
		if err != nil {
			logrus.Fatal("could not determine current user")
		}
		dataDirPath = filepath.Join(u.HomeDir, ".nitrous-solo")
	}

	data.DataDirPath = dataDirPath
	data.Load()

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
	if _, ok := data.D.Users[login]; ok {
		fmt.Printf("User \"%s\" already exists, updating instead of creating...\n", login)
		update = true
	}

	data.D.Users[login] = ud
	if err := data.Save(); err != nil {
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

	if _, ok := data.D.Users[login]; !ok {
		logrus.Fatalf(`User "%s" not found`, login)
	}
	delete(data.D.Users, login)
	if err := data.Save(); err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("Successfully deleted user \"%s\".\n", login)
}
