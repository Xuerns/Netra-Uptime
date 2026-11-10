import { test, expect, type BrowserContext, type Page } from '@playwright/test';
import {
	registerAndLogin,
	createTestOrg,
	generateTestUser,
	type TestUser,
	type CreatedOrg
} from './helpers/test-utils';

test.describe.serial('Multi-Role RBAC & Invitation Verification Flow', () => {
	let ownerContext: BrowserContext;
	let adminContext: BrowserContext;
	let memberContext: BrowserContext;

	let ownerPage: Page;
	let adminPage: Page;
	let memberPage: Page;

	let ownerUser: TestUser;
	let adminUser: TestUser;
	let memberUser: TestUser;

	let org: CreatedOrg;
	let createdMonitorName: string;
	let createdSpName: string;

	test.beforeAll(async ({ browser }) => {
		// Create 3 independent browser contexts to simulate 3 distinct users simultaneously
		ownerContext = await browser.newContext();
		adminContext = await browser.newContext();
		memberContext = await browser.newContext();

		ownerPage = await ownerContext.newPage();
		adminPage = await adminContext.newPage();
		memberPage = await memberContext.newPage();

		// Generate test credentials
		ownerUser = generateTestUser();
		adminUser = generateTestUser();
		memberUser = generateTestUser();
	});

	test.afterAll(async () => {
		await ownerContext.close();
		await adminContext.close();
		await memberContext.close();
	});

	test('Step 1: Self-register all three accounts (Owner, Admin, Member)', async () => {
		// Register Owner
		await registerAndLogin(ownerPage, ownerUser);
		await expect(ownerPage).toHaveURL(/\/dashboard\/org/);

		// Register Admin-to-be
		await registerAndLogin(adminPage, adminUser);
		await expect(adminPage).toHaveURL(/\/dashboard\/org/);

		// Register Member-to-be
		await registerAndLogin(memberPage, memberUser);
		await expect(memberPage).toHaveURL(/\/dashboard\/org/);
	});

	test('Step 2: Owner provisions new organization workspace', async () => {
		org = await createTestOrg(ownerPage);
		expect(org.id).toBeTruthy();
		await expect(ownerPage).toHaveURL(new RegExp(`/dashboard/org/${org.id}/monitors`));
	});

	test('Step 3: Owner invites Admin and Member via Invite Modal', async () => {
		// Navigate to Members page
		await ownerPage.goto(`/dashboard/org/${org.id}/members`);
		await ownerPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/members`));

		// Click Invite member button
		const inviteBtn = ownerPage.getByRole('button', { name: /invite member/i });
		await expect(inviteBtn).toBeVisible({ timeout: 5000 });
		await inviteBtn.click();

		const dialog = ownerPage.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole('heading', { name: 'Invite Members' })).toBeVisible();

		// Search and add Admin User
		const searchInput = dialog.locator('#search-user-input');
		await searchInput.fill(adminUser.name);

		const adminOption = dialog.locator('ul button', { hasText: adminUser.name });
		await expect(adminOption).toBeVisible({ timeout: 7000 });
		await adminOption.click();

		// Set Admin role
		const adminRoleBtn = dialog.getByRole('button', { name: 'Admin', exact: true });
		await expect(adminRoleBtn).toBeVisible();
		await adminRoleBtn.click();

		// Search and add Member User
		await searchInput.fill(memberUser.name);
		const memberOption = dialog.locator('ul button', { hasText: memberUser.name });
		await expect(memberOption).toBeVisible({ timeout: 7000 });
		await memberOption.click();

		// Send invitations
		const sendBtn = dialog.getByRole('button', { name: /send invitation/i });
		await expect(sendBtn).toBeEnabled();
		await sendBtn.click();

		// Modal closes upon successful dispatch
		await expect(dialog).not.toBeVisible({ timeout: 10000 });
	});

	test('Step 4: Admin receives notification and accepts invitation', async () => {
		await adminPage.goto('/dashboard/notification');
		await adminPage.waitForURL(/\/dashboard\/notification/);

		// Locate invitation card
		const inviteCard = adminPage.locator('.bg-card', {
			hasText: org.name
		});
		await expect(inviteCard).toBeVisible({ timeout: 10000 });
		await expect(inviteCard.getByText('Role: admin', { exact: false })).toBeVisible();

		// Click Accept button
		const acceptBtn = inviteCard.getByRole('button', { name: /accept/i });
		await acceptBtn.click();

		// Verify confirmation toast
		await expect(adminPage.getByText(/Invitation accepted/i)).toBeVisible({ timeout: 5000 });

		// Navigate into the organization
		await adminPage.goto('/dashboard/org');
		const orgCard = adminPage.locator('button', {
			has: adminPage.getByRole('heading', { name: org.name, exact: true })
		});
		await expect(orgCard).toBeVisible({ timeout: 7000 });
		await orgCard.click();
		await adminPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`), {
			timeout: 10000
		});
	});

	test('Step 5: Member receives notification and accepts invitation', async () => {
		await memberPage.goto('/dashboard/notification');
		await memberPage.waitForURL(/\/dashboard\/notification/);

		// Locate invitation card
		const inviteCard = memberPage.locator('.bg-card', {
			hasText: org.name
		});
		await expect(inviteCard).toBeVisible({ timeout: 10000 });
		await expect(inviteCard.getByText('Role: member', { exact: false })).toBeVisible();

		// Click Accept button
		const acceptBtn = inviteCard.getByRole('button', { name: /accept/i });
		await acceptBtn.click();

		// Verify confirmation toast
		await expect(memberPage.getByText(/Invitation accepted/i)).toBeVisible({ timeout: 5000 });

		// Navigate into the organization
		await memberPage.goto('/dashboard/org');
		const orgCard = memberPage.locator('button', {
			has: memberPage.getByRole('heading', { name: org.name, exact: true })
		});
		await expect(orgCard).toBeVisible({ timeout: 7000 });
		await orgCard.click();
		await memberPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`), {
			timeout: 10000
		});
	});

	test('Step 6: Owner verifies all 3 members and tests role management & org update', async () => {
		await ownerPage.goto(`/dashboard/org/${org.id}/members`);
		await ownerPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/members`));

		const membersContent = ownerPage.locator('.members-page');
		// Verify Owner entry
		await expect(membersContent.getByText(ownerUser.name)).toBeVisible();
		await expect(membersContent.getByText('Owner', { exact: true })).toBeVisible();

		// Verify Admin entry
		await expect(membersContent.getByText(adminUser.name)).toBeVisible();
		await expect(membersContent.getByText('Admin', { exact: true })).toBeVisible();

		// Verify Member entry
		await expect(membersContent.getByText(memberUser.name)).toBeVisible();
		await expect(membersContent.getByText('Member', { exact: true })).toBeVisible();

		// Test Role Promote: promote Member to Admin
		const memberRow = membersContent
			.locator('.divide-y > div')
			.filter({ hasText: memberUser.name });
		await memberRow.locator('button').click();
		const promoteOption = ownerPage.getByRole('menuitem', { name: /make admin/i });
		await expect(promoteOption).toBeVisible({ timeout: 5000 });
		await promoteOption.click();
		await expect(ownerPage.getByText(/promoted to admin/i).last()).toBeVisible({ timeout: 5000 });
		await expect(memberRow.getByText('Admin', { exact: true })).toBeVisible({ timeout: 5000 });

		// Test Role Demote: demote Member back to Member
		await memberRow.locator('button').click();
		const demoteMemberOption = ownerPage.getByRole('menuitem', { name: /demote to member/i });
		await expect(demoteMemberOption).toBeVisible({ timeout: 5000 });
		await demoteMemberOption.click();
		await expect(ownerPage.getByText(/demoted to member/i).last()).toBeVisible({ timeout: 5000 });
		await expect(memberRow.getByText('Member', { exact: true })).toBeVisible({ timeout: 5000 });

		// Test Role Demote on Admin user: demote Admin to Member
		const adminRow = membersContent.locator('.divide-y > div').filter({ hasText: adminUser.name });
		await adminRow.locator('button').click();
		const demoteAdminOption = ownerPage.getByRole('menuitem', { name: /demote to member/i });
		await expect(demoteAdminOption).toBeVisible({ timeout: 5000 });
		await demoteAdminOption.click();
		await expect(ownerPage.getByText(/demoted to member/i).last()).toBeVisible({ timeout: 5000 });
		await expect(adminRow.getByText('Member', { exact: true })).toBeVisible({ timeout: 5000 });

		// Restore Admin role for Step 7: promote Admin back to Admin
		await adminRow.locator('button').click();
		const rePromoteAdminOption = ownerPage.getByRole('menuitem', { name: /make admin/i });
		await expect(rePromoteAdminOption).toBeVisible({ timeout: 5000 });
		await rePromoteAdminOption.click();
		await expect(ownerPage.getByText(/promoted to admin/i).last()).toBeVisible({ timeout: 5000 });
		await expect(adminRow.getByText('Admin', { exact: true })).toBeVisible({ timeout: 5000 });

		// Update Org Settings
		await ownerPage.goto(`/dashboard/org/${org.id}/settings`);
		const nameInput = ownerPage.locator('#name');
		await expect(nameInput).toBeVisible();
		const updatedName = `Renamed ${Date.now().toString().slice(-4)}`;
		await nameInput.fill(updatedName);
		await ownerPage.getByRole('button', { name: 'Save Changes' }).click();
		await expect(ownerPage.getByText(/Organization updated successfully/i)).toBeVisible({
			timeout: 5000
		});
	});

	test('Step 7: Admin verifies operational capabilities (Create Monitor, Status Page) and RBAC boundaries', async () => {
		// Admin creates a Monitor
		await adminPage.goto(`/dashboard/org/${org.id}/monitors`);
		await adminPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		const addBtn = adminPage.getByRole('button', { name: 'Add monitor' }).first();
		await expect(addBtn).toBeVisible();
		await addBtn.click();
		await adminPage.getByRole('menuitem', { name: /http\/api/i }).click();

		const monitorDialog = adminPage.getByRole('dialog');
		await expect(monitorDialog).toBeVisible();
		createdMonitorName = `Admin Mon ${Date.now().toString().slice(-4)}`;
		await monitorDialog.locator('#monitor-name').fill(createdMonitorName);
		await monitorDialog.locator('#monitor-target').fill('https://httpbin.org/status/200');

		const submitBtn = monitorDialog.getByRole('button', { name: 'Create Monitor' });
		await expect(submitBtn).toBeEnabled();
		await submitBtn.click();
		await expect(monitorDialog).not.toBeVisible({ timeout: 10000 });
		await expect(adminPage.getByRole('heading', { name: createdMonitorName })).toBeVisible({
			timeout: 10000
		});

		// Admin creates a Status Page
		await adminPage.goto(`/dashboard/org/${org.id}/statuspage`);
		await adminPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		const addSpBtn = adminPage.getByRole('button', { name: 'Add status page' }).first();
		await expect(addSpBtn).toBeVisible();
		await addSpBtn.click();

		const spDialog = adminPage.getByRole('dialog');
		await expect(spDialog).toBeVisible();
		createdSpName = `Status ${Date.now().toString().slice(-4)}`;
		const spSubdomain = `sp-${Date.now().toString().slice(-5)}`;
		await spDialog.locator('#sp-title').fill(createdSpName);
		await spDialog.locator('#sp-subdomain').fill(spSubdomain);

		// Select the monitored service
		const monitorCard = spDialog.locator('button', { hasText: createdMonitorName });
		await expect(monitorCard).toBeVisible();
		await monitorCard.click();

		await spDialog.getByRole('button', { name: 'Create Status Page' }).click();
		await expect(spDialog).not.toBeVisible({ timeout: 10000 });

		// Admin operational capability: Create Alert Rule Template
		const alertTemplateRes = await adminPage.request.post(
			`/api/v1/organization/${org.id}/alert-templates`,
			{
				data: {
					name: `Admin Alert Policy ${Date.now()}`,
					description: 'Standard alert rules created by admin',
					rules: [
						{
							name: 'High Latency Alert',
							metric: 'response_time_ms',
							operator: '>',
							target_value: '500',
							consecutive: 3,
							severity: 'warning'
						}
					]
				}
			}
		);
		expect(alertTemplateRes.status()).toBe(201);

		// Admin operational capability: Create Notification Channel
		const channelRes = await adminPage.request.post(
			`/api/v1/organization/${org.id}/notifications/channels`,
			{
				data: {
					name: `Admin Email Alerts ${Date.now()}`,
					type: 'email',
					config: {
						email: `admin-alerts-${Date.now()}@example.com`
					}
				}
			}
		);
		if (channelRes.status() !== 201) {
			console.error('channelRes failed with body:', await channelRes.text());
		}
		expect(channelRes.status()).toBe(201);

		// Admin RBAC Boundary Checks: Direct API security enforcement (must return 403 Forbidden)
		// 1. Admin cannot disband the organization
		const disbandRes = await adminPage.request.delete(`/api/v1/organization/${org.id}/disband`);
		expect(disbandRes.status()).toBe(403);

		// 2. Admin cannot kick members
		const kickRes = await adminPage.request.delete(`/api/v1/organization/${org.id}/kick`, {
			data: JSON.stringify('00000000-0000-0000-0000-000000000000')
		});
		expect(kickRes.status()).toBe(403);

		// 3. Admin cannot promote members
		const promoteRes = await adminPage.request.put(`/api/v1/organization/${org.id}/promote`, {
			data: { user_id: '00000000-0000-0000-0000-000000000000' }
		});
		expect(promoteRes.status()).toBe(403);

		// 4. Admin cannot demote members
		const demoteRes = await adminPage.request.put(`/api/v1/organization/${org.id}/demote`, {
			data: { user_id: '00000000-0000-0000-0000-000000000000' }
		});
		expect(demoteRes.status()).toBe(403);

		// 5. Admin cannot update organization settings
		const updateOrgRes = await adminPage.request.put(`/api/v1/organization/${org.id}/update`, {
			data: { name: 'Admin Renamed Org' }
		});
		expect(updateOrgRes.status()).toBe(403);

		// Admin UI RBAC Boundary Check: Settings page
		await adminPage.goto(`/dashboard/org/${org.id}/settings`);
		await adminPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/settings`));
		await expect(
			adminPage.getByRole('button', { name: /disband organization/i })
		).not.toBeVisible();
		await expect(adminPage.getByRole('button', { name: /transfer/i })).not.toBeVisible();
		await expect(adminPage.locator('#name')).not.toBeVisible();

		// Admin UI RBAC Boundary Check: Members page
		await adminPage.goto(`/dashboard/org/${org.id}/members`);
		await adminPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/members`));
		await expect(adminPage.locator('.members-page [role="menu"]')).not.toBeVisible();
	});

	test('Step 8: Member verifies read-only access and strict RBAC enforcement', async () => {
		// 1. Monitors view & permissions
		await memberPage.goto(`/dashboard/org/${org.id}/monitors`);
		await memberPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/monitors`));

		// Member can view existing monitor created by Admin
		await expect(memberPage.getByRole('heading', { name: createdMonitorName })).toBeVisible({
			timeout: 10000
		});

		// Member should NOT see "Add monitor" button
		const addMonitorBtn = memberPage.getByRole('button', { name: 'Add monitor' });
		await expect(addMonitorBtn).not.toBeVisible();

		// 2. Status Pages view & permissions
		await memberPage.goto(`/dashboard/org/${org.id}/statuspage`);
		await memberPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/statuspage`));

		// Member can view existing status page created by Admin
		await expect(memberPage.getByText(createdSpName)).toBeVisible({ timeout: 10000 });

		// Member should NOT see "Add status page" button
		const addSpBtn = memberPage.getByRole('button', { name: 'Add status page' });
		await expect(addSpBtn).not.toBeVisible();

		// 3. Members list view & permissions
		await memberPage.goto(`/dashboard/org/${org.id}/members`);
		await memberPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/members`));

		// Member can see all members listed in members-page
		const membersList = memberPage.locator('.members-page');
		await expect(membersList.getByText(ownerUser.name).first()).toBeVisible({ timeout: 10000 });
		await expect(membersList.getByText(adminUser.name).first()).toBeVisible({ timeout: 10000 });
		await expect(membersList.getByText(memberUser.name).first()).toBeVisible({ timeout: 10000 });

		// Member should NOT see "Invite member" CTA button
		const inviteBtn = memberPage.getByRole('button', { name: /invite member/i });
		await expect(inviteBtn).not.toBeVisible();

		// Member should NOT see administrative action dropdown menus
		const memberActionMenus = memberPage.locator('.members-page [aria-haspopup="menu"]');
		await expect(memberActionMenus).toHaveCount(0);

		// 4. Audit logs view
		await memberPage.goto(`/dashboard/org/${org.id}/logs`);
		await memberPage.waitForURL(new RegExp(`/dashboard/org/${org.id}/logs`));
		await expect(
			memberPage
				.getByText(/No Audit Logs Found/i)
				.or(memberPage.getByPlaceholder(/Search audit logs/i))
		).toBeVisible({ timeout: 10000 });

		// 5. Strict RBAC Boundary Checks via Direct API calls
		// POST /monitors must return 403 Forbidden
		const forbiddenCreateMonitor = await memberPage.request.post(
			`/api/v1/organization/${org.id}/monitors`,
			{
				data: {
					name: 'Hacker Monitor',
					type: 'http',
					target: 'https://evil.com',
					interval: 60,
					timeout: 10
				}
			}
		);
		expect(forbiddenCreateMonitor.status()).toBe(403);

		// POST /statuspage must return 403 Forbidden
		const forbiddenCreateStatusPage = await memberPage.request.post(
			`/api/v1/organization/${org.id}/statuspage`,
			{
				data: {
					title: 'Hacker Status Page',
					subdomain: `hack-${Date.now().toString().slice(-5)}`,
					monitors: []
				}
			}
		);
		expect(forbiddenCreateStatusPage.status()).toBe(403);

		// POST /invite must return 403 Forbidden
		const forbiddenInvite = await memberPage.request.post(`/api/v1/organization/${org.id}/invite`, {
			data: {
				invitations: [{ invitee_id: 'some-user-id', role: 'member' }]
			}
		});
		expect(forbiddenInvite.status()).toBe(403);

		// DELETE /disband must return 403 Forbidden
		const forbiddenDisband = await memberPage.request.delete(
			`/api/v1/organization/${org.id}/disband`
		);
		expect(forbiddenDisband.status()).toBe(403);
	});
});
