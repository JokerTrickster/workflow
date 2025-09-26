#!/bin/bash

# Project Standards CLI - Unified interface for all standards tools
# Usage: ./standards-cli.sh <command> [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to print colored output
print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}  Project Standards CLI${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show available commands
show_commands() {
    print_header
    echo "Available commands:"
    echo ""
    echo "  init <path> [name] [org]    Initialize new project with standards"
    echo "  sync <path> [--force]       Sync existing project with latest standards"
    echo "  validate <path> [--fix]     Validate project compliance"
    echo "  status <path>               Show project standards status"
    echo "  list                        List all managed projects"
    echo "  update                      Update standards templates"
    echo ""
    echo "Examples:"
    echo "  $0 init /path/to/new-project my-app my-org"
    echo "  $0 sync /path/to/existing-project"
    echo "  $0 validate /path/to/project --fix"
    echo "  $0 status /path/to/project"
    echo ""
    echo "Options:"
    echo "  --help, -h                  Show this help message"
    echo "  --version, -v               Show version information"
}

# Function to show version
show_version() {
    echo "Project Standards CLI v1.0.0"
    echo "Built for workflow project management"
}

# Function to initialize a new project
cmd_init() {
    local project_path="$1"
    local project_name="$2"
    local github_org="$3"

    if [ -z "$project_path" ]; then
        print_error "Project path is required"
        echo "Usage: $0 init <path> [name] [org]"
        exit 1
    fi

    print_status "Initializing project: $project_path"
    "$SCRIPT_DIR/init-project.sh" "$project_path" "$project_name" "$github_org"
}

# Function to sync existing project
cmd_sync() {
    local project_path="$1"
    local force_flag="$2"

    if [ -z "$project_path" ]; then
        print_error "Project path is required"
        echo "Usage: $0 sync <path> [--force]"
        exit 1
    fi

    print_status "Syncing project: $project_path"
    if [ "$force_flag" = "--force" ]; then
        "$SCRIPT_DIR/sync-standards.sh" "$project_path" --force
    else
        "$SCRIPT_DIR/sync-standards.sh" "$project_path"
    fi
}

# Function to validate project
cmd_validate() {
    local project_path="$1"
    local fix_flag="$2"

    if [ -z "$project_path" ]; then
        print_error "Project path is required"
        echo "Usage: $0 validate <path> [--fix]"
        exit 1
    fi

    print_status "Validating project: $project_path"
    if [ "$fix_flag" = "--fix" ]; then
        "$SCRIPT_DIR/validate-standards.sh" "$project_path" --fix --report
    else
        "$SCRIPT_DIR/validate-standards.sh" "$project_path" --report
    fi
}

# Function to show project status
cmd_status() {
    local project_path="$1"

    if [ -z "$project_path" ]; then
        print_error "Project path is required"
        echo "Usage: $0 status <path>"
        exit 1
    fi

    project_path=$(realpath "$project_path")
    print_status "Project Status: $(basename "$project_path")"
    echo "Path: $project_path"
    echo ""

    # Check basic compliance
    if [ -f "$project_path/CLAUDE.md" ]; then
        echo "✅ CLAUDE.md exists"

        # Check last updated
        local last_updated=$(grep "Last updated:" "$project_path/CLAUDE.md" 2>/dev/null | sed 's/.*Last updated: *//' || echo "Unknown")
        echo "📅 Last updated: $last_updated"
    else
        echo "❌ CLAUDE.md missing"
    fi

    if [ -d "$project_path/.git" ]; then
        echo "✅ Git repository"
        local current_branch=$(git -C "$project_path" branch --show-current 2>/dev/null || echo "unknown")
        echo "🌿 Current branch: $current_branch"
    else
        echo "❌ Not a Git repository"
    fi

    if [ -f "$project_path/.gitignore" ]; then
        echo "✅ .gitignore exists"
    else
        echo "❌ .gitignore missing"
    fi

    if [ -f "$project_path/README.md" ]; then
        echo "✅ README.md exists"
    else
        echo "⚠️  README.md missing"
    fi

    # Quick validation
    echo ""
    print_status "Running quick validation..."
    "$SCRIPT_DIR/validate-standards.sh" "$project_path" 2>/dev/null || true
}

# Function to list managed projects
cmd_list() {
    print_status "Searching for managed projects..."

    # Look for projects with CLAUDE.md in common directories
    local common_dirs=("/Users/$USER/project" "/Users/$USER/projects" "/Users/$USER/workspace" "/Users/$USER/dev")
    local found_projects=()

    for dir in "${common_dirs[@]}"; do
        if [ -d "$dir" ]; then
            while IFS= read -r -d '' claude_file; do
                local project_dir=$(dirname "$claude_file")
                if grep -q "Project Standards Template\|standards system" "$claude_file" 2>/dev/null; then
                    found_projects+=("$project_dir")
                fi
            done < <(find "$dir" -name "CLAUDE.md" -type f -print0 2>/dev/null)
        fi
    done

    if [ ${#found_projects[@]} -eq 0 ]; then
        print_status "No managed projects found"
        echo "Initialize a project with: $0 init /path/to/project"
    else
        echo "Found ${#found_projects[@]} managed project(s):"
        echo ""
        for project in "${found_projects[@]}"; do
            local project_name=$(basename "$project")
            local last_updated="Unknown"
            if [ -f "$project/CLAUDE.md" ]; then
                last_updated=$(grep "Last updated:" "$project/CLAUDE.md" 2>/dev/null | sed 's/.*Last updated: *//' || echo "Unknown")
            fi
            echo "  📁 $project_name ($project)"
            echo "     📅 Last updated: $last_updated"
            echo ""
        done
    fi
}

# Function to update standards templates
cmd_update() {
    print_status "Updating standards templates..."

    # Check if we're in the workflow project
    local templates_dir="$SCRIPT_DIR/../../templates"
    if [ ! -d "$templates_dir" ]; then
        print_error "Templates directory not found. Are you running from the workflow project?"
        exit 1
    fi

    # Check for updates (this would normally pull from git)
    if [ -d "$SCRIPT_DIR/../../../.git" ]; then
        print_status "Checking for template updates..."
        local current_dir=$(pwd)
        cd "$SCRIPT_DIR/../../.."

        local changes=$(git status --porcelain .claude/templates/ 2>/dev/null || echo "")
        if [ -n "$changes" ]; then
            print_status "Local changes detected in templates:"
            git status --short .claude/templates/
        else
            print_success "Templates are up to date"
        fi

        cd "$current_dir"
    else
        print_status "Not in a git repository, skipping update check"
    fi

    print_success "Update check completed"
}

# Function to run interactive mode
interactive_mode() {
    print_header
    echo "Interactive mode - Select an action:"
    echo ""
    echo "1) Initialize new project"
    echo "2) Sync existing project"
    echo "3) Validate project"
    echo "4) Show project status"
    echo "5) List managed projects"
    echo "6) Exit"
    echo ""
    read -p "Enter your choice [1-6]: " choice

    case $choice in
        1)
            read -p "Enter project path: " project_path
            read -p "Enter project name (optional): " project_name
            read -p "Enter GitHub organization (optional): " github_org
            cmd_init "$project_path" "$project_name" "$github_org"
            ;;
        2)
            read -p "Enter project path: " project_path
            read -p "Force sync? [y/N]: " force
            if [[ $force =~ ^[Yy]$ ]]; then
                cmd_sync "$project_path" "--force"
            else
                cmd_sync "$project_path"
            fi
            ;;
        3)
            read -p "Enter project path: " project_path
            read -p "Auto-fix issues? [y/N]: " fix
            if [[ $fix =~ ^[Yy]$ ]]; then
                cmd_validate "$project_path" "--fix"
            else
                cmd_validate "$project_path"
            fi
            ;;
        4)
            read -p "Enter project path: " project_path
            cmd_status "$project_path"
            ;;
        5)
            cmd_list
            ;;
        6)
            echo "Goodbye!"
            exit 0
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

# Main function
main() {
    local command="$1"

    case "$command" in
        init)
            shift
            cmd_init "$@"
            ;;
        sync)
            shift
            cmd_sync "$@"
            ;;
        validate)
            shift
            cmd_validate "$@"
            ;;
        status)
            shift
            cmd_status "$@"
            ;;
        list)
            cmd_list
            ;;
        update)
            cmd_update
            ;;
        --help|-h)
            show_commands
            ;;
        --version|-v)
            show_version
            ;;
        "")
            interactive_mode
            ;;
        *)
            print_error "Unknown command: $command"
            show_commands
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"