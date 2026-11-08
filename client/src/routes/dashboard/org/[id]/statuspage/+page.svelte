<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Globe, Plus, TriangleAlert, Search } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import StatusPageItem from '$lib/components/statuspage/StatusPageItem.svelte';
	import { DataSkeleton } from '$lib/components/common';
	import StatusPageModal from '$lib/components/statuspage/StatusPageModal.svelte';
	import MaintenanceModal from '$lib/components/statuspage/MaintenanceModal.svelte';
	import ComponentGroupModal from '$lib/components/statuspage/ComponentGroupModal.svelte';
	import EmbedBadgeModal from '$lib/components/statuspage/EmbedBadgeModal.svelte';
	import ConfirmDialog from '$lib/components/common/ConfirmDialog.svelte';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	const orgId = $derived(($page.params.id as string) || '');

	// ── State ─────────────────────────────────────────────────────────────────
	let statusPages = $state<StatusPage[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');

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
	let isDeleteConfirmOpen = $state(false);
	let deletingStatusPage = $state<StatusPage | null>(null);

	// ── Derived ───────────────────────────────────────────────────────────────
	// Sync with layout top bar
	$effect(() => {
		void $page.url.pathname;
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
		} catch (err: unknown) {
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
		} catch {
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
		try {
			await StatusPageAPI.delete(orgId, deletingStatusPage.ID);
			toast.success('Status page deleted successfully');
			isDeleteConfirmOpen = false;
			deletingStatusPage = null;
			await loadStatusPages(true);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete status page'));
		}
	}

	onMount(() => {
		loadStatusPages();
	});
</script>

<svelte:head>
	<title>Status Pages | Netra Uptime</title>
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
	</header>

	<!-- ── Loading ──────────────────────────────────────────────────────────── -->
	{#if isLoading}
		<DataSkeleton variant="list" count={6} />
		<!-- ── Error state ──────────────────────────────────────────────────────── -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-card"
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
				class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-card"
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
				class="flex flex-col items-center text-center py-16 border border-border rounded-lg bg-card"
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

			<!-- ── List view ─────────────────────────────────────────────────────── -->
		{:else}
			<div class="border border-border rounded-md overflow-hidden">
				<div
					class="hidden md:grid md:grid-cols-[minmax(160px,1.5fr)_minmax(180px,2fr)_1.2fr_90px_100px_80px] gap-4 px-4 py-2 border-b border-border bg-table-header text-[11px] font-semibold tracking-wider uppercase text-muted-foreground items-center"
					aria-hidden="true"
				>
					<div>Status Page</div>
					<div>Description</div>
					<div>Subdomain</div>
					<div>Visibility</div>
					<div>Monitors</div>
					<div class="text-right">Actions</div>
				</div>

				<div class="divide-y divide-border/60 bg-card">
					{#each filtered as sp (sp.ID)}
						<StatusPageItem
							{sp}
							onEdit={openEditModal}
							onDelete={openDeleteConfirm}
							onMaintenance={openMaintenanceModal}
							onGroups={openGroupModal}
							onEmbedBadge={openEmbedModal}
						/>
					{/each}
				</div>
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
<ConfirmDialog
	bind:open={isDeleteConfirmOpen}
	title="Delete Status Page"
	description="Are you sure you want to delete '{deletingStatusPage?.Title}'? The status page will be permanently removed."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={handleDelete}
/>
