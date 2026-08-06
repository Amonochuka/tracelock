package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"

	"golang.org/x/term"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	emailFlag := flag.String("email", "", "admin email address")
	passwordFlag := flag.String("password", "", "new admin password (use only in secure environments)")
	confirmFlag := flag.String("confirm", "", "password confirmation (optional)")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	email := strings.TrimSpace(*emailFlag)
	if email == "" {
		fmt.Print("Admin email: ")
		emailInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("read admin email: ", err)
		}
		email = strings.TrimSpace(emailInput)
	}

	password := strings.TrimSpace(*passwordFlag)
	confirmation := strings.TrimSpace(*confirmFlag)

	if password == "" {
		password = readPassword("New password: ")
		confirmation = readPassword("Confirm password: ")
	}

	if confirmation == "" {
		confirmation = password
	}

	if password != confirmation {
		log.Fatal("passwords do not match")
	}

	service := auth.NewUserService(auth.NewUserAuth(database), nil)
	if err := service.ResetAdminPassword(email, password); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin password reset. Existing refresh sessions have been revoked.")
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatal("read password: ", err)
	}
	return strings.TrimSpace(string(bytePassword))
}
