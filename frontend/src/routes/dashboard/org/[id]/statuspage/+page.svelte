<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { flip } from 'svelte/animate';
	import {
		Globe,
		Plus,
		TriangleAlert,
		Search,
		LayoutGrid,
		List
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import StatusPageCard from '$lib/components/statuspage/StatusPageCard.svelte';
	import StatusPageListRow from '$lib/components/statuspage/StatusPageListRow.svelte';
	import StatusPageSkeleton from '$lib/components/skeletons/StatusPageSkeleton.svelte';
	import AddStatusPageModal from '$lib/components/statuspage/AddStatusPageModal.svelte';

	// ── Types ─────────────────────────────────────────────────────────────────
	type ViewMode = 'grid' | 'list';

	const orgId = $derived(($page.params.id as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let statusPages = $state<StatusPage[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');
	let searchQuery = $state('');

	// Add Modal
	let isAddModalOpen = $state(false);
	let monitors = $state<Monitor[]>([]);

	// ── Derived ───────────────────────────────────────────────────────────────
	let filtered = $derived(
		statusPages.filter((s) => {
			const q = searchQuery.toLowerCase().trim();
			return !q || s.Title.toLowerCase().includes(q) || s.Subdomain.toLowerCase().includes(q);
		})
	);

	// ── Data ──────────────────────────────────────────────────────────────────
	async function loadStatusPages(silent = false) {
		if (!orgId) return;
		if (!silent) isLoading = true;
		fetchError = '';
		try {
			const res = await StatusPageAPI.getAll(orgId);
			statusPages = res.data ?? [];
		} catch (err: any) {
			fetchError = err.message || 'Failed to fetch status pages.';
		} finally {
			isLoading = false;
		}
	}

	async function openAddModal() {
		if (!orgId) return;
		try {
			const res = await MonitorAPI.getAll(orgId);
			monitors = res.data ?? [];
		} catch (e) {
			monitors = [];
		}
		isAddModalOpen = true;
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

<div class="statuspages-page text-foreground antialiased">
	<!-- ── Page Header ──────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-headline-lg font-semibold leading-tight tracking-tight text-foreground mb-1">
				Status Pages
			</h1>
			<p class="text-[14px] leading-relaxed text-muted-foreground max-w-[52ch]">
				Manage and monitor status pages for this organization.
			</p>
		</div>

		<div class="flex items-center gap-3">
			<!-- View toggle -->
			{#if statusPages.length > 0}
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
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-foreground'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
					>
						<List class="size-4" />
					</button>
				</div>
			{/if}

			<Button
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium rounded-lg"
				onclick={openAddModal}
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
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-white/3"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium rounded-lg"
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
					class="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/50 pointer-events-none"
				/>
				<input
					type="search"
					placeholder="Search status pages…"
					bind:value={searchQuery}
					class="w-full h-9 pl-8 pr-3 bg-foreground/10 border border-border rounded-lg text-[13px] text-foreground placeholder:text-muted-foreground/35 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-colors"
				/>
			</div>
		{/if}

		<!-- Empty state (no status pages) -->
		{#if statusPages.length === 0}
			<div
				class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-white/3"
			>
				<div
					class="flex items-center justify-center size-14 rounded-lg bg-primary/10 border border-primary/20 mb-4 text-primary"
				>
					<Globe class="size-6" />
				</div>
				<h3 class="text-xl font-semibold text-foreground mb-1.5">No status pages yet</h3>
				<p class="text-[14px] text-muted-foreground max-w-[38ch] mb-6 leading-relaxed">
					Create a status page to share system status and uptime with your customers.
				</p>
				<Button
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium rounded-lg"
					onclick={openAddModal}
				>
					<Plus class="size-4" />
					Add status page
				</Button>
			</div>

			<!-- Search / filter no-results -->
		{:else if filtered.length === 0}
			<div
				class="flex flex-col items-center text-center py-16 border border-border rounded-lg bg-white/2"
			>
				<Search class="size-6 text-muted-foreground/30 mb-3" />
				<p class="text-[14px] text-muted-foreground">
					No status pages match
					<strong class="text-foreground font-medium"
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
			</div>
			<!-- ── Grid view ─────────────────────────────────────────────────────── -->
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
				class="flex flex-col divide-y divide-border border border-border rounded-lg overflow-hidden bg-white/2"
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

<AddStatusPageModal
	bind:isOpen={isAddModalOpen}
	{orgId}
	{monitors}
	onSuccess={() => loadStatusPages(true)}
/>

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

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
