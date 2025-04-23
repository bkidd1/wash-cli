#!/bin/bash

# Build the project
echo "Building wash-cli..."
go build -o wash ./cmd/wash

# Make the binary executable
chmod +x wash

echo "Build complete! Binary is available at ./wash" 