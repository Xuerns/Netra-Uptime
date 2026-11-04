import { test, expect, type Page } from '@playwright/test';
import { registerAndLogin, createTestOrg, type CreatedOrg } from './helpers/test-utils';

test.describe.serial('Alert Templates Management Lifecycle', () => {
	let page: Page;
	let org: CreatedOrg;

	const timestamp = Date.now();
	const templateName = `Web Services Policy ${timestamp}`;
	let templateDescription = 'Standard alert rules for production web services';

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await registerAndLogin(page);
		// Fresh test organization with 0 templates
		org = await createTestOrg(page, `Alerts Org ${timestamp.toString(36).slice(-4)}`);
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display alert templates list with default template or empty state on a fresh organization', async () => {
		// Navigate via sidebar link to Alert Templates
		const alertTemplatesLink = page.getByRole('link', { name: 'Alert Templates' }).first();
		await expect(alertTemplatesLink).toBeVisible();
		await alertTemplatesLink.click();

		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates`));

		// Check page header
		await expect(page.getByRole('heading', { name: 'Alert Templates' }).first()).toBeVisible();

		// Check either seeded default template or empty state
		const defaultTemplate = page.getByRole('link', { name: 'Standard Web Service' });
		const emptyHeading = page.getByRole('heading', { name: 'No alert templates yet' });
		await expect(defaultTemplate.or(emptyHeading)).toBeVisible();

		// Check create template button
		const createBtn = page.getByRole('button', { name: 'Create Template' }).first();
		await expect(createBtn).toBeVisible();
	});

	test('should open create modal, fill metadata, set default, and redirect to detail page', async () => {
		await page.goto(`/dashboard/org/${org.id}/alert-templates`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates`));

		// Open Create Template modal
		const createBtn = page.getByRole('button', { name: 'Create Template' }).first();
		await createBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Create Alert Template' })).toBeVisible();

		// Fill name and description
		await dialog.locator('#template-name').fill(templateName);
		await dialog.locator('#template-desc').fill(templateDescription);

		// Set as organization default
		const defaultCheckbox = dialog.locator('#template-is-default');
		await defaultCheckbox.check();

		// Submit creation
		const submitBtn = dialog.getByRole('button', { name: 'Create & Configure Rules' });
		await expect(submitBtn).toBeEnabled();
		await submitBtn.click();

		// Verify automatic redirection to the template detail page
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates/[^/]+`), {
			timeout: 10000
		});

		// Verify header elements on detail page
		await expect(page.getByRole('heading', { name: templateName })).toBeVisible();
		await expect(page.getByText('Default')).toBeVisible();
		await expect(page.getByText(templateDescription)).toBeVisible();

		// Verify empty rules state
		await expect(page.getByRole('heading', { name: 'No alert rules defined yet' })).toBeVisible();
	});

	test('should add a latency threshold alert rule via AlertRuleModal', async () => {
		// Ensure we are on the template detail page
		await expect(page.getByRole('heading', { name: templateName })).toBeVisible();

		// Click "Add Rule" button
		const addRuleBtn = page.getByRole('button', { name: /add rule|add first rule/i }).first();
		await expect(addRuleBtn).toBeVisible();
		await addRuleBtn.click();

		// Modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Add Alert Rule' })).toBeVisible();

		// Fill rule name
		await dialog.locator('#rule-name-input').fill('High Latency Warning');

		// Select Metric: Latency (ms)
		await dialog.locator('#rule-metric').selectOption('response_time_ms');

		// Select Operator: > (Greater than)
		await dialog.locator('#rule-operator').selectOption('>');

		// Fill Target Value
		await dialog.locator('#rule-target').fill('1500');

		// Fill Consecutive Breaches
		await dialog.locator('#rule-consecutive').fill('3');

		// Select Severity: Warning
		await dialog.getByRole('button', { name: 'Warning' }).click();

		// Submit rule creation
		const submitRuleBtn = dialog.getByRole('button', { name: 'Add Rule' });
		await expect(submitRuleBtn).toBeEnabled();
		await submitRuleBtn.click();

		// Dialog closes and toast appears
		await expect(dialog).not.toBeVisible({ timeout: 5000 });
		await expect(page.getByText(/alert rule added to template/i)).toBeVisible({ timeout: 5000 });

		// Verify rule is listed in the rules section
		await expect(page.getByText('High Latency Warning')).toBeVisible();
		await expect(page.getByText('warning', { exact: true })).toBeVisible();
		await expect(page.getByText(/response_time_ms\s*>\s*1500/)).toBeVisible();
		await expect(page.getByText(/3x consecutive/)).toBeVisible();
	});

	test('should toggle pause and activate on an alert rule', async () => {
		await expect(page.getByText('High Latency Warning')).toBeVisible();

		// Pause the rule
		const pauseBtn = page.getByRole('button', { name: 'Pause' });
		await expect(pauseBtn).toBeVisible();
		await pauseBtn.click();

		// Verify "Paused" badge and sonner toast
		await expect(page.getByText(/rule paused/i)).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('Paused', { exact: true })).toBeVisible();

		// Reactivate the rule
		const activateBtn = page.getByRole('button', { name: 'Activate' });
		await expect(activateBtn).toBeVisible();
		await activateBtn.click();

		// Verify toast and badge removal
		await expect(page.getByText(/rule activated/i)).toBeVisible({ timeout: 5000 });
		await expect(page.getByText('Paused', { exact: true })).not.toBeVisible();
	});

	test('should add a critical status code alert rule', async () => {
		// Open Add Rule modal
		const addRuleBtn = page.getByRole('button', { name: 'Add Rule' }).first();
		await addRuleBtn.click();

		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();

		// Fill status code rule
		await dialog.locator('#rule-name-input').fill('5xx Server Error Alert');
		await dialog.locator('#rule-metric').selectOption('status_code');
		await dialog.locator('#rule-operator').selectOption('==');
		await dialog.locator('#rule-target').fill('5xx');
		await dialog.locator('#rule-consecutive').fill('2');

		// Severity is critical by default
		const submitRuleBtn = dialog.getByRole('button', { name: 'Add Rule' });
		await submitRuleBtn.click();

		// Verify both rules are visible
		await expect(dialog).not.toBeVisible({ timeout: 5000 });
		await expect(page.getByText('5xx Server Error Alert')).toBeVisible();
		await expect(page.getByText('critical', { exact: true })).toBeVisible();
		await expect(page.getByText('High Latency Warning')).toBeVisible();
	});

	test('should edit template metadata and verify updates', async () => {
		// Click "Edit Info" button in detail header
		const editBtn = page.getByRole('button', { name: 'Edit Info' });
		await expect(editBtn).toBeVisible();
		await editBtn.click();

		// Modal opens
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Edit Alert Template' })).toBeVisible();

		// Update description
		const updatedDescription = 'Updated alert rules for production web applications';
		await dialog.locator('#template-desc').fill(updatedDescription);

		// Submit changes
		const saveBtn = dialog.getByRole('button', { name: 'Save Changes' });
		await expect(saveBtn).toBeEnabled();
		await saveBtn.click();

		// Modal closes and updated description is visible
		await expect(dialog).not.toBeVisible({ timeout: 5000 });
		await expect(page.getByText(/template updated successfully/i)).toBeVisible({ timeout: 5000 });
		await expect(page.getByText(updatedDescription)).toBeVisible();

		templateDescription = updatedDescription;
	});

	test('should verify template displayed on alert templates list page', async () => {
		// Navigate back to the templates list
		await page.getByRole('link', { name: 'Back to Alert Templates' }).click();
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates`));

		// Find row containing the created template
		const templateRow = page.locator('div.group', {
			has: page.getByRole('link', { name: templateName, exact: true })
		});
		await expect(templateRow).toBeVisible();
		await expect(templateRow.getByText('Default')).toBeVisible();
		await expect(templateRow.getByText(/2 rules/)).toBeVisible();
	});

	test('should delete template with confirmation dialog and verify removal', async () => {
		// Navigate to template detail page
		await page.getByRole('link', { name: templateName, exact: true }).click();
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates/[^/]+`));

		// Click "Delete" button in detail header
		const deleteBtn = page.getByRole('button', { name: 'Delete', exact: true });
		await expect(deleteBtn).toBeVisible();
		await deleteBtn.click();

		// Confirm dialog opens
		const alertDialog = page.getByRole('alertdialog');
		await expect(alertDialog).toBeVisible();
		await expect(alertDialog.getByRole('heading', { name: 'Delete Alert Template' })).toBeVisible();
		await expect(
			alertDialog.getByText(/Are you sure you want to delete this alert template/i)
		).toBeVisible();

		// Confirm delete action
		const confirmBtn = alertDialog.getByRole('button', { name: 'Delete Template' });
		await confirmBtn.click();

		// Redirected back to templates list page
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/alert-templates`), {
			timeout: 10000
		});

		// Verify template is no longer displayed
		await expect(page.getByRole('link', { name: templateName, exact: true })).not.toBeVisible({
			timeout: 10000
		});
	});
});
