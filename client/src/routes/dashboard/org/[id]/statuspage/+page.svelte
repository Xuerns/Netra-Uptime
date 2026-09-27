<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { flip } from 'svelte/animate';
	import { Globe, Plus, TriangleAlert, Search, LayoutGrid, List, Trash2 } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import StatusPageItem from '$lib/components/statuspage/StatusPageItem.svelte';
	import StatusPageSkeleton from '$lib/components/skeletons/StatusPageSkeleton.svelte';
	import StatusPageModal from '$lib/components/statuspage/StatusPageModal.svelte';
	import MaintenanceModal from '$lib/components/statuspage/MaintenanceModal.svelte';
	import ComponentGroupModal from '$lib/components/statuspage/ComponentGroupModal.svelte';
	import EmbedBadgeModal from '$lib/components/statuspage/EmbedBadgeModal.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	// ── Types ─────────────────────────────────────────────────────────────────
	type ViewMode = 'grid' | 'list';

	const orgId = $derived(($page.params.id as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let statusPages = $state<StatusPage[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');
	let viewMode = $state<ViewMode>('grid');

	// Add Modal
	let isAddModalOpen = $state(false);
	let monitors = $state<Monitor[]>([]);

	// Edit Modal
	let isEditModalOpen = $state(false);
	let editingStatusPage = $state<StatusPage | null>(null);

	// Maintenance Modal
	let isMaintenanceModalOpen = $state(false);
	let maintenanceStatusPage = $state<StatusPage | null>(null);

	// Component Groups Modal
	let isGroupModalOpen = $state(false);
	let groupStatusPage = $state<StatusPage | null>(null);

	// Embed Badge Modal
	let isEmbedModalOpen = $state(false);
	let embedStatusPage = $state<StatusPage | null>(null);

	// Delete Confirmation
	let isDeletingId = $state<string | null>(null);
	let isDeleteConfirmOpen = $state(false);
	let deletingStatusPage = $state<StatusPage | null>(null);

	// ── Derived ───────────────────────────────────────────────────────────────
	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			showSearch: statusPages.length > 0,
			searchPlaceholder: 'Search status pages...',
			actionType: 'button',
			actionLabel: 'Add status page',
			onAction: openAddModal,
			badge: {
				text: `${statusPages.length} Status page${statusPages.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	let filtered = $derived(
		statusPages.filter((s) => {
			const q = headerStore.searchQuery.toLowerCase().trim();
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
			if (editingStatusPage) {
				editingStatusPage = statusPages.find((s) => s.ID === editingStatusPage?.ID) || null;
			}
			if (maintenanceStatusPage) {
				maintenanceStatusPage = statusPages.find((s) => s.ID === maintenanceStatusPage?.ID) || null;
			}
			if (groupStatusPage) {
				groupStatusPage = statusPages.find((s) => s.ID === groupStatusPage?.ID) || null;
			}
		} catch (err: any) {
			fetchError = parseApiError(err, 'Failed to fetch status pages.');
		} finally {
			isLoading = false;
		}
	}

	async function loadMonitors() {
		if (!orgId) return;
		try {
			const res = await MonitorAPI.getAll(orgId);
			monitors = res.data ?? [];
		} catch (e) {
			monitors = [];
		}
	}

	async function openAddModal() {
		await loadMonitors();
		isAddModalOpen = true;
	}

	async function openEditModal(sp: StatusPage) {
		await loadMonitors();
		editingStatusPage = sp;
		isEditModalOpen = true;
	}

	async function openMaintenanceModal(sp: StatusPage) {
		await loadMonitors();
		maintenanceStatusPage = sp;
		isMaintenanceModalOpen = true;
	}

	async function openGroupModal(sp: StatusPage) {
		await loadMonitors();
		groupStatusPage = sp;
		isGroupModalOpen = true;
	}

	function openEmbedModal(sp: StatusPage) {
		embedStatusPage = sp;
		isEmbedModalOpen = true;
	}

	function openDeleteConfirm(sp: StatusPage) {
		deletingStatusPage = sp;
		isDeleteConfirmOpen = true;
	}

	async function handleDelete() {
		if (!deletingStatusPage || !orgId) return;
		isDeletingId = deletingStatusPage.ID;
		try {
			await StatusPageAPI.delete(orgId, deletingStatusPage.ID);
			toast.success('Status page deleted successfully');
			isDeleteConfirmOpen = false;
			deletingStatusPage = null;
			await loadStatusPages(true);
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete status page'));
		} finally {
			isDeletingId = null;
		}
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
	<title>Status Pages | PulseWatch</title>
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
						>{headerStore.searchQuery === '' ? '' : `"${headerStore.searchQuery}"`}</strong
					>
				</p>
				<button
					class="mt-3 text-[13px] text-primary hover:underline"
					onclick={() => {
						headerStore.searchQuery = '';
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
						<StatusPageItem
							{sp}
							viewMode="grid"
							onEdit={openEditModal}
							onDelete={openDeleteConfirm}
							onMaintenance={openMaintenanceModal}
							onGroups={openGroupModal}
							onEmbedBadge={openEmbedModal}
						/>
					</div>
				{/each}
			</div>

			<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div class="flex flex-col divide-y divide-border border border-border rounded-lg bg-white/2">
				<div class="list-row list-row--header" aria-hidden="true">
					<span>Status Page</span>
					<span>Subdomain</span>
					<span>Visibility</span>
					<span>Monitors</span>
					<span>Actions</span>
				</div>

				{#each filtered as sp (sp.ID)}
					<StatusPageItem
						{sp}
						viewMode="list"
						onEdit={openEditModal}
						onDelete={openDeleteConfirm}
						onMaintenance={openMaintenanceModal}
						onGroups={openGroupModal}
						onEmbedBadge={openEmbedModal}
					/>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<StatusPageModal
	bind:isOpen={isAddModalOpen}
	{orgId}
	{monitors}
	onSuccess={() => loadStatusPages(true)}
/>

<StatusPageModal
	bind:isOpen={isEditModalOpen}
	{orgId}
	{monitors}
	statusPage={editingStatusPage}
	onSuccess={() => loadStatusPages(true)}
/>

<MaintenanceModal
	bind:isOpen={isMaintenanceModalOpen}
	{orgId}
	statusPage={maintenanceStatusPage}
	{monitors}
/>

<ComponentGroupModal
	bind:isOpen={isGroupModalOpen}
	{orgId}
	statusPage={groupStatusPage}
	{monitors}
	onUpdated={() => loadStatusPages(true)}
/>

<EmbedBadgeModal bind:isOpen={isEmbedModalOpen} statusPage={embedStatusPage} />

<!-- ── Delete Confirmation Dialog ──────────────────────────────────────────── -->
{#if isDeleteConfirmOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm w-max-[400px]"
		onclick={(e) => {
			if (e.target === e.currentTarget) isDeleteConfirmOpen = false;
		}}
	>
		<div
			class="bg-card border border-border rounded-2xl shadow-xl p-6 w-full max-w-sm mx-4 flex flex-col gap-4"
		>
			<div class="flex items-center gap-3">
				<div>
					<h2 class="text-base font-semibold text-foreground">Delete Status Page</h2>
					<p class="text-[13px] text-muted-foreground">This action cannot be undone.</p>
				</div>
			</div>

			<p class="text-[13px] text-muted-foreground leading-relaxed">
				Are you sure you want to delete <strong class="text-foreground font-medium"
					>"{deletingStatusPage?.Title}"</strong
				>? The status page will be permanently removed.
			</p>

			<div class="flex items-center gap-2 justify-end">
				<Button
					type="button"
					variant="ghost"
					class="text-foreground h-9 px-4 hover:text-foreground hover:bg-accent"
					onclick={() => (isDeleteConfirmOpen = false)}
					disabled={isDeletingId !== null}
				>
					Cancel
				</Button>
				<Button
					type="button"
					class="h-9 px-4 bg-destructive/10 border border-destructive/20 text-destructive hover:bg-destructive hover:border-destructive hover:text-white active:scale-[0.98] transition-all duration-200 font-medium"
					onclick={handleDelete}
					disabled={isDeletingId !== null}
				>
					{isDeletingId !== null ? 'Deleting...' : 'Delete'}
				</Button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ── List view ────────────────────────────────────────────────────────── */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 1fr 100px 100px 74px;
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
