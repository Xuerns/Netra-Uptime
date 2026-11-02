<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { Building2, Plus, TriangleAlert, LayoutGrid, List } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { OrgAPI } from '$lib/api/organization';
	import { DataSkeleton } from '$lib/components/common';
	import { activeOrgId, organizations } from '$lib/stores/organization';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { CreateOrgModal, OrgCard, OrgListItem } from '$lib/components/org';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// ── Types ────────────────────────────────────────────────────────────────────
	interface Organization {
		ID: string;
		Name: string;
		Slug: string;
		CreatedAt: string;
		Monitors?: Array<{ ID?: string; Status: string }>;
		OrganizationMembers?: Array<unknown>;
	}

	// ── State ────────────────────────────────────────────────────────────────────
	let orgs = $state<Organization[]>([]);
	let isLoading = $state(false);
	let fetchError = $state('');
	let viewMode = $state<'grid' | 'list'>('grid');
	let dialogOpen = $state(false);

	// Sync with layout header bar
	$effect(() => {
		headerStore.set({
			title: 'Organizations',
			showSearch: true,
			searchPlaceholder: 'Search organizations...',
			actionType: 'button',
			actionLabel: 'New organization',
			onAction: () => openDialog()
		});
	});

	let filteredOrgs = $derived(
		orgs.filter((org) => {
			const q = headerStore.searchQuery.toLowerCase().trim();
			if (!q) return true;
			return (
				(org.Name ?? '').toLowerCase().includes(q) || (org.Slug ?? '').toLowerCase().includes(q)
			);
		})
	);

	$effect(() => {
		if (data?.orgs) {
			const list = data.orgs as Organization[];
			untrack(() => {
				orgs = list;
			});
			organizations.set(list.map((o) => ({ id: o.ID, name: o.Name, slug: o.Slug })));
		}
	});

	async function loadOrgs() {
		isLoading = true;
		fetchError = '';
		try {
			const res = await OrgAPI.getAll();
			orgs = res.data ?? [];

			// Update global store
			organizations.set(orgs.map((o) => ({ id: o.ID, name: o.Name, slug: o.Slug })));
		} catch (err: unknown) {
			fetchError = parseApiError(err, 'Failed to load organizations.');
		} finally {
			isLoading = false;
		}
	}

	function openDialog() {
		dialogOpen = true;
	}

	async function handleOrgCreated(newOrg: { ID?: string } | null | undefined) {
		await loadOrgs();
		if (newOrg?.ID) {
			selectOrg(newOrg.ID);
		}
	}

	function selectOrg(orgId: string) {
		activeOrgId.set(orgId);
		localStorage.setItem('netra_last_org', orgId);
		goto(`/dashboard/org/${orgId}/monitors`);
	}

	function setViewMode(mode: 'grid' | 'list') {
		viewMode = mode;
		localStorage.setItem('orgViewMode', mode);
	}

	onMount(() => {
		const savedMode = localStorage.getItem('orgViewMode');
		if (savedMode === 'grid' || savedMode === 'list') {
			viewMode = savedMode;
		}
	});
</script>

<svelte:head>
	<title>Select Organization | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased">
	<!-- Subheader: Counter/Description & View Mode Toggle -->
	<div class="flex flex-wrap items-center justify-between gap-4 mb-6">
		<p class="text-sm text-muted-foreground">
			{#if orgs.length > 0}
				{#if headerStore.searchQuery.trim()}
					Showing {filteredOrgs.length} of {orgs.length} organization{orgs.length !== 1 ? 's' : ''}
				{:else}
					Select an organization ({orgs.length}) to manage its monitors, incidents, and status
					pages.
				{/if}
			{:else}
				Select an organization to manage its monitors, incidents, and status pages.
			{/if}
		</p>
		{#if orgs.length > 0}
			<div class="flex items-center bg-foreground/10 border border-border rounded-lg p-0.5">
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

	<!-- Loading state -->
	{#if isLoading}
		<DataSkeleton variant="cards" {viewMode} count={6} />

		<!-- Fetch error state -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-card"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium"
				onclick={loadOrgs}
			>
				Try again
			</Button>
		</div>

		<!-- Empty state: No orgs in account -->
	{:else if orgs.length === 0}
		<div
			class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-card"
		>
			<div
				class="flex items-center justify-center size-14 rounded-lg bg-primary/10 border border-primary/20 mb-4 text-primary"
			>
				<Building2 class="size-6" />
			</div>
			<h3 class="text-xl font-semibold text-foreground mb-1.5">No organizations yet</h3>
			<p class="text-body-md text-muted-foreground max-w-[40ch] mb-6">
				Create an organization to start adding monitors, inviting your team, and watching uptime
				together.
			</p>
			<Button
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium"
				onclick={openDialog}
			>
				<Plus class="size-4" />
				Create organization
			</Button>
		</div>

		<!-- Empty state: Search query yielded 0 matches -->
	{:else if filteredOrgs.length === 0}
		<div
			class="flex flex-col items-center justify-center text-center py-20 px-4 border border-border rounded-lg bg-card"
		>
			<div
				class="flex items-center justify-center size-12 rounded-lg bg-primary/10 border border-primary/20 mb-3 text-primary"
			>
				<Building2 class="size-5" />
			</div>
			<h3 class="text-base font-semibold text-foreground mb-1">No organizations found</h3>
			<p class="text-xs text-muted-foreground max-w-[40ch] mb-4">
				No organizations match &ldquo;{headerStore.searchQuery}&rdquo;. Try a different keyword or
				clear the search.
			</p>
			<Button
				variant="outline"
				class="h-8 text-xs font-medium"
				onclick={() => {
					headerStore.searchQuery = '';
				}}
			>
				Clear search
			</Button>
		</div>

		<!-- Org list -->
	{:else}
		{#if viewMode === 'grid'}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(430px,1fr))] gap-5">
				{#each filteredOrgs as org (org.ID)}
					<OrgCard {org} onSelect={selectOrg} />
				{/each}
			</div>
		{:else}
			<div class="flex flex-col gap-2">
				{#each filteredOrgs as org (org.ID)}
					<OrgListItem {org} onSelect={selectOrg} />
				{/each}
			</div>
		{/if}
	{/if}
</div>

<!-- ── Create Organization Dialog ──────────────────────────────────────────── -->
<CreateOrgModal bind:open={dialogOpen} onCreated={handleOrgCreated} />
