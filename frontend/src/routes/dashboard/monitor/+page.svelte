<script lang="ts">
	import { onMount } from 'svelte';
	import { flip } from 'svelte/animate';
	import {
		Activity,
		Plus,
		TriangleAlert,
		Search,
		LayoutGrid,
		List,
		RefreshCw,
		Pause,
		ShieldCheck
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import MonitorCard from '$lib/components/monitor/MonitorCard.svelte';
	import MonitorListRow from '$lib/components/monitor/MonitorListRow.svelte';
	import MonitorSkeleton from '$lib/components/skeletons/MonitorSkeleton.svelte';
	import { wsClient, type MonitorUpdatePayload } from '$lib/utils/ws';

	// ── Types ─────────────────────────────────────────────────────────────────
	type StatusFilter = 'all' | 'up' | 'down' | 'paused';
	type ViewMode = 'grid' | 'list';

	// ── State ─────────────────────────────────────────────────────────────────
	let monitors = $state<Monitor[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let statusFilter = $state<StatusFilter>('all');
	let searchQuery = $state('');
	let isRefreshing = $state(false);

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
	async function loadMonitors(silent = false) {
		if (!silent) isLoading = true;
		else isRefreshing = true;
		fetchError = '';
		try {
			const res = await MonitorAPI.getAllByUser();
			monitors = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to fetch monitors.';
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	async function handleRefresh() {
		if (isRefreshing) return;
		await loadMonitors(true);
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

				// Reaktif update pada array `monitors`
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
	<title>Monitors • PulseWatch</title>
</svelte:head>

<div class="monitors-page text-on-surface antialiased">
	<!-- ── Page Header ──────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[28px] font-semibold leading-tight tracking-tight text-on-surface mb-1">
				Monitors
			</h1>
			<p class="text-[14px] leading-relaxed text-on-surface-variant max-w-[52ch]">
				All monitors across your organizations. Spot issues before they become incidents.
			</p>
		</div>

		<div class="flex items-center gap-2 flex-wrap">
			<!-- Refresh -->
			<button
				class="flex items-center justify-center size-9 rounded-lg border border-white/10 bg-white/3 text-on-surface-variant hover:text-on-surface hover:bg-white/6 transition-colors disabled:opacity-40"
				onclick={handleRefresh}
				disabled={isRefreshing}
				aria-label="Refresh monitors"
			>
				<RefreshCw class="size-4 {isRefreshing ? 'animate-spin' : ''}" />
			</button>

			<!-- Search + Filter Chips — only when we have data -->
			{#if monitors.length > 0}
				<!-- Search bar -->
				<div class="relative">
					<Search
						class="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-on-surface-variant/50 pointer-events-none"
					/>
					<input
						type="search"
						placeholder="Search monitors…"
						bind:value={searchQuery}
						class="w-[200px] h-9 pl-8 pr-3 bg-black/20 border border-white/10 rounded-lg text-[13px] text-on-surface placeholder:text-on-surface-variant/35 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-colors"
					/>
				</div>

				<!-- Status filter chips -->
				<div class="flex items-center gap-1.5" role="group" aria-label="Filter by status">
					<button
						class="stat-chip {statusFilter === 'all' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'all')}
						aria-pressed={statusFilter === 'all'}
					>
						<Activity class="size-3.5" />
						All <span class="stat-chip__count">{counts.total}</span>
					</button>

					<button
						class="stat-chip stat-chip--up {statusFilter === 'up' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'up')}
						aria-pressed={statusFilter === 'up'}
					>
						<ShieldCheck class="size-3.5" />
						Up <span class="stat-chip__count">{counts.up}</span>
					</button>

					<button
						class="stat-chip stat-chip--down {statusFilter === 'down' ? 'stat-chip--active' : ''}"
						onclick={() => (statusFilter = 'down')}
						aria-pressed={statusFilter === 'down'}
					>
						<TriangleAlert class="size-3.5" />
						Down <span class="stat-chip__count">{counts.down}</span>
					</button>

					<button
						class="stat-chip stat-chip--paused {statusFilter === 'paused'
							? 'stat-chip--active'
							: ''}"
						onclick={() => (statusFilter = 'paused')}
						aria-pressed={statusFilter === 'paused'}
					>
						<Pause class="size-3.5" />
						Paused <span class="stat-chip__count">{counts.paused}</span>
					</button>
				</div>

				<!-- Divider -->
				<div class="w-px h-5 bg-white/10"></div>

				<!-- View toggle -->
				<div
					class="flex items-center bg-black/20 border border-white/10 rounded-lg p-0.5"
					role="group"
					aria-label="View mode"
				>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'grid'
							? 'bg-white/10 text-on-surface shadow-sm'
							: 'text-on-surface-variant hover:text-on-surface hover:bg-white/5'}"
						onclick={() => setViewMode('grid')}
						aria-label="Grid view"
						aria-pressed={viewMode === 'grid'}
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-on-surface shadow-sm'
							: 'text-on-surface-variant hover:text-on-surface hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
						aria-pressed={viewMode === 'list'}
					>
						<List class="size-4" />
					</button>
				</div>
			{/if}

			<Button
				class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium shadow-sm"
				href="/dashboard/organization"
			>
				<Plus class="size-4" />
				Add monitor
			</Button>
		</div>
	</header>

	<!-- ── Loading Skeleton ──────────────────────────────────────────────────── -->
	{#if isLoading}
		<MonitorSkeleton {viewMode} />

		<!-- ── Error state ────────────────────────────────────────────────────────── -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-white/8 rounded-xl bg-white/3"
		>
			<TriangleAlert class="size-8 text-error" />
			<p class="text-sm text-on-surface-variant">{fetchError}</p>
			<Button
				class="bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white hover:bg-[#2A2A2B] active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium shadow-sm"
				onclick={() => loadMonitors()}
			>
				Try again
			</Button>
		</div>

		<!-- ── Data ──────────────────────────────────────────────────────────────── -->
	{:else}
		<!-- Empty state (no monitors at all) -->
		{#if monitors.length === 0}
			<div
				class="flex flex-col items-center text-center py-24 px-4 border border-white/8 rounded-xl bg-white/3"
			>
				<div
					class="flex items-center justify-center size-14 rounded-xl bg-primary/10 border border-primary/20 mb-4 text-primary"
				>
					<Activity class="size-6" />
				</div>
				<h3 class="text-xl font-semibold text-on-surface mb-1.5">No monitors yet</h3>
				<p class="text-[14px] text-on-surface-variant max-w-[38ch] mb-6 leading-relaxed">
					Create a monitor inside an organization to start tracking uptime and response times.
				</p>
				<Button
					class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium shadow-sm"
					href="/dashboard/organization"
				>
					<Plus class="size-4" />
					Go to organizations
				</Button>
			</div>

			<!-- Search / filter no-results -->
		{:else if filtered.length === 0}
			<div
				class="flex flex-col items-center text-center py-16 border border-white/6 rounded-xl bg-white/2"
			>
				<Search class="size-6 text-on-surface-variant/30 mb-3" />
				<p class="text-[14px] text-on-surface-variant">
					No monitors match
					<strong class="text-on-surface font-medium"
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
				<div data-lenis-prevent class="grid grid-cols-[repeat(auto-fill,minmax(450px,1fr))] gap-6">
					{#each filtered as monitor (monitor.ID)}
						<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
							<MonitorCard {monitor} />
						</div>
					{/each}
				</div>
			</div>

			<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div class="monitors-scroll-area">
				<div
					class="flex flex-col divide-y divide-white/6 border border-white/8 rounded-xl overflow-hidden bg-white/2"
				>
					<!-- Header row -->
					<div class="list-row list-row--header" aria-hidden="true">
						<span>Monitor</span>
						<span>Status</span>
						<span>Response</span>
						<span>Uptime 30d</span>
						<span>Interval</span>
						<span>Last check</span>
						<span></span>
					</div>

					{#each filtered as monitor (monitor.ID)}
						<MonitorListRow {monitor} />
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	/* ── Stat chips ───────────────────────────────────────────────────────── */
	.stat-chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 3px 10px;
		border-radius: 6px;
		font-size: 12.5px;
		font-weight: 500;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.03);
		color: var(--color-on-surface-variant);
		transition:
			background 140ms ease,
			border-color 140ms ease,
			color 140ms ease;
		cursor: pointer;
	}

	.stat-chip:hover:not(.stat-chip--active) {
		background: rgba(255, 255, 255, 0.07);
		color: var(--color-on-surface);
	}

	.stat-chip--active {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.18);
		color: var(--color-on-surface);
	}

	.stat-chip--down.stat-chip--active {
		background: rgba(255, 107, 107, 0.12);
		border-color: rgba(255, 107, 107, 0.3);
		color: #ff6b6b;
	}

	.stat-chip--up.stat-chip--active {
		background: rgba(78, 222, 163, 0.1);
		border-color: rgba(78, 222, 163, 0.28);
		color: #4edea3;
	}

	.stat-chip--paused.stat-chip--active {
		background: rgba(136, 136, 153, 0.1);
		border-color: rgba(136, 136, 153, 0.25);
		color: #888899;
	}

	.stat-chip__count {
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		font-size: 12px;
		opacity: 0.75;
	}

	/* ── List view ────────────────────────────────────────────────────────── */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 90px 76px 90px 64px 96px 56px;
		align-items: center;
		gap: 12px;
		padding: 10px 16px;
		font-size: 13px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: rgba(255, 255, 255, 0.025);
	}

	.list-row--header {
		padding: 7px 16px;
		font-size: 11px;
		font-weight: 500;
		color: var(--color-on-surface-variant);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		background: rgba(255, 255, 255, 0.02);
		opacity: 0.65;
	}

	.list-row--header:hover {
		background: rgba(255, 255, 255, 0.02);
	}

	/* ── Monitors page layout ────────────────────────────────────────────── */
	.monitors-page {
		/* 48px topbar (h-12) + 24px padding-top + 24px padding-bottom dari p-6 layout */
		height: calc(100vh - 96px);
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* ── Monitors scroll area ─────────────────────────────────────────────── */
	.monitors-scroll-area {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 4px;
		scrollbar-width: thin;
		scrollbar-color: rgba(255, 255, 255, 0.12) transparent;
	}

	.monitors-scroll-area::-webkit-scrollbar {
		width: 5px;
	}

	.monitors-scroll-area::-webkit-scrollbar-track {
		background: transparent;
	}

	.monitors-scroll-area::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.12);
		border-radius: 999px;
	}

	.monitors-scroll-area::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.2);
	}

	/* ── Reduced motion ───────────────────────────────────────────────────── */
	@media (prefers-reduced-motion: reduce) {
		.stat-chip,
		.list-row {
			transition: none;
		}
	}
</style>
