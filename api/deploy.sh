#!/bin/bash

# Set default stage if not provided as an argument
STAGE=${1:-dev}

# Additional check for production stage
if [ "$STAGE" == "prod" ]; then
    echo "🚨 PRODUCTION DEPLOYMENT WARNING 🚨"
    read -p "Have you reviewed and updated CORS headers? (yes/no): " cors_confirmation

    if [ "$cors_confirmation" != "yes" ]; then
        echo "Deployment aborted. Please update CORS headers before proceeding."
        exit 1
    fi
fi

FULL_STACK_NAME="BookIt-${STAGE}"

# The GOOS and GOARCH are set for AWS Lambda's Linux environment.
GOOS="linux"
GOARCH="arm64"

# Array of source directories where your main.go files are located.
SRC_DIRS=()
for dir in ./cmd/*; do
    if [ -d "$dir" ]; then
        SRC_DIRS+=("$dir")
    fi
done

# Loop through each source directory
for SRC_DIR in "${SRC_DIRS[@]}"; do
    # The name of the binary output; typically "bootstrap" for AWS Lambda.
    OUTPUT_NAME="bootstrap"

    # The output directory where the compiled binary will be placed.
    OUT_DIR="$SRC_DIR"

    # Extract the last directory name for creating a unique ZIP file name
    DIR_NAME=$(basename "$SRC_DIR")

    # Navigate to the source directory
    cd "$SRC_DIR" || exit

    # Compile the Go application
    echo "Compiling the Go application in $SRC_DIR for stage: $STAGE..."
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go

    # Ensure the binary is executable
    chmod 755 "./$OUTPUT_NAME"

    # Navigate back to the original directory
    cd ../..
done

# Deploy using AWS SAM with the stage parameter
sam deploy --stack-name="$FULL_STACK_NAME" --parameter-overrides StageName="$STAGE"
