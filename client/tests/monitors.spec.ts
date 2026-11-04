import { test, expect, type Page } from '@playwright/test';
import { registerAndLogin, createTestOrg, type CreatedOrg } from './helpers/test-utils';

test.describe.serial('Monitor Management Lifecycle', () => {
	let page: Page;
	let org: CreatedOrg;

	const timestamp = Date.now();
	let httpMonitorName = `HTTP Web App ${timestamp}`;
	const pingMonitorName = `Ping Gateway ${timestamp}`;
	const tcpMonitorName = `TCP Port ${timestamp}`;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await page.setViewportSize({ width: 1280, height: 900 });
		await registerAndLogin(page);
		// Create a fresh organization dedicated to this test suite
		org = await createTestOrg(page, `Monitors Org ${timestamp.toString(36).slice(-4)}`);
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display empty state for monitors on a fresh organization', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Verify empty state UI elements
		await expect(page.getByRole('heading', { name: 'No monitors yet' })).toBeVisible();
		await expect(
			page.getByText(
				'Create a monitor to start tracking uptime and response times for your services.'
			)
		).toBeVisible();

		// Verify empty state "Add monitor" dropdown button
		const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
		await expect(addBtn).toBeVisible();
	});

	test('should enforce form validations on short name and target in create modal', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Open Add Monitor dropdown and select HTTP/API
		const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
		await addBtn.click();
		const httpOption = page.getByRole('menuitem', { name: /http\/api/i });
		await expect(httpOption).toBeVisible();
		await httpOption.click();

		// Modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Add HTTP/API Monitor' })).toBeVisible();

		const nameInput = dialog.locator('#monitor-name');
		const targetInput = dialog.locator('#monitor-target');
		const submitBtn = dialog.getByRole('button', { name: 'Create Monitor' });

		// Test short name validation (min 4 characters, 500ms debounce)
		await nameInput.fill('abc');
		await expect(dialog.getByText('Name minimum 4 characters')).toBeVisible({ timeout: 5000 });
		await expect(submitBtn).toBeDisabled();

		// Test short target validation (min 4 characters, 500ms debounce)
		await targetInput.fill('a');
		await expect(dialog.getByText('Target minimum 4 characters')).toBeVisible({ timeout: 5000 });
		await expect(submitBtn).toBeDisabled();

		// Cancel dialog to reset state
		await dialog.getByRole('button', { name: 'Cancel' }).click();
		await expect(dialog).not.toBeVisible();
	});

	test('should create an HTTP monitor with valid configurations', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Open Add Monitor modal
		const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
		await addBtn.click();
		await page.getByRole('menuitem', { name: /http\/api/i }).click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		// Fill required valid fields
		await dialog.locator('#monitor-name').fill(httpMonitorName);
		await dialog.locator('#monitor-target').fill('https://httpbin.org/status/200');

		// Wait for debounce validation to clear
		await expect(dialog.getByText('Name minimum 4 characters')).not.toBeVisible();
		await expect(dialog.getByText('Target minimum 4 characters')).not.toBeVisible();

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Monitor' });
		await expect(submitBtn).toBeEnabled({ timeout: 5000 });
		await submitBtn.click();

		// Modal closes upon successful creation
		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify monitor card appears on page
		await expect(page.getByRole('heading', { name: httpMonitorName })).toBeVisible({
			timeout: 10000
		});
		await expect(page.getByText('https://httpbin.org/status/200')).toBeVisible();
	});

	test('should create a Ping (ICMP) monitor', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Open Add Monitor modal for Ping
		const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
		await addBtn.click();
		await page.getByRole('menuitem', { name: /ping/i }).click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Add Ping Monitor' })).toBeVisible();

		// Verify ping target label
		await expect(dialog.getByText(/Hostname \/ IP/)).toBeVisible();

		// Fill ping details
		await dialog.locator('#monitor-name').fill(pingMonitorName);
		await dialog.locator('#monitor-target').fill('1.1.1.1');

		// Submit
		const submitBtn = dialog.getByRole('button', { name: 'Create Monitor' });
		await expect(submitBtn).toBeEnabled({ timeout: 5000 });
		await submitBtn.click();

		// Modal closes and monitor appears
		await expect(dialog).not.toBeVisible({ timeout: 10000 });
		await expect(page.getByRole('heading', { name: pingMonitorName })).toBeVisible({
			timeout: 10000
		});
		await expect(page.getByText('1.1.1.1')).toBeVisible();
	});

	test('should create a TCP port monitor', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Open Add Monitor modal for TCP
		const addBtn = page.getByRole('button', { name: 'Add monitor' }).first();
		await addBtn.click();
		await page.getByRole('menuitem', { name: /tcp \(port\)/i }).click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Add TCP Port Monitor' })).toBeVisible();

		// Verify TCP target label
		await expect(dialog.getByText(/Hostname \/ IP:Port/)).toBeVisible();

		// Fill TCP details
		await dialog.locator('#monitor-name').fill(tcpMonitorName);
		await dialog.locator('#monitor-target').fill('1.1.1.1:53');

		// Submit
		const submitBtn = dialog.getByRole('button', { name: 'Create Monitor' });
		await expect(submitBtn).toBeEnabled({ timeout: 5000 });
		await submitBtn.click();

		// Modal closes and monitor appears
		await expect(dialog).not.toBeVisible({ timeout: 10000 });
		await expect(page.getByRole('heading', { name: tcpMonitorName })).toBeVisible({
			timeout: 10000
		});
	});

	test('should verify monitors overview stats, view mode toggle, and search filter', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Health overview KPI strip must be visible
		await expect(page.locator('[data-purpose="health-overview"]')).toBeVisible();

		// Filter bar buttons should be visible
		await expect(page.getByRole('button', { name: /all/i })).toBeVisible();
		await expect(page.getByRole('button', { name: 'Grid view' })).toBeVisible();
		await expect(page.getByRole('button', { name: 'List view' })).toBeVisible();

		// Switch to List view
		await page.getByRole('button', { name: 'List view' }).click();
		await expect(page.locator('table, [role="table"], .divide-y').first()).toBeVisible();

		// Switch back to Grid view
		await page.getByRole('button', { name: 'Grid view' }).click();
		await expect(page.getByRole('heading', { name: httpMonitorName })).toBeVisible();

		// Test search filtering via top header search input
		const searchInput = page.locator('header input[type="search"]');
		await expect(searchInput).toBeVisible();
		await searchInput.fill(httpMonitorName);

		// HTTP monitor remains, other monitors filtered out
		await expect(page.getByRole('heading', { name: httpMonitorName })).toBeVisible();
		await expect(page.getByRole('heading', { name: pingMonitorName })).not.toBeVisible();

		// Clear search query
		await searchInput.fill('');
		await expect(page.getByRole('heading', { name: pingMonitorName })).toBeVisible();
	});

	test('should navigate to monitor detail page, verify stats grid, period tabs, and manual check', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Click on HTTP monitor heading link
		const monitorLink = page.getByRole('link', { name: httpMonitorName, exact: true });
		await expect(monitorLink).toBeVisible();
		await monitorLink.click();

		// Redirection to /dashboard/org/[id]/monitors/[monitorId]
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors/[^/]+`));

		// Check Detail Header details
		await expect(page.getByRole('heading', { name: httpMonitorName })).toBeVisible();
		await expect(page.getByText('https://httpbin.org/status/200')).toBeVisible();

		// Check Stats Grid section
		await expect(page.getByText('Uptime Rate')).toBeVisible();
		await expect(page.getByText('Avg Latency')).toBeVisible();
		await expect(page.getByText('Min / Max Peak')).toBeVisible();
		await expect(page.getByText('Total Outages')).toBeVisible();

		// Check Performance charts section
		await expect(page.getByRole('region', { name: 'Performance charts' })).toBeVisible();

		// Test period selector tabs (24H, 7D, 30D, 90D, 1Y)
		const periodTabs = ['24H', '7D', '30D', '90D', '1Y'];
		for (const tabName of periodTabs) {
			const tab = page.getByRole('tab', { name: tabName });
			await expect(tab).toBeVisible();
			await tab.click();
			await expect(tab).toHaveAttribute('data-state', 'active');
		}

		// Trigger manual check ("Check Now" button)
		const checkNowBtn = page.getByRole('button', { name: 'Check Now' });
		if (await checkNowBtn.isVisible()) {
			await checkNowBtn.click();
			// The button triggers an immediate check and sonner toast notifies completion
			await expect(page.getByText(/check completed/i)).toBeVisible({ timeout: 15000 });
		}

		// Toggle pause / resume on detail page
		const pauseBtn = page.getByRole('button', { name: 'Pause' });
		if (await pauseBtn.isVisible()) {
			await pauseBtn.click();
			await expect(page.getByText(/monitor paused/i)).toBeVisible({ timeout: 5000 });

			const resumeBtn = page.getByRole('button', { name: 'Resume' });
			await expect(resumeBtn).toBeVisible();
			await resumeBtn.click();
			await expect(page.getByText(/monitor resumed/i)).toBeVisible({ timeout: 5000 });
		}
	});

	test('should edit monitor configuration from the monitors grid', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Find HTTP monitor card
		const card = page.locator('article.monitor-card', {
			has: page.getByRole('heading', { name: httpMonitorName })
		});
		await expect(card).toBeVisible();

		// Hover card to reveal options button, then click options
		await card.hover();
		const optionsBtn = card.getByLabel('Monitor options');
		await optionsBtn.click();

		// Click "Edit monitor" item
		const editItem = page.getByRole('menuitem', { name: 'Edit monitor' });
		await expect(editItem).toBeVisible();
		await editItem.click();

		// Edit Modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Edit Monitor' })).toBeVisible();

		// Update name
		const updatedName = `${httpMonitorName} Updated`;
		const nameInput = dialog.locator('#monitor-name');
		await nameInput.fill(updatedName);

		// Wait for debounce validation
		await expect(dialog.getByText('Name minimum 4 characters')).not.toBeVisible();

		// Submit update
		const saveBtn = dialog.getByRole('button', { name: 'Save Changes' });
		await expect(saveBtn).toBeEnabled({ timeout: 5000 });
		await saveBtn.scrollIntoViewIfNeeded();
		await saveBtn.click();

		// Modal closes and updated name is visible
		await expect(dialog).not.toBeVisible({ timeout: 10000 });
		await expect(page.getByRole('heading', { name: updatedName })).toBeVisible({
			timeout: 10000
		});

		httpMonitorName = updatedName;
	});

	test('should delete a monitor with confirmation dialog and verify removal', async () => {
		await page.goto(`/dashboard/org/${org.id}/monitors`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Locate TCP monitor card
		const card = page.locator('article.monitor-card', {
			has: page.getByRole('heading', { name: tcpMonitorName })
		});
		await expect(card).toBeVisible();

		// Open options menu
		await card.hover();
		const optionsBtn = card.getByLabel('Monitor options');
		await optionsBtn.click();

		// Click "Delete monitor"
		const deleteItem = page.getByRole('menuitem', { name: 'Delete monitor' });
		await expect(deleteItem).toBeVisible();
		await deleteItem.click();

		// Confirmation dialog opens
		const alertDialog = page.getByRole('alertdialog');
		await expect(alertDialog).toBeVisible();
		await expect(alertDialog.getByRole('heading', { name: 'Delete Monitor' })).toBeVisible();
		await expect(
			alertDialog.getByText(new RegExp(`permanently delete '${tcpMonitorName}'`, 'i'))
		).toBeVisible();

		// Confirm delete
		const confirmDeleteBtn = alertDialog.getByRole('button', { name: 'Delete' });
		await confirmDeleteBtn.click();

		// Dialog closes and monitor is removed from page
		await expect(alertDialog).not.toBeVisible({ timeout: 5000 });
		await expect(page.getByRole('heading', { name: tcpMonitorName })).not.toBeVisible({
			timeout: 10000
		});
	});
});
