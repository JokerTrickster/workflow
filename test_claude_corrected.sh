#!/bin/bash

# Test the corrected enhanced Claude CLI integration
echo "Testing corrected enhanced Claude CLI..."

# Change to a test directory in gallery_ios
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

# Create the enhanced prompt like our updated code does
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
Create a simple test file called 'claude_test.txt' with today's date and a message confirming Claude's file modification capabilities

IMPLEMENTATION REQUIREMENTS:
1. Start implementing immediately - no planning phase needed
2. Make actual file changes using the available tools
3. Test your implementation to ensure it works
4. Commit your changes with a descriptive message
5. Do not ask for confirmation - proceed with implementation"

echo "Working directory: $(pwd)"
echo "Files before execution:"
ls -la *.txt 2>/dev/null || echo "No txt files found"

echo -e "\nExecuting Claude CLI with corrected flags..."

# Run Claude CLI with the corrected flags
claude --dangerously-skip-permissions --permission-mode bypassPermissions --print "$ENHANCED_PROMPT"

echo -e "\nFiles after execution:"
ls -la *.txt 2>/dev/null || echo "No txt files found"

# Check git status to see if any changes were made
echo -e "\nGit status:"
git status --porcelain 2>/dev/null || echo "Git status check failed"