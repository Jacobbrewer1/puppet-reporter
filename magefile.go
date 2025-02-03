//go:build mage
// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/hashicorp/vault/api"
	"github.com/jacobbrewer1/utils"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	imageAppSeparator = "/"
)

var dockerRegistry = os.Getenv("DOCKER_REGISTRY")
var envTags = os.Getenv("TAGS")
var toPushEnv = os.Getenv("DOCKER_PUSH")

var toPush = false

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Images() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	toPush, _ = strconv.ParseBool(toPushEnv)
	fmt.Println("Push images set to: ", toPush)

	// Get all directory names in the ./cmd directory
	cmds, err := os.ReadDir("./cmd")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	wg := new(sync.WaitGroup)
	multiErr := utils.NewMultiError()

	// Iterate over each directory
	for _, cmd := range cmds {
		if cmd.IsDir() {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				if err := buildImage(cli, name); err != nil {
					multiErr.Add(fmt.Errorf("failed to build image for %s: %w", name, err))
				}

				if toPush {
					if err := pushImage(name); err != nil {
						multiErr.Add(fmt.Errorf("failed to push image for %s: %w", name, err))
					}
				}
			}(cmd.Name())
		}
	}

	wg.Wait()

	if multiErr.Err() != nil {
		for _, err := range multiErr.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	return nil
}

func buildImage(cli *client.Client, appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := imageTags(applicationDockerRegistry)
	fmt.Println(tags)

	cmd := exec.Command("docker", "build")

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-t", tag)
	}

	cmd.Args = append(cmd.Args, ".")
	cmd.Args = append(cmd.Args, "--build-arg", "APP_NAME="+appName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}

func pushImage(appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := imageTags(applicationDockerRegistry)
	fmt.Println(tags)
	for _, tag := range tags {
		cmd := exec.Command("docker", "push", tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
	}

	return nil
}

func imageTags(registry string) []string {
	envSplit := strings.Split(envTags, ",")
	tags := make([]string, 0)
	for _, tag := range envSplit {
		tags = append(tags, registry+":"+tag)
	}
	return tags
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself
func Clean() error {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin")
	cmd := exec.Command("docker", "compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type apiLintResponse struct {
	Error struct {
		Results []interface{} `json:"results"`
		Summary struct {
			Total   int           `json:"total"`
			Entries []interface{} `json:"entries"`
		} `json:"summary"`
	} `json:"error"`
	Warning struct {
		Results []interface{} `json:"results"`
		Summary struct {
			Total   int           `json:"total"`
			Entries []interface{} `json:"entries"`
		} `json:"summary"`
	} `json:"warning"`
	Info struct {
		Results []interface{} `json:"results"`
		Summary struct {
			Total   int           `json:"total"`
			Entries []interface{} `json:"entries"`
		} `json:"summary"`
	} `json:"info"`
	Hint struct {
		Results []interface{} `json:"results"`
		Summary struct {
			Total   int           `json:"total"`
			Entries []interface{} `json:"entries"`
		} `json:"summary"`
	} `json:"hint"`
	HasResults  bool `json:"hasResults"`
	ImpactScore struct {
		CategorizedSummary struct {
			Usability  int `json:"usability"`
			Security   int `json:"security"`
			Robustness int `json:"robustness"`
			Evolution  int `json:"evolution"`
			Overall    int `json:"overall"`
		} `json:"categorizedSummary"`
		ScoringData []interface{} `json:"scoringData"`
	} `json:"impactScore"`
}

func APILint() error {
	fmt.Println("Linting API spec...")

	if err := installOpenAPILint(); err != nil {
		return fmt.Errorf("failed to install openapi-lint: %w", err)
	}

	// Get all routes.yaml files in the ./pkg/apis/specs directory
	routes, err := os.ReadDir("./pkg/apis/specs")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	specs := make([]string, 0)
	for _, route := range routes {
		if route.IsDir() {
			specs = append(specs, route.Name())
		}
	}

	failed := false
	failedSpecs := make([]string, 0, len(specs))

	for _, spec := range specs {
		cmd := exec.Command("lint-openapi", "-c", "./openapi-lint-config.yaml", "-s", "./pkg/apis/specs/"+spec+"/routes.yaml")
		got, err := cmd.Output()
		if err != nil && err.Error() != "exit status 1" {
			return fmt.Errorf("failed to lint API spec: %w", err)
		}

		fmt.Println(string(got))

		resp := new(apiLintResponse)
		if err := json.Unmarshal(got, resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if resp.ImpactScore.CategorizedSummary.Overall < 100 ||
			resp.ImpactScore.CategorizedSummary.Usability < 100 ||
			resp.ImpactScore.CategorizedSummary.Security < 100 ||
			resp.ImpactScore.CategorizedSummary.Robustness < 100 ||
			resp.ImpactScore.CategorizedSummary.Evolution < 100 {
			failed = true
			failedSpecs = append(failedSpecs, spec)
			fmt.Println("API spec linting failed for", spec)

			if err := os.Rename("./routes-validator-report.md", "./routes-validator-report-"+spec+".md"); err != nil {
				return fmt.Errorf("failed to rename report file: %w", err)
			}
			continue
		}

		fmt.Println("Usability:", resp.ImpactScore.CategorizedSummary.Usability)
		fmt.Println("Security:", resp.ImpactScore.CategorizedSummary.Security)
		fmt.Println("Robustness:", resp.ImpactScore.CategorizedSummary.Robustness)
		fmt.Println("Evolution:", resp.ImpactScore.CategorizedSummary.Evolution)
		fmt.Println("Overall:", resp.ImpactScore.CategorizedSummary.Overall)

		fmt.Println("API spec linting passed for", spec)
		fmt.Println("Removing report file...")
		if err := os.Remove("./routes-validator-report.md"); err != nil {
			return fmt.Errorf("failed to remove report file: %w", err)
		}
	}

	if failed {
		// Combine all reports into a single file
		fmt.Println("Combining reports...")
		if _, err := os.Create("./routes-validator-report.md"); err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}

		builder := new(strings.Builder)

		// Combine all reports into a single file
		for i, spec := range failedSpecs {
			report, err := os.ReadFile("./routes-validator-report-" + spec + ".md")
			if err != nil {
				return fmt.Errorf("failed to read report file: %w", err)
			}

			builder.WriteString(string(report))

			if i != len(failedSpecs)-1 {
				builder.WriteString("\n\n---\n\n")
			}

			if err := os.Remove("./routes-validator-report-" + spec + ".md"); err != nil {
				return fmt.Errorf("failed to remove report file: %w", err)
			}
		}

		if err := os.WriteFile("./routes-validator-report.md", []byte(builder.String()), 0644); err != nil {
			return fmt.Errorf("failed to write report file: %w", err)
		}

		return fmt.Errorf("API spec linting failed")
	}

	fmt.Println("API spec linting passed")

	return nil
}

func installOpenAPILint() error {
	fmt.Println("Installing OpenAPI Lint...")

	// Is the linter already installed?
	_, err := exec.LookPath("lint-openapi")
	if err == nil {
		cmd := exec.Command("npm", "install", "-g", "ibm-openapi-validator")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install ibm-openapi-validator: %w", err)
		}
	}

	// Is the ruleset already installed?
	depsCmd := exec.Command("npm", "install", "@ibm-cloud/openapi-ruleset")
	if err := depsCmd.Run(); err != nil {
		return fmt.Errorf("failed to install openapi-ruleset: %w", err)
	}

	return nil
}

// Set up the local environment and provision the necessary resources
func LocalSetup() error {
	fmt.Println("Initializing local environment...")

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start local environment: %w", err)
	}

	fmt.Println("Waiting for MariaDB to be ready...")
	time.Sleep(10 * time.Second) // Give MariaDB time to start

	fmt.Println("Setting up Vault...")
	if err := vaultSetup(); err != nil {
		//mg.Deps(Clean)
		return fmt.Errorf("failed to setup vault: %w", err)
	}
	fmt.Println("Vault setup successfully!")

	fmt.Println("Setting up local Database...")
	if err := setupLocalDatabase(); err != nil {
		//mg.Deps(Clean)
		return fmt.Errorf("failed to setup local database: %w", err)
	}
	fmt.Println("Local Database setup successfully!")

	fmt.Println("Local environment initialized successfully!")
	return nil
}

func setupLocalDatabase() error {
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

func vaultSetup() error {
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
		"allowed_roles":            "readonly",
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

func getLocalVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	client.SetToken("root")

	return client, nil
}
