import { expect, type Page } from '@playwright/test';
import { execSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';

export interface TestUser {
	name: string;
	email: string;
	password: string;
}

export interface CreatedOrg {
	name: string;
	slug: string;
	id: string;
}

/**
 * Generate a unique test user with timestamp and random string.
 * Ensures no collisions across test runs.
 */
export function generateTestUser(): TestUser {
	const timestamp = Date.now();
	const randomStr = Math.random().toString(36).substring(2, 8);
	return {
		name: `QA Tester ${randomStr}`,
		email: `test_${timestamp}_${randomStr}@example.com`,
		password: `P@ssword123_${randomStr}`
	};
}

/**
 * Generate a unique organization name and slug within length constraints.
 * Name: 3-24 characters. Slug: 6-24 characters.
 */
export function generateTestOrg(): { name: string; slug: string } {
	const randomStr = Math.random().toString(36).substring(2, 7);
	return {
		name: `Org ${randomStr}`,
		slug: `org-${randomStr}`
	};
}

/**
 * Register a user and log them in, waiting until they reach /dashboard/org.
 * Returns the test user credentials used.
 */
export async function registerAndLogin(page: Page, user?: TestUser): Promise<TestUser> {
	const testUser = user || generateTestUser();

	// Step 1: Register
	await page.goto('/auth/register');
	await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible();

	await page.locator('#name').fill(testUser.name);
	await page.locator('#email').fill(testUser.email);
	await page.locator('#password').fill(testUser.password);
	await page.locator('#confirmPassword').fill(testUser.password);

	await page.getByRole('button', { name: 'Create Account' }).click();

	// Registration redirects to /auth/login
	await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
	await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();

	// Step 2: Login
	await page.locator('#email').fill(testUser.email);
	await page.locator('#password').fill(testUser.password);
	await page.getByRole('button', { name: 'Login' }).click();

	// Successful login redirects to /dashboard/org
	await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

	return testUser;
}

/**
 * Create an organization through the UI modal, click its card, and navigate into its workspace.
 * Returns the organization name, slug, and extracted ID.
 */
export async function createTestOrg(page: Page, customName?: string): Promise<CreatedOrg> {
	const { name: defaultName } = generateTestOrg();
	const orgName = customName || defaultName;

	// Ensure we are on /dashboard/org
	if (!page.url().includes('/dashboard/org')) {
		await page.goto('/dashboard/org');
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });
	}

	// Click button to open create org modal
	const createButton = page
		.getByRole('button', { name: /create organization|new organization/i })
		.first();
	await expect(createButton).toBeVisible({ timeout: 5000 });
	await createButton.click();

	// Verify modal is open
	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	await expect(dialog.getByRole('heading', { name: 'New organization' })).toBeVisible();

	// Fill organization name
	const nameInput = dialog.locator('#modalOrgName');
	await nameInput.fill(orgName);

	// Wait for auto-slug generation
	const slugInput = dialog.locator('#modalOrgSlug');
	await expect(slugInput).not.toHaveValue('', { timeout: 5000 });
	const slug = await slugInput.inputValue();

	// Submit creation
	const submitBtn = dialog.getByRole('button', { name: 'Create organization' });
	await expect(submitBtn).toBeEnabled({ timeout: 5000 });
	await submitBtn.click();

	// Modal closes and the new organization appears in the list on /dashboard/org
	await expect(dialog).not.toBeVisible({ timeout: 5000 });

	// Locate and click the organization card to navigate into its workspace
	const orgCard = page.locator('button', {
		has: page.getByRole('heading', { name: orgName, exact: true })
	});
	await expect(orgCard).toBeVisible({ timeout: 5000 });
	await orgCard.click();

	// Wait for redirection into workspace monitors page
	await page.waitForURL(/\/dashboard\/org\/[^/]+\/monitors/, { timeout: 10000 });

	const currentUrl = page.url();
	const match = currentUrl.match(/\/dashboard\/org\/([^/]+)\/monitors/);
	const id = match ? match[1] : '';

	return {
		name: orgName,
		slug,
		id
	};
}

/**
 * Create a simple HTTP monitor for test flows that require an existing monitor.
 */
export async function createTestMonitor(
	page: Page,
	orgId: string,
	customName?: string
): Promise<{ id: string; name: string }> {
	const timestamp = Date.now();
	const monitorName = customName || `Monitor ${timestamp}`;

	await page.goto(`/dashboard/org/${orgId}/monitors`);
	await page.waitForURL(new RegExp(`/dashboard/org/${orgId}/monitors`));

	const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
	await addBtn.click();
	await page.getByRole('menuitem', { name: /http\/api/i }).click();

	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();

	await dialog.locator('#monitor-name').fill(monitorName);
	await dialog.locator('#monitor-target').fill('https://httpbin.org/status/200');

	const submitBtn = dialog.getByRole('button', { name: 'Create Monitor' });
	await expect(submitBtn).toBeEnabled({ timeout: 5000 });
	await submitBtn.click();

	await expect(dialog).not.toBeVisible({ timeout: 10000 });
	await expect(page.getByRole('heading', { name: monitorName })).toBeVisible({ timeout: 10000 });

	const monitorLink = page.locator(`a:has-text("${monitorName}")`).first();
	let id = '';
	if (await monitorLink.isVisible()) {
		const href = await monitorLink.getAttribute('href');
		id = href?.split('/').pop() || '';
	}

	return { id, name: monitorName };
}

export interface InsertIncidentParams {
	monitorId: string;
	summary: string;
	status?: 'ongoing' | 'resolved';
	severity?: 'critical' | 'high' | 'medium' | 'low';
	startedAt?: Date;
	resolvedAt?: Date;
}

/**
 * Insert an incident record into the PostgreSQL test container.
 * Returns the created incident UUID.
 */
export function insertTestIncident(params: InsertIncidentParams): string {
	const id = randomUUID();
	const status = params.status || 'ongoing';
	const severity = params.severity || 'critical';
	const startedAt = (params.startedAt || new Date()).toISOString();
	const resolvedAt = params.resolvedAt ? `'${params.resolvedAt.toISOString()}'` : 'NULL';
	const summary = params.summary.replace(/'/g, "''");

	const sql = `INSERT INTO incidents (id, monitor_id, status, severity, started_at, resolved_at, summary) VALUES ('${id}', '${params.monitorId}', '${status}', '${severity}', '${startedAt}', ${resolvedAt}, '${summary}');`;
	execSync(`docker exec -i netra_uptime_postgres psql -U netra -d netra_uptime -c "${sql}"`);
	return id;
}
