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
		Pause,
		ShieldCheck,
		ChevronDown,
		ChevronRight
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
	let openDropdownId = $state<string | null>(null);
	let collapsedOrgs = $state<Set<string>>(new Set());

	// ── Derived ───────────────────────────────────────────────────────────────
	type OrgGroup = { name: string; slug: string; monitors: Monitor[] };

	let grouped = $derived.by<OrgGroup[]>(() => {
		const map = new Map<string, { slug: string; monitors: Monitor[] }>();
		for (const m of filtered) {
			const orgName = m.Organization?.Name ?? 'Unknown';
			const orgSlug = m.Organization?.Slug ?? '';
			const entry = map.get(orgName) ?? { slug: orgSlug, monitors: [] };
			entry.monitors.push(m);
			map.set(orgName, entry);
		}
		return Array.from(map.entries()).map(([name, data]) => ({
			name,
			slug: data.slug,
			monitors: data.monitors
		}));
	});

	function toggleOrg(name: string) {
		collapsedOrgs = new Set(
			collapsedOrgs.has(name)
				? [...collapsedOrgs].filter((n) => n !== name)
				: [...collapsedOrgs, name]
		);
	}

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
		try {
			await MonitorAPI.update(m.OrganizationID, m.ID, {
				is_active: m.Status === 'paused'
			});
			await loadMonitors(true);
		} catch (err: any) {
			alert(err.message || 'Failed to toggle monitor status');
		}
	}

	async function handleDeleteMonitor(id: string) {
		const m = monitors.find((item) => item.ID === id);
		if (!m) return;
		try {
			await MonitorAPI.delete(m.OrganizationID, m.ID);
			await loadMonitors(true);
		} catch (err: any) {
			alert(err.message || 'Failed to delete monitor');
		}
	}

	async function loadMonitors(silent = false) {
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const res = await MonitorAPI.getAllByUser();
			monitors = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to fetch monitors.';
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

<div class="monitors-page text-foreground antialiased">
	<!-- ── Page Header ──────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[28px] font-semibold leading-tight tracking-tight text-foreground mb-1">
				Monitors
			</h1>
			<p class="text-[14px] leading-relaxed text-muted-foreground max-w-[52ch]">
				All monitors across your organizations. Spot issues before they become incidents.
			</p>
		</div>

		<div class="flex items-center gap-2 flex-wrap">
			<!-- Search + Filter Chips — only when we have data -->
			{#if monitors.length > 0}
				<!-- Search bar -->
				<div class="relative">
					<Search
						class="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/50 pointer-events-none"
					/>
					<input
						type="search"
						placeholder="Search monitors…"
						bind:value={searchQuery}
						class="w-[200px] h-9 pl-8 pr-3 bg-foreground/10 border border-border rounded-lg text-[13px] text-foreground placeholder:text-muted-foreground/35 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-colors"
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
					class="flex items-center bg-foreground/10 border border-border rounded-lg p-0.5"
					role="group"
					aria-label="View mode"
				>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'grid'
							? 'bg-white/10 text-foreground shadow-sm'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('grid')}
						aria-label="Grid view"
						aria-pressed={viewMode === 'grid'}
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-foreground shadow-sm'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
						aria-pressed={viewMode === 'list'}
					>
						<List class="size-4" />
					</button>
				</div>
			{/if}

			<Button
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium shadow-sm"
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
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-xl bg-white/3"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium shadow-sm"
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
				class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-xl bg-white/3"
			>
				<div
					class="flex items-center justify-center size-14 rounded-xl bg-primary/10 border border-primary/20 mb-4 text-primary"
				>
					<Activity class="size-6" />
				</div>
				<h3 class="text-xl font-semibold text-foreground mb-1.5">No monitors yet</h3>
				<p class="text-[14px] text-muted-foreground max-w-[38ch] mb-6 leading-relaxed">
					Create a monitor inside an organization to start tracking uptime and response times.
				</p>
				<Button
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium shadow-sm"
					href="/dashboard/organization"
				>
					<Plus class="size-4" />
					Go to organizations
				</Button>
			</div>

			<!-- Search / filter no-results -->
		{:else if filtered.length === 0}
			<div
				class="flex flex-col items-center text-center py-16 border border-border rounded-xl bg-white/2"
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
				<div data-lenis-prevent class="flex flex-col gap-1">
					{#each grouped as orgGroup, i (orgGroup.name)}
						<section class={`org-group ${i > 0 ? 'mt-1' : ''}`}>
							<button class="org-group__header" onclick={() => toggleOrg(orgGroup.name)}>
								<div class="flex items-center gap-2">
									{#if collapsedOrgs.has(orgGroup.name)}
										<ChevronRight class="size-3.5 text-muted-foreground/60" />
									{:else}
										<ChevronDown class="size-3.5 text-muted-foreground/60" />
									{/if}
									<span class="org-group__name">{orgGroup.name}</span>
								</div>
								<span class="org-group__count">{orgGroup.monitors.length}</span>
							</button>
							{#if !collapsedOrgs.has(orgGroup.name)}
								<div class="org-group__grid">
									{#each orgGroup.monitors as monitor (monitor.ID)}
										<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
											<MonitorCard {monitor} />
										</div>
									{/each}
								</div>
							{/if}
						</section>
					{/each}
				</div>
			</div>

			<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div class="monitors-scroll-area">
				<div data-lenis-prevent class="flex flex-col gap-1">
					{#each grouped as orgGroup, i (orgGroup.name)}
						<section class={`org-group ${i > 0 ? 'mt-1' : ''}`}>
							<button class="org-group__header" onclick={() => toggleOrg(orgGroup.name)}>
								<div class="flex items-center gap-2">
									{#if collapsedOrgs.has(orgGroup.name)}
										<ChevronRight class="size-3.5 text-muted-foreground/60" />
									{:else}
										<ChevronDown class="size-3.5 text-muted-foreground/60" />
									{/if}
									<span class="org-group__name">{orgGroup.name}</span>
								</div>
								<span class="org-group__count">{orgGroup.monitors.length}</span>
							</button>
							{#if !collapsedOrgs.has(orgGroup.name)}
								<div
									class="flex flex-col divide-y divide-border border border-border rounded-xl overflow-hidden bg-white/2"
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

									{#each orgGroup.monitors as monitor (monitor.ID)}
										<MonitorListRow {monitor} />
									{/each}
								</div>
							{/if}
						</section>
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
			border-color 140ms ease,
			color 140ms ease;
		cursor: pointer;
	}

	.stat-chip:hover:not(.stat-chip--active) {
		background: var(--color-accent);
		color: var(--color-foreground);
	}

	.stat-chip--active {
		background: var(--color-accent);
		border-color: var(--color-border);
		color: var(--color-foreground);
		font-weight: 700;
	}

	.stat-chip__count {
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		font-size: 11.5px;
		padding: 1px 6px;
		border-radius: 999px;
		background: var(--color-muted);
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
		background: var(--color-accent);
	}

	.list-row--header {
		padding: 7px 16px;
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

	/* ── Monitors page layout ────────────────────────────────────────────── */
	.monitors-page {
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
		scrollbar-color: var(--color-muted-foreground) transparent;
	}

	.monitors-scroll-area::-webkit-scrollbar {
		width: 5px;
	}

	.monitors-scroll-area::-webkit-scrollbar-track {
		background: transparent;
	}

	.monitors-scroll-area::-webkit-scrollbar-thumb {
		background: var(--color-muted-foreground);
		opacity: 0.3;
		border-radius: 999px;
	}

	.monitors-scroll-area::-webkit-scrollbar-thumb:hover {
		opacity: 0.5;
	}

	/* ── Org groups ──────────────────────────────────────────────────────── */
	.org-group__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 8px 12px;
		margin-bottom: 10px;
		border-radius: 8px;
		background: var(--color-secondary);
		border: 1px solid var(--color-border);
		cursor: pointer;
		transition: background 140ms ease;
	}

	.org-group__header:hover {
		background: var(--color-accent);
	}

	.org-group__name {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-foreground);
	}

	.org-group__count {
		font-size: 11px;
		font-weight: 600;
		color: var(--color-muted-foreground);
		padding: 1px 7px;
		border-radius: 999px;
		background: var(--color-muted);
		border: 1px solid var(--color-border);
	}

	.org-group__grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(450px, 1fr));
		gap: 6px;
	}

	/* ── Reduced motion ───────────────────────────────────────────────────── */
	@media (prefers-reduced-motion: reduce) {
		.stat-chip,
		.list-row {
			transition: none;
		}
	}
</style>
