<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		Users,
		UserPlus,
		MoreVertical,
		ShieldCheck,
		Shield,
		Trash2,
		ShieldAlert,
		ShieldX,
		Search,
		TriangleAlert
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import InviteMemberModal from '$lib/components/InviteMemberModal.svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';

	interface MemberUser {
		ID: string;
		Name: string;
		AvatarURL?: string;
	}

	interface Member {
		ID: string;
		OrganizationID: string;
		UserID: string;
		Role: string;
		User: MemberUser;
	}

	type RoleFilter = 'all' | 'owner' | 'admin' | 'member';

	const orgId = $derived(($page.params.id as string) || '');
	const currentUserId = $derived($page.data?.user?.id || $page.data?.user?.ID || '');

	let members = $state<Member[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let roleFilter = $state<RoleFilter>('all');

	// Modals
	let isInviteOpen = $state(false);
	let isKickConfirmOpen = $state(false);
	let memberToKick = $state<Member | null>(null);
	let isKicking = $state(false);

	let owners = $derived(members.filter((m) => m.Role === 'owner'));
	let admins = $derived(members.filter((m) => m.Role === 'admin'));
	let regularMembers = $derived(members.filter((m) => m.Role === 'member'));

	const currentMember = $derived(members.find((m) => m.UserID === currentUserId));
	const currentUserRole = $derived(currentMember?.Role || 'member');
	const isOwner = $derived(currentUserRole === 'owner');
	const isAdmin = $derived(currentUserRole === 'admin');
	const canManage = $derived(isOwner || isAdmin);

	// Sync with layout header bar
	$effect(() => {
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

	function initials(name: string): string {
		const parts = (name || '').trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) return 'U';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	function avatarColor(name: string): string {
		const colors = [
			'bg-emerald-500/20 text-emerald-400',
			'bg-blue-500/20 text-blue-400',
			'bg-violet-500/20 text-violet-400',
			'bg-amber-500/20 text-amber-400',
			'bg-rose-500/20 text-rose-400',
			'bg-cyan-500/20 text-cyan-400'
		];
		let hash = 0;
		for (let i = 0; i < (name || '').length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}

	async function loadMembers(silent = false) {
		if (!orgId) return;
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const res = await OrgAPI.getMembers(orgId);
			members = res.data ?? [];
		} catch (err: any) {
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
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to promote member'));
		}
	}

	async function handleDemote(userId: string) {
		if (!orgId) return;
		try {
			await OrgAPI.demoteToMember(orgId, userId);
			toast.success('Admin demoted to member');
			await loadMembers(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to demote admin'));
		}
	}

	function promptKick(m: Member) {
		memberToKick = m;
		isKickConfirmOpen = true;
	}

	async function confirmKick() {
		if (!memberToKick || !orgId) return;
		isKicking = true;
		try {
			await OrgAPI.kickMember(orgId, memberToKick.UserID);
			toast.success(`${memberToKick.User?.Name || 'Member'} removed from organization`);
			isKickConfirmOpen = false;
			memberToKick = null;
			await loadMembers(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to remove member'));
		} finally {
			isKicking = false;
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
	<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-2">
		<div
			class="flex items-center gap-1.5 p-1 bg-card border border-border rounded-lg text-xs font-medium"
		>
			<button
				type="button"
				class="px-3 py-1.5 rounded-md transition-colors {roleFilter === 'all'
					? 'bg-secondary text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => (roleFilter = 'all')}
			>
				All <span class="ml-1 text-[11px] font-mono text-muted-foreground">({members.length})</span>
			</button>
			<button
				type="button"
				class="px-3 py-1.5 rounded-md transition-colors {roleFilter === 'owner'
					? 'bg-secondary text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => (roleFilter = 'owner')}
			>
				Owners <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({owners.length})</span
				>
			</button>
			<button
				type="button"
				class="px-3 py-1.5 rounded-md transition-colors {roleFilter === 'admin'
					? 'bg-secondary text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => (roleFilter = 'admin')}
			>
				Admins <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({admins.length})</span
				>
			</button>
			<button
				type="button"
				class="px-3 py-1.5 rounded-md transition-colors {roleFilter === 'member'
					? 'bg-secondary text-foreground font-semibold'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => (roleFilter = 'member')}
			>
				Members <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({regularMembers.length})</span
				>
			</button>
		</div>
	</div>

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
		<div class="border border-border rounded-lg bg-card divide-y divide-border">
			{#each filteredMembers as m (m.ID)}
				<div class="flex items-center justify-between p-4 hover:bg-muted/20 transition-colors">
					<!-- User Info -->
					<div class="flex items-center gap-3.5 min-w-0">
						{#if m.User?.AvatarURL}
							<img
								src={m.User.AvatarURL}
								alt={m.User?.Name || 'User'}
								class="size-9 rounded-lg object-cover border border-border shrink-0"
							/>
						{:else}
							<div
								class="size-9 rounded-lg flex items-center justify-center font-bold text-xs shrink-0 border border-border {avatarColor(
									m.User?.Name || 'User'
								)}"
							>
								{initials(m.User?.Name || 'User')}
							</div>
						{/if}

						<div class="min-w-0">
							<div class="flex items-center gap-2">
								<span class="text-sm font-semibold text-foreground truncate">
									{m.User?.Name || 'Unknown User'}
								</span>
								{#if m.UserID === currentUserId}
									<span
										class="px-1.5 py-0.5 rounded text-[10px] bg-secondary text-muted-foreground font-mono"
									>
										You
									</span>
								{/if}
							</div>
							<div class="flex items-center gap-1.5 mt-0.5">
								<span class="text-xs text-muted-foreground capitalize">{m.Role}</span>
							</div>
						</div>
					</div>

					<!-- Role Pill & Actions -->
					<div class="flex items-center gap-3 shrink-0">
						<!-- Role Indicator -->
						{#if m.Role === 'owner'}
							<div
								class="flex items-center gap-1 text-amber-400 text-xs font-medium px-2.5 py-1 rounded-md bg-amber-500/10 border border-amber-500/20"
							>
								<ShieldCheck class="size-3.5" />
								<span>Owner</span>
							</div>
						{:else if m.Role === 'admin'}
							<div
								class="flex items-center gap-1 text-emerald-400 text-xs font-medium px-2.5 py-1 rounded-md bg-emerald-500/10 border border-emerald-500/20"
							>
								<Shield class="size-3.5" />
								<span>Admin</span>
							</div>
						{:else}
							<div
								class="flex items-center gap-1 text-muted-foreground text-xs font-medium px-2.5 py-1 rounded-md bg-secondary border border-border"
							>
								<span>Member</span>
							</div>
						{/if}

						<!-- Member Action Menu -->
						{#if canManage && m.UserID !== currentUserId && m.Role !== 'owner'}
							<DropdownMenu.Root>
								<DropdownMenu.Trigger
									class="p-1.5 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary transition-colors focus:outline-none cursor-pointer"
								>
									<MoreVertical class="size-4" />
								</DropdownMenu.Trigger>
								<DropdownMenu.Content
									align="end"
									class="bg-card border-border text-foreground min-w-[150px] p-1 rounded-lg"
								>
									{#if m.Role === 'member'}
										<DropdownMenu.Item
											class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
											onclick={() => handlePromote(m.UserID)}
										>
											<ShieldAlert class="mr-2 size-3.5 text-emerald-400" /> Make Admin
										</DropdownMenu.Item>
									{:else if m.Role === 'admin' && isOwner}
										<DropdownMenu.Item
											class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
											onclick={() => handleDemote(m.UserID)}
										>
											<ShieldX class="mr-2 size-3.5 text-amber-400" /> Demote to Member
										</DropdownMenu.Item>
									{/if}

									<DropdownMenu.Item
										class="hover:bg-destructive/10 text-destructive cursor-pointer rounded-md p-2 text-xs"
										onclick={() => promptKick(m)}
									>
										<Trash2 class="mr-2 size-3.5" /> Remove from Org
									</DropdownMenu.Item>
								</DropdownMenu.Content>
							</DropdownMenu.Root>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- ── Invite Member Modal ────────────────────────────────────────────────── -->
<InviteMemberModal bind:open={isInviteOpen} {orgId} onInvited={() => loadMembers(true)} />

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
