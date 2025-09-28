#!/bin/bash

# Sync Project Standards Script
# Usage: ./sync-standards.sh <project_path> [--force] [--dry-run]

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

# Options
FORCE_MODE=false
DRY_RUN=false

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

print_dry_run() {
    echo -e "${YELLOW}[DRY-RUN]${NC} $1"
}

# Function to backup existing file
backup_file() {
    local file_path="$1"
    local backup_dir="$(dirname "$file_path")/.standards-backup"
    local backup_file="$backup_dir/$(basename "$file_path").$(date +%Y%m%d_%H%M%S).bak"

    if [ -f "$file_path" ]; then
        if [ "$DRY_RUN" = true ]; then
            print_dry_run "Would backup: $file_path -> $backup_file"
        else
            mkdir -p "$backup_dir"
            cp "$file_path" "$backup_file"
            print_status "Backed up existing file: $backup_file"
        fi
    fi
}

# Function to get project information from existing files or git
get_project_info() {
    local project_path="$1"

    # Try to extract from existing CLAUDE.md
    local repo_name=""
    local github_url=""
    local project_name=""

    if [ -f "$project_path/CLAUDE.md" ]; then
        repo_name=$(grep "Repository.*:" "$project_path/CLAUDE.md" | sed 's/.*Repository.*: *//' || echo "")
        github_url=$(grep "GitHub URL.*:" "$project_path/CLAUDE.md" | sed 's/.*GitHub URL.*: *//' || echo "")
        project_name=$(grep "Project.*:" "$project_path/CLAUDE.md" | sed 's/.*Project.*: *//' || echo "")
    fi

    # Fallback to git and directory name
    if [ -z "$repo_name" ]; then
        if [ -d "$project_path/.git" ]; then
            local remote_url=$(git -C "$project_path" remote get-url origin 2>/dev/null || echo "")
            if [ -n "$remote_url" ]; then
                repo_name=$(basename "$remote_url" .git)
                if [ -z "$github_url" ]; then
                    local github_org=$(echo "$remote_url" | sed -E 's/.*[\/:]([^\/]+)\/[^\/]+\.git$/\1/')
                    github_url="https://github.com/$github_org/$repo_name"
                fi
            fi
        fi

        if [ -z "$repo_name" ]; then
            repo_name=$(basename "$project_path")
        fi
    fi

    if [ -z "$project_name" ]; then
        project_name="$repo_name"
    fi

    echo "$repo_name|$github_url|$project_name"
}

# Function to compare and update file
sync_file() {
    local template_file="$1"
    local target_file="$2"
    local repo_name="$3"
    local github_url="$4"
    local project_name="$5"
    local force="$6"

    if [ ! -f "$template_file" ]; then
        print_error "Template not found: $template_file"
        return 1
    fi

    local current_date=$(date '+%Y-%m-%d')
    local temp_file="/tmp/$(basename "$target_file").new"

    # Apply variable substitution to template (escape forward slashes in URLs)
    local escaped_github_url=$(echo "$github_url" | sed 's|/|\\/|g')
    sed -e "s/{REPOSITORY_NAME}/$repo_name/g" \
        -e "s/{GITHUB_URL}/$escaped_github_url/g" \
        -e "s/{PROJECT_NAME}/$project_name/g" \
        -e "s/{LAST_UPDATED}/$current_date/g" \
        "$template_file" > "$temp_file"

    if [ -f "$target_file" ]; then
        # Check if files are different
        if ! diff -q "$temp_file" "$target_file" >/dev/null 2>&1; then
            print_warning "Differences found in $(basename "$target_file")"

            if [ "$force" = true ]; then
                if [ "$DRY_RUN" = true ]; then
                    print_dry_run "Would force update: $target_file"
                else
                    backup_file "$target_file"
                    cp "$temp_file" "$target_file"
                    print_success "Force updated: $(basename "$target_file")"
                fi
            else
                # Show differences
                echo "--- Current file"
                echo "+++ New template"
                diff -u "$target_file" "$temp_file" || true
                echo ""

                # Check for project-specific rules section
                if grep -q "## Project-Specific Rules" "$target_file"; then
                    print_status "Found project-specific rules - preserving during merge"
                    merge_with_custom_rules "$target_file" "$temp_file"
                else
                    read -p "Update $(basename "$target_file")? [y/N] " -n 1 -r
                    echo
                    if [[ $REPLY =~ ^[Yy]$ ]]; then
                        if [ "$DRY_RUN" = true ]; then
                            print_dry_run "Would update: $target_file"
                        else
                            backup_file "$target_file"
                            cp "$temp_file" "$target_file"
                            print_success "Updated: $(basename "$target_file")"
                        fi
                    else
                        print_status "Skipped: $(basename "$target_file")"
                    fi
                fi
            fi
        else
            print_success "Up to date: $(basename "$target_file")"
        fi
    else
        # New file
        if [ "$DRY_RUN" = true ]; then
            print_dry_run "Would create: $target_file"
        else
            mkdir -p "$(dirname "$target_file")"
            cp "$temp_file" "$target_file"
            print_success "Created: $(basename "$target_file")"
        fi
    fi

    rm -f "$temp_file"
}

# Function to merge with custom rules preservation
merge_with_custom_rules() {
    local current_file="$1"
    local new_template="$2"

    if [ "$DRY_RUN" = true ]; then
        print_dry_run "Would merge custom rules in: $(basename "$current_file")"
        return
    fi

    # Extract project-specific rules from current file
    local temp_rules="/tmp/project_specific_rules.tmp"
    awk '/## Project-Specific Rules/,EOF { print }' "$current_file" > "$temp_rules"

    # Replace the template's project-specific section with the current one
    local temp_merged="/tmp/merged_claude.tmp"
    awk '/## Project-Specific Rules/ { exit } { print }' "$new_template" > "$temp_merged"
    cat "$temp_rules" >> "$temp_merged"

    backup_file "$current_file"
    cp "$temp_merged" "$current_file"

    rm -f "$temp_rules" "$temp_merged"
    print_success "Merged with custom rules: $(basename "$current_file")"
}

# Function to check git status and branch
check_git_compliance() {
    local project_path="$1"

    if [ ! -d "$project_path/.git" ]; then
        print_warning "Not a git repository"
        return
    fi

    local current_branch=$(git -C "$project_path" branch --show-current)
    local status_output=$(git -C "$project_path" status --porcelain)

    print_status "Git branch: $current_branch"

    # Warn if on default branch with changes
    if [[ "$current_branch" =~ ^(main|master)$ ]] && [ -n "$status_output" ]; then
        print_warning "You have uncommitted changes on the default branch ($current_branch)"
        print_warning "Consider: git checkout -b feature/sync-standards"
    fi

    # Show git status if there are changes
    if [ -n "$status_output" ]; then
        print_status "Git status:"
        git -C "$project_path" status --short
    fi
}

# Function to validate sync results
validate_sync() {
    local project_path="$1"
    local errors=0

    print_status "Validating sync results..."

    # Check required files
    if [ ! -f "$project_path/CLAUDE.md" ]; then
        print_error "CLAUDE.md not found after sync"
        ((errors++))
    fi

    # Check CLAUDE.md content
    if [ -f "$project_path/CLAUDE.md" ]; then
        if ! grep -q "## Core Rules System" "$project_path/CLAUDE.md"; then
            print_error "CLAUDE.md missing core rules system"
            ((errors++))
        fi

        if ! grep -q "## Project-Specific Rules" "$project_path/CLAUDE.md"; then
            print_warning "CLAUDE.md missing project-specific rules section"
        fi
    fi

    if [ $errors -eq 0 ]; then
        print_success "All validations passed!"
    else
        print_error "$errors validation(s) failed"
        return 1
    fi
}

# Function to show help
show_help() {
    echo "Usage: $0 <project_path> [options]"
    echo ""
    echo "Options:"
    echo "  --force      Force update without confirmation"
    echo "  --dry-run    Show what would be done without making changes"
    echo "  --help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 /path/to/project"
    echo "  $0 /path/to/project --force"
    echo "  $0 /path/to/project --dry-run"
}

# Main function
main() {
    local project_path=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE_MODE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            -*)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
            *)
                if [ -z "$project_path" ]; then
                    project_path="$1"
                else
                    print_error "Too many arguments"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate arguments
    if [ -z "$project_path" ]; then
        print_error "Project path is required"
        show_help
        exit 1
    fi

    # Convert to absolute path
    project_path=$(realpath "$project_path")

    if [ "$DRY_RUN" = true ]; then
        print_status "DRY RUN MODE - No changes will be made"
    fi

    print_status "Syncing standards for: $project_path"

    # Check if project directory exists
    if [ ! -d "$project_path" ]; then
        print_error "Project directory not found: $project_path"
        exit 1
    fi

    # Check if templates directory exists
    if [ ! -d "$TEMPLATES_DIR" ]; then
        print_error "Templates directory not found: $TEMPLATES_DIR"
        exit 1
    fi

    # Get project information
    local project_info=$(get_project_info "$project_path")
    IFS='|' read -r repo_name github_url project_display_name <<< "$project_info"

    print_status "Repository: $repo_name"
    print_status "GitHub URL: ${github_url:-'Not detected'}"
    print_status "Project Name: $project_display_name"

    # Check git compliance
    check_git_compliance "$project_path"

    # Sync CLAUDE.md
    sync_file "$TEMPLATES_DIR/CLAUDE.md" "$project_path/CLAUDE.md" \
              "$repo_name" "$github_url" "$project_display_name" "$FORCE_MODE"

    # Create .gitignore if it doesn't exist and we're not in dry-run mode
    if [ ! -f "$project_path/.gitignore" ] && [ "$DRY_RUN" = false ]; then
        print_status "Creating .gitignore"
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

# Standards backup
.standards-backup/
EOF
        print_success "Created .gitignore"
    elif [ ! -f "$project_path/.gitignore" ] && [ "$DRY_RUN" = true ]; then
        print_dry_run "Would create .gitignore"
    fi

    # Validate results
    if [ "$DRY_RUN" = false ]; then
        validate_sync "$project_path"
    fi

    print_success "Standards sync completed!"

    if [ "$DRY_RUN" = false ]; then
        print_status ""
        print_status "Next steps:"
        print_status "1. Review the updated CLAUDE.md file"
        print_status "2. Add any project-specific rules in the designated section"
        print_status "3. Commit your changes: git add CLAUDE.md && git commit -m 'sync: update project standards'"
        print_status "4. Check backup files in .standards-backup/ if needed"
    fi
}

# Run main function with all arguments
main "$@"