import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
	test('should render login page with form elements and page title', async ({ page }) => {
		await page.goto('/auth/login');

		await expect(page).toHaveTitle(/Netra Uptime/);
		await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();

		await expect(page.locator('#email')).toBeVisible();
		await expect(page.locator('#password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();
	});

	test('should navigate to sign up and forgot password pages', async ({ page }) => {
		await page.goto('/auth/login');

		await page.getByRole('link', { name: 'Sign up' }).click();
		await expect(page).toHaveURL(/\/auth\/register/);

		await page.goto('/auth/login');
		await page.getByRole('link', { name: 'Forgot password?' }).click();
		await expect(page).toHaveURL(/\/auth\/forgot/);
	});

	test('should display error message on invalid login credentials', async ({ page }) => {
		// Intercept only the POST API request to avoid blocking page navigation
		await page.route('**/auth/login', async (route) => {
			if (route.request().method() === 'POST') {
				await route.fulfill({
					status: 401,
					contentType: 'application/json',
					body: JSON.stringify({ message: 'Invalid email or password' })
				});
			} else {
				await route.continue();
			}
		});

		await page.goto('/auth/login');
		await page.fill('#email', 'admin@example.com');
		await page.fill('#password', 'WrongPassword123!');
		await page.getByRole('button', { name: 'Login' }).click();

		const errorBanner = page.locator('.text-destructive');
		await expect(errorBanner).toBeVisible();
	});
});
