<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { flip } from 'svelte/animate';
	import {
		Activity,
		Plus,
		TriangleAlert,
		Search,
		LayoutGrid,
		List,
		Pause,
		ShieldCheck,
		Globe,
		Radio,
		Cpu
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import MonitorCard from '$lib/components/monitor/MonitorCard.svelte';
	import MonitorListRow from '$lib/components/monitor/MonitorListRow.svelte';
	import MonitorSkeleton from '$lib/components/skeletons/MonitorSkeleton.svelte';
	import AddMonitorModal from '$lib/components/monitor/AddMonitorModal.svelte';
	import EditMonitorModal from '$lib/components/monitor/EditMonitorModal.svelte';
	import { wsClient, type MonitorUpdatePayload } from '$lib/utils/ws';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	// ── Types ─────────────────────────────────────────────────────────────────
	type StatusFilter = 'all' | 'up' | 'down' | 'paused';
	type ViewMode = 'grid' | 'list';

	// ── Route param ───────────────────────────────────────────────────────────
	const orgId = $derived(($page.params.id as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let monitors = $state<Monitor[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let statusFilter = $state<StatusFilter>('all');
	let searchQuery = $state('');
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
		pending: monitors.filter((m) => m.Status === 'pending').length
	});

	let filtered = $derived(
		monitors.filter((m) => {
			const matchesStatus = statusFilter === 'all' || m.Status === statusFilter;
			const q = searchQuery.toLowerCase().trim();
			const matchesSearch =
				!q || m.Name.toLowerCase().includes(q) || m.Target.toLowerCase().includes(q);
			return matchesStatus && matchesSearch;
		})
	);

	// ── Data ──────────────────────────────────────────────────────────────────
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

		return () => unsubscribe();
	});
</script>

<svelte:head>
	<title>Monitors | PulseWatch</title>
</svelte:head>

<div class="monitors-page text-foreground antialiased">
	<!-- ── Page Header ──────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[28px] font-semibold leading-tight tracking-tight text-foreground mb-1">
				Monitors
			</h1>
			<p class="text-[14px] leading-relaxed text-muted-foreground max-w-[52ch]">
				Track uptime and response times. Spot issues before they become incidents.
			</p>
		</div>

		<div class="flex items-center gap-2 flex-wrap">
			{#if monitors.length > 0}
				<!-- Search bar -->
				<div class="relative">
					<Search
						class="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/50 pointer-events-none"
					/>
					<input
						type="search"
						placeholder="Search monitors..."
						bind:value={searchQuery}
						class="h-9 pl-8 pr-3 w-52 bg-foreground/10 border border-border rounded-lg text-[13px] text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-colors"
					/>
				</div>

				<!-- Status filter pills -->
				<div
					class="flex items-center bg-foreground/10 border border-border rounded-lg p-0.5"
					role="group"
					aria-label="Filter by status"
				>
					<button
						class="stat-chip {statusFilter === 'all' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'all')}
					>
						All
						<span class="stat-count">{counts.total}</span>
					</button>

					<button
						class="stat-chip stat-chip--up {statusFilter === 'up' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'up')}
					>
						Up
						<span class="stat-count">{counts.up}</span>
					</button>

					<button
						class="stat-chip stat-chip--down {statusFilter === 'down' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'down')}
					>
						Down
						<span class="stat-count">{counts.down}</span>
					</button>

					<button
						class="stat-chip stat-chip--paused {statusFilter === 'paused'
							? 'stat-chip--active'
							: ''}"
						onclick={() => (statusFilter = 'paused')}
					>
						Paused
						<span class="stat-count">{counts.paused}</span>
					</button>
				</div>

				<!-- View mode toggle -->
				<div
					class="flex items-center bg-foreground/10 border border-border rounded-lg p-0.5"
					role="group"
					aria-label="View mode"
				>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'grid'
							? 'bg-white/10 text-foreground'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('grid')}
						aria-label="Grid view"
						aria-pressed={viewMode === 'grid'}
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-foreground'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
						aria-pressed={viewMode === 'list'}
					>
						<List class="size-4" />
					</button>
				</div>
			{/if}

			<!-- Add Monitor Dropdown -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 font-medium text-sm focus:outline-none"
				>
					<Plus class="size-4" />
					<span>Add monitor</span>
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					align="end"
					class="bg-card border-border text-foreground min-w-[160px] p-1.5 rounded-lg"
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
						<Radio class="mr-2 size-4 text-primary" /> Ping
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="hover:bg-white/10 cursor-pointer rounded-md p-2"
						onclick={() => openAddMonitor('tcp')}
					>
						<Cpu class="mr-2 size-4 text-primary" /> TCP (Port)
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
	</header>

	<!-- ── Loading Skeleton ──────────────────────────────────────────────────── -->
	{#if isLoading}
		<MonitorSkeleton {viewMode} />

		<!-- ── Error state ────────────────────────────────────────────────────────── -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-white/3"
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

		<!-- ── Data ──────────────────────────────────────────────────────────────── -->
	{:else}
		<!-- Empty state -->
		{#if monitors.length === 0}
			<div
				class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-white/3"
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
						class="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 font-medium text-sm focus:outline-none"
					>
						<Plus class="size-4" />
						<span>Add monitor</span>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						align="center"
						class="bg-card border-border text-foreground min-w-[160px] p-1.5 rounded-lg"
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
							<Radio class="mr-2 size-4 text-primary" /> Ping
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="hover:bg-white/10 cursor-pointer rounded-md p-2"
							onclick={() => openAddMonitor('tcp')}
						>
							<Cpu class="mr-2 size-4 text-primary" /> TCP (Port)
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>

			<!-- Search / filter no-results -->
		{:else if filtered.length === 0}
			<div
				class="flex flex-col items-center text-center py-16 border border-border rounded-lg bg-white/2"
			>
				<Search class="size-6 text-muted-foreground/30 mb-3" />
				<p class="text-[14px] text-muted-foreground">
					No monitors match
					<strong class="text-foreground font-medium"
						>{searchQuery === '' ? '' : `"${searchQuery}"`}</strong
					>
					{statusFilter !== 'all' ? `with status "${statusFilter}"` : ''}
				</p>
				<button
					class="mt-3 text-[13px] text-primary hover:underline"
					onclick={() => {
						searchQuery = '';
						statusFilter = 'all';
					}}
				>
					Clear filters
				</button>
			</div>

			<!-- ── Grid view ─────────────────────────────────────────────────────── -->
		{:else if viewMode === 'grid'}
			<div class="monitors-scroll-area">
				<div data-lenis-prevent class="org-group__grid">
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

			<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div class="monitors-scroll-area">
				<div
					data-lenis-prevent
					class="flex flex-col divide-y divide-border border border-border rounded-lg bg-white/2"
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
	/* ── Stat chips ───────────────────────────────────────────────────────── */
	.stat-chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 32px;
		padding: 0 12px;
		border-radius: 8px;
		font-size: 12.5px;
		font-weight: 500;
		border: 1px solid transparent;
		background: transparent;
		color: var(--color-muted-foreground);
		transition:
			background 140ms ease,
			color 140ms ease,
			border-color 140ms ease;
		cursor: pointer;
		white-space: nowrap;
	}

	.stat-chip:hover {
		background: rgba(255, 255, 255, 0.05);
		color: var(--color-foreground);
	}

	.stat-chip--active {
		background: rgba(255, 255, 255, 0.09);
		border-color: rgba(255, 255, 255, 0.08);
		color: var(--color-foreground);
		font-weight: 600;
	}

	.stat-count {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 18px;
		height: 18px;
		padding: 0 4px;
		border-radius: 9999px;
		font-size: 11px;
		font-family: var(--font-mono, monospace);
		background: rgba(255, 255, 255, 0.07);
		color: inherit;
		line-height: 1;
	}

	.stat-chip--active .stat-count {
		background: rgba(255, 255, 255, 0.14);
	}

	.stat-chip--up.stat-chip--active {
		border-color: rgba(78, 222, 163, 0.25);
		color: #4edea3;
	}
	.stat-chip--up.stat-chip--active .stat-count {
		background: rgba(78, 222, 163, 0.15);
	}

	.stat-chip--down.stat-chip--active {
		border-color: rgba(255, 107, 107, 0.25);
		color: #ff6b6b;
	}
	.stat-chip--down.stat-chip--active .stat-count {
		background: rgba(255, 107, 107, 0.15);
	}

	.stat-chip--paused.stat-chip--active {
		border-color: rgba(255, 255, 255, 0.15);
		color: #aaaabb;
	}

	/* ── Layout & Grids ───────────────────────────────────────────────────── */
	.monitors-scroll-area {
		width: 100%;
	}

	.org-group__grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
		gap: 16px;
	}

	@media (max-width: 640px) {
		.org-group__grid {
			grid-template-columns: 1fr;
		}
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
