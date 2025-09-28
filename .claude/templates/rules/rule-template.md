# {Category Name} Rules

## {Rule Name} {Priority_Emoji}

### Rule Definition
- **{Rule Name}**: {Clear and specific description of what the rule requires}
- **Rationale**: {Why this rule is important - business/technical justification}
- **Implementation**: {How to apply this rule in practice - specific steps or patterns}
- **Validation**: {How to check if this rule is being followed - automated or manual checks}

### Examples
**Good Examples ✅:**
```
{Show concrete examples of code/patterns that follow this rule}
```

**Bad Examples ❌:**
```
{Show concrete examples of code/patterns that violate this rule}
```

### Validation Checklist
- [ ] {Specific item to check for compliance}
- [ ] {Another specific item to check}
- [ ] {Additional validation criteria}

### Automation Opportunities
- **Linting Rules**: {ESLint, Prettier, or other tool configurations}
- **Git Hooks**: {Pre-commit or pre-push hook scripts}
- **CI/CD Checks**: {Automated pipeline validations}

### Related Rules
- {Reference to other rules that work together with this one}
- {Link to conflicting rules that need to be considered}

### Exceptions
- **When to deviate**: {Rare cases where this rule might not apply}
- **Approval process**: {Who needs to approve exceptions}
- **Documentation requirement**: {How to document approved exceptions}

---

## Template Usage Guide

### Priority Selection
- 🔴 **CRITICAL**: Security, data safety, production breaks - never compromise
- 🟡 **IMPORTANT**: Quality, maintainability, professionalism - strong preference
- 🟢 **RECOMMENDED**: Optimization, style, best practices - apply when practical

### Category Guidelines
- **Workflow**: Git processes, branching, PRs, deployment
- **Code Quality**: Structure, naming, patterns, maintainability
- **Safety**: Security, data protection, system stability
- **Testing**: Test strategy, structure, execution, coverage
- **Performance**: Optimization, monitoring, scalability
- **Custom**: Domain-specific rules (API, Database, UI, etc.)

### Writing Effective Rules

#### ✅ Do
- Be specific and actionable
- Include concrete examples
- Explain the "why" not just the "what"
- Make validation criteria clear
- Consider automation possibilities
- Think about exceptions upfront

#### ❌ Avoid
- Vague or subjective language
- Rules that can't be validated
- Overly restrictive rules without good reason
- Rules that conflict with existing standards
- Missing rationale or examples
- One-size-fits-all mentality

### Example Complete Rule

## Database Connection Rules 🔴

### Rule Definition
- **Connection Cleanup Required**: All database connections must be properly closed after use
- **Rationale**: Prevents connection pool exhaustion which can cause service outages
- **Implementation**: Use try-finally blocks or connection pooling with automatic cleanup
- **Validation**: Monitor connection pool metrics and code review for proper cleanup patterns

### Examples
**Good Examples ✅:**
```javascript
async function getUser(id) {
  let connection;
  try {
    connection = await pool.getConnection();
    return await connection.query('SELECT * FROM users WHERE id = ?', [id]);
  } finally {
    if (connection) connection.close();
  }
}
```

**Bad Examples ❌:**
```javascript
async function getUser(id) {
  const connection = await pool.getConnection();
  return await connection.query('SELECT * FROM users WHERE id = ?', [id]);
  // Connection never closed - memory leak!
}
```

### Validation Checklist
- [ ] All database query functions include connection cleanup
- [ ] try-finally blocks are used for connection management
- [ ] Connection pool monitoring shows stable connection counts
- [ ] No connection timeout errors in production logs

### Automation Opportunities
- **Linting Rules**: Custom ESLint rule to detect unclosed connections
- **Git Hooks**: Pre-commit hook to scan for connection patterns
- **CI/CD Checks**: Integration tests that monitor connection pool health

### Related Rules
- Error Handling Rules (proper exception handling with cleanup)
- Resource Management Rules (general resource cleanup principles)

### Exceptions
- **When to deviate**: When using ORM with built-in connection management
- **Approval process**: Tech lead approval for ORM adoption
- **Documentation requirement**: Document ORM connection handling in project README