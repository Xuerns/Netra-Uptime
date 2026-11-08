<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Users, TriangleAlert } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import InviteMemberModal from '$lib/components/InviteMemberModal.svelte';
	import MemberListItem, { type Member } from '$lib/components/member/MemberListItem.svelte';
	import MemberRoleFilter from '$lib/components/member/MemberRoleFilter.svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';
	import {
		getCurrentUserId,
		findMemberByUserId,
		getMemberRole,
		canManageOrg
	} from '$lib/utils/role';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const orgId = $derived(($page.params.id as string) || '');
	const currentUserId = $derived(getCurrentUserId(data?.user ?? $page.data?.user));

	let members = $state<Member[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let roleFilter = $state<string>('all');

	// Modals
	let isInviteOpen = $state(false);
	let isKickConfirmOpen = $state(false);
	let memberToKick = $state<Member | null>(null);

	let owners = $derived(members.filter((m) => m.Role === 'owner'));
	let admins = $derived(members.filter((m) => m.Role === 'admin'));
	let regularMembers = $derived(members.filter((m) => m.Role === 'member'));

	let roleCounts = $derived({
		total: members.length,
		owners: owners.length,
		admins: admins.length,
		members: regularMembers.length
	});

	const currentMember = $derived(findMemberByUserId(members, currentUserId));
	const currentUserRole = $derived(getMemberRole(currentMember, 'member'));
	const isOwner = $derived(currentUserRole === 'owner');
	const canManage = $derived(canManageOrg(currentUserRole));

	// Sync with layout header bar
	$effect(() => {
		void $page.url.pathname;
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search members by name...',
			actionType: canManage ? 'member' : null,
			onAction: () => {
				isInviteOpen = true;
			},
			badge: {
				text: `${members.length} Member${members.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	let filteredMembers = $derived(
		members.filter((m) => {
			if (roleFilter !== 'all' && m.Role !== roleFilter) {
				return false;
			}
			const q = headerStore.searchQuery.toLowerCase().trim();
			if (!q) return true;
			const name = (m.User?.Name || '').toLowerCase();
			return name.includes(q);
		})
	);

	async function loadMembers(silent = false) {
		if (!orgId) return;
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const res = await OrgAPI.getMembers(orgId);
			members = res.data ?? [];
		} catch (err: unknown) {
			fetchError = parseApiError(err, 'Failed to load organization members.');
		} finally {
			isLoading = false;
		}
	}

	async function handlePromote(userId: string) {
		if (!orgId) return;
		try {
			await OrgAPI.promoteToAdmin(orgId, userId);
			toast.success('Member promoted to admin');
			await loadMembers(true);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to promote member'));
		}
	}

	async function handleDemote(userId: string) {
		if (!orgId) return;
		try {
			await OrgAPI.demoteToMember(orgId, userId);
			toast.success('Admin demoted to member');
			await loadMembers(true);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to demote admin'));
		}
	}

	function promptKick(m: Member) {
		memberToKick = m;
		isKickConfirmOpen = true;
	}

	async function confirmKick() {
		if (!memberToKick || !orgId) return;
		try {
			await OrgAPI.kickMember(orgId, memberToKick.UserID);
			toast.success(`${memberToKick.User?.Name || 'Member'} removed from organization`);
			isKickConfirmOpen = false;
			memberToKick = null;
			await loadMembers(true);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to remove member'));
		}
	}

	onMount(() => {
		loadMembers();
	});
</script>

<svelte:head>
	<title>Members | Netra Uptime</title>
</svelte:head>

<div class="members-page text-foreground antialiased space-y-6">
	<!-- ── Subheader & Filter Tabs ─────────────────────────────────────────── -->
	<MemberRoleFilter
		bind:value={roleFilter}
		counts={roleCounts}
		onChange={(role) => (roleFilter = role)}
	/>

	<!-- ── Content Area ──────────────────────────────────────────────────────── -->
	{#if isLoading}
		<DataSkeleton variant="list" count={5} />
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-20 border border-border rounded-lg bg-card"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium rounded-lg"
				onclick={() => loadMembers()}
			>
				Try again
			</Button>
		</div>
	{:else if filteredMembers.length === 0}
		<div
			class="flex flex-col items-center justify-center gap-3 py-20 border border-border rounded-lg bg-card text-center px-4"
		>
			<Users class="size-10 text-muted-foreground/40" />
			<h3 class="text-sm font-semibold text-foreground">No members found</h3>
			<p class="text-xs text-muted-foreground max-w-2xl">
				{headerStore.searchQuery
					? `No members matching "${headerStore.searchQuery}". Try a different search.`
					: 'No members under the selected filter.'}
			</p>
		</div>
	{:else}
		<!-- Members List -->
		<div class="border border-border rounded-md bg-card divide-y divide-border">
			{#each filteredMembers as m (m.ID)}
				<MemberListItem
					member={m}
					{currentUserId}
					{isOwner}
					{canManage}
					onPromote={handlePromote}
					onDemote={handleDemote}
					onKick={promptKick}
				/>
			{/each}
		</div>
	{/if}
</div>

<!-- ── Invite Member Modal ────────────────────────────────────────────────── -->
<InviteMemberModal bind:open={isInviteOpen} {orgId} {isOwner} onInvited={() => loadMembers(true)} />

<!-- ── Confirm Kick Dialog ────────────────────────────────────────────────── -->
<ConfirmDialog
	bind:open={isKickConfirmOpen}
	title="Remove member"
	description="Are you sure you want to remove {memberToKick?.User?.Name ||
		'this member'} from the organization? They will lose access to all monitors and resources."
	confirmText="Remove member"
	destructive={true}
	onConfirm={confirmKick}
/>
