<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import {
		ScrollText,
		Search,
		ChevronLeft,
		ChevronRight,
		Activity,
		Users,
		Globe,
		CheckCircle2,
		AlertCircle,
		ListFilter,
		X
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { AuditAPI, type AuditLogItem } from '$lib/api/audit';
	import { debounce } from '$lib/helper/utils';
	import LogsSkeleton from '$lib/components/skeletons/LogsSkeleton.svelte';
	import { EmptyState, Pagination } from '$lib/components/common';
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
	let selectedCategory = $state<'ALL' | 'MONITOR' | 'MEMBER' | 'STATUSPAGE'>('ALL');
	let selectedAction = $state('');

	// ── Category Filters Definition ────────────────────────────────────────────
	const categories = [
		{ id: 'ALL', label: 'All Actions', icon: ScrollText },
		{ id: 'MONITOR', label: 'Monitors', icon: Activity },
		{ id: 'MEMBER', label: 'Members', icon: Users },
		{ id: 'STATUSPAGE', label: 'Status Pages', icon: Globe }
	];

	// ── Action Filter Definitions (grouped by category) ────────────────────────
	const actionFilterGroups = [
		{
			label: 'Monitor',
			actions: [
				{ value: 'MONITOR_CREATE', label: 'Monitor Create' },
				{ value: 'MONITOR_UPDATE', label: 'Monitor Update' },
				{ value: 'MONITOR_DELETE', label: 'Monitor Delete' }
			]
		},
		{
			label: 'Organization',
			actions: [
				{ value: 'ORG_UPDATE', label: 'Org Update' },
				{ value: 'ORG_DISBAND', label: 'Org Disband' },
				{ value: 'ORG_TRANSFER_OWNERSHIP', label: 'Org Transfer' }
			]
		},
		{
			label: 'Members',
			actions: [
				{ value: 'ORG_MEMBER_JOIN', label: 'Member Join' },
				{ value: 'ORG_MEMBER_INVITE', label: 'Member Invite' },
				{ value: 'ORG_MEMBER_PROMOTE', label: 'Member Promote' },
				{ value: 'ORG_MEMBER_DEMOTE', label: 'Member Demote' },
				{ value: 'ORG_MEMBER_KICK', label: 'Member Kick' },
				{ value: 'ORG_MEMBER_LEAVE', label: 'Member Leave' },
				{ value: 'ORG_INVITE_REJECT', label: 'Invite Reject' }
			]
		},
		{
			label: 'Status Page',
			actions: [
				{ value: 'STATUSPAGE_CREATE', label: 'Page Create' },
				{ value: 'STATUSPAGE_UPDATE', label: 'Page Update' },
				{ value: 'STATUSPAGE_DELETE', label: 'Page Delete' }
			]
		},
		{
			label: 'API Key',
			actions: [
				{ value: 'API_KEY_CREATE', label: 'Key Create' },
				{ value: 'API_KEY_REVOKE', label: 'Key Revoke' }
			]
		}
	];

	const categoryToGroupLabels: Record<string, string[]> = {
		ALL: ['Monitor', 'Organization', 'Members', 'Status Page', 'API Key'],
		MONITOR: ['Monitor'],
		MEMBER: ['Organization', 'Members'],
		STATUSPAGE: ['Status Page']
	};

	let visibleActionGroups = $derived(
		actionFilterGroups.filter((g) =>
			(categoryToGroupLabels[selectedCategory] ?? []).includes(g.label)
		)
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
			act.includes('LEAVE')
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
			let entityTypeFilter = '';
			if (selectedCategory !== 'ALL') {
				switch (selectedCategory) {
					case 'MONITOR':
						entityTypeFilter = 'monitor';
						break;
					case 'MEMBER':
						entityTypeFilter = 'organization_member';
						break;
					case 'STATUSPAGE':
						entityTypeFilter = 'status_page';
						break;
				}
			}

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

	function handleCategoryChange(cat: 'ALL' | 'MONITOR' | 'MEMBER' | 'STATUSPAGE') {
		selectedCategory = cat;
		currentPage = 1;

		if (selectedAction) {
			const allowedLabels = categoryToGroupLabels[cat] ?? [];
			const allowedGroups = actionFilterGroups.filter((g) => allowedLabels.includes(g.label));
			const allAllowedActions = allowedGroups.flatMap((g) => g.actions.map((a) => a.value));
			if (!allAllowedActions.includes(selectedAction)) {
				selectedAction = '';
			}
		}

		fetchAuditLogs(false);
	}

	function handleSearchInput() {
		debouncedFetchLogs();
	}

	function handleActionFilterChange(action: string) {
		selectedAction = action;
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages || p === currentPage) return;
		currentPage = p;
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
			// First run on mount: initialize tracking without triggering duplicate fetch
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
	<title>Audit Logs | PulseWatch</title>
</svelte:head>

<div class="text-foreground antialiased mx-auto w-full flex-1 flex flex-col min-h-full">
	<!-- Controls: Search & Category Chips -->
	<div class="flex flex-col gap-4 mb-6">
		<div
			class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-3 bg-card border border-border rounded-lg"
		>
			<!-- Category Filter Buttons -->
			<div class="flex items-center gap-1 overflow-x-auto pb-1 sm:pb-0">
				{#each categories as cat (cat.id)}
					{@const isSelected = selectedCategory === cat.id}
					<button
						type="button"
						class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs transition-all whitespace-nowrap border cursor-pointer {isSelected
							? 'font-bold text-foreground bg-white/10 border-border'
							: 'font-medium text-foreground/40 border-transparent bg-transparent hover:text-foreground/70 hover:bg-white/5'}"
						onclick={() => handleCategoryChange(cat.id as any)}
					>
						<cat.icon class="size-3.5 {isSelected ? 'text-primary' : 'text-foreground/40'}" />
						{cat.label}
					</button>
				{/each}
			</div>
		</div>
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
		<LogsSkeleton />
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
			class="border border-border rounded-lg bg-white/3 backdrop-blur-md overflow-hidden flex flex-col max-h-[calc(100vh-16rem)] transition-opacity duration-200 {isFetching
				? 'opacity-50 pointer-events-none'
				: 'opacity-100'}"
		>
			<!-- Table Column Headers -->
			<div class="log-table-header sticky top-0 bg-card z-10">
				<div class="log-col-actor">Actor</div>
				<div class="log-col-action">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="focus:outline-none w-full">
							<span
								class="flex items-center gap-1.5 transition-colors {selectedAction
									? 'text-primary'
									: 'text-inherit hover:text-foreground'}"
							>
								<span>{selectedAction || 'Action'}</span>
								{#if selectedAction}
									<!-- svelte-ignore a11y_no_static_element_interactions -->
									<span
										role="button"
										tabindex="-1"
										class="size-3.5 rounded-full bg-primary/20 hover:bg-primary/40 flex items-center justify-center transition-colors"
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
									<ListFilter class="size-3 opacity-50" />
								{/if}
							</span>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							data-lenis-prevent
							class="w-56 max-h-[360px] overflow-y-auto bg-secondary border-border text-foreground rounded-lg p-1"
							align="start"
							sideOffset={6}
						>
							<!-- All Actions option -->
							<DropdownMenu.Item
								class="cursor-pointer flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-white/8 text-sm font-medium {selectedAction ===
								''
									? 'text-primary bg-white/5'
									: 'text-slate-200'}"
								onclick={() => handleActionFilterChange('')}
							>
								<span>All Actions</span>
								{#if selectedAction === ''}
									<CheckCircle2 class="size-3.5 text-primary shrink-0" />
								{/if}
							</DropdownMenu.Item>

							{#each visibleActionGroups as group (group.label)}
								<DropdownMenu.Separator class="bg-white/8 my-1" />
								<DropdownMenu.Label
									class="text-[10px] text-muted-foreground font-medium px-2.5 py-1 uppercase tracking-wider"
								>
									{group.label}
								</DropdownMenu.Label>
								{#each group.actions as action (action.label)}
									{@const style = getActionStyle(action.value)}
									<DropdownMenu.Item
										class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs {selectedAction ===
										action.value
											? 'text-primary bg-white/5'
											: 'text-slate-300'}"
										onclick={() => handleActionFilterChange(action.value)}
									>
										<span class="font-mono {style}">{action.label}</span>
										{#if selectedAction === action.value}
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
