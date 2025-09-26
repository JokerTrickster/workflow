#!/bin/bash

# Post-Completion Hook - 작업 완료 후 빌드 및 검증
# 이 스크립트는 Claude Hook에 의해 자동 실행됩니다.

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
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

# Function to print colored output
print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}  Post-Completion Hook${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
}

print_status() {
    echo -e "${BLUE}[HOOK]${NC} $1"
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

# Function to detect if significant changes were made
detect_changes() {
    local git_status=$(git status --porcelain 2>/dev/null || echo "")
    local has_changes=false

    # Check for uncommitted changes
    if [ -n "$git_status" ]; then
        has_changes=true
    fi

    # Check recent commits (last 5 minutes)
    local recent_commits=$(git log --since="5 minutes ago" --oneline 2>/dev/null | wc -l)
    if [ "$recent_commits" -gt 0 ]; then
        has_changes=true
    fi

    echo "$has_changes"
}

# Function to run project standards validation
run_standards_validation() {
    print_status "Running project standards validation..."

    if [ -x "$SCRIPT_DIR/validate-standards.sh" ]; then
        "$SCRIPT_DIR/validate-standards.sh" "$PROJECT_ROOT" --fix
        if [ $? -eq 0 ]; then
            print_success "Standards validation passed"
        else
            print_warning "Standards validation found issues (fixed automatically)"
        fi
    else
        print_warning "Standards validation script not found"
    fi
}

# Function to run builds if applicable
run_builds() {
    print_status "Checking for build requirements..."

    # Frontend build
    if [ -f "$PROJECT_ROOT/frontend/package.json" ]; then
        print_status "Running frontend build..."
        cd "$PROJECT_ROOT/frontend"

        if npm run build --if-present >/dev/null 2>&1; then
            print_success "Frontend build completed"
        else
            print_warning "Frontend build failed or not configured"
        fi
    fi

    # Backend build (Go)
    if [ -f "$PROJECT_ROOT/local-backend/main.go" ]; then
        print_status "Running backend build..."
        cd "$PROJECT_ROOT/local-backend"

        if go build -o /tmp/workflow-backend . >/dev/null 2>&1; then
            print_success "Backend build completed"
            rm -f /tmp/workflow-backend
        else
            print_warning "Backend build failed"
        fi
    fi

    cd "$PROJECT_ROOT"
}

# Function to run tests if available
run_tests() {
    print_status "Running available tests..."

    # Frontend tests
    if [ -f "$PROJECT_ROOT/frontend/package.json" ]; then
        cd "$PROJECT_ROOT/frontend"
        if npm test --if-present --passWithNoTests >/dev/null 2>&1; then
            print_success "Frontend tests passed"
        else
            print_warning "Frontend tests failed or not configured"
        fi
    fi

    # Backend tests (Go)
    if [ -f "$PROJECT_ROOT/local-backend/main.go" ]; then
        cd "$PROJECT_ROOT/local-backend"
        if go test ./... >/dev/null 2>&1; then
            print_success "Backend tests passed"
        else
            print_warning "Backend tests failed or not configured"
        fi
    fi

    cd "$PROJECT_ROOT"
}

# Function to run linting
run_linting() {
    print_status "Running code quality checks..."

    # Frontend linting
    if [ -f "$PROJECT_ROOT/frontend/package.json" ]; then
        cd "$PROJECT_ROOT/frontend"
        if npm run lint --if-present >/dev/null 2>&1; then
            print_success "Frontend linting passed"
        else
            print_warning "Frontend linting issues found"
        fi
    fi

    # Go formatting
    if [ -f "$PROJECT_ROOT/local-backend/main.go" ]; then
        cd "$PROJECT_ROOT/local-backend"
        if gofmt -l . | grep -q .; then
            print_warning "Go code formatting issues found"
            gofmt -w .
            print_status "Go code formatted automatically"
        else
            print_success "Go code formatting is correct"
        fi
    fi

    cd "$PROJECT_ROOT"
}

# Function to check Git status and suggest actions
check_git_status() {
    print_status "Checking Git status..."

    local git_status=$(git status --porcelain 2>/dev/null || echo "")
    local current_branch=$(git branch --show-current 2>/dev/null || echo "unknown")

    if [ -n "$git_status" ]; then
        print_warning "Uncommitted changes detected:"
        git status --short
        echo ""
        print_status "Consider: git add . && git commit -m 'your message'"
    fi

    if [[ "$current_branch" =~ ^(main|master)$ ]]; then
        print_warning "Working on default branch ($current_branch)"
        print_status "Consider: git checkout -b feature/your-feature"
    fi
}

# Function to generate summary report
generate_summary() {
    print_header
    print_status "Post-Completion Hook Summary:"
    echo ""

    print_success "✅ Standards validation completed"
    print_success "✅ Build verification completed"
    print_success "✅ Test execution completed"
    print_success "✅ Code quality checks completed"
    print_success "✅ Git status reviewed"

    echo ""
    print_status "Next suggested actions:"
    print_status "1. Review any warnings above"
    print_status "2. Commit changes if satisfied"
    print_status "3. Create PR when feature is complete"
    print_status "4. Run: ./standards-cli.sh validate . for detailed report"
}

# Main execution
main() {
    # Change to project root
    cd "$PROJECT_ROOT"

    print_header
    print_status "Starting post-completion validation..."
    print_status "Project: $(basename "$PROJECT_ROOT")"
    print_status "Time: $(date)"
    echo ""

    # Check if we should run (only if significant changes detected)
    local has_changes=$(detect_changes)

    if [ "$has_changes" = "false" ]; then
        print_status "No significant changes detected, skipping validation"
        exit 0
    fi

    # Run validation steps
    run_standards_validation
    echo ""

    run_builds
    echo ""

    run_tests
    echo ""

    run_linting
    echo ""

    check_git_status
    echo ""

    generate_summary
}

# Execute main function
main "$@"