package main

import (
	"fmt"
	"log"
	"os"
	"styx/internal/server"
	"styx/internal/setup"

	"github.com/spf13/pflag"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "setup" {
		db, err := setup.Run()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[*] Server setup, Registering...\n")

		setupArgs := pflag.NewFlagSet("setup", pflag.ContinueOnError)
		username := setupArgs.StringP("user", "u", "", "your username")
		password := setupArgs.StringP("pass", "p", "", "your password")
		domain := setupArgs.StringP("domain", "d", "", "backend domain")

		setupArgs.Parse(os.Args[1:])

		if err := setup.Register(*username, *password, *domain, db); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("[*] Successfully Registered, you can now run '%s run'\n", os.Args[0])
		return
	} else if len(os.Args) >= 2 && os.Args[1] == "run" {
		s, err := server.New()
		if err != nil {
			log.Fatal(err)
		}

		if err := s.Run(); err != nil {
			log.Fatal(err)
		}
	}

}
