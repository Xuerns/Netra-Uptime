import { test, expect, type Page } from '@playwright/test';
import {
	registerAndLogin,
	createTestOrg,
	generateTestOrg,
	type TestUser,
	type CreatedOrg
} from './helpers/test-utils';

test.describe.serial('Organization Management Lifecycle', () => {
	let page: Page;
	let user: TestUser;
	let createdOrg: CreatedOrg;
	let defaultOrgId = '';
	let defaultOrgName = '';

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		user = await registerAndLogin(page);
		defaultOrgName = `${user.name}Workspace`;

		// Extract default organization ID by clicking into its workspace
		const defaultCard = page.locator('button', {
			has: page.getByRole('heading', { name: defaultOrgName, exact: true })
		});
		await expect(defaultCard).toBeVisible({ timeout: 5000 });
		await defaultCard.click();
		await page.waitForURL(/\/dashboard\/org\/[^/]+\/monitors/, { timeout: 10000 });

		const match = page.url().match(/\/dashboard\/org\/([^/]+)\/monitors/);
		defaultOrgId = match ? match[1] : '';
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display organization list view and loaded state for registered user', async () => {
		await page.goto('/dashboard/org');
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

		// Verify loaded state with default workspace
		await expect(page.getByText(/Select an organization/i)).toBeVisible();
		await expect(page.getByRole('button', { name: 'New organization' })).toBeVisible();
		await expect(page.getByRole('button', { name: 'Grid view' })).toBeVisible();
		await expect(page.getByRole('button', { name: 'List view' })).toBeVisible();

		// Verify organization card for default workspace
		await expect(page.getByRole('heading', { name: defaultOrgName })).toBeVisible();
	});

	test('should validate organization modal inputs, auto-generate slug, and cancel creation', async () => {
		await page.goto('/dashboard/org');
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

		// Open create organization modal
		const newOrgBtn = page.getByRole('button', { name: 'New organization' });
		await newOrgBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'New organization' })).toBeVisible();

		const nameInput = dialog.locator('#modalOrgName');
		const slugInput = dialog.locator('#modalOrgSlug');
		const submitBtn = dialog.getByRole('button', { name: 'Create organization' });

		// Name too short validation
		await nameInput.fill('ab');
		await expect(dialog.getByText('Name minimum 3 characters')).toBeVisible({ timeout: 5000 });
		await expect(submitBtn).toBeDisabled();

		// Enter valid name and check auto-slug generation
		const { name: validName } = generateTestOrg();
		await nameInput.fill(validName);

		// Slug must be auto-populated
		await expect(slugInput).not.toHaveValue('', { timeout: 5000 });
		const generatedSlug = await slugInput.inputValue();
		expect(generatedSlug.length).toBeGreaterThanOrEqual(6);

		// Cancel modal
		await dialog.getByRole('button', { name: 'Cancel' }).click();
		await expect(dialog).not.toBeVisible();
	});

	test('should create a new organization and navigate into its workspace', async () => {
		await page.goto('/dashboard/org');
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

		// Create organization
		createdOrg = await createTestOrg(page);

		// Verify navigation into the organization workspace
		await expect(page).toHaveURL(new RegExp(`/dashboard/org/${createdOrg.id}/monitors`));
		await expect(page.getByRole('link', { name: 'Monitors' }).first()).toBeVisible();
		await expect(page.getByRole('link', { name: 'Members' }).first()).toBeVisible();
		await expect(page.getByRole('link', { name: 'Settings' }).first()).toBeVisible();
	});

	test('should allow owner to update organization name in settings', async () => {
		// Navigate to settings page
		await page.goto(`/dashboard/org/${createdOrg.id}/settings`);
		await expect(page).toHaveTitle(/Organization Settings/);
		await expect(page.getByText('General Information')).toBeVisible();

		// Verify existing org name
		const nameInput = page.locator('#name');
		await expect(nameInput).toHaveValue(createdOrg.name);

		// Update name (<= 24 chars limit)
		const updatedName = `Renamed ${Date.now().toString().slice(-4)}`;
		await nameInput.fill(updatedName);

		// Save changes
		await page.getByRole('button', { name: 'Save Changes' }).click();

		// Verify success toast
		await expect(page.getByText(/Organization updated successfully/i)).toBeVisible({
			timeout: 5000
		});

		createdOrg.name = updatedName;
	});

	test('should display owner in organization members table with search input', async () => {
		// Navigate to members page
		await page.goto(`/dashboard/org/${createdOrg.id}/members`);
		await expect(page).toHaveTitle(/Members/);

		// Verify search input exists
		const searchInput = page.getByPlaceholder('Search members by name...');
		await expect(searchInput).toBeVisible();

		// Verify owner appears in members table
		const membersContent = page.locator('.members-page');
		await expect(membersContent.getByText(user.name)).toBeVisible();
		await expect(membersContent.getByText('Owner', { exact: true })).toBeVisible();
		await expect(membersContent.getByText('You', { exact: true })).toBeVisible();
	});

	test('should disband organizations and display empty state when zero organizations remain', async () => {
		// Disband created organization
		await page.goto(`/dashboard/org/${createdOrg.id}/settings`);
		await page.getByRole('button', { name: 'Disband Organization' }).click();
		const dialog1 = page.getByRole('alertdialog');
		await expect(dialog1.getByRole('heading', { name: 'Disband Organization' })).toBeVisible();
		await dialog1.locator('#confirmName').fill(createdOrg.name);
		await dialog1.getByRole('button', { name: 'Disband Organization' }).click();
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

		// Disband default workspace
		await page.goto(`/dashboard/org/${defaultOrgId}/settings`);
		await page.getByRole('button', { name: 'Disband Organization' }).click();
		const dialog2 = page.getByRole('alertdialog');
		await expect(dialog2.getByRole('heading', { name: 'Disband Organization' })).toBeVisible();
		await dialog2.locator('#confirmName').fill(defaultOrgName);
		await dialog2.getByRole('button', { name: 'Disband Organization' }).click();
		await page.waitForURL(/\/dashboard\/org/, { timeout: 10000 });

		// Verify empty state is displayed now that 0 organizations remain
		await expect(page.getByRole('heading', { name: 'No organizations yet' })).toBeVisible({
			timeout: 5000
		});
		await expect(page.getByText(/Create an organization to start adding monitors/i)).toBeVisible();
		await expect(page.getByRole('button', { name: 'Create organization' })).toBeVisible();
	});
});
