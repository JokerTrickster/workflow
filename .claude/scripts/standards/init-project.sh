#!/bin/bash

# Project Standards Initialization Script
# Usage: ./init-project.sh <project_path> [project_name] [github_org]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")/templates"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get project information
get_project_info() {
    local project_path="$1"
    local project_name="${2:-$(basename "$project_path")}"
    local github_org="${3:-}"

    # Get repository name from git remote or directory name
    local repo_name="$project_name"
    if [ -d "$project_path/.git" ]; then
        local remote_url=$(git -C "$project_path" remote get-url origin 2>/dev/null || echo "")
        if [ -n "$remote_url" ]; then
            repo_name=$(basename "$remote_url" .git)
            if [ -z "$github_org" ]; then
                github_org=$(echo "$remote_url" | sed -E 's/.*[\/:]([^\/]+)\/[^\/]+\.git$/\1/')
            fi
        fi
    fi

    # Construct GitHub URL
    local github_url=""
    if [ -n "$github_org" ] && [ -n "$repo_name" ]; then
        github_url="https://github.com/$github_org/$repo_name"
    fi

    echo "$repo_name|$github_url|$project_name"
}

# Function to apply template with variable substitution
apply_template() {
    local template_file="$1"
    local target_file="$2"
    local repo_name="$3"
    local github_url="$4"
    local project_name="$5"
    local current_date=$(date '+%Y-%m-%d')

    print_status "Applying template $template_file -> $target_file"

    # Create target directory if it doesn't exist
    mkdir -p "$(dirname "$target_file")"

    # Apply variable substitution (escape forward slashes in URLs)
    local escaped_github_url=$(echo "$github_url" | sed 's|/|\\/|g')
    sed -e "s/{REPOSITORY_NAME}/$repo_name/g" \
        -e "s/{GITHUB_URL}/$escaped_github_url/g" \
        -e "s/{PROJECT_NAME}/$project_name/g" \
        -e "s/{LAST_UPDATED}/$current_date/g" \
        "$template_file" > "$target_file"

    print_success "Template applied: $(basename "$target_file")"
}

# Function to setup Git workflow
setup_git_workflow() {
    local project_path="$1"

    if [ ! -d "$project_path/.git" ]; then
        print_warning "Not a git repository. Initializing git..."
        git -C "$project_path" init
        git -C "$project_path" branch -M main
    fi

    # Check current branch
    local current_branch=$(git -C "$project_path" branch --show-current)
    local default_branch=$(git -C "$project_path" symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")

    print_status "Current branch: $current_branch"
    print_status "Default branch: $default_branch"

    # Warn if working on default branch
    if [ "$current_branch" = "$default_branch" ] || [ "$current_branch" = "main" ] || [ "$current_branch" = "master" ]; then
        print_warning "You are on the default branch ($current_branch)"
        print_warning "Standards require feature branches. Consider: git checkout -b feature/setup-standards"
    fi

    # Create .gitignore if it doesn't exist
    if [ ! -f "$project_path/.gitignore" ]; then
        print_status "Creating basic .gitignore"
        cat > "$project_path/.gitignore" << 'EOF'
# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
*.o
*.so
*.exe

# Environment variables
.env
.env.local
.env.*.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
*.tmp
*.temp
EOF
        print_success "Created .gitignore"
    fi
}

# Function to create directory structure
create_project_structure() {
    local project_path="$1"

    print_status "Creating standard project structure"

    # Create common directories
    mkdir -p "$project_path"/{docs,tests,scripts}

    # Create .claude directory if it doesn't exist
    if [ ! -d "$project_path/.claude" ]; then
        mkdir -p "$project_path/.claude"
        print_success "Created .claude directory"
    fi

    # Create basic README if it doesn't exist
    if [ ! -f "$project_path/README.md" ]; then
        local project_name=$(basename "$project_path")
        cat > "$project_path/README.md" << EOF
# $project_name

## Overview
[Brief description of the project]

## Setup
1. Clone the repository
2. Install dependencies
3. Run the project

## Development Standards
This project follows standardized development rules defined in \`CLAUDE.md\`.

## Contributing
1. Create a feature branch from main
2. Make your changes
3. Ensure all tests pass
4. Create a pull request
EOF
        print_success "Created basic README.md"
    fi
}

# Function to validate installation
validate_installation() {
    local project_path="$1"

    print_status "Validating installation..."

    local errors=0

    # Check required files
    if [ ! -f "$project_path/CLAUDE.md" ]; then
        print_error "CLAUDE.md not found"
        ((errors++))
    fi

    if [ ! -f "$project_path/.gitignore" ]; then
        print_error ".gitignore not found"
        ((errors++))
    fi

    if [ ! -d "$project_path/.git" ]; then
        print_error "Git repository not initialized"
        ((errors++))
    fi

    if [ $errors -eq 0 ]; then
        print_success "All validations passed!"
    else
        print_error "$errors validation(s) failed"
        return 1
    fi
}

# Main function
main() {
    local project_path="$1"
    local project_name="$2"
    local github_org="$3"

    # Validate arguments
    if [ -z "$project_path" ]; then
        print_error "Usage: $0 <project_path> [project_name] [github_org]"
        exit 1
    fi

    # Convert to absolute path
    project_path=$(realpath "$project_path")

    print_status "Initializing project standards for: $project_path"

    # Check if templates directory exists
    if [ ! -d "$TEMPLATES_DIR" ]; then
        print_error "Templates directory not found: $TEMPLATES_DIR"
        exit 1
    fi

    # Get project information
    local project_info=$(get_project_info "$project_path" "$project_name" "$github_org")
    IFS='|' read -r repo_name github_url project_display_name <<< "$project_info"

    print_status "Repository: $repo_name"
    print_status "GitHub URL: ${github_url:-'Not available'}"
    print_status "Project Name: $project_display_name"

    # Create project directory if it doesn't exist
    if [ ! -d "$project_path" ]; then
        print_status "Creating project directory: $project_path"
        mkdir -p "$project_path"
    fi

    # Setup project structure
    create_project_structure "$project_path"

    # Apply CLAUDE.md template
    if [ -f "$TEMPLATES_DIR/CLAUDE.md" ]; then
        apply_template "$TEMPLATES_DIR/CLAUDE.md" "$project_path/CLAUDE.md" \
                      "$repo_name" "$github_url" "$project_display_name"
    else
        print_error "CLAUDE.md template not found"
        exit 1
    fi

    # Setup Git workflow
    setup_git_workflow "$project_path"

    # Validate installation
    validate_installation "$project_path"

    print_success "Project standards initialization completed!"
    print_status ""
    print_status "Next steps:"
    print_status "1. Review and customize CLAUDE.md for your project needs"
    print_status "2. Add project-specific rules in the 'Project-Specific Rules' section"
    print_status "3. Create a feature branch: git checkout -b feature/your-feature"
    print_status "4. Commit your changes and create a pull request"

    if [ -n "$github_url" ]; then
        print_status "5. Visit your repository: $github_url"
    fi
}

# Run main function with all arguments
main "$@"