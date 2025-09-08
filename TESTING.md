# Testing Guide

## Overview

This project uses a comprehensive testing strategy with multiple layers:

- **Unit Tests**: Jest + React Testing Library
- **Integration Tests**: API endpoint testing
- **E2E Tests**: Playwright for browser automation
- **Accessibility Tests**: axe-core with Playwright
- **Performance Tests**: Lighthouse CI integration

## Getting Started

### Prerequisites

```bash
cd frontend
npm install
```

### Running Tests

#### Unit Tests
```bash
# Run all unit tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

#### E2E Tests
```bash
# Run all E2E tests
npm run test:e2e

# Run E2E tests with UI
npm run test:e2e:ui

# Run specific test file
npx playwright test auth.spec.ts

# Run tests in specific browser
npx playwright test --project=chromium
```

## Test Structure

### Unit Tests (`src/__tests__/`)
```
src/__tests__/
├── components/        # Component tests
│   ├── ErrorMessage.test.tsx
│   └── SearchFilter.test.tsx
├── hooks/            # Hook tests
│   └── useErrorRecovery.test.ts
├── utils/            # Utility function tests
│   └── errorHandler.test.ts
└── integration/      # Integration tests
    └── api.test.ts
```

### E2E Tests (`tests/e2e/`)
```
tests/e2e/
├── auth.spec.ts           # Authentication flow
├── accessibility.spec.ts  # WCAG compliance
└── performance.spec.ts    # Core Web Vitals
```

## Writing Tests

### Unit Test Example
```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorMessage } from '../../components/ErrorMessage'

test('should show retry button when error is retryable', () => {
  const mockOnRetry = jest.fn()
  render(<ErrorMessage error={retryableError} onRetry={mockOnRetry} />)
  
  const retryButton = screen.getByRole('button', { name: /다시 시도/i })
  expect(retryButton).toBeInTheDocument()
  
  fireEvent.click(retryButton)
  expect(mockOnRetry).toHaveBeenCalledTimes(1)
})
```

### E2E Test Example
```typescript
import { test, expect } from '@playwright/test'

test('should handle authentication flow', async ({ page }) => {
  await page.goto('/login')
  
  await expect(page.locator('text=Sign in')).toBeVisible()
  await page.locator('button:has-text("Continue with GitHub")').click()
  
  // Assertions for expected behavior
})
```

### Accessibility Test Example
```typescript
import { injectAxe, checkA11y } from 'axe-playwright'

test('should be accessible', async ({ page }) => {
  await injectAxe(page)
  await page.goto('/dashboard')
  
  await checkA11y(page, undefined, {
    detailedReport: true,
    detailedReportOptions: { html: true },
  })
})
```

## Test Coverage

### Coverage Goals
- **Lines**: >80%
- **Functions**: >80%
- **Branches**: >80%
- **Statements**: >80%

### Excluded from Coverage
- Configuration files
- Type definitions
- Story files
- Test files themselves

## Accessibility Testing

### WCAG 2.1 AA Compliance
Our accessibility tests verify:

- **Perceivable**: Alt text, color contrast, text scaling
- **Operable**: Keyboard navigation, focus management, touch targets
- **Understandable**: Clear labels, error messages, instructions
- **Robust**: Valid HTML, ARIA attributes, screen reader support

### Key Requirements
- Touch targets ≥44x44px
- Color contrast ratio ≥4.5:1
- Keyboard navigation for all interactive elements
- Screen reader announcements for dynamic content
- Proper heading hierarchy

## Performance Testing

### Core Web Vitals Targets
- **LCP** (Largest Contentful Paint): <2.5s
- **FID** (First Input Delay): <100ms
- **CLS** (Cumulative Layout Shift): <0.1

### Lighthouse Scores
- **Performance**: >95
- **Accessibility**: >95
- **Best Practices**: >95
- **SEO**: >90

## Continuous Integration

### GitHub Actions
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run unit tests
        run: npm run test:coverage
      
      - name: Run E2E tests
        run: npm run test:e2e
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

## Debugging Tests

### Jest Debugging
```bash
# Run specific test file
npm test ErrorMessage.test.tsx

# Run tests in debug mode
node --inspect-brk node_modules/.bin/jest --runInBand

# Run with verbose output
npm test -- --verbose
```

### Playwright Debugging
```bash
# Run tests in headed mode
npx playwright test --headed

# Run with debugger
npx playwright test --debug

# Generate trace
npx playwright test --trace on
```

## Best Practices

### Unit Tests
1. **AAA Pattern**: Arrange, Act, Assert
2. **Descriptive Names**: Test what, not how
3. **Single Responsibility**: One assertion per test
4. **Mock External Dependencies**: Keep tests isolated
5. **Test Behavior**: Focus on user interactions

### E2E Tests
1. **Page Object Model**: Organize selectors
2. **Wait Strategies**: Use explicit waits
3. **Data Independence**: Don't depend on test order
4. **Error Handling**: Test negative scenarios
5. **Browser Support**: Test across browsers

### Accessibility Tests
1. **Automated + Manual**: Tools catch 30-40% of issues
2. **Keyboard Testing**: Tab through all interactions
3. **Screen Reader**: Test with actual assistive tech
4. **Color Vision**: Test with color blindness simulators
5. **Mobile Touch**: Verify touch target sizes

## Troubleshooting

### Common Issues

#### Jest
- **Module not found**: Check `moduleNameMapping` in jest.config.js
- **Async tests**: Use `await` with async test functions
- **Mock issues**: Clear mocks between tests

#### Playwright
- **Element not found**: Use `waitFor` methods
- **Timeouts**: Increase timeout or check selectors
- **Flaky tests**: Add proper wait conditions

#### Accessibility
- **Color contrast**: Check actual computed colors
- **Focus management**: Test keyboard navigation
- **Screen reader**: Verify ARIA attributes

### Getting Help
1. Check test logs and error messages
2. Use browser dev tools for debugging
3. Review official documentation
4. Check GitHub issues for known problems

## Resources

- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Playwright Documentation](https://playwright.dev/)
- [axe-core Rules](https://github.com/dequelabs/axe-core/blob/develop/doc/rule-descriptions.md)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)