import { test, expect, type Page } from '@playwright/test';
import {
	registerAndLogin,
	createTestOrg,
	createTestMonitor,
	insertTestIncident,
	type CreatedOrg
} from './helpers/test-utils';

test.describe.serial('Incident Management Lifecycle', () => {
	let page: Page;
	let org: CreatedOrg;
	let monitorId = '';
	let monitorName = '';
	let ongoingIncidentId = '';

	const timestamp = Date.now();
	const ongoingSummary = `Production Gateway Outage ${timestamp}`;
	const resolvedSummary = `API Latency Spike Resolved ${timestamp}`;
	const statusUpdateMessage = `Root cause identified: upstream connection timeout. Applying failover route at ${timestamp}.`;
	const secondUpdateMessage = `Database failover completed successfully. Error rates returning to baseline.`;

	test.beforeAll(async ({ browser }) => {
		page = await browser.newPage();
		await page.setViewportSize({ width: 1280, height: 900 });
		await registerAndLogin(page);

		// Create dedicated organization for incident tests
		org = await createTestOrg(page, `Incident Org ${timestamp.toString(36).slice(-4)}`);

		// Create a monitor in this organization to attach incidents to
		monitorName = `Core API Gateway ${timestamp}`;
		const monitor = await createTestMonitor(page, org.id, monitorName);
		monitorId = monitor.id;

		// Seed 1 ongoing incident (critical / Sev-1) and 1 resolved incident (medium / Sev-3)
		ongoingIncidentId = insertTestIncident({
			monitorId,
			summary: ongoingSummary,
			status: 'ongoing',
			severity: 'critical',
			startedAt: new Date(Date.now() - 30 * 60 * 1000) // 30 minutes ago
		});

		insertTestIncident({
			monitorId,
			summary: resolvedSummary,
			status: 'resolved',
			severity: 'medium',
			startedAt: new Date(Date.now() - 120 * 60 * 1000), // 2 hours ago
			resolvedAt: new Date(Date.now() - 60 * 60 * 1000) // 1 hour ago
		});
	});

	test.afterAll(async () => {
		await page.close();
	});

	test('should display incident command board with active incidents and metrics', async () => {
		// Navigate to Incidents board
		await page.goto(`/dashboard/org/${org.id}/incidents`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		// Verify page title in browser tab
		await expect(page).toHaveTitle(/Incident Command Board/);

		// Verify Active and Resolved filter tabs
		const activeTab = page.getByRole('button', { name: /Active/ });
		const resolvedTab = page.getByRole('button', { name: /Resolved/ });
		await expect(activeTab).toBeVisible();
		await expect(resolvedTab).toBeVisible();

		// Verify ongoing incident is rendered on the Active tab
		await expect(page.getByText(ongoingSummary)).toBeVisible();

		// Verify severity badge for critical incident
		await expect(page.getByText('Sev-1 Critical').first()).toBeVisible();

		// Verify linked monitor tag
		await expect(page.getByText(monitorName).first()).toBeVisible();

		// Verify Command Room action link
		const commandRoomLink = page.getByRole('link', { name: 'Command Room' }).first();
		await expect(commandRoomLink).toBeVisible();
	});

	test('should toggle between Active and Resolved tabs', async () => {
		await page.goto(`/dashboard/org/${org.id}/incidents`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		// Click Resolved tab
		const resolvedTab = page.getByRole('button', { name: /Resolved/ });
		await resolvedTab.click();

		// Verify resolved incident is displayed
		await expect(page.getByText(resolvedSummary)).toBeVisible();
		await expect(page.getByText(ongoingSummary)).not.toBeVisible();

		// Verify resolved incident metadata
		await expect(page.getByText('Sev-3 Medium').first()).toBeVisible();
		await expect(page.getByText(/Resolved:/).first()).toBeVisible();

		// Switch back to Active tab
		const activeTab = page.getByRole('button', { name: /Active/ });
		await activeTab.click();

		// Verify active incident returns
		await expect(page.getByText(ongoingSummary)).toBeVisible();
		await expect(page.getByText(resolvedSummary)).not.toBeVisible();
	});

	test('should filter incidents by severity code using dropdown', async () => {
		await page.goto(`/dashboard/org/${org.id}/incidents`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		// Open Severity filter dropdown
		const severityDropdown = page.getByRole('button', { name: /severity/i });
		await expect(severityDropdown).toBeVisible();
		await severityDropdown.click();

		// Select Sev 1 (Critical) filter
		const sev1Option = page.getByRole('menuitem', { name: /Sev 1 \(Critical\)/i });
		await expect(sev1Option).toBeVisible();
		await sev1Option.click();

		// Verify ongoing incident (critical) is still visible
		await expect(page.getByText(ongoingSummary)).toBeVisible();

		// Switch severity to Sev 4 (Low) where 0 incidents exist
		await page.getByRole('button', { name: /critical/i }).click();
		const sev4Option = page.getByRole('menuitem', { name: /Sev 4 \(Low\)/i });
		await expect(sev4Option).toBeVisible();
		await sev4Option.click();

		// Verify empty filter state
		await expect(page.getByRole('heading', { name: 'No Incidents Found' })).toBeVisible();
		await expect(
			page.getByText('No incident records match your current search query or filter criteria.')
		).toBeVisible();

		// Reset severity filter to All Severities
		await page.getByRole('button', { name: /low/i }).click();
		await page.getByRole('menuitem', { name: 'All Severities' }).click();

		// Verify incident is visible again
		await expect(page.getByText(ongoingSummary)).toBeVisible();
	});

	test('should filter incidents by search query in top bar', async () => {
		await page.goto(`/dashboard/org/${org.id}/incidents`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		const searchInput = page.getByPlaceholder('Search incidents by summary, monitor...');
		await expect(searchInput).toBeVisible();

		// Search for non-matching string
		await searchInput.fill('NonExistentServiceOutageXYZ');
		await expect(page.getByRole('heading', { name: 'No Incidents Found' })).toBeVisible();

		// Clear search query
		await searchInput.fill('');
		await expect(page.getByText(ongoingSummary)).toBeVisible();
	});

	test('should navigate to incident detail command room and view metadata', async () => {
		await page.goto(`/dashboard/org/${org.id}/incidents`);
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		// Click on the incident summary link
		const incidentLink = page.getByRole('link', { name: ongoingSummary });
		await incidentLink.click();

		// Redirection to incident detail page
		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents/${ongoingIncidentId}`), {
			timeout: 10000
		});

		// Verify incident detail header
		await expect(page.getByRole('heading', { name: ongoingSummary })).toBeVisible();
		await expect(page.getByText('ONGOING')).toBeVisible();
		await expect(page.getByText('Sev-1 Critical')).toBeVisible();

		// Verify timeline and responders sections
		await expect(page.getByRole('heading', { name: 'Timeline' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Command Responders' })).toBeVisible();

		// Verify "Join Incident Bridge" button is visible
		const joinBtn = page.getByRole('button', { name: 'Join Incident Bridge' });
		await expect(joinBtn).toBeVisible();
	});

	test('should join the incident command bridge as an assignee', async () => {
		// Ensure we are on the incident detail page
		await expect(page.getByRole('heading', { name: ongoingSummary })).toBeVisible();

		const joinBtn = page.getByRole('button', { name: 'Join Incident Bridge' });
		await expect(joinBtn).toBeVisible();
		await joinBtn.click();

		// Verify state changes to joined status badge
		await expect(page.getByText('You are on this bridge')).toBeVisible({ timeout: 10000 });

		// Verify user appears in Command Responders column
		const respondersSection = page.locator('section', { hasText: 'Command Responders' });
		await expect(respondersSection.getByText('responder', { exact: true })).toBeVisible();
	});

	test('should post updates to the incident timeline and verify feed', async () => {
		// Ensure we are on the incident detail page
		await expect(page.getByRole('heading', { name: ongoingSummary })).toBeVisible();

		// Verify Post Status Update form is visible
		await expect(page.getByRole('heading', { name: 'Post Status Update' })).toBeVisible();
		const textarea = page.getByPlaceholder(
			'Describe current status, root cause findings, or mitigation steps...'
		);
		const submitUpdateBtn = page.getByRole('button', { name: 'Post Status Update' });

		// Submit button should be disabled when textarea is empty
		await expect(submitUpdateBtn).toBeDisabled();

		// Post first status update
		await textarea.fill(statusUpdateMessage);
		await expect(submitUpdateBtn).toBeEnabled();
		await submitUpdateBtn.click();

		// Verify first update appears in timeline feed
		await expect(page.getByText(statusUpdateMessage)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(/1 update/)).toBeVisible();

		// Post second status update
		await textarea.fill(secondUpdateMessage);
		await expect(submitUpdateBtn).toBeEnabled();
		await submitUpdateBtn.click();

		// Verify second update appears and count increases
		await expect(page.getByText(secondUpdateMessage)).toBeVisible({ timeout: 10000 });
		await expect(page.getByText(/2 updates/)).toBeVisible();
	});

	test('should navigate back to incident board and verify updated assignee count', async () => {
		// Click back button
		const backLink = page.getByRole('link', { name: 'Back to Incidents' });
		await backLink.click();

		await page.waitForURL(new RegExp(`/dashboard/org/${org.id}/incidents`));

		// Verify incident row still present and reflects the joined assignee
		await expect(page.getByText(ongoingSummary)).toBeVisible();
		await expect(page.getByText('1 Assignee')).toBeVisible();
	});
});
