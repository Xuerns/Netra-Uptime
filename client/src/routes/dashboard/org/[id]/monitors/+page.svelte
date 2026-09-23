<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { flip } from 'svelte/animate';
	import {
		Activity,
		Plus,
		TriangleAlert,
		LayoutGrid,
		List,
		Globe,
		Radio,
		Cpu,
		Search
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import MonitorCard from '$lib/components/monitor/MonitorCard.svelte';
	import MonitorListRow from '$lib/components/monitor/MonitorListRow.svelte';
	import MonitorSkeleton from '$lib/components/skeletons/MonitorSkeleton.svelte';
	import MonitorOverviewStats from '$lib/components/monitor/MonitorOverviewStats.svelte';
	import AddMonitorModal from '$lib/components/monitor/AddMonitorModal.svelte';
	import EditMonitorModal from '$lib/components/monitor/EditMonitorModal.svelte';
	import { wsClient, type MonitorUpdatePayload } from '$lib/utils/ws';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	// ── Types ─────────────────────────────────────────────────────────────────
	type StatusFilter = 'all' | 'up' | 'flagged' | 'paused';
	type ViewMode = 'grid' | 'list';

	// ── Route param & Organization ─────────────────────────────────────────────
	const orgId = $derived(($page.params.id as string) || '');
	const orgName = $derived($page.data?.org?.Name || $page.data?.org?.name || 'Workspace');

	// ── State ─────────────────────────────────────────────────────────────────
	let monitors = $state<Monitor[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let statusFilter = $state<StatusFilter>('all');
	let timeWindow = $state<'30d' | '24h'>('30d');
	let openDropdownId = $state<string | null>(null);

	// Add & Edit Modals
	let isAddMonitorOpen = $state(false);
	let selectedMonitorType = $state<'http' | 'ping' | 'tcp' | null>(null);
	let isEditMonitorOpen = $state(false);
	let selectedMonitorToEdit = $state<Monitor | null>(null);

	function openAddMonitor(type: 'http' | 'ping' | 'tcp') {
		selectedMonitorType = type;
		isAddMonitorOpen = true;
	}

	function handleEditMonitor(m: Monitor) {
		selectedMonitorToEdit = m;
		isEditMonitorOpen = true;
	}

	// ── Derived ───────────────────────────────────────────────────────────────
	let counts = $derived({
		total: monitors.length,
		up: monitors.filter((m) => m.Status === 'up').length,
		down: monitors.filter((m) => m.Status === 'down').length,
		paused: monitors.filter((m) => m.Status === 'paused').length,
		flagged: monitors.filter(
			(m) =>
				m.Status === 'down' ||
				m.DomainStatus === 'warning' ||
				m.DomainStatus === 'error' ||
				m.SSLStatus === 'warning' ||
				m.SSLStatus === 'error' ||
				m.SSLStatus === 'expired'
		).length
	});

	// Sync with unified layout header bar
	$effect(() => {
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search monitors (URL, tag, name)...',
			actionType: 'monitor',
			onAction: (type: 'http' | 'ping' | 'tcp') => openAddMonitor(type),
			badge:
				counts.total === 0
					? { text: 'No monitors', variant: 'slate' }
					: counts.down > 0
						? {
								text: `${counts.down} service${counts.down > 1 ? 's' : ''} degraded`,
								variant: 'rose'
							}
						: counts.flagged > 0
							? {
									text: `${counts.flagged} notice${counts.flagged > 1 ? 's' : ''}`,
									variant: 'amber'
								}
							: { text: 'All systems monitored', variant: 'emerald' }
		});
	});

	let filtered = $derived(
		monitors.filter((m) => {
			let matchesStatus = true;
			if (statusFilter === 'up') {
				matchesStatus = m.Status === 'up';
			} else if (statusFilter === 'paused') {
				matchesStatus = m.Status === 'paused';
			} else if (statusFilter === 'flagged') {
				matchesStatus =
					m.Status === 'down' ||
					m.DomainStatus === 'warning' ||
					m.DomainStatus === 'error' ||
					m.SSLStatus === 'warning' ||
					m.SSLStatus === 'error' ||
					m.SSLStatus === 'expired';
			}
			const q = headerStore.searchQuery.toLowerCase().trim();
			const matchesSearch =
				!q || m.Name.toLowerCase().includes(q) || m.Target.toLowerCase().includes(q);
			return matchesStatus && matchesSearch;
		})
	);

	// ── Actions ───────────────────────────────────────────────────────────────
	function toggleDropdown(id: string) {
		openDropdownId = openDropdownId === id ? null : id;
	}

	async function handleTogglePause(m: Monitor) {
		if (!orgId) return;
		try {
			await MonitorAPI.update(orgId, m.ID, {
				is_active: m.Status === 'paused'
			});
			await loadMonitors(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to toggle monitor status'));
		}
	}

	async function handleDeleteMonitor(id: string) {
		if (!orgId) return;
		const m = monitors.find((item) => item.ID === id);
		if (!m) return;
		try {
			await MonitorAPI.delete(orgId, m.ID);
			toast.success('Monitor deleted successfully');
			await loadMonitors(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete monitor'));
		}
	}

	async function loadMonitors(silent = false) {
		if (!orgId) return;
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const res = await MonitorAPI.getAll(orgId);
			monitors = res.data ?? [];
		} catch (err: any) {
			fetchError = parseApiError(err, 'Failed to fetch monitors.');
		} finally {
			isLoading = false;
		}
	}

	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		localStorage.setItem('monitorViewMode', mode);
	}

	onMount(() => {
		loadMonitors();
		const saved = localStorage.getItem('monitorViewMode');
		if (saved === 'grid' || saved === 'list') viewMode = saved;

		// Listen to WebSocket updates
		const unsubscribe = wsClient.subscribe((msg) => {
			if (msg.Type === 'MONITOR_UPDATED') {
				const payload = msg.Payload as MonitorUpdatePayload;

				const index = monitors.findIndex((m) => m.ID === payload.id);
				if (index !== -1) {
					monitors[index].Status = payload.status;
					monitors[index].ResponseTimeMs = payload.response_time;
					monitors[index].LastCheckedAt = payload.last_checked;
				}
			}
		});

		return () => {
			unsubscribe();
		};
	});


</script>

<svelte:head>
	<title>Monitors | PulseWatch</title>
</svelte:head>

<div class="monitors-page text-foreground antialiased space-y-6">
	<!-- ── Loading Skeleton ──────────────────────────────────────────────────── -->
	{#if isLoading}
		<MonitorSkeleton {viewMode} />

		<!-- ── Error state ────────────────────────────────────────────────────────── -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-card"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium rounded-lg"
				onclick={() => loadMonitors()}
			>
				Try again
			</Button>
		</div>

		<!-- ── Main Data ─────────────────────────────────────────────────────────── -->
	{:else}
		{#if monitors.length === 0}
			<!-- Empty state -->
			<div
				class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-card"
			>
				<div
					class="flex items-center justify-center size-14 rounded-lg bg-primary/10 border border-primary/20 mb-4 text-primary"
				>
					<Activity class="size-6" />
				</div>
				<h3 class="text-xl font-semibold text-foreground mb-1.5">No monitors yet</h3>
				<p class="text-[14px] text-muted-foreground max-w-[38ch] mb-6 leading-relaxed">
					Create a monitor to start tracking uptime and response times for your services.
				</p>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="inline-flex items-center gap-1.5 bg-emerald-500 hover:bg-emerald-400 text-zinc-950 px-3.5 py-1.5 rounded-lg font-semibold text-xs transition-colors active:scale-95 cursor-pointer focus:outline-none"
					>
						<Plus class="size-3.5 stroke-[2.5]" />
						<span>Add monitor</span>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						align="center"
						class="bg-card border-border text-foreground min-w-[160px] p-1.5 rounded-lg"
					>
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
							onclick={() => openAddMonitor('http')}
						>
							<Globe class="mr-2 size-3.5 text-emerald-400" /> HTTP/API
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
							onclick={() => openAddMonitor('ping')}
						>
							<Radio class="mr-2 size-3.5 text-emerald-400" /> Ping
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
							onclick={() => openAddMonitor('tcp')}
						>
							<Cpu class="mr-2 size-3.5 text-emerald-400" /> TCP (Port)
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>
		{:else}
			<!-- BEGIN: OverviewHeroStrip -->
			<MonitorOverviewStats {monitors} />

			<!-- BEGIN: FilterToolbar -->
			<div
				class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 pt-2"
				data-purpose="monitors-filter-bar"
			>
				<!-- Status Filter Pills -->
				<div
					class="flex items-center gap-1 bg-surface-sidebar p-1 rounded-lg border border-border text-xs"
					role="group"
					aria-label="Filter by status"
				>
					<button
						class="px-3 py-1 font-medium rounded-md transition-colors {statusFilter === 'all'
							? 'bg-muted text-foreground'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (statusFilter = 'all')}
					>
						All <span class="ml-1 text-muted-foreground font-mono text-[11px]">{counts.total}</span>
					</button>

					<button
						class="px-3 py-1 font-medium rounded-md transition-colors {statusFilter === 'up'
							? 'bg-muted text-foreground'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (statusFilter = 'up')}
					>
						Operational <span class="ml-1 text-emerald-400 font-mono text-[11px]">{counts.up}</span>
					</button>

					<button
						class="px-3 py-1 font-medium rounded-md transition-colors {statusFilter === 'flagged'
							? 'bg-muted text-foreground'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (statusFilter = 'flagged')}
					>
						DNS Flagged <span class="ml-1 text-amber-400 font-mono text-[11px]"
							>{counts.flagged}</span
						>
					</button>

					<button
						class="px-3 py-1 font-medium rounded-md transition-colors {statusFilter === 'paused'
							? 'bg-muted text-foreground'
							: 'text-muted-foreground hover:text-foreground'}"
						onclick={() => (statusFilter = 'paused')}
					>
						Paused <span class="ml-1 text-muted-foreground font-mono text-[11px]"
							>{counts.paused}</span
						>
					</button>
				</div>

				<!-- View Controls & Time Range -->
				<div class="flex items-center gap-2 self-stretch sm:self-auto justify-end">
					<div
						class="flex items-center bg-surface-sidebar border border-border rounded-lg p-0.5 text-xs"
					>
						<span class="px-2.5 py-1 text-muted-foreground font-medium">Window:</span>
						<button
							class="px-2 py-1 rounded font-mono text-[11px] transition-colors {timeWindow === '30d'
								? 'bg-muted text-foreground'
								: 'text-muted-foreground hover:text-foreground'}"
							onclick={() => (timeWindow = '30d')}
						>
							30d
						</button>
						<button
							class="px-2 py-1 rounded font-mono text-[11px] transition-colors {timeWindow === '24h'
								? 'bg-muted text-foreground'
								: 'text-muted-foreground hover:text-foreground'}"
							onclick={() => (timeWindow = '24h')}
						>
							24h
						</button>
					</div>

					<!-- View Mode Toggle -->
					<div class="flex items-center bg-surface-sidebar border border-border rounded-lg p-0.5">
						<button
							class="p-1.5 rounded transition-colors {viewMode === 'grid'
								? 'bg-muted text-foreground'
								: 'text-muted-foreground hover:text-foreground'}"
							onclick={() => setViewMode('grid')}
							aria-label="Grid view"
							aria-pressed={viewMode === 'grid'}
						>
							<LayoutGrid class="size-3.5" />
						</button>
						<button
							class="p-1.5 rounded transition-colors {viewMode === 'list'
								? 'bg-muted text-foreground'
								: 'text-muted-foreground hover:text-foreground'}"
							onclick={() => setViewMode('list')}
							aria-label="List view"
							aria-pressed={viewMode === 'list'}
						>
							<List class="size-3.5" />
						</button>
					</div>
				</div>
			</div>

			<!-- Search / filter no-results -->
			{#if filtered.length === 0}
				<div
					class="flex flex-col items-center text-center py-16 border border-border rounded-lg bg-card"
				>
					<Search class="size-6 text-muted-foreground/30 mb-3" />
					<p class="text-[14px] text-muted-foreground">
						No monitors match
						<strong class="text-foreground font-medium"
							>{headerStore.searchQuery === '' ? '' : `"${headerStore.searchQuery}"`}</strong
						>
						{statusFilter !== 'all' ? `with filter "${statusFilter}"` : ''}
					</p>
					<button
						class="mt-3 text-[13px] text-primary hover:underline"
						onclick={() => {
							headerStore.searchQuery = '';
							statusFilter = 'all';
						}}
					>
						Clear filters
					</button>
				</div>

				<!-- ── Grid view (3 columns per contohpage.html) ───────────────────────── -->
			{:else if viewMode === 'grid'}
				<div class="monitors-scroll-area">
					<div data-lenis-prevent class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
						{#each filtered as monitor (monitor.ID)}
							<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
								<MonitorCard
									{monitor}
									{openDropdownId}
									{toggleDropdown}
									{handleDeleteMonitor}
									{handleEditMonitor}
									{handleTogglePause}
								/>
							</div>
						{/each}
					</div>
				</div>

				<!-- ── List view ──────────────────────────────────────────────────────── -->
			{:else}
				<div class="monitors-scroll-area">
					<div
						data-lenis-prevent
						class="flex flex-col divide-y divide-border border border-border rounded-lg bg-card"
					>
						<!-- Header row -->
						<div class="list-row list-row--header" aria-hidden="true">
							<span>Monitor</span>
							<span>Status</span>
							<span>Response</span>
							<span>Uptime 30d</span>
							<span>Interval</span>
							<span>Last check</span>
							<span class="text-right">Actions</span>
						</div>

						{#each filtered as monitor (monitor.ID)}
							<MonitorListRow
								{monitor}
								{openDropdownId}
								{toggleDropdown}
								{handleDeleteMonitor}
								{handleEditMonitor}
								{handleTogglePause}
							/>
						{/each}
					</div>
				</div>
			{/if}
		{/if}
	{/if}
</div>

<AddMonitorModal
	bind:isOpen={isAddMonitorOpen}
	monitorType={selectedMonitorType}
	{orgId}
	onSuccess={() => loadMonitors(true)}
/>

<EditMonitorModal
	bind:isOpen={isEditMonitorOpen}
	monitor={selectedMonitorToEdit}
	{orgId}
	onSuccess={() => loadMonitors(true)}
/>

<style>
	/* ── Layout & Grids ───────────────────────────────────────────────────── */
	.monitors-scroll-area {
		width: 100%;
	}

	/* ── List view ────────────────────────────────────────────────────────── */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 100px 90px 90px 80px 110px 80px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: var(--color-accent);
	}

	.list-row--header {
		padding: 10px 20px;
		font-size: 11px;
		font-weight: 500;
		color: var(--color-muted-foreground);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		background: var(--color-muted);
		opacity: 0.65;
	}

	.list-row--header:hover {
		background: var(--color-muted);
	}

	@media (max-width: 900px) {
		.list-row {
			grid-template-columns: 1fr 100px 90px 80px;
		}
		.list-row > :nth-child(4),
		.list-row > :nth-child(5),
		.list-row > :nth-child(6) {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
