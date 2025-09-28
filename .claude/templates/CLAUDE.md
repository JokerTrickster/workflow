# CLAUDE.md - Project Standards Template

> Think carefully and implement the most concise solution that changes as little code as possible.

## Project Information
- **Repository**: {REPOSITORY_NAME}
- **GitHub URL**: {GITHUB_URL}
- **Project**: {PROJECT_NAME}

## Core Rules System
This project follows standardized development rules. Rules are organized by priority and domain.

### Rule Priority System
- 🔴 **CRITICAL**: Never compromise - Security, data safety, production breaks
- 🟡 **IMPORTANT**: Strong preference - Quality, maintainability, professionalism
- 🟢 **RECOMMENDED**: Apply when practical - Optimization, style, best practices

## Base Rules (Always Applied)

### Workflow Rules 🟡
- **Branch Strategy**: Always create feature branch from default branch → work → PR
- **Commit Pattern**: Descriptive messages, no "fix", "update", "changes"
- **Quality Gates**: Run lint/typecheck before completion
- **Task Planning**: Use TodoWrite for >3 step operations
- **Evidence-Based**: All claims must be verifiable through testing or documentation

### Code Quality Rules 🔴
- **NO PARTIAL IMPLEMENTATION**: Complete all started features
- **NO CODE DUPLICATION**: Check existing codebase, reuse functions
- **NO DEAD CODE**: Use it or delete it completely
- **NO INCONSISTENT NAMING**: Follow existing patterns
- **NO MIXED CONCERNS**: Proper separation of responsibilities

### Safety Rules 🔴
- **Framework Respect**: Check dependencies before using libraries
- **Pattern Adherence**: Follow existing project conventions
- **Resource Management**: Close connections, clear timeouts, clean up
- **Git Safety**: Never work directly on main/master branch

### Testing Rules 🟡
- **Test Everything**: Implement tests for every function
- **Real Tests**: No mocks, accurate tests that reveal flaws
- **Test Structure**: Organized in tests/, __tests__/, or test/ directories
- **Verbose Testing**: Design tests for debugging visibility

### Professional Standards 🟡
- **Honest Assessment**: No marketing language, provide real trade-offs
- **Critical Feedback**: Point out problems respectfully
- **Clean Workspace**: Remove temp files, maintain professional structure
- **Documentation**: Only create when explicitly requested

## Project-Specific Rules

<!-- Add your custom rules here following the format:
### Custom Rule Category {PRIORITY_EMOJI}
- **Rule Name**: Description and rationale
- **Implementation**: How to apply this rule
- **Validation**: How to check compliance
-->

## Extension Guidelines

### Adding New Rules
1. **Identify Category**: Workflow, Quality, Safety, Testing, or Custom
2. **Set Priority**: 🔴 Critical, 🟡 Important, 🟢 Recommended
3. **Define Clearly**: What, Why, How, and Validation method
4. **Test Impact**: Ensure rule doesn't conflict with existing ones

### Rule Template
```markdown
### {Category} Rules {Priority}
- **{Rule Name}**: {Clear description}
- **Rationale**: {Why this rule matters}
- **Implementation**: {How to apply}
- **Validation**: {How to check compliance}
```

### Custom Rule Examples
```markdown
### API Rules 🟡
- **Response Format**: Always return consistent JSON structure
- **Error Handling**: Use standardized error codes and messages
- **Authentication**: Validate tokens on every protected endpoint

### Database Rules 🔴
- **Transaction Safety**: Wrap related operations in transactions
- **Connection Pooling**: Never leave connections open
- **Migration Safety**: Always test migrations in staging first
```

## Validation Checklist
Before marking any task complete:
- [ ] Code follows project naming conventions
- [ ] No partial implementations remain
- [ ] Tests written and passing
- [ ] Lint/typecheck passes
- [ ] Feature branch created and PR ready
- [ ] Documentation updated if needed
- [ ] Workspace cleaned of temporary files

---
*This file is managed by the workflow project standards system.*
*Last updated: {LAST_UPDATED}*