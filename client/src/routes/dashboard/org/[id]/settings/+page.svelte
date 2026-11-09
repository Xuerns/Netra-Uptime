<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { AlertTriangle } from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import { DataSkeleton } from '$lib/components/common';
	import { OrgAPI } from '$lib/api/organization';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { logger } from '$lib/utils/logger';
	import { toast } from 'svelte-sonner';
	import {
		TransferOwnershipModal,
		DisbandOrgModal,
		OrgGeneralSettingsCard,
		OrgDangerZoneCard
	} from '$lib/components/org';
	import type { Member } from '$lib/components/org/TransferOwnershipModal.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { getCurrentUserId } from '$lib/utils/role';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// Ensure top navigation header is clean
	$effect(() => {
		headerStore.reset();
	});

	const orgId = $derived(data.orgId || ($page.params.id as string) || '');
	const currentUserId = $derived(getCurrentUserId(data?.user ?? $page.data?.user));

	let isLoading = $state(true);
	let hasAccess = $state(false);
	let currentUserRole = $state('');
	const isOwner = $derived(
		currentUserRole ? currentUserRole === 'owner' : (data?.isOwner ?? false)
	);
	let eligibleMembers = $state<Member[]>([]);

	let isTransferModalOpen = $state(false);
	let isDisbandModalOpen = $state(false);
	let isLeaveModalOpen = $state(false);

	let orgName = $state('');
	let orgSlug = $state('');

	onMount(async () => {
		try {
			// Fetch members to check role
			const membersRes = await OrgAPI.getMembers(orgId);
			const members: Member[] = membersRes.data || [];
			const currentUser = members.find((m) => m.UserID === currentUserId);

			if (currentUser) {
				hasAccess = true;
				currentUserRole = currentUser.Role ?? '';
				eligibleMembers = members.filter((m) => m.UserID !== currentUserId);

				if (currentUserRole === 'owner') {
					// Fetch org details only if owner
					const orgsRes = await OrgAPI.getAll();
					const org = orgsRes.data?.find((o: { ID?: string }) => o.ID === orgId);
					if (org) {
						orgName = org.Name;
						orgSlug = org.Slug;
					}
				}
			}
		} catch (error) {
			logger.error('Failed to load organization data:', error);
			toast.error('Failed to load organization settings');
		} finally {
			isLoading = false;
		}
	});

	async function leaveOrganization() {
		try {
			await OrgAPI.leaveOrganization(orgId);
			if (typeof window !== 'undefined') {
				localStorage.removeItem('netra_last_org');
			}
			toast.success('You have left the organization');
			goto('/dashboard/org');
		} catch (error: unknown) {
			toast.error(parseApiError(error, 'Failed to leave organization'));
		} finally {
			isLeaveModalOpen = false;
		}
	}
</script>

<svelte:head>
	<title>Organization Settings | Netra Uptime</title>
</svelte:head>

<div class="p-6 max-w-4xl mx-auto space-y-6">
	{#if isLoading}
		<DataSkeleton variant="form" />
	{:else if !hasAccess}
		<Card.Root class="bg-card border-red-900/50">
			<Card.Content class="p-8 text-center space-y-4">
				<div class="mx-auto size-12 bg-red-500/10 rounded-full flex items-center justify-center">
					<AlertTriangle class="size-6 text-red-500" />
				</div>
				<h2 class="text-xl font-semibold text-foreground">Access Denied</h2>
				<p class="text-muted-foreground max-w-md mx-auto">
					You do not have permission to view this page. You must be an active member of this
					organization.
				</p>
				<div class="pt-4">
					<Button
						href="/dashboard/org"
						class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground transition-colors cursor-pointer"
					>
						Back to Organizations
					</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{:else}
		<!-- General Settings (Owner Only) -->
		{#if isOwner}
			<OrgGeneralSettingsCard
				{orgId}
				initialName={orgName}
				initialSlug={orgSlug}
				onUpdated={(name, slug) => {
					orgName = name;
					orgSlug = slug;
				}}
			/>
		{/if}

		<!-- Danger Zone -->
		<OrgDangerZoneCard
			{isOwner}
			onOpenTransfer={() => (isTransferModalOpen = true)}
			onOpenDisband={() => (isDisbandModalOpen = true)}
			onOpenLeave={() => (isLeaveModalOpen = true)}
		/>
	{/if}
</div>

<!-- Modals -->
<TransferOwnershipModal bind:open={isTransferModalOpen} {orgId} {eligibleMembers} />

<DisbandOrgModal bind:open={isDisbandModalOpen} {orgId} {orgName} />

<ConfirmDialog
	bind:open={isLeaveModalOpen}
	title="Leave Organization"
	confirmText="Leave Organization"
	destructive={true}
	onConfirm={leaveOrganization}
>
	<div class="space-y-4">
		<p class="text-muted-foreground">
			Are you sure you want to leave this organization? You will no longer have access to any
			monitors or status pages, and this action cannot be undone unless you are invited back.
		</p>
	</div>
</ConfirmDialog>
