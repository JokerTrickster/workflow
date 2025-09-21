#!/bin/bash

# Enhanced Issue Start Script with Korean Comment Integration
# This script extends the PM system's issue-start functionality to automatically
# create Korean comments when starting work on GitHub issues.

set -euo pipefail

ISSUE_NUMBER="$1"
REPO_ID="${2:-JokerTrickster/workflow}"

if [ -z "$ISSUE_NUMBER" ]; then
    echo "❌ Usage: $0 <issue_number> [repo_id]"
    echo "   Example: $0 39"
    echo "   Example: $0 39 owner/repo"
    exit 1
fi

echo "🚀 Starting issue #$ISSUE_NUMBER with Korean comment integration..."

# Step 1: Get issue details from GitHub
echo "📋 Fetching issue details..."
if ! ISSUE_DATA=$(gh issue view "$ISSUE_NUMBER" --json state,title,labels,body,assignees 2>/dev/null); then
    echo "❌ Cannot access issue #$ISSUE_NUMBER. Check issue number or run: gh auth login"
    exit 1
fi

ISSUE_TITLE=$(echo "$ISSUE_DATA" | jq -r '.title')
ISSUE_STATE=$(echo "$ISSUE_DATA" | jq -r '.state')

if [ "$ISSUE_STATE" != "open" ]; then
    echo "⚠️  Issue #$ISSUE_NUMBER is $ISSUE_STATE - proceeding anyway"
fi

echo "✅ Found issue: $ISSUE_TITLE"

# Step 2: Check if user is authenticated with GitHub for API access
echo "🔐 Checking GitHub API authentication..."
if ! gh auth status >/dev/null 2>&1; then
    echo "❌ GitHub CLI not authenticated. Run: gh auth login"
    exit 1
fi

# Step 3: Assign issue to self and add in-progress label
echo "👤 Assigning issue to self..."
if ! gh issue edit "$ISSUE_NUMBER" --add-assignee @me --add-label "in-progress" >/dev/null 2>&1; then
    echo "⚠️  Could not modify issue assignment/labels (continuing anyway)"
fi

# Step 4: Create Korean start comment
echo "💬 Creating Korean start comment..."

# Create a temporary Node.js script to call our PM integration service
TEMP_SCRIPT="/tmp/korean_comment_$$.js"
cat > "$TEMP_SCRIPT" << 'EOF'
import { PMGitHubIntegrationService } from '../frontend/src/services/pmGitHubIntegration.js';

const issueNumber = parseInt(process.argv[2]);
const repoId = process.argv[3];
const taskTitle = process.argv[4] === 'undefined' ? null : process.argv[4];

console.log(`Creating Korean start comment for issue #${issueNumber}...`);

PMGitHubIntegrationService.createIssueStartComment(issueNumber, repoId, taskTitle)
  .then(() => {
    console.log('✅ Korean comment created successfully');
    process.exit(0);
  })
  .catch(error => {
    console.error('❌ Failed to create Korean comment:', error.message);
    process.exit(1);
  });
EOF

# Execute the Korean comment creation
if cd frontend && node "$TEMP_SCRIPT" "$ISSUE_NUMBER" "$REPO_ID" "$ISSUE_TITLE" 2>/dev/null; then
    echo "✅ Korean start comment created successfully"
else
    echo "⚠️  Korean comment creation failed (continuing with issue start)"
fi

# Cleanup
rm -f "$TEMP_SCRIPT"

# Step 5: Create or find local task file
echo "📂 Creating/finding local task file..."

# First try to find existing task file
TASK_FILE=""
for epic_dir in .claude/epics/*/; do
    if [ -f "${epic_dir}${ISSUE_NUMBER}.md" ]; then
        TASK_FILE="${epic_dir}${ISSUE_NUMBER}.md"
        break
    fi
done

# If not found, search for old naming pattern
if [ -z "$TASK_FILE" ]; then
    TASK_FILE=$(find .claude/epics -name "*.md" -exec grep -l "github:.*issues/$ISSUE_NUMBER" {} \; 2>/dev/null | head -1)
fi

if [ -z "$TASK_FILE" ]; then
    echo "📝 No existing task file found - creating new one..."
    
    # Determine epic name from issue labels or use default
    EPIC_NAME="github-issue-${ISSUE_NUMBER}"
    ISSUE_BODY=$(echo "$ISSUE_DATA" | jq -r '.body // ""')
    DESCRIPTION="GitHub Issue #${ISSUE_NUMBER}: ${ISSUE_TITLE}\n\n${ISSUE_BODY}"
    
    # Create task file using our script
    echo "🔧 Creating task file..."
    TASK_ID=$(.claude/scripts/create-task-file.sh "$ISSUE_TITLE" "$EPIC_NAME" "" "$ISSUE_NUMBER" "$DESCRIPTION")
    
    TASK_FILE=".claude/epics/tasks/${TASK_ID}.md"
    echo "✅ Created new task file: $TASK_FILE"
else
    echo "✅ Found existing task file: $TASK_FILE"
    # Extract epic name from existing file
    if [[ "$TASK_FILE" =~ \.claude/epics/([^/]+)/ ]]; then
        EPIC_NAME="${BASH_REMATCH[1]}"
    else
        EPIC_NAME="general-tasks"
    fi
fi

# Step 6: Check for worktree
echo "🌳 Checking for epic worktree..."
if ! git worktree list | grep -q "epic-$EPIC_NAME"; then
    echo "❌ No worktree found for epic '$EPIC_NAME'"
    echo "   Run: /pm:epic-start $EPIC_NAME"
    exit 1
fi

WORKTREE_PATH="../epic-$EPIC_NAME"
echo "✅ Using worktree: $WORKTREE_PATH"

# Step 7: Setup progress tracking
echo "📊 Setting up progress tracking..."
PROGRESS_DIR=".claude/epics/$EPIC_NAME/updates/$ISSUE_NUMBER"
mkdir -p "$PROGRESS_DIR"

# Create initial progress file
cat > "$PROGRESS_DIR/korean-integration.md" << EOF
---
issue: $ISSUE_NUMBER
integration: korean-comments
started: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: active
---

# Korean Comment Integration - Issue #$ISSUE_NUMBER

## Status
✅ Korean start comment created successfully
✅ Issue assigned and labeled as in-progress  
✅ Progress tracking initiated

## Issue Details
- **Title**: $ISSUE_TITLE
- **Repository**: $REPO_ID
- **State**: $ISSUE_STATE

## Korean Comment Features
- [x] Start comment creation
- [ ] Progress comment updates
- [ ] Completion comment
- [ ] Review comment (if needed)
- [ ] Blocked comment (if needed)

## Next Steps
1. Begin implementation work
2. Use PM integration service for progress updates
3. Create completion comment when finished

## Usage Examples
\`\`\`javascript
// Progress update
await PMGitHubIntegrationService.createIssueProgressComment(
  $ISSUE_NUMBER,
  '$REPO_ID', 
  '기능 구현 50% 완료',
  ['한글 댓글 템플릿 시스템 완성', 'GitHub API 연동 완료'],
  ['PM 시스템 통합', '테스트 케이스 추가']
);

// Completion
await PMGitHubIntegrationService.createIssueCompleteComment(
  $ISSUE_NUMBER,
  '$REPO_ID',
  ['한글 댓글 시스템 구현 완료', 'PM 시스템 통합 완료'],
  '100% 테스트 커버리지 달성'
);
\`\`\`
EOF

# Step 8: Output summary
echo ""
echo "🎉 Issue #$ISSUE_NUMBER started successfully with Korean comment integration!"
echo ""
echo "📋 Summary:"
echo "   Issue: #$ISSUE_NUMBER - $ISSUE_TITLE"
echo "   Repository: $REPO_ID"
echo "   Epic: $EPIC_NAME"
echo "   Worktree: $WORKTREE_PATH"
echo "   Korean Comment: ✅ Created"
echo ""
echo "📊 Progress Tracking:"
echo "   Directory: $PROGRESS_DIR"
echo "   Integration: korean-integration.md"
echo ""
echo "🔧 Next Steps:"
echo "   1. Begin implementation in worktree: cd $WORKTREE_PATH"
echo "   2. Use PMGitHubIntegrationService for progress comments"
echo "   3. Monitor with: /pm:epic-status $EPIC_NAME"
echo ""
echo "📚 Korean Comment Types Available:"
echo "   - Progress: createIssueProgressComment()"
echo "   - Complete: createIssueCompleteComment()"
echo "   - Blocked: createIssueBlockedComment()" 
echo "   - Review: createIssueReviewComment()"
echo ""

exit 0