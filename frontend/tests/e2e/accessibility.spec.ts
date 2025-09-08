import { test, expect } from '@playwright/test';
import { injectAxe, checkA11y } from 'axe-playwright';

test.describe('Accessibility Tests', () => {
  test.beforeEach(async ({ page }) => {
    await injectAxe(page);
  });

  test('login page should be accessible', async ({ page }) => {
    await page.goto('/login');
    await checkA11y(page, undefined, {
      detailedReport: true,
      detailedReportOptions: { html: true },
    });
  });

  test('dashboard page should be accessible', async ({ page }) => {
    await page.goto('/dashboard');
    await checkA11y(page, undefined, {
      detailedReport: true,
      detailedReportOptions: { html: true },
    });
  });

  test('keyboard navigation should work', async ({ page }) => {
    await page.goto('/login');
    
    // Tab through interactive elements
    await page.keyboard.press('Tab');
    
    const loginButton = page.locator('button:has-text("Continue with GitHub")');
    await expect(loginButton).toBeFocused();
    
    // Enter should activate the button
    await page.keyboard.press('Enter');
    // This would trigger the login process
  });

  test('search filter should be keyboard accessible', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Focus should be able to reach search input
    await page.keyboard.press('Tab');
    
    const searchInput = page.locator('input[placeholder*="Search repositories"]');
    if (await searchInput.isVisible()) {
      await searchInput.focus();
      await searchInput.fill('test');
      await expect(searchInput).toHaveValue('test');
    }
  });

  test('filter badges should have proper ARIA labels', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check if filter badges have proper accessibility attributes
    const clearButtons = page.locator('button[aria-label*="Clear"]');
    const count = await clearButtons.count();
    
    for (let i = 0; i < count; i++) {
      const button = clearButtons.nth(i);
      await expect(button).toHaveAttribute('aria-label');
      
      // Check minimum touch target size
      const box = await button.boundingBox();
      if (box) {
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.height).toBeGreaterThanOrEqual(44);
      }
    }
  });

  test('error messages should be announced to screen readers', async ({ page }) => {
    await page.goto('/login');
    
    // Mock an error response
    await page.route('**/auth/**', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Authentication failed' }),
      });
    });
    
    await page.locator('button:has-text("Continue with GitHub")').click();
    
    // Error should be announced (role="alert" or aria-live)
    const errorMessage = page.locator('[role="alert"], [aria-live="polite"], [aria-live="assertive"]');
    await expect(errorMessage).toBeVisible();
  });

  test('color contrast should meet WCAG standards', async ({ page }) => {
    await page.goto('/dashboard');
    
    // This would require color contrast analysis
    // For now, we rely on axe-core which includes contrast checks
    await checkA11y(page, undefined, {
      rules: {
        'color-contrast': { enabled: true },
        'color-contrast-enhanced': { enabled: true },
      },
    });
  });

  test('images should have alt text', async ({ page }) => {
    await page.goto('/dashboard');
    
    const images = page.locator('img');
    const imageCount = await images.count();
    
    for (let i = 0; i < imageCount; i++) {
      const image = images.nth(i);
      const alt = await image.getAttribute('alt');
      const role = await image.getAttribute('role');
      
      // Images should have alt text or role="presentation"
      expect(alt !== null || role === 'presentation').toBeTruthy();
    }
  });
});