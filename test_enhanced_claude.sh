#!/bin/bash

# Test script for enhanced Claude integration
echo "Testing enhanced Claude integration..."

# Submit a simple test task
echo "Submitting test task..."
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": "Create a simple README.md file with a welcome message and basic project information for this repository",
    "repository_name": "gallery_ios",
    "provider": "claude",
    "interactive": false,
    "working_dir": "",
    "cmd": ""
  }'

echo -e "\n\nTask submitted! Check the logs to see the enhanced Claude integration in action."