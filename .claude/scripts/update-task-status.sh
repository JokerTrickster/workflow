#!/bin/bash

# Script to update task status in task files
# Usage: ./update-task-status.sh <task-id> <status> [tokens] [branch] [pr-url]

set -e

# Configuration
TASKS_DIR=".claude/epics/tasks"

# Arguments
TASK_ID="$1"
STATUS="$2"
TOKENS="${3:-0}"
BRANCH="$4"
PR_URL="$5"

# Validation
if [ -z "$TASK_ID" ] || [ -z "$STATUS" ]; then
    echo "Error: Task ID and status are required"
    echo "Usage: $0 <task-id> <status> [tokens] [branch] [pr-url]"
    echo "Status options: pending, in_progress, completed, failed"
    exit 1
fi

# Validate status
case "$STATUS" in
    pending|in_progress|completed|failed)
        ;;
    *)
        echo "Error: Invalid status. Use: pending, in_progress, completed, failed"
        exit 1
        ;;
esac

# Find task file
TASK_FILE="$TASKS_DIR/${TASK_ID}.md"

if [ ! -f "$TASK_FILE" ]; then
    echo "Error: Task file not found: $TASK_FILE"
    exit 1
fi

# Get current timestamp
CURRENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create temporary file for processing
TEMP_FILE="${TASK_FILE}.tmp"

# Process the file to update frontmatter
awk -v status="$STATUS" -v tokens="$TOKENS" -v branch="$BRANCH" -v pr_url="$PR_URL" -v timestamp="$CURRENT_TIME" '
BEGIN { 
    in_frontmatter = 0
    found_status = 0
    found_tokens = 0
    found_started = 0
    found_completed = 0
    found_updated = 0
    found_branch = 0
    found_pr = 0
}
/^---$/ { 
    if (in_frontmatter == 0) {
        in_frontmatter = 1
    } else {
        # End of frontmatter - add missing fields
        if (!found_status) print "status: \"" status "\""
        if (!found_tokens) print "tokensUsed: " tokens
        if (!found_updated) print "updatedAt: \"" timestamp "\""
        if (status == "in_progress" && !found_started) print "startedAt: \"" timestamp "\""
        if (status == "completed" && !found_completed) print "completedAt: \"" timestamp "\""
        if (branch != "" && !found_branch) print "branch: \"" branch "\""
        if (pr_url != "" && !found_pr) print "prUrl: \"" pr_url "\""
        in_frontmatter = 0
    }
    print
    next
}
in_frontmatter && /^status:/ { 
    print "status: \"" status "\""
    found_status = 1
    next
}
in_frontmatter && /^tokensUsed:/ { 
    print "tokensUsed: " tokens
    found_tokens = 1
    next
}
in_frontmatter && /^updatedAt:/ { 
    print "updatedAt: \"" timestamp "\""
    found_updated = 1
    next
}
in_frontmatter && /^startedAt:/ && status == "in_progress" { 
    if ($2 == "" || $2 == "\"\"") {
        print "startedAt: \"" timestamp "\""
    } else {
        print
    }
    found_started = 1
    next
}
in_frontmatter && /^completedAt:/ && status == "completed" { 
    print "completedAt: \"" timestamp "\""
    found_completed = 1
    next
}
in_frontmatter && /^branch:/ && branch != "" { 
    print "branch: \"" branch "\""
    found_branch = 1
    next
}
in_frontmatter && /^prUrl:/ && pr_url != "" { 
    print "prUrl: \"" pr_url "\""
    found_pr = 1
    next
}
{ print }
' "$TASK_FILE" > "$TEMP_FILE"

# Replace original file
mv "$TEMP_FILE" "$TASK_FILE"

# Add work log entry
echo "" >> "$TASK_FILE"
echo "### $(date '+%Y-%m-%d %H:%M:%S') - Status Update" >> "$TASK_FILE"
echo "- Status changed to: **$STATUS**" >> "$TASK_FILE"
[ "$TOKENS" != "0" ] && echo "- Tokens used: $TOKENS" >> "$TASK_FILE"
[ -n "$BRANCH" ] && echo "- Branch: $BRANCH" >> "$TASK_FILE"
[ -n "$PR_URL" ] && echo "- PR: $PR_URL" >> "$TASK_FILE"

echo "✅ Task status updated successfully!"
echo "📁 File: $TASK_FILE"
echo "🆔 Task ID: $TASK_ID"
echo "📊 Status: $STATUS"
[ "$TOKENS" != "0" ] && echo "🪙 Tokens: $TOKENS"
[ -n "$BRANCH" ] && echo "🌿 Branch: $BRANCH"
[ -n "$PR_URL" ] && echo "🔗 PR: $PR_URL"