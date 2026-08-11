<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { flip } from 'svelte/animate';
	import {
		Eye,
		Bell,
		Edit2,
		Search,
		X,
		Check,
		MoreVertical,
		Trash2,
		ArrowUpRight,
		Plus,
		Activity,
		AlertCircle,
		LayoutGrid,
		Globe,
		Terminal,
		TriangleAlert,
		UserPlus,
		Pencil,
		Settings,
		CreditCard
	} from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import { OrgAPI } from '$lib/api/organization';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import { debounce } from '$lib/helper/utils';
	import MembersSidebar from './MembersSidebar.svelte';
	import AddMonitorModal from './AddMonitorModal.svelte';
	import EditMonitorModal from './EditMonitorModal.svelte';
	import MonitorCard from './MonitorCard.svelte';
	import StatusPageTab from './StatusPageTab.svelte';
	import APIKeysTab from './APIKeysTab.svelte';
	import { wsClient, type MonitorUpdatePayload } from '$lib/utils/ws';
	import { toast } from 'svelte-sonner';
	// ── Types ─────────────────────────────────────────────────────────────────
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

	// ── Route param ───────────────────────────────────────────────────────────
	const orgId = $page.params.id as string;

	// ── State ─────────────────────────────────────────────────────────────────
	let members = $state<Member[]>([]);
	let monitors = $state<Monitor[]>([]);
	let isLoadingMembers = $state(true);
	let isLoadingMonitors = $state(true);
	let membersError = $state('');
	let monitorsError = $state('');

	let activeTab = $state('monitors');

	let panelOpen = $state(true);
	let openDropdownId = $state<string | null>(null);

	// ── Invite dialog state ────────────────────────────────────────────────────
	interface InviteUser {
		id: string;
		name: string;
		avatar_url?: string;
	}
	let inviteOpen = $state(false);
	let inviteQuery = $state('');
	let inviteResults = $state<InviteUser[]>([]);
	let inviteList = $state<InviteUser[]>([]);
	let isSearchingInvite = $state(false);
	let isSubmittingInvite = $state(false);
	let inviteError = $state('');
	let showInviteDropdown = $state(false);

	// ── Monitor State ────────────────────────────────────────────────────────
	let isAddMonitorOpen = $state(false);
	let selectedMonitorType = $state<'http' | 'ping' | 'tcp' | null>(null);

	// ── Status Page State ────────────────────────────────────────────────────
	let isAddStatusPageOpen = $state(false);

	let orgName = $state('Monitors');

	async function loadOrgDetails() {
		try {
			const res = await OrgAPI.getAll();
			const org = res.data?.find((o: any) => o.ID === orgId);
			if (org) {
				orgName = org.Name;
			}
		} catch (e) {
			console.error('Failed to load org details', e);
		}
	}

	function openAddMonitor(type: 'http' | 'ping' | 'tcp') {
		selectedMonitorType = type;
		isAddMonitorOpen = true;
	}

	let isEditMonitorOpen = $state(false);
	let selectedMonitorToEdit = $state<Monitor | null>(null);

	function openEditMonitor(monitor: Monitor) {
		selectedMonitorToEdit = monitor;
		isEditMonitorOpen = true;
	}

	// ── Derived ──────────────────────────────────────────────────────────────
	let currentUserRole = $derived(members.find((m: Member) => m.UserID === $page.data.user?.id)?.Role);
	let upCount = $derived(monitors.filter((m) => m.Status === 'up').length);
	let downCount = $derived(monitors.filter((m) => m.Status === 'down').length);

	// ── Helpers ──────────────────────────────────────────────────────────────
	function initials(name: string): string {
		const parts = name.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	function avatarColor(name: string): string {
		const colors = [
			'bg-emerald-500/20 text-emerald-400',
			'bg-blue-500/20 text-blue-400',
			'bg-violet-500/20 text-violet-400',
			'bg-orange-500/20 text-orange-400',
			'bg-rose-500/20 text-rose-400',
			'bg-cyan-500/20 text-cyan-400'
		];
		let hash = 0;
		for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
		return colors[Math.abs(hash) % colors.length];
	}

	function typeBadgeClass(type: string): string {
		switch (type.toLowerCase()) {
			case 'http':
				return 'bg-blue-500/14 text-blue-300';
			case 'tcp':
				return 'bg-violet-500/14 text-violet-300';
			case 'ping':
				return 'bg-white/8 text-on-surface-variant';
			default:
				return 'bg-white/8 text-on-surface-variant';
		}
	}

	function toggleDropdown(id: string) {
		openDropdownId = openDropdownId === id ? null : id;
	}

	function closeDropdowns() {
		openDropdownId = null;
	}

	// ── Invite helpers ────────────────────────────────────────────────────────
	let filteredInviteResults = $derived(
		inviteResults.filter((r) => !inviteList.some((u) => u.id === r.id))
	);

	async function searchInviteUsers(query: string) {
		if (!query.trim()) {
			inviteResults = [];
			isSearchingInvite = false;
			return;
		}
		try {
			const res = await OrgAPI.searchPersons(query, 1, orgId);
			inviteResults = (res.data ?? []).map((u: any) => ({
				id: u.id ?? u.ID,
				name: u.name ?? u.Name,
				avatar_url: u.avatar_url ?? u.AvatarURL
			}));
		} catch {
			inviteResults = [];
		} finally {
			isSearchingInvite = false;
		}
	}

	const debouncedSearchInvite = debounce(searchInviteUsers, 400);

	function handleInviteInput() {
		showInviteDropdown = true;
		if (!inviteQuery.trim()) {
			inviteResults = [];
			isSearchingInvite = false;
			return;
		}
		isSearchingInvite = true;
		debouncedSearchInvite(inviteQuery);
	}

	function addToInviteList(user: InviteUser) {
		if (inviteList.some((u) => u.id === user.id)) return;
		inviteList = [...inviteList, user];
		inviteQuery = '';
		inviteResults = [];
		showInviteDropdown = false;
	}

	function removeFromInviteList(id: string) {
		inviteList = inviteList.filter((u) => u.id !== id);
	}

	function openInviteDialog() {
		inviteQuery = '';
		inviteResults = [];
		inviteList = [];
		inviteError = '';
		showInviteDropdown = false;
		inviteOpen = true;
	}

	function closeInviteDialog() {
		inviteOpen = false;
	}

	async function handleSubmitInvite() {
		if (inviteList.length === 0) return;
		isSubmittingInvite = true;
		inviteError = '';
		try {
			for (const user of inviteList) {
				await OrgAPI.inviteMember(orgId, { invitee_id: user.id, role: 'member' });
			}
			toast.success('Invitations sent successfully');
			inviteOpen = false;
			await loadMembers();
		} catch (err: any) {
			inviteError = err.message ?? 'Failed to send invitations.';
			toast.error(inviteError);
		} finally {
			isSubmittingInvite = false;
		}
	}

	// ── Data Fetching ─────────────────────────────────────────────────────────
	async function loadMembers() {
		isLoadingMembers = true;
		membersError = '';
		try {
			const res = await OrgAPI.getMembers(orgId);
			members = res.data ?? [];
		} catch (err: any) {
			membersError = err.message ?? 'Failed to load members.';
		} finally {
			isLoadingMembers = false;
		}
	}

	async function loadMonitors() {
		isLoadingMonitors = true;
		monitorsError = '';
		try {
			const res = await MonitorAPI.getAll(orgId);
			monitors = res.data ?? [];
		} catch (err: any) {
			monitorsError = err.message ?? 'Failed to load monitors.';
		} finally {
			isLoadingMonitors = false;
		}
	}

	// ── Actions ───────────────────────────────────────────────────────────────
	async function handleKickMember(memberId: string) {
		closeDropdowns();
		try {
			await OrgAPI.kickMember(orgId, memberId);
			await loadMembers();
			toast.success('Member removed successfully');
		} catch (err: any) {
			console.error('Failed to kick member:', err);
			toast.error(err.message ?? 'Failed to remove member');
		}
	}

	async function handlePromoteMember(memberId: string) {
		closeDropdowns();
		try {
			await OrgAPI.promoteToAdmin(orgId, memberId);
			await loadMembers();
			toast.success('Member promoted to admin successfully');
		} catch (err: any) {
			console.error('Failed to promote member:', err);
			toast.error(err.message ?? 'Failed to promote member');
		}
	}

	async function handleDemoteMember(memberId: string) {
		closeDropdowns();
		try {
			await OrgAPI.demoteToMember(orgId, memberId);
			await loadMembers();
			toast.success('Admin demoted to member successfully');
		} catch (err: any) {
			console.error('Failed to demote member:', err);
			toast.error(err.message ?? 'Failed to demote member');
		}
	}

	async function handleDeleteMonitor(monitorId: string) {
		closeDropdowns();
		try {
			await MonitorAPI.delete(orgId, monitorId);
			await loadMonitors();
			toast.success('Monitor deleted successfully');
		} catch (err: any) {
			console.error('Failed to delete monitor:', err);
			toast.error(err.message ?? 'Failed to delete monitor');
		}
	}

	onMount(() => {
		loadOrgDetails();
		loadMembers();
		loadMonitors();

		// Listen to WebSocket updates
		const unsubscribe = wsClient.subscribe((msg) => {
			if (msg.Type === 'MONITOR_UPDATED') {
				const payload = msg.Payload as MonitorUpdatePayload;

				const index = monitors.findIndex((m) => m.ID === payload.id);
				if (index !== -1) {
					monitors[index].Status = payload.status;
					monitors[index].FailureCount = payload.failure_count;
					monitors[index].ResponseTimeMs = payload.response_time;
					monitors[index].LastCheckedAt = payload.last_checked;
				}
			}
		});

		return () => unsubscribe();
	});
</script>

<svelte:head>
	<title>Monitors · PulseWatch</title>
</svelte:head>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="flex items-stretch h-[calc(100vh-3rem)] -m-6 text-on-surface antialiased"
	onclick={closeDropdowns}
>
	<!-- ── Left Column (Header & Main Content) ────────────────────────────────── -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- ── Page Header ────────────────────────────────────────────────────── -->
		<header
			class="flex items-center justify-between h-[72px] px-6 border-b border-white/8 shrink-0"
		>
			<div>
				<div class="flex items-center gap-3">
					<h1 class="text-2xl font-semibold leading-tight tracking-tight text-on-surface">
						{orgName}
					</h1>
					<a
						href="/dashboard/organization/{orgId}/settings"
						class="text-on-surface-variant hover:text-on-surface transition-colors p-1 mt-1"
						title="Organization Settings"
					>
					</a>
				</div>
				<p class="font-mono text-[13px] text-on-surface-variant mt-1">
					{#if isLoadingMonitors}
						Loading monitors...
					{:else}
						{monitors.length} monitor{monitors.length !== 1 ? 's' : ''}
						{#if monitors.length > 0}
							· <span class="text-primary">{upCount} up</span>
							{#if downCount > 0}
								· <span class="text-error">{downCount} down</span>
							{/if}
						{/if}
					{/if}
				</p>
			</div>

			<div class="flex items-center gap-2">


				<!-- Panel Toggle -->
				<button
					class="flex items-center justify-center size-9 rounded-lg border transition-colors
				{panelOpen
						? 'bg-primary/12 border-primary/40 text-primary'
						: 'bg-white/2 border-white/10 text-on-surface-variant hover:bg-white/7 hover:text-on-surface'}"
					onclick={(e) => {
						e.stopPropagation();
						panelOpen = !panelOpen;
					}}
					aria-label="Toggle members panel"
					aria-pressed={panelOpen}
				>
					<LayoutGrid class="size-[18px]" />
				</button>

				<!-- Add Monitor -->
				{#if activeTab === 'monitors' && currentUserRole === 'owner'}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 font-medium text-sm shadow-sm"
							onclick={(e) => e.stopPropagation()}
						>
							<Plus class="size-4" />
							<span>Add monitor</span>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							align="end"
							class="bg-[#161618] border-white/8 text-white min-w-[160px] p-1.5"
						>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2"
								onclick={() => openAddMonitor('http')}
							>
								<Globe class="mr-2 size-4 text-primary" /> HTTP/API
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2"
								onclick={() => openAddMonitor('ping')}
							>
								<Activity class="mr-2 size-4 text-[#ffc107]" /> Ping
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2"
								onclick={() => openAddMonitor('tcp')}
							>
								<Terminal class="mr-2 size-4 text-[#17a2b8]" /> TCP Port
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}

				{#if activeTab === 'statuspages' && currentUserRole === 'owner'}
					<button
						class="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 font-medium text-sm shadow-sm"
						onclick={(e) => {
							e.stopPropagation();
							isAddStatusPageOpen = true;
						}}
					>
						<Plus class="size-4" />
						<span>Create status page</span>
					</button>
				{/if}

				{#if currentUserRole === 'owner'}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="flex items-center justify-center size-9 rounded-lg border bg-white/2 border-white/10 text-on-surface-variant hover:bg-white/7 hover:text-on-surface transition-colors"
							onclick={(e) => e.stopPropagation()}
						>
							<MoreVertical class="size-[18px]" />
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							align="end"
							class="bg-[#161618] border-white/8 text-white min-w-[200px] p-1.5"
						>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 flex items-center"
								onclick={(e) => {
									e.stopPropagation();
									openInviteDialog();
								}}
							>
								<UserPlus class="mr-2 size-4 text-primary" /> Invite Member
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 flex items-center"
								onclick={() => goto(`/dashboard/organization/${orgId}/subscription`)}
							>
								<CreditCard class="mr-2 size-4 text-primary" /> Billing & Subscription
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 flex items-center"
								onclick={() => goto(`/dashboard/organization/${orgId}/settings`)}
							>
								<Settings class="mr-2 size-4 text-primary" /> Organization Settings
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}
			</div>
		</header>

		<!-- ── Main Content ──────────────────────────────────────────────────────── -->
		<Tabs.Root bind:value={activeTab} class="flex-1 flex flex-col min-h-0 overflow-hidden">
			<div class="px-6 shrink-0 relative pt-1">
				<!-- The border line that sits behind the tabs -->
				<div class="absolute bottom-0 left-0 right-0 h-px bg-white/8"></div>
				
				<Tabs.List class="bg-transparent border-none p-0 h-auto flex gap-1 relative z-10">
					<Tabs.Trigger
						value="monitors"
						class="relative rounded-t-lg rounded-b-none px-5 py-2.5 text-[13px] font-medium transition-all duration-200 shadow-none
						data-[state=active]:bg-white/6 data-[state=active]:text-white data-[state=active]:border data-[state=active]:border-white/8 data-[state=active]:border-b-transparent data-[state=active]:-mb-px
						data-[state=inactive]:bg-transparent data-[state=inactive]:text-on-surface-variant data-[state=inactive]:border data-[state=inactive]:border-transparent data-[state=inactive]:hover:text-white/80 data-[state=inactive]:hover:bg-white/3"
						>Monitors</Tabs.Trigger
					>
					<Tabs.Trigger
						value="statuspages"
						class="relative rounded-t-lg rounded-b-none px-5 py-2.5 text-[13px] font-medium transition-all duration-200 shadow-none
						data-[state=active]:bg-white/6 data-[state=active]:text-white data-[state=active]:border data-[state=active]:border-white/8 data-[state=active]:border-b-transparent data-[state=active]:-mb-px
						data-[state=inactive]:bg-transparent data-[state=inactive]:text-on-surface-variant data-[state=inactive]:border data-[state=inactive]:border-transparent data-[state=inactive]:hover:text-white/80 data-[state=inactive]:hover:bg-white/3"
						>Status Pages</Tabs.Trigger
					>
					<Tabs.Trigger
						value="apikeys"
						class="relative rounded-t-lg rounded-b-none px-5 py-2.5 text-[13px] font-medium transition-all duration-200 shadow-none
						data-[state=active]:bg-white/6 data-[state=active]:text-white data-[state=active]:border data-[state=active]:border-white/8 data-[state=active]:border-b-transparent data-[state=active]:-mb-px
						data-[state=inactive]:bg-transparent data-[state=inactive]:text-on-surface-variant data-[state=inactive]:border data-[state=inactive]:border-transparent data-[state=inactive]:hover:text-white/80 data-[state=inactive]:hover:bg-white/3"
						>API Keys</Tabs.Trigger
					>
				</Tabs.List>
			</div>

			<Tabs.Content
				value="monitors"
				class="flex-1 p-6 overflow-y-auto m-0 outline-none data-[state=inactive]:hidden flex flex-col"
			>
				<!-- Loading monitors -->
				{#if isLoadingMonitors}
					<div class="grid grid-cols-[repeat(auto-fill,minmax(270px,1fr))] gap-6">
						{#each Array(3) as _}
							<div
								class="bg-white/3 border border-white/8 rounded-lg p-5 flex flex-col gap-3.5 animate-pulse"
							>
								<div class="flex justify-between">
									<div class="flex flex-col gap-1.5 flex-1">
										<div class="h-3.5 bg-white/10 rounded w-2/3"></div>
										<div class="h-2.5 bg-white/6 rounded w-full font-mono"></div>
									</div>
									<div class="h-5 w-10 bg-white/8 rounded-full ml-2"></div>
								</div>
								<div class="h-2 bg-white/8 rounded w-1/3"></div>
								<div class="h-9 bg-white/5 rounded w-full"></div>
								<div class="flex justify-between pt-3 border-t border-white/8">
									<div class="h-4 w-14 bg-white/8 rounded"></div>
									<div class="h-4 w-14 bg-white/8 rounded"></div>
								</div>
							</div>
						{/each}
					</div>

					<!-- Monitors error -->
				{:else if monitorsError}
					<div
						class="flex flex-col items-center justify-center gap-3 py-24 border border-white/8 rounded-lg bg-white/3"
					>
						<TriangleAlert class="size-8 text-error" />
						<p class="text-sm text-on-surface-variant">{monitorsError}</p>
						<button
							class="h-9 px-4 text-sm bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 rounded-lg hover:text-white hover:bg-[#2A2A2B] active:scale-[0.98] transition-all duration-200 shadow-sm"
							onclick={loadMonitors}
						>
							Try again
						</button>
					</div>

					<!-- Empty state -->
				{:else if monitors.length === 0}
					<div
						class="flex flex-col items-center justify-center text-center py-24 px-6 gap-3.5 bg-white/3 border border-white/8 rounded-lg backdrop-blur-md"
					>
						<div
							class="flex items-center justify-center size-[60px] rounded-lg bg-primary/8 text-primary"
						>
							<Activity class="size-6" />
						</div>
						<h2 class="text-[19px] font-semibold text-on-surface m-0">No monitors yet</h2>
						<p class="text-sm leading-relaxed text-on-surface-variant max-w-[360px] m-0">
							{#if currentUserRole === 'owner'}
								Add your first monitor to start tracking uptime for websites, APIs, or servers.
							{:else}
								There are no monitors in this organization yet.
							{/if}
						</p>
						{#if currentUserRole === 'owner'}
							<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="flex items-center gap-2 h-10 px-5 rounded-lg bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 font-medium text-[14px] shadow-sm"
							>
								<Plus class="size-4" />
								<span>Add monitor</span>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content
								align="center"
								class="bg-[#161618] border-white/8 text-white min-w-[160px] p-1.5"
							>
								<DropdownMenu.Item
									class="hover:bg-white/10 cursor-pointer rounded-md p-2"
									onclick={() => openAddMonitor('http')}
								>
									<Globe class="mr-2 size-4 text-primary" /> HTTP/API
								</DropdownMenu.Item>
								<DropdownMenu.Item
									class="hover:bg-white/10 cursor-pointer rounded-md p-2"
									onclick={() => openAddMonitor('ping')}
								>
									<Activity class="mr-2 size-4 text-[#ffc107]" /> Ping
								</DropdownMenu.Item>
								<DropdownMenu.Item
									class="hover:bg-white/10 cursor-pointer rounded-md p-2"
									onclick={() => openAddMonitor('tcp')}
								>
									<Terminal class="mr-2 size-4 text-[#17a2b8]" /> TCP Port
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
						{/if}
					</div>

					<!-- Monitor Grid -->
				{:else}
					<div class="grid grid-cols-[repeat(auto-fill,minmax(450px,1fr))] gap-6">
						{#each monitors as monitor (monitor.ID)}
							<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
								<MonitorCard
									{monitor}
									{typeBadgeClass}
									{openDropdownId}
									{toggleDropdown}
									{handleDeleteMonitor}
									handleEditMonitor={openEditMonitor}
								/>
							</div>
						{/each}
					</div>
				{/if}
			</Tabs.Content>

			<Tabs.Content
				value="statuspages"
				class="flex-1 p-6 overflow-y-auto m-0 outline-none data-[state=inactive]:hidden flex flex-col"
			>
				<StatusPageTab {orgId} {monitors} {currentUserRole} bind:isAddModalOpen={isAddStatusPageOpen} />
			</Tabs.Content>

			<Tabs.Content
				value="apikeys"
				class="flex-1 p-6 overflow-y-auto m-0 outline-none data-[state=inactive]:hidden flex flex-col"
			>
				<APIKeysTab {orgId} userRole={currentUserRole || 'member'} />
			</Tabs.Content>
		</Tabs.Root>
	</div>

	<!-- ── Right Sidebar (Members Panel) ──────────────────────────────────────── -->
	<MembersSidebar
		{panelOpen}
		{isLoadingMembers}
		{membersError}
		{members}
		currentUserId={$page.data.user?.id}
		{openDropdownId}
		{toggleDropdown}
		{handleKickMember}
		{handlePromoteMember}
		{handleDemoteMember}
		{avatarColor}
		{initials}
	/>
</div>

<AddMonitorModal
	bind:isOpen={isAddMonitorOpen}
	monitorType={selectedMonitorType}
	{orgId}
	onSuccess={loadMonitors}
/>

<EditMonitorModal
	bind:isOpen={isEditMonitorOpen}
	monitor={selectedMonitorToEdit}
	{orgId}
	onSuccess={loadMonitors}
/>

<!-- ── Invite Member Modal ──────────────────────────────────────────────── -->
{#if inviteOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/55 backdrop-blur-sm"
		onclick={(e) => {
			if (e.target === e.currentTarget) closeInviteDialog();
		}}
	>
		<div
			class="w-full max-w-[420px] bg-[#1c1b1c] border border-white/8 rounded-lg"
			onclick={(e) => e.stopPropagation()}
		>
			<!-- Modal Header -->
			<div class="flex items-center justify-between px-5 pt-5 pb-4">
				<div>
					<h3 class="text-[16px] font-semibold text-on-surface m-0">Invite member</h3>
					<p class="text-[13px] text-on-surface-variant mt-0.5 m-0">
						Search and invite people to this organization.
					</p>
				</div>
				<button
					class="flex items-center justify-center size-7 rounded-lg text-on-surface-variant hover:bg-white/8 hover:text-on-surface transition-colors shrink-0"
					onclick={closeInviteDialog}
					aria-label="Close"
				>
					<X class="size-4" />
				</button>
			</div>

			<div class="h-px bg-white/8"></div>

			<!-- Modal Body -->
			<div class="px-5 py-4 flex flex-col gap-4">
				<!-- Search input -->
				<div class="flex flex-col gap-1.5">
					<label class="text-[12px] font-medium text-on-surface-variant" for="inviteSearch">
						Search by name
					</label>
					<div class="relative">
						<div
							class="flex items-center h-10 rounded-lg border border-white/10 bg-black/25 transition-colors focus-within:border-[#3b82f6] focus-within:ring-1 focus-within:ring-[#3b82f6]/25 gap-2 px-3"
						>
							{#if isSearchingInvite}
								<div
									class="size-4 rounded-full border-2 border-white/20 border-t-primary animate-spin shrink-0"
								></div>
							{:else}
								<Search class="size-4 text-on-surface-variant/50 shrink-0" />
							{/if}
							<input
								id="inviteSearch"
								type="text"
								placeholder="Search by name..."
								autocomplete="off"
								bind:value={inviteQuery}
								oninput={handleInviteInput}
								onfocus={() => {
									if (inviteQuery.trim()) showInviteDropdown = true;
								}}
								onblur={() =>
									setTimeout(() => {
										showInviteDropdown = false;
									}, 150)}
								class="flex-1 h-full bg-transparent border-none outline-none text-on-surface text-[13px] placeholder:text-on-surface-variant/35 min-w-0"
							/>
						</div>

						<!-- Results dropdown -->
						{#if showInviteDropdown && inviteQuery.trim()}
							<div
								class="absolute top-[calc(100%+4px)] left-0 right-0 rounded-lg border border-white/10 bg-[#1c1b1c] z-10 overflow-hidden"
							>
								{#if isSearchingInvite}
									<div class="px-4 py-3 text-[13px] text-on-surface-variant/60">Searching...</div>
								{:else if filteredInviteResults.length > 0}
									<ul class="max-h-52 overflow-y-auto">
										{#each filteredInviteResults as user (user.id)}
											<li>
												<button
													type="button"
													class="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-white/5 transition-colors text-left"
													onmousedown={() => addToInviteList(user)}
												>
													{#if user.avatar_url}
														<img
															src={user.avatar_url}
															alt={user.name}
															class="size-7 rounded-lg object-cover shrink-0"
														/>
													{:else}
														<div
															class="size-7 rounded-lg flex items-center justify-center font-semibold text-[11px] shrink-0 {avatarColor(
																user.name
															)}"
														>
															{initials(user.name)}
														</div>
													{/if}
													<span class="text-[13px] text-on-surface font-medium truncate">
														{user.name}
													</span>
													<Plus class="size-3.5 text-on-surface-variant ml-auto shrink-0" />
												</button>
											</li>
										{/each}
									</ul>
								{:else}
									<div class="px-4 py-4 flex flex-col items-center gap-1.5 text-center">
										<UserPlus class="size-5 text-on-surface-variant/40" />
										<p class="text-[13px] text-on-surface-variant/60 m-0">
											No users found for "{inviteQuery}"
										</p>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<!-- Invite list -->
				{#if inviteList.length > 0}
					<ul class="flex flex-col gap-1.5">
						{#each inviteList as user (user.id)}
							<li
								class="flex items-center gap-3 px-3 py-2 rounded-lg bg-white/4 border border-white/8"
							>
								{#if user.avatar_url}
									<img
										src={user.avatar_url}
										alt={user.name}
										class="size-7 rounded-lg object-cover shrink-0"
									/>
								{:else}
									<div
										class="size-7 rounded-lg flex items-center justify-center font-semibold text-[11px] shrink-0 {avatarColor(
											user.name
										)}"
									>
										{initials(user.name)}
									</div>
								{/if}
								<span class="text-[13px] text-on-surface font-medium flex-1 truncate">
									{user.name}
								</span>
								<span
									class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-white/7 text-on-surface-variant shrink-0"
								>
									Member
								</span>
								<button
									type="button"
									class="size-6 flex items-center justify-center rounded text-on-surface-variant/40 hover:text-on-surface hover:bg-white/8 transition-colors shrink-0"
									onclick={() => removeFromInviteList(user.id)}
									aria-label="Remove {user.name}"
								>
									<X class="size-3.5" />
								</button>
							</li>
						{/each}
					</ul>
				{/if}

				<!-- Error -->
				{#if inviteError}
					<div
						class="flex items-start gap-2 rounded-lg bg-error/8 border border-error/20 px-3 py-2.5"
					>
						<TriangleAlert class="size-3.5 shrink-0 text-error mt-0.5" />
						<p class="text-[13px] text-error leading-snug m-0">{inviteError}</p>
					</div>
				{/if}
			</div>

			<div class="h-px bg-white/8"></div>

			<!-- Modal Footer -->
			<div class="px-5 py-3.5 flex items-center justify-between">
				<p class="text-[12px] text-on-surface-variant/50 m-0">
					{#if inviteList.length > 0}
						{inviteList.length} person{inviteList.length > 1 ? 's' : ''} will be invited
					{:else}
						Select people to invite
					{/if}
				</p>
				<div class="flex items-center gap-2">
					<button
						class="h-8 px-3.5 text-[13px] text-white hover:text-white hover:bg-white/10 rounded-lg transition-colors"
						onclick={closeInviteDialog}
					>
						Cancel
					</button>
					<button
						class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-8 px-3.5 text-[13px] font-medium rounded-lg shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
						disabled={inviteList.length === 0 || isSubmittingInvite}
						onclick={handleSubmitInvite}
					>
						{isSubmittingInvite ? 'Sending...' : 'Send invitations'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes pulseGlow {
		0%,
		100% {
			box-shadow: 0 0 0 4px rgba(78, 222, 163, 0.15);
		}
		50% {
			box-shadow: 0 0 0 7px rgba(78, 222, 163, 0.05);
		}
	}
</style>
