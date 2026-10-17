<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import {
		ScrollText,
		Search,
		ChevronDown,
		Activity,
		Users,
		Globe,
		Building2,
		Wrench,
		FolderTree,
		KeyRound,
		Bell,
		Send,
		AlertTriangle,
		CheckCircle2,
		AlertCircle,
		ListFilter,
		RotateCcw,
		X
	} from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { AuditAPI, type AuditLogItem } from '$lib/api/audit';
	import { debounce } from '$lib/helper/utils';
	import { EmptyState, Pagination, DataSkeleton } from '$lib/components/common';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatTimeAgo, formatDateTime } from '$lib/utils/datetime';

	const orgId = $derived($page.params.id);

	let isInitialLoading = $state(true);
	let isFetching = $state(false);
	let errorMsg = $state('');

	let logs = $state<AuditLogItem[]>([]);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalCount = $state(0);

	let searchQuery = $state('');
	let selectedCategory = $state<string>('ALL');
	let selectedAction = $state('');

	// ── Category Filters Definition ────────────────────────────────────────────
	const categories = [
		{ id: 'ALL', label: 'All Categories', icon: ScrollText, entityTypes: '' },
		{ id: 'MONITOR', label: 'Monitors', icon: Activity, entityTypes: 'monitor' },
		{ id: 'ORGANIZATION', label: 'Organization', icon: Building2, entityTypes: 'organization' },
		{
			id: 'MEMBER',
			label: 'Members & Invites',
			icon: Users,
			entityTypes: 'organization_member,organization_invitation'
		},
		{ id: 'STATUSPAGE', label: 'Status Pages', icon: Globe, entityTypes: 'status_page' },
		{
			id: 'MAINTENANCE',
			label: 'Maintenance',
			icon: Wrench,
			entityTypes: 'scheduled_maintenance'
		},
		{
			id: 'COMPONENT_GROUP',
			label: 'Component Groups',
			icon: FolderTree,
			entityTypes: 'component_group'
		},
		{ id: 'API_KEY', label: 'API Keys', icon: KeyRound, entityTypes: 'api_key' },
		{
			id: 'ALERT_TEMPLATE',
			label: 'Alert Templates & Rules',
			icon: Bell,
			entityTypes: 'alert_template,alert_rule'
		},
		{
			id: 'NOTIFICATION',
			label: 'Notifications',
			icon: Send,
			entityTypes: 'notification_channel'
		},
		{ id: 'INCIDENT', label: 'Incidents', icon: AlertTriangle, entityTypes: 'incident' }
	];

	// ── Action Filter Definitions (grouped by category) ────────────────────────
	const actionFilterGroups = [
		{
			categoryId: 'MONITOR',
			label: 'Monitors',
			icon: Activity,
			actions: [
				{ value: 'MONITOR_CREATE', label: 'Monitor Created' },
				{ value: 'MONITOR_UPDATE', label: 'Monitor Updated' },
				{ value: 'MONITOR_DELETE', label: 'Monitor Deleted' }
			]
		},
		{
			categoryId: 'ORGANIZATION',
			label: 'Organization',
			icon: Building2,
			actions: [
				{ value: 'ORG_UPDATE', label: 'Organization Updated' },
				{ value: 'ORG_DISBAND', label: 'Organization Disbanded' },
				{ value: 'ORG_TRANSFER_OWNERSHIP', label: 'Ownership Transferred' }
			]
		},
		{
			categoryId: 'MEMBER',
			label: 'Members & Invites',
			icon: Users,
			actions: [
				{ value: 'ORG_MEMBER_INVITE', label: 'Member Invited' },
				{ value: 'ORG_MEMBER_JOIN', label: 'Member Joined' },
				{ value: 'ORG_INVITE_REJECT', label: 'Invite Rejected' },
				{ value: 'ORG_MEMBER_PROMOTE', label: 'Member Promoted' },
				{ value: 'ORG_MEMBER_DEMOTE', label: 'Member Demoted' },
				{ value: 'ORG_MEMBER_KICK', label: 'Member Kicked' },
				{ value: 'ORG_MEMBER_LEAVE', label: 'Member Left' }
			]
		},
		{
			categoryId: 'STATUSPAGE',
			label: 'Status Pages',
			icon: Globe,
			actions: [
				{ value: 'STATUSPAGE_CREATE', label: 'Status Page Created' },
				{ value: 'STATUSPAGE_UPDATE', label: 'Status Page Updated' },
				{ value: 'STATUSPAGE_DELETE', label: 'Status Page Deleted' }
			]
		},
		{
			categoryId: 'MAINTENANCE',
			label: 'Scheduled Maintenance',
			icon: Wrench,
			actions: [
				{ value: 'MAINTENANCE_CREATE', label: 'Maintenance Created' },
				{ value: 'MAINTENANCE_UPDATE', label: 'Maintenance Updated' },
				{ value: 'MAINTENANCE_DELETE', label: 'Maintenance Deleted' }
			]
		},
		{
			categoryId: 'COMPONENT_GROUP',
			label: 'Component Groups',
			icon: FolderTree,
			actions: [
				{ value: 'COMPONENT_GROUP_CREATE', label: 'Component Group Created' },
				{ value: 'COMPONENT_GROUP_UPDATE', label: 'Component Group Updated' },
				{ value: 'COMPONENT_GROUP_DELETE', label: 'Component Group Deleted' }
			]
		},
		{
			categoryId: 'API_KEY',
			label: 'API Keys',
			icon: KeyRound,
			actions: [
				{ value: 'API_KEY_CREATE', label: 'API Key Created' },
				{ value: 'API_KEY_REVOKE', label: 'API Key Revoked' }
			]
		},
		{
			categoryId: 'ALERT_TEMPLATE',
			label: 'Alert Templates & Rules',
			icon: Bell,
			actions: [
				{ value: 'ALERT_TEMPLATE_CREATE', label: 'Template Created' },
				{ value: 'ALERT_TEMPLATE_UPDATE', label: 'Template Updated' },
				{ value: 'ALERT_TEMPLATE_DELETE', label: 'Template Deleted' },
				{ value: 'ALERT_RULE_CREATE', label: 'Alert Rule Created' },
				{ value: 'ALERT_RULE_UPDATE', label: 'Alert Rule Updated' },
				{ value: 'ALERT_RULE_DELETE', label: 'Alert Rule Deleted' }
			]
		},
		{
			categoryId: 'NOTIFICATION',
			label: 'Notification Channels',
			icon: Send,
			actions: [
				{ value: 'NOTIFICATION_CHANNEL_CREATE', label: 'Channel Created' },
				{ value: 'NOTIFICATION_CHANNEL_UPDATE', label: 'Channel Updated' },
				{ value: 'NOTIFICATION_CHANNEL_DELETE', label: 'Channel Deleted' }
			]
		},
		{
			categoryId: 'INCIDENT',
			label: 'Incidents',
			icon: AlertTriangle,
			actions: [
				{ value: 'INCIDENT_UPDATE', label: 'Incident Updated' },
				{ value: 'INCIDENT_ASSIGNEE_JOIN', label: 'Assignee Joined' }
			]
		}
	];

	let activeCategory = $derived(categories.find((c) => c.id === selectedCategory) || categories[0]);

	let selectedActionLabel = $derived.by(() => {
		if (!selectedAction) return '';
		for (const group of actionFilterGroups) {
			const found = group.actions.find((a) => a.value === selectedAction);
			if (found) return found.label;
		}
		return selectedAction;
	});

	let visibleActionGroups = $derived(
		selectedCategory === 'ALL'
			? actionFilterGroups
			: actionFilterGroups.filter((g) => g.categoryId === selectedCategory)
	);

	let hasActiveFilters = $derived(
		selectedCategory !== 'ALL' || selectedAction !== '' || searchQuery.trim() !== ''
	);

	function getActionStyle(action: string) {
		const act = action.toUpperCase();
		if (
			act.includes('CREATE') ||
			act.includes('JOIN') ||
			act.includes('UPGRADE') ||
			act.includes('ACCEPT')
		) {
			return 'text-emerald-400';
		}
		if (
			act.includes('UPDATE') ||
			act.includes('INVITE') ||
			act.includes('PROMOTE') ||
			act.includes('DEMOTE')
		) {
			return 'text-sky-400';
		}
		if (
			act.includes('DELETE') ||
			act.includes('KICK') ||
			act.includes('DISBAND') ||
			act.includes('TRANSFER') ||
			act.includes('REJECT') ||
			act.includes('LEAVE') ||
			act.includes('REVOKE')
		) {
			return 'text-rose-400';
		}
		return 'text-muted-foreground';
	}

	function initials(name?: string): string {
		if (!name) return 'US';
		const parts = name.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	let tz = $derived($userTimezone);
	const fmtTimeAgo = formatTimeAgo;
	const fmtFullDate = (iso: string) => formatDateTime(iso, tz);

	async function fetchAuditLogs(showSkeleton = false) {
		if (!orgId) return;
		if (showSkeleton) {
			isInitialLoading = true;
		} else {
			isFetching = true;
		}
		errorMsg = '';

		try {
			const cat = categories.find((c) => c.id === selectedCategory);
			const entityTypeFilter = cat && cat.id !== 'ALL' ? cat.entityTypes : undefined;

			const res = await AuditAPI.getLogs(orgId, {
				page: currentPage,
				limit: 20,
				search: searchQuery.trim() || undefined,
				action: selectedAction || undefined,
				entity_type: entityTypeFilter || undefined
			});

			logs = res.logs ?? [];
			totalCount = res.total_count ?? 0;
			totalPages = res.total_pages ?? 1;
		} catch (err: any) {
			errorMsg = err.message || 'Failed to load audit logs.';
			logs = [];
		} finally {
			isInitialLoading = false;
			isFetching = false;
		}
	}

	const debouncedFetchLogs = debounce(() => {
		currentPage = 1;
		fetchAuditLogs(false);
	}, 350);

	function handleCategoryChange(catId: string) {
		selectedCategory = catId;
		currentPage = 1;

		if (selectedAction) {
			if (catId !== 'ALL') {
				const group = actionFilterGroups.find((g) => g.categoryId === catId);
				if (!group || !group.actions.some((a) => a.value === selectedAction)) {
					selectedAction = '';
				}
			}
		}

		fetchAuditLogs(false);
	}

	function handleActionFilterChange(actionVal: string) {
		selectedAction = actionVal;
		if (actionVal) {
			const parentGroup = actionFilterGroups.find((g) =>
				g.actions.some((a) => a.value === actionVal)
			);
			if (
				parentGroup &&
				selectedCategory !== 'ALL' &&
				selectedCategory !== parentGroup.categoryId
			) {
				selectedCategory = parentGroup.categoryId;
			}
		}
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function resetFilters() {
		selectedCategory = 'ALL';
		selectedAction = '';
		headerStore.searchQuery = '';
		searchQuery = '';
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function clearSearch() {
		headerStore.searchQuery = '';
		searchQuery = '';
		currentPage = 1;
		fetchAuditLogs(false);
	}

	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search audit logs...',
			badge: {
				text: `${totalCount} Event${totalCount !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	let prevSearch = $state<string | null>(null);

	// Only trigger search fetch when searchQuery actually changes after initial mount
	$effect(() => {
		const current = headerStore.searchQuery;
		if (prevSearch === null) {
			prevSearch = current;
			searchQuery = current;
			return;
		}
		if (current !== prevSearch) {
			prevSearch = current;
			searchQuery = current;
			debouncedFetchLogs();
		}
	});

	onMount(() => {
		fetchAuditLogs(true);
	});
</script>

<svelte:head>
	<title>Audit Logs | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased mx-auto w-full flex-1 flex flex-col min-h-full">
	<!-- Filter Controls Toolbar -->
	<div class="flex flex-col gap-2.5 mb-6">
		<div
			class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-2.5 bg-card border border-border rounded-lg"
		>
			<!-- Left: Category & Action Selectors -->
			<div class="flex items-center gap-2 flex-wrap">
				<!-- Category Dropdown Trigger -->
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="focus:outline-none">
						<button
							type="button"
							class="flex items-center gap-2 h-9 px-3 rounded-lg text-xs font-medium border transition-colors cursor-pointer {selectedCategory !==
							'ALL'
								? 'bg-primary/10 border-primary/30 text-primary'
								: 'bg-surface-sidebar border-border text-foreground hover:bg-white/5'}"
						>
							<activeCategory.icon
								class="size-3.5 {selectedCategory !== 'ALL'
									? 'text-primary'
									: 'text-muted-foreground'}"
							/>
							<span>{activeCategory.label}</span>
							<ChevronDown class="size-3 opacity-60 ml-0.5" />
						</button>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						data-lenis-prevent
						class="w-64 max-h-95 overflow-y-auto bg-card border-border text-foreground rounded-lg p-1.5 shadow-xl z-50"
						align="start"
						sideOffset={6}
					>
						<DropdownMenu.Label
							class="text-[10px] text-muted-foreground font-medium px-2 py-1 uppercase tracking-wider"
						>
							Filter by Category
						</DropdownMenu.Label>
						{#each categories as cat (cat.id)}
							{@const isSelected = selectedCategory === cat.id}
							<DropdownMenu.Item
								class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs transition-colors {isSelected
									? 'text-primary bg-white/5 font-semibold'
									: 'text-foreground'}"
								onclick={() => handleCategoryChange(cat.id)}
							>
								<div class="flex items-center gap-2">
									<cat.icon
										class="size-3.5 {isSelected ? 'text-primary' : 'text-muted-foreground'}"
									/>
									<span>{cat.label}</span>
								</div>
								{#if isSelected}
									<CheckCircle2 class="size-3 text-primary shrink-0" />
								{/if}
							</DropdownMenu.Item>
						{/each}
					</DropdownMenu.Content>
				</DropdownMenu.Root>


				<!-- Reset Filters button -->
				{#if hasActiveFilters}
					<button
						type="button"
						class="flex items-center gap-1.5 h-9 px-2.5 text-xs text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg transition-colors cursor-pointer"
						onclick={resetFilters}
					>
						<RotateCcw class="size-3" />
						<span>Reset</span>
					</button>
				{/if}
			</div>

			<!-- Right: Total Event Counter -->
			<div class="flex items-center gap-2 self-end sm:self-auto">
				<span
					class="text-xs font-mono text-muted-foreground px-2.5 py-1 bg-surface-sidebar border border-border rounded-md"
				>
					{totalCount} event{totalCount !== 1 ? 's' : ''}
				</span>
			</div>
		</div>

		<!-- Active Filter Badges Bar (shown only when a filter is applied) -->
		{#if hasActiveFilters}
			<div class="flex items-center gap-2 flex-wrap text-xs px-1">
				<span class="text-[11px] text-muted-foreground font-medium">Active:</span>

				{#if selectedCategory !== 'ALL'}
					<span
						class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-primary/10 border border-primary/20 text-primary text-[11px] font-medium"
					>
						<activeCategory.icon class="size-3" />
						<span>Category: {activeCategory.label}</span>
						<button
							type="button"
							class="hover:text-foreground transition-colors cursor-pointer ml-0.5"
							onclick={() => handleCategoryChange('ALL')}
							aria-label="Clear category filter"
						>
							<X class="size-3" />
						</button>
					</span>
				{/if}

				{#if selectedAction}
					<span
						class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-primary/10 border border-primary/20 text-primary text-[11px] font-medium"
					>
						<ListFilter class="size-3" />
						<span>Action: {selectedActionLabel}</span>
						<button
							type="button"
							class="hover:text-foreground transition-colors cursor-pointer ml-0.5"
							onclick={() => handleActionFilterChange('')}
							aria-label="Clear action filter"
						>
							<X class="size-3" />
						</button>
					</span>
				{/if}

				{#if searchQuery}
					<span
						class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-white/5 border border-border text-foreground text-[11px] font-medium"
					>
						<Search class="size-3 text-muted-foreground" />
						<span>"{searchQuery}"</span>
						<button
							type="button"
							class="text-muted-foreground hover:text-foreground transition-colors cursor-pointer ml-0.5"
							onclick={clearSearch}
							aria-label="Clear search"
						>
							<X class="size-3" />
						</button>
					</span>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Error alert -->
	{#if errorMsg}
		<div
			class="p-4 border border-rose-500/20 bg-rose-500/10 rounded-lg flex items-center gap-3 text-rose-400 text-sm mb-6"
		>
			<AlertCircle class="size-5 shrink-0" />
			<span>{errorMsg}</span>
		</div>
	{/if}

	<!-- Logs List / Table -->
	{#if isInitialLoading}
		<DataSkeleton variant="table" count={8} />
	{:else if logs.length === 0}
		<EmptyState
			icon={ScrollText}
			title="No Audit Logs Found"
			description={searchQuery
				? `No activity matching "${searchQuery}" in this organization.`
				: 'No recorded activity events for this organization yet.'}
		/>
	{:else}
		<div
			class="border border-border rounded-lg bg-card overflow-hidden flex flex-col max-h-[calc(100vh-16rem)] transition-opacity duration-200 {isFetching
				? 'opacity-50 pointer-events-none'
				: 'opacity-100'}"
		>
			<!-- Table Column Headers -->
			<div class="log-table-header sticky top-0 bg-[#0a0a0b] z-10">
				<div class="log-col-actor">Actor</div>
				<div class="log-col-action">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="focus:outline-none w-full">
							<span
								class="flex items-center gap-1.5 transition-colors {selectedAction
									? 'text-primary'
									: 'text-inherit hover:text-foreground'}"
							>
								<span class="truncate">{selectedActionLabel || 'Action'}</span>
								{#if selectedAction}
									<!-- svelte-ignore a11y_no_static_element_interactions -->
									<span
										role="button"
										tabindex="-1"
										class="size-3.5 rounded-full bg-primary/20 hover:bg-primary/40 flex items-center justify-center transition-colors shrink-0"
										onclick={(e) => {
											e.stopPropagation();
											handleActionFilterChange('');
										}}
										onkeydown={(e) => {
											if (e.key === 'Enter') {
												e.stopPropagation();
												handleActionFilterChange('');
											}
										}}
									>
										<X class="size-2.5 text-primary" />
									</span>
								{:else}
									<ListFilter class="size-3 opacity-50 shrink-0" />
								{/if}
							</span>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							data-lenis-prevent
							class="w-72 max-h-90 overflow-y-auto bg-card border-border text-foreground rounded-lg p-1.5 shadow-xl z-50"
							align="start"
							sideOffset={6}
						>
							<!-- All Actions option -->
							<DropdownMenu.Item
								class="cursor-pointer flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-white/8 text-xs font-medium {selectedAction ===
								''
									? 'text-primary bg-white/5 font-semibold'
									: 'text-foreground'}"
								onclick={() => handleActionFilterChange('')}
							>
								<span>All Actions</span>
								{#if selectedAction === ''}
									<CheckCircle2 class="size-3 text-primary shrink-0" />
								{/if}
							</DropdownMenu.Item>

							{#each visibleActionGroups as group (group.categoryId)}
								<DropdownMenu.Separator class="bg-white/8 my-1" />
								<DropdownMenu.Label
									class="text-[10px] text-muted-foreground font-medium px-2.5 py-1 uppercase tracking-wider flex items-center gap-1.5"
								>
									<group.icon class="size-3 opacity-70" />
									<span>{group.label}</span>
								</DropdownMenu.Label>
								{#each group.actions as action (action.value)}
									{@const isSelected = selectedAction === action.value}
									{@const style = getActionStyle(action.value)}
									<DropdownMenu.Item
										class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs {isSelected
											? 'text-primary bg-white/5 font-semibold'
											: 'text-foreground'}"
										onclick={() => handleActionFilterChange(action.value)}
									>
										<span class="font-mono text-[11px] {style}">{action.label}</span>
										{#if isSelected}
											<CheckCircle2 class="size-3 text-primary shrink-0" />
										{/if}
									</DropdownMenu.Item>
								{/each}
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
				<div class="log-col-desc">Event / Details</div>
				<div class="log-col-ip">IP Address</div>
				<div class="log-col-date text-right">Time</div>
			</div>

			<!-- Log Rows (Scrollable) -->
			<div data-lenis-prevent class="overflow-y-auto flex-1 divide-y divide-white/5">
				{#each logs as log (log.id)}
					{@const actionStyle = getActionStyle(log.action)}
					<div
						class="log-table-row hover:bg-white/3 transition-colors text-xs py-3 px-4 flex flex-col md:grid md:grid-cols-[200px_160px_1fr_130px_140px] items-start md:items-center gap-2 md:gap-4"
					>
						<!-- Col 1: Actor -->
						<div class="flex items-center gap-2.5 min-w-0 w-full md:w-auto">
							<div
								class="size-7 rounded-lg bg-primary/10 border border-primary/20 text-primary flex items-center justify-center font-bold text-[11px] shrink-0"
							>
								{#if log.user_avatar_url}
									<img
										src={log.user_avatar_url}
										alt={log.user_name || 'User'}
										class="size-full object-cover rounded-lg"
									/>
								{:else}
									{initials(log.user_name)}
								{/if}
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-foreground font-medium truncate text-xs">
									{log.user_name || 'Unknown'}
								</p>
								<p class="text-[10px] text-muted-foreground font-mono truncate">
									{log.user_email || 'System'}
								</p>
							</div>
						</div>

						<!-- Col 2: Action -->
						<div class="w-full md:w-auto">
							<span
								class="inline-flex items-center px-2 py-0.5 rounded-md font-mono text-[11px] font-semibold {actionStyle} bg-white/5 border border-white/5"
							>
								{log.action}
							</span>
						</div>

						<!-- Col 3: Description -->
						<div class="min-w-0 w-full md:w-auto">
							<p class="text-foreground/90 font-normal truncate leading-relaxed">
								{log.description}
							</p>
							{#if log.details}
								<p class="text-[11px] text-muted-foreground/60 font-mono truncate mt-0.5">
									{log.details}
								</p>
							{/if}
						</div>

						<!-- Col 4: IP Address -->
						<div class="text-[11px] font-mono text-muted-foreground/60 w-full md:w-auto">
							{log.ip_address || '—'}
						</div>

						<!-- Col 5: Timestamp -->
						<div class="text-right w-full md:w-auto" title={fmtFullDate(log.created_at)}>
							<span class="text-xs text-muted-foreground/80 font-medium">
								{fmtTimeAgo(log.created_at)}
							</span>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Pagination Footer -->
		<Pagination
			bind:currentPage
			{totalPages}
			totalItems={totalCount}
			itemName="events"
			onPageChange={() => fetchAuditLogs(false)}
			class="mt-4"
		/>
	{/if}
</div>

<style>
	.log-table-header {
		display: none;
		padding: 0.625rem 1rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: rgba(255, 255, 255, 0.4);
	}

	@media (min-width: 768px) {
		.log-table-header {
			display: grid;
			grid-template-columns: 200px 160px 1fr 130px 140px;
			gap: 1rem;
			align-items: center;
		}
	}
</style>
