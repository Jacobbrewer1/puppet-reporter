//go:build mage
// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
)

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

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

type Linting mg.Namespace

// Apis lints the API spec.
func (l Linting) Apis() error {
	fmt.Println("Linting API spec...")

	if err := l.installOpenApiLint(); err != nil {
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

func (l Linting) installOpenApiLint() error {
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
