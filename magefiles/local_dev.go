//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/vault/api"
	"github.com/jmoiron/sqlx"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

type LocalDev mg.Namespace

// Deps sets up the local development environment.
func (l LocalDev) Deps() error {
	mg.Deps(Clean)
	fmt.Println("Initializing local environment...")

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start local environment: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connected := false
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for local environment to start")
		default:
			fmt.Println("Checking if database is ready...")
			db, err := sqlx.Open("mysql", "root:Password123@tcp(localhost:3306)/puppetreporter")
			if err != nil {
				fmt.Println("Failed to open database connection:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("Waiting for database to start...")
			if err := db.Ping(); err != nil {
				fmt.Println("Failed to ping database:", err)
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("Database is ready! Closing connection...")
			if err := db.Close(); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("Database connection closed!")
			connected = true
		}

		if connected {
			break
		}
	}

	fmt.Println("Setting up Vault...")
	if err := l.vaultInit(); err != nil {
		return fmt.Errorf("failed to setup vault: %w", err)
	}
	fmt.Println("Vault setup successfully!")

	fmt.Println("Setting up local Database...")
	if err := l.setupLocalDatabase(); err != nil {
		return fmt.Errorf("failed to setup local database: %w", err)
	}
	fmt.Println("Local Database setup successfully!")

	fmt.Println("Local environment initialized successfully!")
	return nil
}

// Clean stops and removes the local development environment.
func (l LocalDev) Clean() error {
	fmt.Println("Cleaning...")
	cmd := exec.Command("docker", "compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l LocalDev) vaultInit() error {
	client, err := getLocalVaultClient()
	if err != nil {
		return fmt.Errorf("error getting vault client: %w", err)
	}

	const dbMountPath = "b3-prod-1-db"

	// Enable the Database secrets engine
	if err := client.Sys().Mount(dbMountPath, &api.MountInput{
		Type:        "database",
		Description: "Database secrets engine",
		Config: api.MountConfigInput{
			Options: map[string]string{
				"plugin_name": "mysql-database-plugin",
			},
		},
	}); err != nil {
		return fmt.Errorf("error enabling database secrets engine: %w", err)
	}

	// Create a new connection
	path := dbMountPath + "/config/prod-mysql"

	data := map[string]interface{}{
		"plugin_name":              "mysql-database-plugin",
		"allowed_roles":            "readwrite",
		"connection_url":           "{{username}}:{{password}}@tcp(mariadb:3306)/",
		"username":                 "root",
		"password":                 "Password123",
		"root_rotation_statements": []string{},
	}

	if _, err := client.Logical().Write(path, data); err != nil {
		return fmt.Errorf("error configuring MySQL connection: %w", err)
	}

	fmt.Println("MySQL database connection configured successfully!")

	const roleName = "readwrite"
	rolePath := dbMountPath + "/roles/" + roleName

	roleData := map[string]interface{}{
		"db_name":   "prod-mysql",
		"role_name": roleName,
		"creation_statements": []string{
			"CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';",
			"GRANT ALL PRIVILEGES ON *.* TO '{{name}}'@'%';",
		},
		"default_ttl": "1h",
		"max_ttl":     "24h",
	}

	if _, err := client.Logical().Write(rolePath, roleData); err != nil {
		return fmt.Errorf("error creating MySQL role: %w", err)
	}

	fmt.Println("MySQL role 'readonly' created successfully at:", rolePath)

	return nil
}

func (l LocalDev) setupLocalDatabase() error {
	// Current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			fmt.Println("failed to change directory: %w", err)
		}
	}()

	os.Setenv("DATABASE_URL", "root:Password123@tcp(localhost:3306)/puppetreporter")

	// Set the working directory to the database/migrations directory
	if err := os.Chdir("./database/migrations"); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Run the database migrations
	cmd := exec.Command("goschema", "migrate", "-up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
