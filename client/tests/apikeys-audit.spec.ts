import { test, expect, type Page } from '@playwright/test';
import { registerAndLogin, createTestOrg, type CreatedOrg } from './helpers/test-utils';

test.describe.serial('API Keys and Audit Logs Explorer', () => {
	let page: Page;
	let org: CreatedOrg;
	const timestamp = Date.now();
	const readOnlyKeyName = `ReadOnly Key ${timestamp}`;
	const fullAccessKeyName = `FullAccess Key ${timestamp}`;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await page.setViewportSize({ width: 1280, height: 900 });
		await page.context().grantPermissions(['clipboard-read', 'clipboard-write']);
		await registerAndLogin(page);
		org = await createTestOrg(page, `API Org ${timestamp.toString(36).slice(-4)}`);
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('1. Empty state on fresh organization', async () => {
		await page.goto(`/dashboard/org/${org.id}/api-keys`);
		await page.waitForLoadState('networkidle');

		await expect(page.getByRole('heading', { name: 'API Keys', level: 1 })).toBeVisible();
		await expect(page.getByText('No API keys created yet')).toBeVisible();
		await expect(
			page.getByText('Create an API key to allow CI/CD pipelines, automated scripts')
		).toBeVisible();
		await expect(page.getByRole('button', { name: 'Create First Key' })).toBeVisible();
	});

	test('2. Create Read-Only API key with 30-day expiration', async () => {
		await page.goto(`/dashboard/org/${org.id}/api-keys`);
		await page.waitForLoadState('networkidle');

		await page.getByRole('button', { name: 'Create First Key' }).click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByText('Create New API Key')).toBeVisible();

		await dialog.locator('#keyName').fill(readOnlyKeyName);
		await dialog.getByRole('button', { name: /Read Only/i }).click();
		await dialog.getByRole('combobox').click();
		await page.getByRole('option', { name: '30 Days' }).click();

		await dialog.getByRole('button', { name: 'Generate API Key' }).click();

		// Plaintext reveal modal should open
		const revealModal = page.getByRole('dialog');
		await expect(revealModal.getByText('API Key Created Successfully')).toBeVisible();
		await revealModal.getByRole('button', { name: 'Done' }).click();
		await expect(revealModal).toBeHidden();
	});

	test('3. Plaintext reveal modal verification and key listing', async () => {
		await page.goto(`/dashboard/org/${org.id}/api-keys`);
		await page.waitForLoadState('networkidle');

		// Open modal again to verify full reveal lifecycle
		await page.getByRole('button', { name: 'Create API Key' }).click();
		const dialog = page.getByRole('dialog');
		const testKeyName = `Lifecycle Key ${Date.now()}`;
		await dialog.locator('#keyName').fill(testKeyName);
		await dialog.getByRole('button', { name: 'Generate API Key' }).click();

		const revealModal = page.getByRole('dialog');
		await expect(revealModal.getByText('API Key Created Successfully')).toBeVisible();
		await expect(revealModal.getByText('it will never be displayed again!')).toBeVisible();

		// Verify key starts with prefix pw_live_
		const keyDisplay = revealModal.locator('.font-mono').first();
		await expect(keyDisplay).toBeVisible();
		const keyText = await keyDisplay.innerText();
		expect(keyText).toContain('pw_live_');

		// Click copy button
		const copyButton = revealModal.getByRole('button', { name: /Copy/i });
		await expect(copyButton).toBeVisible();
		await copyButton.click();

		// Dismiss modal
		await revealModal.getByRole('button', { name: 'Done' }).click();
		await expect(revealModal).toBeHidden();

		// Verify key is listed in the dashboard
		const testKeyCard = page.locator('div.p-4', { hasText: testKeyName }).first();
		await expect(testKeyCard).toBeVisible();
		await expect(testKeyCard.getByText('Read Only')).toBeVisible();
		await expect(testKeyCard.getByText(/pw_live_.*?\.\.\.\*\*\*\*/)).toBeVisible();
	});

	test('4. Create Full Access (Write) API key with Never expiration', async () => {
		await page.goto(`/dashboard/org/${org.id}/api-keys`);
		await page.waitForLoadState('networkidle');

		await page.getByRole('button', { name: 'Create API Key' }).click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		await dialog.locator('#keyName').fill(fullAccessKeyName);
		await dialog.getByRole('button', { name: /Full Access/i }).click();
		await dialog.getByRole('combobox').click();
		await page.getByRole('option', { name: /Never/i }).click();

		await dialog.getByRole('button', { name: 'Generate API Key' }).click();

		const revealModal = page.getByRole('dialog');
		await expect(revealModal.getByText('API Key Created Successfully')).toBeVisible();
		await revealModal.getByRole('button', { name: 'Done' }).click();
		await expect(revealModal).toBeHidden();

		// Verify item attributes
		const keyCard = page.locator('div.p-4', { hasText: fullAccessKeyName }).first();
		await expect(keyCard).toBeVisible();
		await expect(keyCard.getByText('Full Access (Read & Write)')).toBeVisible();
		await expect(keyCard.getByText('Expires: —')).toBeVisible();
	});

	test('5. Revoke key with confirmation dialog', async () => {
		await page.goto(`/dashboard/org/${org.id}/api-keys`);
		await page.waitForLoadState('networkidle');

		await expect(page.getByText(readOnlyKeyName)).toBeVisible();

		// Find the row containing readOnlyKeyName and click its revoke button
		const keyRow = page.locator('div.p-4', { hasText: readOnlyKeyName }).first();
		await keyRow.locator('button[title="Revoke API key"]').click();

		// Confirm dialog is an alertdialog
		const alertDialog = page.getByRole('alertdialog');
		await expect(alertDialog).toBeVisible();
		await expect(alertDialog.getByText('Revoke API Key?')).toBeVisible();
		await expect(
			alertDialog.getByText(`Are you sure you want to revoke '${readOnlyKeyName}'?`)
		).toBeVisible();

		await alertDialog.getByRole('button', { name: 'Revoke Key' }).click();

		// Verify the revoked key is removed
		await expect(alertDialog).toBeHidden();
		await expect(page.getByText(readOnlyKeyName)).toBeHidden();
		await expect(page.getByText(fullAccessKeyName)).toBeVisible();
	});

	test('6. Audit logs explorer displays recorded events', async () => {
		await page.goto(`/dashboard/org/${org.id}/logs`);
		await page.waitForLoadState('networkidle');

		// Audit log table should be visible
		const table = page.locator('.log-table-header');
		await expect(table).toBeVisible();
		await expect(table.getByText('Actor')).toBeVisible();
		await expect(table.getByText('Action')).toBeVisible();
		await expect(table.getByText('Event / Details')).toBeVisible();

		// Should have recorded API key create and revoke events
		await expect(page.getByText('API_KEY_CREATE').first()).toBeVisible();
		await expect(page.getByText('API_KEY_REVOKE').first()).toBeVisible();
	});

	test('7. Audit logs Category filter', async () => {
		await page.goto(`/dashboard/org/${org.id}/logs`);
		await page.waitForLoadState('networkidle');

		// Click Category Dropdown Trigger (button with ChevronDown icon)
		const categoryTrigger = page
			.locator('button')
			.filter({ has: page.locator('svg.lucide-chevron-down') })
			.first();
		await categoryTrigger.click();

		// Select API Keys category
		const menu = page.getByRole('menu');
		await expect(menu).toBeVisible();
		await menu.getByText('API Keys', { exact: true }).click();

		// Close dropdown menu with Escape
		await page.keyboard.press('Escape');

		// Verify toolbar shows active category badge and filtered rows
		await expect(page.getByText('Category: API Keys')).toBeVisible();
		await expect(page.getByText('API_KEY_CREATE').first()).toBeVisible();

		// Open dropdown again, click All Categories to reset, then select Monitors
		await categoryTrigger.click();
		await expect(menu).toBeVisible();
		await menu.getByText('All Categories', { exact: true }).click();

		await categoryTrigger.click();
		await expect(menu).toBeVisible();
		await menu.getByText('Monitors', { exact: true }).click();
		await page.keyboard.press('Escape');

		// Since no monitors exist in this org, verify empty state
		await expect(page.getByText('No Audit Logs Found')).toBeVisible();

		// Click Reset button to clear filters
		await page.getByRole('button', { name: 'Reset' }).click();
		await expect(page.getByText('API_KEY_CREATE').first()).toBeVisible();
	});

	test('8. Audit logs Search filter', async () => {
		await page.goto(`/dashboard/org/${org.id}/logs`);
		await page.waitForLoadState('networkidle');

		const searchInput = page.getByPlaceholder('Search audit logs...');
		await expect(searchInput).toBeVisible();

		// Search for fullAccessKeyName
		await searchInput.fill(fullAccessKeyName);
		// Allow debounce and network request to finish
		await page.waitForTimeout(600);

		// Check that matching event description is visible
		await expect(page.getByText(`Created API Key '${fullAccessKeyName}'`)).toBeVisible();
		await expect(page.getByText(`"${fullAccessKeyName}"`)).toBeVisible();

		// Search for non-existent text
		const nonexistentQuery = `nonexistent-query-${Date.now()}`;
		await searchInput.fill(nonexistentQuery);
		await page.waitForTimeout(600);

		await expect(page.getByText('No Audit Logs Found')).toBeVisible();
		await expect(page.getByText(`No activity matching "${nonexistentQuery}"`)).toBeVisible();

		// Clear search using the active filter pill clear button
		await page.getByRole('button', { name: 'Clear search' }).click();
		await page.waitForTimeout(600);

		await expect(page.getByText('API_KEY_CREATE').first()).toBeVisible();
	});
});
