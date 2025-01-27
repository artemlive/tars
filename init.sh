#!/bin/bash

# Define project structure
folders=(
  "cmd/tars"
  "pkg/slack"
  "pkg/sorting"
  "pkg/storage"
  "pkg/utils"
  "internal/core"
  "configs"
  "scripts"
  "docs"
  "tests/integration"
  "tests/unit"
)

files=(
  "cmd/tars/main.go"
  "pkg/slack/client.go"
  "pkg/slack/handlers.go"
  "pkg/slack/models.go"
  "pkg/sorting/sorter.go"
  "pkg/sorting/config.go"
  "pkg/storage/db.go"
  "pkg/storage/queries.go"
  "pkg/utils/logger.go"
  "pkg/utils/config.go"
  "internal/core/bot.go"
  "internal/core/events.go"
  "configs/config.yaml"
  "scripts/build.sh"
  "scripts/run.sh"
  "scripts/deploy.sh"
  "docs/README.md"
  "Makefile"
  "go.mod"
)

# Create folders
for folder in "${folders[@]}"; do
  mkdir -p "$folder"
  echo "Created folder: $folder"
done

# Create files
for file in "${files[@]}"; do
  touch "$file"
  echo "Created file: $file"
done

# Initialize go.mod
if [ ! -f "go.mod" ]; then
  echo "module tars" > go.mod
  echo "go 1.20" >> go.mod
  echo "Initialized go.mod"
fi

# Add basic README content
echo "# TARS - Task and Request Sorter\n\nThis project is a Slack bot written in Go for categorizing and managing requests in an SRE support channel." > docs/README.md

# Make scripts executable
chmod +x scripts/*.sh

# Done
echo "Project structure initialized successfully!"

