import { test, expect, type Page } from '@playwright/test';
import {
	registerAndLogin,
	createTestOrg,
	createTestMonitor,
	type CreatedOrg
} from './helpers/test-utils';

test.describe.serial('Status Page Lifecycle', () => {
	let page: Page;
	let org: CreatedOrg;

	const timestamp = Date.now();
	let monitorName = '';
	const statusPageTitle = `Acme Cloud Reliability ${timestamp}`;
	const subdomain = `acme-${timestamp.toString(36).slice(-6).toLowerCase()}`;
	const statusPageDescription = `Realtime infrastructure status for Acme Cloud`;
	const groupName = 'Core Platform Infrastructure';
	const updatedGroupName = 'Core Infrastructure & Services';
	const maintenanceTitle = `Database Upgrade Window ${timestamp}`;
	const maintenanceDescription = `Upgrading PostgreSQL database server to improve query latency.`;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await page.setViewportSize({ width: 1280, height: 900 });
		await registerAndLogin(page);

		// Create dedicated organization for status page tests
		org = await createTestOrg(page, `Status Org ${timestamp.toString(36).slice(-4)}`);

		// Status page creation requires selecting at least one monitor
		monitorName = `Web API Service ${timestamp}`;
		await createTestMonitor(page, org.id, monitorName);
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display empty state on a fresh organization', async () => {
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Verify page title and header
		await expect(page.getByRole('heading', { name: 'Status Pages', exact: true })).toBeVisible();

		// Verify empty state
		await expect(page.getByRole('heading', { name: 'No status pages yet' })).toBeVisible();
		await expect(
			page.getByText('Create a status page to share system status and uptime with your customers.')
		).toBeVisible();

		// Verify "Add status page" CTA button is visible
		const addBtn = page.getByRole('button', { name: 'Add status page' }).first();
		await expect(addBtn).toBeVisible();
	});

	test('should create a public status page with branding configurations', async () => {
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Open Create Status Page modal
		const addBtn = page.getByRole('button', { name: 'Add status page' }).first();
		await addBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Create Status Page' })).toBeVisible();

		// Fill General & Monitors tab
		await dialog.locator('#sp-title').fill(statusPageTitle);
		await dialog.locator('#sp-subdomain').fill(subdomain);
		await dialog.locator('#sp-desc').fill(statusPageDescription);

		// Select the monitored service
		const monitorCard = dialog.locator('button', { hasText: monitorName });
		await expect(monitorCard).toBeVisible();
		await monitorCard.click();

		// Switch to Branding & Styling tab
		const brandingTab = dialog.getByRole('tab', { name: /branding/i });
		await brandingTab.click();

		// Fill branding details
		await dialog.locator('#sp-logo').fill('https://example.com/logo.svg');
		await dialog.locator('#sp-favicon').fill('https://example.com/favicon.ico');
		await dialog.getByPlaceholder('#4edea3').fill('#10b981');
		await dialog.locator('#sp-css').fill('.custom-status { font-family: sans-serif; }');

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Status Page' });
		await expect(submitBtn).toBeEnabled();
		await submitBtn.click();

		// Dialog closes upon creation
		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify status page item is listed in table
		await expect(page.getByText(statusPageTitle)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(subdomain)).toBeVisible();
		await expect(page.getByText(statusPageDescription)).toBeVisible();
		await expect(page.getByText('Public').first()).toBeVisible();
		await expect(page.getByText('1 monitors', { exact: true })).toBeVisible();

		// Verify link to view status page is present
		const viewLink = page.getByRole('link', { name: 'View status page' }).first();
		await expect(viewLink).toBeVisible();
	});

	test('should open Embed Status Badge modal and verify preview and embed codes', async () => {
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Open 3-dot dropdown menu on the status page row
		const moreBtn = page.getByRole('button', { name: 'More options' }).first();
		await moreBtn.click();

		// Click "Embed Badge" option
		const embedOption = page.getByRole('menuitem', { name: 'Embed Badge' });
		await expect(embedOption).toBeVisible();
		await embedOption.click();

		// Verify Embed Badge modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Embed Status Badge' })).toBeVisible();
		await expect(dialog.getByText(statusPageTitle).first()).toBeVisible();

		// Verify Live Badge Preview section
		const previewImg = dialog.getByRole('img', { name: 'Live Status Badge' });
		await expect(previewImg).toBeVisible();

		// Verify embed snippet sections
		await expect(dialog.getByText('Markdown (for GitHub README)')).toBeVisible();
		await expect(dialog.getByText('HTML (for Websites & Portals)')).toBeVisible();
		await expect(dialog.getByText('Direct SVG Badge URL')).toBeVisible();

		// Close modal by pressing Escape
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();
	});

	test('should manage component groups and assign monitors via ComponentGroupModal', async () => {
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Open 3-dot menu and select Groups
		const moreBtn = page.getByRole('button', { name: 'More options' }).first();
		await moreBtn.click();
		await page.getByRole('menuitem', { name: 'Groups' }).click();

		// Verify Component Groups modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Component Groups' })).toBeVisible();

		// Add new component group
		await dialog.locator('#g-name').fill(groupName);
		const addGroupBtn = dialog.getByRole('button', { name: 'Add Group' });
		await expect(addGroupBtn).toBeEnabled();
		await addGroupBtn.click();

		// Verify group appears in list
		await expect(dialog.getByText(groupName)).toBeVisible({ timeout: 10000 });

		// Switch to Assign Monitors tab
		const assignTab = dialog.getByRole('tab', { name: /assign monitors/i });
		await assignTab.click();

		// Assign our monitor to the newly created group
		await expect(dialog.getByText(monitorName)).toBeVisible();
		const groupTrigger = dialog.getByRole('combobox').first();
		await groupTrigger.click();
		await page.getByRole('option', { name: groupName }).click();

		// Switch back to Manage Groups tab to edit group name
		const manageTab = dialog.getByRole('tab', { name: /manage groups/i });
		await manageTab.click();

		// Click Edit Group button for this group
		const editGroupBtn = dialog.getByRole('button', { name: 'Edit Group' }).first();
		await editGroupBtn.click();

		// Verify form switches to Edit Group
		await expect(dialog.getByRole('heading', { name: 'Edit Group' })).toBeVisible();
		await dialog.locator('#g-name').fill(updatedGroupName);
		const updateGroupBtn = dialog.getByRole('button', { name: 'Update Group' });
		await updateGroupBtn.click();

		// Verify group name updated in list
		await expect(dialog.getByText(updatedGroupName)).toBeVisible({ timeout: 10000 });

		// Close modal
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();
	});

	test('should schedule maintenance and update its status via MaintenanceModal', async () => {
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Open 3-dot menu and select Maintenance
		const moreBtn = page.getByRole('button', { name: 'More options' }).first();
		await moreBtn.click();
		await page.getByRole('menuitem', { name: 'Maintenance' }).click();

		// Verify Maintenance modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Scheduled Maintenance' })).toBeVisible();

		// Click "Schedule" button to open creation form
		const scheduleBtn = dialog.getByRole('button', { name: 'Schedule' });
		await scheduleBtn.click();

		// Fill maintenance form
		await dialog.locator('#m-title').fill(maintenanceTitle);
		await dialog.locator('#m-desc').fill(maintenanceDescription);

		// Select affected monitor
		const monitorCheckbox = dialog.locator('button', { hasText: monitorName });
		await monitorCheckbox.click();

		// Save scheduled maintenance
		const saveBtn = dialog.getByRole('button', { name: 'Save Maintenance' });
		await saveBtn.click();

		// Verify maintenance item appears in list with Scheduled badge
		await expect(dialog.getByText(maintenanceTitle)).toBeVisible({ timeout: 10000 });
		await expect(dialog.getByText('Scheduled', { exact: true })).toBeVisible();

		// Edit maintenance to set status to in_progress
		const editBtn = dialog.getByRole('button', { name: 'Edit' }).first();
		await editBtn.click();

		const statusSelect = dialog.getByRole('combobox').first();
		await expect(statusSelect).toBeVisible();
		await statusSelect.click();
		await page.getByRole('option', { name: 'In Progress' }).click();

		const updateBtn = dialog.getByRole('button', { name: 'Update Maintenance' });
		await updateBtn.click();

		// Verify status badge changed to In Progress
		await expect(dialog.getByText('In Progress', { exact: true })).toBeVisible({
			timeout: 10000
		});

		// Close modal
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();
	});

	test('should navigate to public status page view and verify all public sections', async () => {
		// Navigate to public status page route
		await page.goto(`/status/${subdomain}`);
		await page.waitForURL(new RegExp(`/status/${subdomain}`));

		// Verify public header with title and description
		await expect(page.getByText(statusPageTitle).first()).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(statusPageDescription).first()).toBeVisible();

		// Verify overall status banner
		await expect(page.getByRole('heading', { name: 'All Systems Operational' })).toBeVisible();

		// Verify component group section
		await expect(page.getByText(updatedGroupName)).toBeVisible();
		await expect(page.getByText(monitorName).first()).toBeVisible();

		// Verify active/in-progress maintenance banner
		await expect(page.getByRole('heading', { name: 'Ongoing Maintenance' })).toBeVisible();
		await expect(page.getByText(maintenanceTitle)).toBeVisible();
		await expect(page.getByText(maintenanceDescription)).toBeVisible();
		await expect(page.getByText('In Progress').first()).toBeVisible();
	});

	test('should delete status page with confirmation dialog and verify 404', async () => {
		// Return to status page admin dashboard
		await page.goto(`/dashboard/org/${org.id}/statuspage`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// First remove scheduled maintenance to clean up child relations
		const maintMoreBtn = page.getByRole('button', { name: 'More options' }).first();
		await maintMoreBtn.click();
		await page.getByRole('menuitem', { name: 'Maintenance' }).click();

		const modalDialog = page.getByRole('dialog');
		await expect(modalDialog).toBeVisible();
		const deleteMaintBtn = modalDialog.locator('button[title="Delete Maintenance"]').first();
		await deleteMaintBtn.click();
		await expect(modalDialog.getByText('No scheduled maintenances')).toBeVisible({
			timeout: 10000
		});
		await page.keyboard.press('Escape');
		await expect(modalDialog).not.toBeVisible();

		// Open 3-dot dropdown menu to delete status page
		const deleteMoreBtn = page.getByRole('button', { name: 'More options' }).first();
		await deleteMoreBtn.click();

		// Click Delete
		const deleteOption = page.getByRole('menuitem', { name: 'Delete' });
		await expect(deleteOption).toBeVisible();
		await deleteOption.click();

		// Confirm delete in ConfirmDialog (destructive alertdialog)
		const confirmDialog = page.getByRole('alertdialog');
		await expect(confirmDialog).toBeVisible();
		await expect(confirmDialog.getByRole('heading', { name: 'Delete Status Page' })).toBeVisible();

		const confirmDeleteBtn = confirmDialog.getByRole('button', { name: 'Delete' });
		await confirmDeleteBtn.click();

		// Dialog closes and empty state returns
		await expect(confirmDialog).not.toBeVisible({ timeout: 10000 });
		await expect(page.getByRole('heading', { name: 'No status pages yet' })).toBeVisible({
			timeout: 10000
		});

		// Verify public view now displays 404 not found page
		await page.goto(`/status/${subdomain}`);
		await expect(page.getByText('Status Page not found')).toBeVisible({ timeout: 10000 });
	});
});
