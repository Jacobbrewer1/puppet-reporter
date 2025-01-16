#!/bin/bash

# Check that the IBM OpenAPI Linter is installed
if ! command -v lint-openapi >/dev/null; then
  gum style --foreground 196 "Error: IBM OpenAPI Linter is not installed. Please install the linter by following the instructions at https://github.com/IBM/openapi-validator"
  exit 1
fi

# Find all routes.yaml files in the ./pkg/codegen/apis directory
routesFiles=$(find ./pkg/codegen/apis -name "routes.yaml")

score=0

# Lint each routes.yaml file
for file in $routesFiles; do
  rm -rf ./lint-output.json

  lint-openapi -c ./openapi-lint-config.yaml -s "$file" >./lint-output.json

  # Make ./pkg/codegen/apis/api/routes.yaml -> api
  apiName=$(echo "$file" | sed -e 's/.*\/\(.*\)\/routes.yaml/\1/')

  mv ./routes-validator-report.md "./routes-validator-report-$apiName.md"

  score=$(jq '.impactScore.categorizedSummary.overall' ./lint-output.json)
done

# Combine all the reports into a single file. Separate each report with a horizontal rule.
cat ./routes-validator-report-*.md >./routes-validator-report.md

if [ "$score" -lt 100 ]; then
  echo "IBM OpenAPI Linter found issues with the OpenAPI specification. Please fix the issues and try again."
  exit 1
fi
