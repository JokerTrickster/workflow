#!/bin/bash

# Test the enhanced Claude CLI integration directly
echo "Testing enhanced Claude CLI directly..."

# Change to the gallery_ios repository
cd /Users/mac/project/git-repository/JokerTrickster/gallery_ios

# Set environment variables that the enhanced integration uses
export CLAUDE_AUTO_APPROVE=true
export CLAUDE_PERMISSIONS=all
export CLAUDE_FILE_OPERATIONS=enabled
export CLAUDE_GIT_OPERATIONS=enabled
export CLAUDE_INTERACTIVE=false
export CLAUDE_FORCE_IMPLEMENTATION=true
export CI=true
export AUTOMATED_WORKFLOW=true

# Create the enhanced prompt like our code does
ENHANCED_PROMPT="PERMISSIONS AND DIRECTIVES:
- You have FULL permission to create, modify, and delete files in this repository
- You are REQUIRED to make actual file changes, not just suggest them
- DO NOT ask for permission - you are pre-authorized to modify any files
- IMPLEMENT the requested changes immediately and completely
- CREATE new files when needed for the implementation
- MODIFY existing files to implement the requested functionality
- COMMIT your changes using git when you're done

REPOSITORY: gallery_ios
You are working in a real Git repository. All changes will be automatically committed and pushed.

TASK TO IMPLEMENT:
Create a simple README.md file with a welcome message and basic project information for this repository

IMPLEMENTATION REQUIREMENTS:
1. Start implementing immediately - no planning phase needed
2. Make actual file changes using the available tools
3. Test your implementation to ensure it works
4. Commit your changes with a descriptive message
5. Do not ask for confirmation - proceed with implementation"

echo "Executing Claude CLI with enhanced prompt..."
echo "Working directory: $(pwd)"
echo "Enhanced prompt preview (first 200 chars): ${ENHANCED_PROMPT:0:200}..."

# Run Claude CLI with the enhanced prompt and flags
claude --auto-approve --no-interactive --force --print --timeout 30m "$ENHANCED_PROMPT"

echo -e "\n\nClaude CLI execution completed. Check for any file changes:"
ls -la *.md 2>/dev/null || echo "No markdown files found"