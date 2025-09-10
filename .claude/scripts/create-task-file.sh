#!/bin/bash

# Script to automatically create task files when starting work
# Usage: ./create-task-file.sh <title> <epic> [branch] [github-issue] [description]

set -e

# Configuration
TASKS_DIR=".claude/epics/tasks"
TEMPLATE_FILE="$TASKS_DIR/task-template.md"

# Arguments
TITLE="$1"
EPIC="${2:-general-tasks}"
BRANCH="$3"
GITHUB_ISSUE="$4"
DESCRIPTION="${5:-No description provided}"

# Validation
if [ -z "$TITLE" ]; then
    echo "Error: Title is required"
    echo "Usage: $0 <title> [epic] [branch] [github-issue] [description]"
    exit 1
fi

# Create tasks directory if it doesn't exist
mkdir -p "$TASKS_DIR"

# Generate task ID
TIMESTAMP=$(date +%s)
RANDOM_SUFFIX=$(openssl rand -hex 3 2>/dev/null || echo $(date +%N | tail -c 7))
TASK_ID="task-${TIMESTAMP}-${RANDOM_SUFFIX}"

# Generate filename
FILENAME="${TASK_ID}.md"
FILEPATH="$TASKS_DIR/$FILENAME"

# Check if file already exists
if [ -f "$FILEPATH" ]; then
    echo "Error: Task file already exists: $FILEPATH"
    exit 1
fi

# Get current timestamp in ISO format
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create task content
cat > "$FILEPATH" << EOF
---
id: "$TASK_ID"
title: "$TITLE"
status: "pending"
epic: "$EPIC"
branch: "$BRANCH"
createdAt: "$CURRENT_TIME"
updatedAt: "$CURRENT_TIME"
startedAt: ""
completedAt: ""
tokensUsed: 0
githubIssue: $GITHUB_ISSUE
prUrl: ""
buildStatus: ""
lintStatus: ""
---

# Task: $TITLE

## Epic
$EPIC

## Description
$DESCRIPTION

## Progress
- [ ] Task created ✓
- [ ] Analysis completed
- [ ] Implementation started
- [ ] Testing completed
- [ ] Task completed

## Work Log
### $(date '+%Y-%m-%d %H:%M:%S')
- Task created automatically
- Ready to begin work

## Notes
- Created: $CURRENT_TIME
- Epic: $EPIC
${BRANCH:+- Branch: $BRANCH}
${GITHUB_ISSUE:+- GitHub Issue: #$GITHUB_ISSUE}
EOF

echo "✅ Task file created successfully!"
echo "📁 File: $FILEPATH"
echo "🆔 Task ID: $TASK_ID"
echo "🎯 Epic: $EPIC"
echo "📝 Title: $TITLE"

# Output the task ID for other scripts to use
echo "$TASK_ID"