#!/bin/bash

# Project Standards Validation Tool
# Usage: ./validate-standards.sh <project_path> [--fix] [--report]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Options
FIX_MODE=false
REPORT_MODE=false
VALIDATION_ERRORS=0
VALIDATION_WARNINGS=0

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    ((VALIDATION_WARNINGS++))
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((VALIDATION_ERRORS++))
}

print_fix() {
    echo -e "${GREEN}[FIXED]${NC} $1"
}

# Function to validate CLAUDE.md existence and structure
validate_claude_md() {
    local project_path="$1"
    local claude_file="$project_path/CLAUDE.md"

    print_status "Validating CLAUDE.md..."

    if [ ! -f "$claude_file" ]; then
        print_error "CLAUDE.md not found"
        if [ "$FIX_MODE" = true ]; then
            print_status "Run: ./init-project.sh $project_path to create CLAUDE.md"
        fi
        return 1
    fi

    # Check required sections
    local required_sections=(
        "## Core Rules System"
        "## Base Rules"
        "## Project-Specific Rules"
        "## Validation Checklist"
    )

    local missing_sections=()
    for section in "${required_sections[@]}"; do
        if ! grep -q "$section" "$claude_file"; then
            missing_sections+=("$section")
        fi
    done

    if [ ${#missing_sections[@]} -eq 0 ]; then
        print_success "CLAUDE.md structure is valid"
    else
        for section in "${missing_sections[@]}"; do
            print_error "Missing section: $section"
        done
        if [ "$FIX_MODE" = true ]; then
            print_status "Run: ./sync-standards.sh $project_path to fix structure"
        fi
    fi

    # Check for project information
    if ! grep -q "Repository.*:" "$claude_file"; then
        print_warning "Repository information missing in CLAUDE.md"
    fi

    if ! grep -q "GitHub URL.*:" "$claude_file"; then
        print_warning "GitHub URL missing in CLAUDE.md"
    fi
}

# Function to validate Git workflow compliance
validate_git_workflow() {
    local project_path="$1"

    print_status "Validating Git workflow..."

    if [ ! -d "$project_path/.git" ]; then
        print_error "Not a Git repository"
        if [ "$FIX_MODE" = true ]; then
            print_status "Initializing Git repository..."
            git -C "$project_path" init
            git -C "$project_path" branch -M main
            print_fix "Initialized Git repository"
        fi
        return 1
    fi

    # Check current branch
    local current_branch=$(git -C "$project_path" branch --show-current)
    local status_output=$(git -C "$project_path" status --porcelain)

    if [[ "$current_branch" =~ ^(main|master)$ ]] && [ -n "$status_output" ]; then
        print_warning "Working on default branch ($current_branch) with uncommitted changes"
        print_status "Consider: git checkout -b feature/your-feature-name"
    fi

    # Check for .gitignore
    if [ ! -f "$project_path/.gitignore" ]; then
        print_warning ".gitignore not found"
        if [ "$FIX_MODE" = true ]; then
            print_status "Creating basic .gitignore..."
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
            print_fix "Created .gitignore"
        fi
    else
        print_success "Git workflow setup is valid"
    fi
}

# Function to validate code quality
validate_code_quality() {
    local project_path="$1"

    print_status "Validating code quality..."

    # Check for TODO/FIXME comments
    local todos=$(find "$project_path" -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.go" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" | \
                  xargs grep -n "TODO\|FIXME\|XXX\|HACK" 2>/dev/null || true)

    if [ -n "$todos" ]; then
        print_warning "Found TODO/FIXME comments:"
        echo "$todos" | head -5
        if [ $(echo "$todos" | wc -l) -gt 5 ]; then
            echo "... and $(( $(echo "$todos" | wc -l) - 5 )) more"
        fi
    else
        print_success "No TODO/FIXME comments found"
    fi

    # Check for debugging statements
    local debug_statements=$(find "$project_path" -name "*.js" -o -name "*.ts" -o -name "*.py" | \
                            xargs grep -n "console\.log\|print(\|debugger\|var_dump" 2>/dev/null || true)

    if [ -n "$debug_statements" ]; then
        print_warning "Found debugging statements:"
        echo "$debug_statements" | head -3
        if [ $(echo "$debug_statements" | wc -l) -gt 3 ]; then
            echo "... and $(( $(echo "$debug_statements" | wc -l) - 3 )) more"
        fi
    fi

    # Check for temporary files
    local temp_files=$(find "$project_path" -name "*.tmp" -o -name "*.temp" -o -name "*~" -o -name ".DS_Store" 2>/dev/null || true)

    if [ -n "$temp_files" ]; then
        print_warning "Found temporary files:"
        echo "$temp_files"
        if [ "$FIX_MODE" = true ]; then
            print_status "Removing temporary files..."
            echo "$temp_files" | xargs rm -f
            print_fix "Removed temporary files"
        fi
    fi
}

# Function to validate file organization
validate_file_organization() {
    local project_path="$1"

    print_status "Validating file organization..."

    # Check for misplaced test files
    local misplaced_tests=$(find "$project_path" -name "*.test.*" -not -path "*/tests/*" -not -path "*/__tests__/*" -not -path "*/test/*" 2>/dev/null || true)

    if [ -n "$misplaced_tests" ]; then
        print_warning "Found test files outside test directories:"
        echo "$misplaced_tests"
        print_status "Consider moving to tests/, __tests__/, or test/ directories"
    fi

    # Check for README.md
    if [ ! -f "$project_path/README.md" ]; then
        print_warning "README.md not found"
        if [ "$FIX_MODE" = true ]; then
            local project_name=$(basename "$project_path")
            print_status "Creating basic README.md..."
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
            print_fix "Created basic README.md"
        fi
    fi
}

# Function to validate security
validate_security() {
    local project_path="$1"

    print_status "Validating security..."

    # Check for potential secrets in code
    local secret_patterns=(
        "password\s*=\s*['\"][^'\"]{8,}"
        "api[_-]?key\s*[=:]\s*['\"][^'\"]{8,}"
        "secret\s*[=:]\s*['\"][^'\"]{8,}"
        "token\s*[=:]\s*['\"][^'\"]{8,}"
    )

    local secrets_found=false
    for pattern in "${secret_patterns[@]}"; do
        local matches=$(find "$project_path" -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.go" | \
                       xargs grep -i -E "$pattern" 2>/dev/null || true)
        if [ -n "$matches" ]; then
            if [ "$secrets_found" = false ]; then
                print_warning "Potential secrets found in code:"
                secrets_found=true
            fi
            echo "$matches" | head -3
        fi
    done

    if [ "$secrets_found" = false ]; then
        print_success "No obvious secrets found in code"
    fi

    # Check if .env is in .gitignore
    if [ -f "$project_path/.gitignore" ]; then
        if ! grep -q "\.env" "$project_path/.gitignore"; then
            print_warning ".env files not in .gitignore"
            if [ "$FIX_MODE" = true ]; then
                echo ".env" >> "$project_path/.gitignore"
                echo ".env.local" >> "$project_path/.gitignore"
                print_fix "Added .env to .gitignore"
            fi
        fi
    fi
}

# Function to check package.json scripts
validate_package_scripts() {
    local project_path="$1"
    local package_json="$project_path/package.json"

    if [ -f "$package_json" ]; then
        print_status "Validating package.json scripts..."

        # Check for common quality scripts
        local required_scripts=("lint" "format" "test")
        local missing_scripts=()

        for script in "${required_scripts[@]}"; do
            if ! grep -q "\"$script\"" "$package_json"; then
                missing_scripts+=("$script")
            fi
        done

        if [ ${#missing_scripts[@]} -eq 0 ]; then
            print_success "Package.json has required scripts"
        else
            for script in "${missing_scripts[@]}"; do
                print_warning "Missing npm script: $script"
            done
        fi
    fi
}

# Function to generate validation report
generate_report() {
    local project_path="$1"
    local report_file="$project_path/standards-validation-report.md"

    if [ "$REPORT_MODE" = true ]; then
        print_status "Generating validation report..."

        cat > "$report_file" << EOF
# Project Standards Validation Report

Generated: $(date)
Project: $(basename "$project_path")

## Summary
- Errors: $VALIDATION_ERRORS
- Warnings: $VALIDATION_WARNINGS

## Validation Results

### CLAUDE.md Structure
$(validate_claude_md "$project_path" 2>&1 || true)

### Git Workflow
$(validate_git_workflow "$project_path" 2>&1 || true)

### Code Quality
$(validate_code_quality "$project_path" 2>&1 || true)

### File Organization
$(validate_file_organization "$project_path" 2>&1 || true)

### Security
$(validate_security "$project_path" 2>&1 || true)

## Recommendations

1. Review and fix all ERROR items immediately
2. Address WARNING items when possible
3. Run validation regularly: \`./validate-standards.sh $project_path\`
4. Use \`--fix\` flag to auto-fix some issues: \`./validate-standards.sh $project_path --fix\`

## Next Steps

- [ ] Fix all validation errors
- [ ] Address validation warnings
- [ ] Set up GitHub Actions workflow
- [ ] Review project-specific rules in CLAUDE.md
EOF

        print_success "Report generated: $report_file"
    fi
}

# Function to show help
show_help() {
    echo "Usage: $0 <project_path> [options]"
    echo ""
    echo "Options:"
    echo "  --fix        Auto-fix issues where possible"
    echo "  --report     Generate a detailed validation report"
    echo "  --help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 /path/to/project"
    echo "  $0 /path/to/project --fix"
    echo "  $0 /path/to/project --report"
    echo "  $0 /path/to/project --fix --report"
}

# Main function
main() {
    local project_path=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --fix)
                FIX_MODE=true
                shift
                ;;
            --report)
                REPORT_MODE=true
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

    print_status "Validating project standards for: $project_path"

    if [ "$FIX_MODE" = true ]; then
        print_status "Auto-fix mode enabled"
    fi

    # Check if project directory exists
    if [ ! -d "$project_path" ]; then
        print_error "Project directory not found: $project_path"
        exit 1
    fi

    # Run validations
    validate_claude_md "$project_path"
    validate_git_workflow "$project_path"
    validate_code_quality "$project_path"
    validate_file_organization "$project_path"
    validate_security "$project_path"
    validate_package_scripts "$project_path"

    # Generate report if requested
    generate_report "$project_path"

    # Summary
    echo ""
    print_status "Validation Summary:"
    if [ $VALIDATION_ERRORS -eq 0 ]; then
        print_success "✅ No errors found"
    else
        print_error "❌ $VALIDATION_ERRORS error(s) found"
    fi

    if [ $VALIDATION_WARNINGS -eq 0 ]; then
        print_success "✅ No warnings"
    else
        print_warning "⚠️  $VALIDATION_WARNINGS warning(s) found"
    fi

    # Exit with appropriate code
    if [ $VALIDATION_ERRORS -gt 0 ]; then
        exit 1
    fi
}

# Run main function with all arguments
main "$@"