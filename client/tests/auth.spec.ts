import { test, expect, type Page } from '@playwright/test';
import { generateTestUser, type TestUser } from './helpers/test-utils';

test.describe('Landing Page Redirection', () => {
	test('should redirect root url to /auth/login for unauthenticated users', async ({ page }) => {
		await page.goto('/');

		await expect(page).toHaveURL(/\/auth\/login/);
		await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();
	});
});

test.describe('Login Form Validation', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/auth/login');
		await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();
	});

	test('should render form elements and navigation links correctly', async ({ page }) => {
		await expect(page.locator('#email')).toBeVisible();
		await expect(page.locator('#password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();

		await expect(page.getByRole('link', { name: 'Forgot password?' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Sign up' })).toBeVisible();
	});

	test('should display validation error on invalid email format', async ({ page }) => {
		await page.locator('#email').fill('not-an-email');

		await expect(page.getByText('Invalid email format')).toBeVisible({ timeout: 5000 });
	});

	test('should display validation error on short password', async ({ page }) => {
		await page.locator('#password').fill('short');

		await expect(page.getByText('Password minimum 8 characters')).toBeVisible({ timeout: 5000 });
	});

	test('should not submit with empty fields due to required attributes', async ({ page }) => {
		await page.getByRole('button', { name: 'Login' }).click();

		await expect(page).toHaveURL(/\/auth\/login/);
	});

	test('should display error banner on invalid credentials', async ({ page }) => {
		const nonExistentEmail = `nonexistent_${Date.now()}@example.com`;
		await page.locator('#email').fill(nonExistentEmail);
		await page.locator('#password').fill('IncorrectPassword123!');
		await page.getByRole('button', { name: 'Login' }).click();

		const errorBanner = page.locator('.text-destructive');
		await expect(errorBanner).toBeVisible({ timeout: 5000 });
	});
});

test.describe('Forgot Password Flow', () => {
	test('should display confirmation screen upon submitting reset request', async ({ page }) => {
		await page.goto('/auth/forgot');
		await expect(page.getByRole('heading', { name: 'Forgot your password?' })).toBeVisible();

		const targetEmail = `reset_${Date.now()}@example.com`;
		await page.locator('#email').fill(targetEmail);
		await page.getByRole('button', { name: 'Send reset link' }).click();

		await expect(page.getByRole('heading', { name: 'Check your email' })).toBeVisible({
			timeout: 10000
		});
		await expect(page.getByText(targetEmail)).toBeVisible();
		await expect(page.getByRole('button', { name: 'Try a different email' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Back to login' }).first()).toBeVisible();
	});
});

test.describe('Registration Form Validation', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/auth/register');
		await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible();
	});

	test('should display validation error when name is too short', async ({ page }) => {
		await page.locator('#name').fill('ab');

		await expect(page.getByText('Name minimum 3 characters')).toBeVisible({ timeout: 5000 });
	});

	test('should display validation error when email format is invalid', async ({ page }) => {
		await page.locator('#email').fill('invalid-email-address');

		await expect(page.getByText('Invalid email format')).toBeVisible({ timeout: 5000 });
	});

	test('should display validation error when passwords do not match', async ({ page }) => {
		await page.locator('#password').fill('ValidPassword123!');
		await page.locator('#confirmPassword').fill('DifferentPassword123!');

		await expect(page.getByText('Password does not match')).toBeVisible({ timeout: 5000 });
	});
});

test.describe.serial('User Registration, Login, and Profile Management', () => {
	let page: Page;
	let user: TestUser;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		user = generateTestUser();
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should register a new account and log in successfully', async () => {
		// Register
		await page.goto('/auth/register');
		await page.locator('#name').fill(user.name);
		await page.locator('#email').fill(user.email);
		await page.locator('#password').fill(user.password);
		await page.locator('#confirmPassword').fill(user.password);
		await page.getByRole('button', { name: 'Create Account' }).click();

		// Verify redirect to login
		await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
		await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();

		// Log in
		await page.locator('#email').fill(user.email);
		await page.locator('#password').fill(user.password);
		await page.getByRole('button', { name: 'Login' }).click();

		// Verify redirect to dashboard
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });
	});

	test('should update user profile display name', async () => {
		// Navigate to profile page (already authenticated in session)
		await page.goto('/dashboard/profile');
		await expect(page.getByRole('heading', { name: 'Profile Settings' })).toBeVisible();

		// Verify initial display name
		const nameInput = page.locator('#name');
		await expect(nameInput).toHaveValue(user.name);

		// Update name
		const updatedName = `Updated User ${Date.now().toString().slice(-4)}`;
		await nameInput.fill(updatedName);
		await page.getByRole('button', { name: 'Save Changes' }).click();

		// Verify toast notification and updated heading
		await expect(page.getByText(/Profile updated successfully/i)).toBeVisible({ timeout: 5000 });
		await expect(page.locator('h3', { hasText: updatedName })).toBeVisible({ timeout: 5000 });
	});
});
