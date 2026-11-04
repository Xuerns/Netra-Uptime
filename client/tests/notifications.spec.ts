import { test, expect, type Page } from '@playwright/test';
import { registerAndLogin, createTestOrg, type CreatedOrg } from './helpers/test-utils';

test.describe.serial('Notification Channels Lifecycle', () => {
	let page: Page;
	let org: CreatedOrg;

	const timestamp = Date.now();
	const emailChannelName = `DevOps Email Alerts ${timestamp}`;
	const emailRecipient = `alerts-${timestamp}@example.com`;
	const slackChannelName = `Slack SRE Incident Room ${timestamp}`;
	const slackWebhookUrl = `https://hooks.slack.com/services/T00/B00/X${timestamp}`;
	const discordChannelName = `Discord Ops Channel ${timestamp}`;
	const discordWebhookUrl = `https://discord.com/api/webhooks/1234567890/token-${timestamp}`;
	const webhookChannelName = `Custom Ingestion Webhook ${timestamp}`;
	const customWebhookUrl = `https://example.com/alerts/webhook-${timestamp}`;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await page.setViewportSize({ width: 1280, height: 900 });
		await registerAndLogin(page);

		// Dedicated organization for notification channel tests
		org = await createTestOrg(page, `Notify Org ${timestamp.toString(36).slice(-4)}`);
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display empty state on a fresh organization', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Check page header
		await expect(page.getByRole('heading', { name: 'Notifications' })).toBeVisible();

		// Check tabs
		const channelsTab = page.getByRole('tab', { name: /channels \(0\)/i });
		const historyTab = page.getByRole('tab', { name: /history logs/i });
		await expect(channelsTab).toBeVisible();
		await expect(historyTab).toBeVisible();

		// Check empty state
		await expect(page.getByRole('heading', { name: 'No notification channels yet' })).toBeVisible();
		await expect(
			page.getByText('Add Email, Slack Incoming Webhooks, Discord, or Custom Webhooks')
		).toBeVisible();

		// Verify "Add Channel" button is available
		const addBtn = page.getByRole('button', { name: 'Add Channel' }).first();
		await expect(addBtn).toBeVisible();
	});

	test('should create an Email notification channel', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Open Add Channel modal
		const addBtn = page.getByRole('button', { name: 'Add Channel' }).first();
		await addBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Add Notification Channel' })).toBeVisible();

		// Fill channel name
		await dialog.locator('#ch-name').fill(emailChannelName);

		// Verify Email is selected by default and fill recipient
		const emailInput = dialog.locator('#ch-email');
		await expect(emailInput).toBeVisible();
		await emailInput.fill(emailRecipient);

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Channel' });
		await expect(submitBtn).toBeEnabled();
		await submitBtn.click();

		// Modal closes
		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify channel is rendered in the list
		await expect(page.getByText(emailChannelName)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('Email').first()).toBeVisible();
		await expect(page.getByText(emailRecipient)).toBeVisible();

		// Verify status switch is Enabled
		const switchBtn = page.getByRole('switch', { name: new RegExp(emailChannelName) });
		await expect(switchBtn).toBeVisible();
		await expect(switchBtn).toHaveAttribute('aria-checked', 'true');
	});

	test('should create a Slack webhook notification channel', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Open Add Channel modal
		const addBtn = page.getByRole('button', { name: 'Add Channel' }).first();
		await addBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		// Fill channel name
		await dialog.locator('#ch-name').fill(slackChannelName);

		// Select Slack channel type
		const slackTypeBtn = dialog.getByRole('button', { name: 'Slack' });
		await slackTypeBtn.click();

		// Fill Slack webhook URL
		const slackInput = dialog.locator('#ch-slack');
		await expect(slackInput).toBeVisible();
		await slackInput.fill(slackWebhookUrl);

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Channel' });
		await submitBtn.click();

		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify Slack channel in list with masked webhook URL
		await expect(page.getByText(slackChannelName)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('Slack').first()).toBeVisible();
		await expect(page.getByText('https://hooks.slack.com/***')).toBeVisible();
	});

	test('should create a Discord webhook notification channel', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Open Add Channel modal
		const addBtn = page.getByRole('button', { name: 'Add Channel' }).first();
		await addBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		// Fill channel name
		await dialog.locator('#ch-name').fill(discordChannelName);

		// Select Discord channel type
		const discordTypeBtn = dialog.getByRole('button', { name: 'Discord' });
		await discordTypeBtn.click();

		// Fill Discord webhook URL
		const discordInput = dialog.locator('#ch-discord');
		await expect(discordInput).toBeVisible();
		await discordInput.fill(discordWebhookUrl);

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Channel' });
		await submitBtn.click();

		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify Discord channel in list with masked webhook URL
		await expect(page.getByText(discordChannelName)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('Discord').first()).toBeVisible();
		await expect(page.getByText('https://discord.com/***')).toBeVisible();
	});

	test('should create a Custom Webhook notification channel', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Open Add Channel modal
		const addBtn = page.getByRole('button', { name: 'Add Channel' }).first();
		await addBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		// Fill channel name
		await dialog.locator('#ch-name').fill(webhookChannelName);

		// Select Webhook channel type
		const webhookTypeBtn = dialog.getByRole('button', { name: 'Webhook' });
		await webhookTypeBtn.click();

		// Fill Webhook destination URL
		const webhookInput = dialog.locator('#ch-webhook');
		await expect(webhookInput).toBeVisible();
		await webhookInput.fill(customWebhookUrl);

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create Channel' });
		await submitBtn.click();

		await expect(dialog).not.toBeVisible({ timeout: 10000 });

		// Verify Webhook channel in list with masked URL
		await expect(page.getByText(webhookChannelName)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('Webhook').first()).toBeVisible();
		await expect(page.getByText('https://example.com/***')).toBeVisible();

		// Verify tab reflects 4 channels
		await expect(page.getByRole('tab', { name: /channels \(4\)/i })).toBeVisible();
	});

	test('should toggle active status of a notification channel', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Locate switch for the email channel
		const switchBtn = page.getByRole('switch', { name: new RegExp(emailChannelName) });
		await expect(switchBtn).toBeVisible();
		await expect(switchBtn).toHaveAttribute('aria-checked', 'true');
		await expect(switchBtn.getByText('Enabled')).toBeVisible();

		// Click to disable
		await switchBtn.click();

		// Verify status updates to Disabled
		await expect(switchBtn).toHaveAttribute('aria-checked', 'false', { timeout: 10000 });
		await expect(switchBtn.getByText('Disabled')).toBeVisible();

		// Click to re-enable
		await switchBtn.click();

		// Verify status returns to Enabled
		await expect(switchBtn).toHaveAttribute('aria-checked', 'true', { timeout: 10000 });
		await expect(switchBtn.getByText('Enabled')).toBeVisible();
	});

	test('should delete a notification channel with confirmation dialog', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Find the row for the custom webhook channel
		const channelRow = page
			.locator('div.grid')
			.filter({ has: page.getByText(webhookChannelName, { exact: true }) });
		await expect(channelRow).toBeVisible();

		// Click delete channel button inside this row
		const deleteBtn = channelRow.getByRole('button', { name: 'Delete channel' });
		await expect(deleteBtn).toBeVisible();
		await deleteBtn.click();

		// Verify confirmation alert dialog
		const dialog = page.getByRole('alertdialog');
		await expect(dialog).toBeVisible();
		await expect(
			dialog.getByRole('heading', { name: 'Delete notification channel?' })
		).toBeVisible();
		await expect(dialog.getByText(new RegExp(webhookChannelName))).toBeVisible();

		// Confirm deletion
		const confirmBtn = dialog.getByRole('button', { name: 'Delete' });
		await expect(confirmBtn).toBeVisible();
		await confirmBtn.click();

		// Dialog closes and channel is removed
		await expect(dialog).not.toBeVisible({ timeout: 10000 });
		await expect(page.getByText(webhookChannelName)).not.toBeVisible();

		// Verify tab counter decreases to 3
		await expect(page.getByRole('tab', { name: /channels \(3\)/i })).toBeVisible();
	});

	test('should switch to notification history logs tab and verify empty state', async () => {
		await page.goto(`/dashboard/org/${org.id}/notifications`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/notifications`));

		// Switch to History Logs tab
		const historyTab = page.getByRole('tab', { name: /history logs/i });
		await historyTab.click();

		// Verify history tab content
		await expect(page.getByRole('heading', { name: 'No notification logs' })).toBeVisible();
		await expect(
			page.getByText('History will appear here when monitor alerts are triggered')
		).toBeVisible();

		// Switch back to Channels tab
		const channelsTab = page.getByRole('tab', { name: /channels/i });
		await channelsTab.click();

		// Verify remaining channels still visible
		await expect(page.getByText(emailChannelName)).toBeVisible();
		await expect(page.getByText(slackChannelName)).toBeVisible();
		await expect(page.getByText(discordChannelName)).toBeVisible();
	});
});
