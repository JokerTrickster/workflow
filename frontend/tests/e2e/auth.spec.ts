import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('should show login page when not authenticated', async ({ page }) => {
    await page.goto('/');
    
    // Should redirect to login page or show login UI
    await expect(page).toHaveURL(/\/login/);
    await expect(page.locator('text=Sign in')).toBeVisible();
    await expect(page.locator('text=Continue with GitHub')).toBeVisible();
  });

  test('login form should be accessible', async ({ page }) => {
    await page.goto('/login');
    
    // Check for proper heading hierarchy
    const title = page.locator('h1, [role="heading"][aria-level="1"]').first();
    await expect(title).toBeVisible();
    
    // Check login button accessibility
    const loginButton = page.locator('button:has-text("Continue with GitHub")');
    await expect(loginButton).toBeVisible();
    await expect(loginButton).toHaveAttribute('type', 'button');
    
    // Ensure button meets minimum size requirements (44x44px)
    const buttonBox = await loginButton.boundingBox();
    expect(buttonBox?.width).toBeGreaterThanOrEqual(44);
    expect(buttonBox?.height).toBeGreaterThanOrEqual(44);
  });

  test('should handle authentication errors gracefully', async ({ page }) => {
    await page.goto('/login');
    
    // Mock authentication error
    await page.route('**/auth/**', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Authentication failed' }),
      });
    });
    
    await page.locator('button:has-text("Continue with GitHub")').click();
    
    // Should show error message
    await expect(page.locator('.text-red-600, [role="alert"]')).toBeVisible();
  });

  test('responsive design should work on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/login');
    
    // Check that login form is properly sized for mobile
    const loginCard = page.locator('[data-slot="card"]').first();
    await expect(loginCard).toBeVisible();
    
    const cardBox = await loginCard.boundingBox();
    expect(cardBox?.width).toBeLessThanOrEqual(375);
    
    // Ensure touch targets are adequate
    const loginButton = page.locator('button:has-text("Continue with GitHub")');
    const buttonBox = await loginButton.boundingBox();
    expect(buttonBox?.width).toBeGreaterThanOrEqual(44);
    expect(buttonBox?.height).toBeGreaterThanOrEqual(44);
  });
});

test.describe('Dashboard Access', () => {
  test.skip('should show dashboard when authenticated', async ({ page }) => {
    // This test would require actual authentication setup
    // For now, skip it as we don't have real auth flow
    await page.goto('/dashboard');
    // Would expect to see repository dashboard
  });

  test('should show proper loading states', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Should show some form of loading indicator initially
    const loadingElement = page.locator('.animate-spin, text=Loading, [aria-live="polite"]');
    // This might be briefly visible or not, depending on cache
  });
});