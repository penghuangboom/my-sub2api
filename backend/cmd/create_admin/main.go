package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	cfg, err := config.LoadForBootstrap()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := ent.Open("postgres", cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		log.Printf("Schema migrate warning: %v", err)
	}

	email := "admin@example.com"
	password := "admin123"

	if len(os.Args) >= 3 {
		email = strings.TrimSpace(os.Args[1])
		password = strings.TrimSpace(os.Args[2])
	}

	total, _ := client.User.Query().Count(ctx)
	fmt.Printf("Total users in DB: %d\n", total)

	exists, err := client.User.Query().Where(dbuser.EmailEQ(email)).First(ctx)
	if err == nil && exists != nil {
		fmt.Printf("User already exists: %s (id=%s)\n", exists.Email, exists.ID)
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	u, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash(string(hashed)).
		SetRole("admin").
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	fmt.Printf("Admin user created: %s (id=%s)\n", u.Email, u.ID)
	return nil
}
