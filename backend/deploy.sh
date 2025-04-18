#!/bin/bash

# Build the project
npm run build

# Create deployment package
zip -r function.zip dist/ node_modules/

# Upload to AWS Lambda (replace with your function name)
aws lambda update-function-code \
    --function-name wash-analyzer \
    --zip-file fileb://function.zip

# Clean up
rm function.zip

echo "Deployment complete!" 