<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { flip } from 'svelte/animate';
	import { Activity, Plus, TriangleAlert, Globe, Radio, Cpu, Search } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { MonitorAPI, type Monitor, type FleetOverviewResponse } from '$lib/api/monitor';
	import MonitorItem from '$lib/components/monitor/MonitorItem.svelte';
	import MonitorFilterBar from '$lib/components/monitor/MonitorFilterBar.svelte';
	import MonitorListTable from '$lib/components/monitor/MonitorListTable.svelte';
	import { DataSkeleton } from '$lib/components/common';
	import MonitorOverviewStats from '$lib/components/monitor/MonitorOverviewStats.svelte';
	import MonitorModal from '$lib/components/monitor/MonitorModal.svelte';
	import { wsClient, type MonitorUpdatePayload } from '$lib/utils/ws';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	let { data }: { data: PageData } = $props();

	// ── Types ─────────────────────────────────────────────────────────────────
	type StatusFilter = 'all' | 'up' | 'down' | 'flagged' | 'paused';
	type ViewMode = 'grid' | 'list';

	// ── Route param & Organization ─────────────────────────────────────────────
	const orgId = $derived(data.orgId || ($page.params.id as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let monitors = $state<Monitor[]>([]);
	let overview = $state<FleetOverviewResponse | null>(null);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let statusFilter = $state<StatusFilter>('all');

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
		down: monitors.filter((m) => m.Status === 'down' || m.Status === 'degraded').length,
		paused: monitors.filter((m) => m.Status === 'paused').length,
		flagged: monitors.filter(
			(m) =>
				m.Status !== 'down' &&
				m.Status !== 'degraded' &&
				(m.DomainStatus === 'warning' ||
					m.DomainStatus === 'expired' ||
					m.SSLStatus === 'warning' ||
					m.SSLStatus === 'expired')
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
			} else if (statusFilter === 'down') {
				matchesStatus = m.Status === 'down' || m.Status === 'degraded';
			} else if (statusFilter === 'paused') {
				matchesStatus = m.Status === 'paused';
			} else if (statusFilter === 'flagged') {
				matchesStatus =
					m.Status !== 'down' &&
					m.Status !== 'degraded' &&
					(m.DomainStatus === 'warning' ||
						m.DomainStatus === 'expired' ||
						m.SSLStatus === 'warning' ||
						m.SSLStatus === 'expired');
			}
			const q = headerStore.searchQuery.toLowerCase().trim();
			const matchesSearch =
				!q || m.Name.toLowerCase().includes(q) || m.Target.toLowerCase().includes(q);
			return matchesStatus && matchesSearch;
		})
	);

	async function handleTogglePause(m: Monitor) {
		if (!orgId) return;
		try {
			const willResume = m.Status === 'paused';
			await MonitorAPI.update(orgId, m.ID, {
				is_active: willResume
			});
			toast.success(willResume ? 'Monitor resumed' : 'Monitor paused');
			await loadMonitors(true);
		} catch (err: unknown) {
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
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete monitor'));
		}
	}

	async function loadMonitors(silent = false) {
		if (!orgId) return;
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const [monitorsRes, overviewRes] = await Promise.all([
				MonitorAPI.getAll(orgId),
				MonitorAPI.getFleetOverview(orgId, 30).catch((err) => {
					console.error('Failed to load fleet overview', err);
					return null;
				})
			]);
			monitors = monitorsRes.data ?? [];
			if (overviewRes && overviewRes.data) {
				overview = overviewRes.data;
			}
		} catch (err: unknown) {
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
	<title>Monitors | Netra Uptime</title>
</svelte:head>

<div class="monitors-page text-foreground antialiased space-y-6">
	<!-- ── Loading Skeleton ──────────────────────────────────────────────────── -->
	{#if isLoading}
		<DataSkeleton variant="monitor" {viewMode} count={6} />

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
						class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 rounded-lg font-medium text-xs cursor-pointer focus:outline-none"
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
			<MonitorOverviewStats {monitors} {overview} />

			<!-- BEGIN: FilterToolbar -->
			<MonitorFilterBar
				{statusFilter}
				{viewMode}
				{counts}
				onFilterChange={(s) => (statusFilter = s)}
				onViewModeChange={(m) => setViewMode(m)}
			/>

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

				<!-- ── Grid view (up to 4 columns on wide screens) ─────────────────── -->
			{:else if viewMode === 'grid'}
				<div class="monitors-scroll-area">
					<div
						data-lenis-prevent
						class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 min-[1400px]:grid-cols-4 2xl:grid-cols-4 gap-4"
					>
						{#each filtered as monitor (monitor.ID)}
							<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
								<MonitorItem
									{monitor}
									viewMode="grid"
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
				<MonitorListTable
					monitors={filtered}
					onDelete={handleDeleteMonitor}
					onEdit={handleEditMonitor}
					onTogglePause={handleTogglePause}
				/>
			{/if}
		{/if}
	{/if}
</div>

<MonitorModal
	bind:isOpen={isAddMonitorOpen}
	monitorType={selectedMonitorType}
	{orgId}
	onSuccess={() => loadMonitors(true)}
/>

<MonitorModal
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
</style>
