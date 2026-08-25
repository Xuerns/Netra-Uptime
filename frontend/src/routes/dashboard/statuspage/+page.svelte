<script lang="ts">
	import { onMount } from 'svelte';
	import { flip } from 'svelte/animate';
	import {
		Globe,
		Plus,
		TriangleAlert,
		Search,
		LayoutGrid,
		List,
		RefreshCw,
		ExternalLink
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import StatusPageCard from '$lib/components/statuspage/StatusPageCard.svelte';
	import StatusPageListRow from '$lib/components/statuspage/StatusPageListRow.svelte';
	import StatusPageSkeleton from '$lib/components/skeletons/StatusPageSkeleton.svelte';

	// ── Types ─────────────────────────────────────────────────────────────────
	type ViewMode = 'grid' | 'list';

	// ── State ─────────────────────────────────────────────────────────────────
	let statusPages = $state<StatusPage[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let searchQuery = $state('');
	let isRefreshing = $state(false);

	// ── Derived ───────────────────────────────────────────────────────────────
	let counts = $derived({
		total: statusPages.length,
		public: statusPages.filter((s) => s.IsPublic).length,
		private: statusPages.filter((s) => !s.IsPublic).length
	});

	let filtered = $derived(
		statusPages.filter((s) => {
			const q = searchQuery.toLowerCase().trim();
			return !q || s.Title.toLowerCase().includes(q) || s.Subdomain.toLowerCase().includes(q);
		})
	);

	// ── Data ──────────────────────────────────────────────────────────────────
	async function loadStatusPages(silent = false) {
		if (!silent) isLoading = true;
		else isRefreshing = true;
		fetchError = '';
		try {
			const res = await StatusPageAPI.getAllByUser();
			statusPages = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to fetch status pages.';
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}

	async function handleRefresh() {
		if (isRefreshing) return;
		await loadStatusPages(true);
	}

	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		localStorage.setItem('statusPageViewMode', mode);
	}

	onMount(() => {
		loadStatusPages();
		const saved = localStorage.getItem('statusPageViewMode');
		if (saved === 'grid' || saved === 'list') viewMode = saved;
	});
</script>

<svelte:head>
	<title>Status Pages • PulseWatch</title>
</svelte:head>

<div class="statuspages-page text-on-surface antialiased">
	<!-- ── Page Header ──────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-headline-lg font-semibold leading-tight tracking-tight text-on-surface mb-1">
				Status Pages
			</h1>
			<p class="text-[14px] leading-relaxed text-on-surface-variant max-w-[52ch]">
				Manage and monitor all status pages across your organizations.
			</p>
		</div>

		<div class="flex items-center gap-3">
			<!-- Refresh -->
			<button
				class="flex items-center justify-center size-9 rounded-lg border border-white/10 bg-white/3 text-on-surface-variant hover:text-on-surface hover:bg-white/6 transition-colors disabled:opacity-40"
				onclick={handleRefresh}
				disabled={isRefreshing}
				aria-label="Refresh status pages"
			>
				<RefreshCw class="size-4 {isRefreshing ? 'animate-spin' : ''}" />
			</button>

			<!-- View toggle -->
			{#if statusPages.length > 0}
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
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-on-surface shadow-sm'
							: 'text-on-surface-variant hover:text-on-surface hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
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
				Add status page
			</Button>
		</div>
	</header>

	<!-- ── Loading ──────────────────────────────────────────────────────────── -->
	{#if isLoading}
		<StatusPageSkeleton {viewMode} />
	<!-- ── Error state ──────────────────────────────────────────────────────── -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-white/8 rounded-xl bg-white/3"
		>
			<TriangleAlert class="size-8 text-error" />
			<p class="text-sm text-on-surface-variant">{fetchError}</p>
			<Button
				class="bg-[#1A1A1B] border border-[#2A2A2B] text-slate-300 hover:text-white hover:bg-[#2A2A2B] active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium shadow-sm"
				onclick={() => loadStatusPages()}
			>
				Try again
			</Button>
		</div>

	<!-- ── Data ────────────────────────────────────────────────────────────── -->
	{:else}
		{#if statusPages.length > 0}
			<!-- Search bar -->
			<div class="relative w-full max-w-80 mb-6">
				<Search
					class="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-on-surface-variant/50 pointer-events-none"
				/>
				<input
					type="search"
					placeholder="Search status pages…"
					bind:value={searchQuery}
					class="w-full h-9 pl-8 pr-3 bg-black/20 border border-white/10 rounded-lg text-[13px] text-on-surface placeholder:text-on-surface-variant/35 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-colors"
				/>
			</div>
		{/if}

		<!-- Empty state (no status pages) -->
		{#if statusPages.length === 0}
			<div
				class="flex flex-col items-center text-center py-24 px-4 border border-white/8 rounded-xl bg-white/3"
			>
				<div
					class="flex items-center justify-center size-14 rounded-xl bg-primary/10 border border-primary/20 mb-4 text-primary"
				>
					<Globe class="size-6" />
				</div>
				<h3 class="text-xl font-semibold text-on-surface mb-1.5">No status pages yet</h3>
				<p class="text-[14px] text-on-surface-variant max-w-[38ch] mb-6 leading-relaxed">
					Create a status page inside an organization to share system status with your customers.
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
					No status pages match
					<strong class="text-on-surface font-medium"
						>{searchQuery === '' ? '' : `"${searchQuery}"`}</strong
					>
				</p>
				<button
					class="mt-3 text-[13px] text-primary hover:underline"
					onclick={() => {
						searchQuery = '';
					}}
				>
					Clear filters
				</button>
			</div>		<!-- ── Grid view ─────────────────────────────────────────────────────── -->
		{:else if viewMode === 'grid'}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(450px,1fr))] gap-6">
				{#each filtered as sp (sp.ID)}
					<div animate:flip={{ duration: 250 }} class="flex w-full h-full">
						<StatusPageCard {sp} />
					</div>
				{/each}
			</div>

		<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div
				class="flex flex-col divide-y divide-white/6 border border-white/8 rounded-xl overflow-hidden bg-white/2"
			>
				<div class="list-row list-row--header" aria-hidden="true">
					<span>Status Page</span>
					<span>Subdomain</span>
					<span>Visibility</span>
					<span>Monitors</span>
					<span>Actions</span>
				</div>

				{#each filtered as sp (sp.ID)}
					<StatusPageListRow {sp} />
				{/each}
			</div>
		{/if}
	{/if}
</div>

<style>
	/* ── List view ────────────────────────────────────────────────────────── */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 1fr 100px 100px 60px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: rgba(255, 255, 255, 0.025);
	}

	.list-row--header {
		padding: 10px 20px;
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

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
