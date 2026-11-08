<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { IncidentAPI, type Incident } from '$lib/api/incident';
	import { userTimezone } from '$lib/stores/timezone';
	import { Pagination, DataSkeleton } from '$lib/components/common';
	import { AlertTriangle, CheckCircle2 } from 'lucide-svelte';
	import IncidentToolbar from '$lib/components/incident/IncidentToolbar.svelte';
	import IncidentListItem from '$lib/components/incident/IncidentListItem.svelte';

	let { data }: { data: PageData } = $props();

	let allIncidents: Incident[] = $state([]);
	let isLoading = $state(true);
	let error = $state('');
	let currentOrgId = $derived(data.orgId || ($page.params.id as string) || '');

	let activeFilter = $state<'active' | 'resolved'>('active');
	let severityFilter = $state<'all' | 'sev1' | 'sev2' | 'sev3' | 'sev4'>('all');
	let currentPage = $state(1);

	// Adaptive page size: fit as many rows as the available viewport height allows
	const ESTIMATED_ROW_HEIGHT = 59;
	const MIN_ITEMS_PER_PAGE = 6;
	const DEFAULT_ITEMS_PER_PAGE = 10;
	let listContainerHeight = $state(0);
	let itemsPerPage = $derived(
		listContainerHeight > 0
			? Math.max(MIN_ITEMS_PER_PAGE, Math.floor(listContainerHeight / ESTIMATED_ROW_HEIGHT))
			: DEFAULT_ITEMS_PER_PAGE
	);

	async function loadIncidents(orgId: string) {
		const res = await IncidentAPI.getAll(orgId);
		allIncidents = res.data || [];
	}

	function handleSelectActiveFilter(filter: 'active' | 'resolved') {
		activeFilter = filter;
		currentPage = 1;
	}

	function handleSelectSeverityFilter(sev: 'all' | 'sev1' | 'sev2' | 'sev3' | 'sev4') {
		severityFilter = sev;
		currentPage = 1;
	}

	// Filter & Derived logic
	let activeCount = $derived(allIncidents.filter((inc) => inc.Status === 'ongoing').length);
	let resolvedCount = $derived(allIncidents.filter((inc) => inc.Status === 'resolved').length);

	let filteredIncidents = $derived.by(() => {
		return allIncidents.filter((inc) => {
			// Status Filter (Only active or resolved)
			if (activeFilter === 'active' && inc.Status !== 'ongoing') return false;
			if (activeFilter === 'resolved' && inc.Status !== 'resolved') return false;

			// Severity Filter
			if (severityFilter !== 'all' && inc.SeverityCode !== severityFilter) return false;

			// Search Query Filter
			const q = headerStore.searchQuery.toLowerCase().trim();
			if (q) {
				const matchSummary = inc.Summary?.toLowerCase().includes(q);
				const matchMonitor = inc.Monitor?.Name?.toLowerCase().includes(q);
				const matchTarget = inc.Monitor?.Target?.toLowerCase().includes(q);
				const matchId = inc.ID.toLowerCase().includes(q);
				if (!matchSummary && !matchMonitor && !matchTarget && !matchId) return false;
			}

			return true;
		});
	});

	let totalPages = $derived(Math.max(1, Math.ceil(filteredIncidents.length / itemsPerPage)));
	let safeCurrentPage = $derived(Math.min(Math.max(1, currentPage), totalPages));

	let paginatedIncidents = $derived.by(() => {
		const start = (safeCurrentPage - 1) * itemsPerPage;
		return filteredIncidents.slice(start, start + itemsPerPage);
	});

	// Sync with layout top bar
	$effect(() => {
		void $page.url.pathname;
		const ongoingCount = allIncidents.filter((i) => i.Status !== 'resolved').length;
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search incidents by summary, monitor...',
			badge: {
				text:
					ongoingCount === 0
						? 'All systems operational'
						: `${ongoingCount} Active incident${ongoingCount > 1 ? 's' : ''}`,
				variant: ongoingCount === 0 ? 'emerald' : 'rose'
			}
		});
	});

	onMount(async () => {
		try {
			if (currentOrgId) {
				await loadIncidents(currentOrgId);
			}
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load incidents';
		} finally {
			isLoading = false;
		}
	});

	let tz = $derived($userTimezone);
</script>

<svelte:head>
	<title>Incident Command Board | Netra Uptime</title>
</svelte:head>

{#if isLoading}
	<DataSkeleton variant="list" count={5} />
{:else if error}
	<div class="p-4 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg">
		{error}
	</div>
{:else}
	<!-- GitHub Style Table List Box (fills available viewport height) -->
	<div class="flex flex-col h-[calc(100vh-6rem)] min-h-120">
		<div
			class="flex flex-col flex-1 min-h-0 bg-card border border-border rounded-md overflow-hidden"
		>
			<!-- Table Header Bar with GitHub-Style Active/Resolved Toggles -->
			<div class="shrink-0">
				<IncidentToolbar
					{activeFilter}
					{severityFilter}
					{activeCount}
					{resolvedCount}
					totalFiltered={filteredIncidents.length}
					onSelectActiveFilter={handleSelectActiveFilter}
					onSelectSeverityFilter={handleSelectSeverityFilter}
				/>
			</div>

			<!-- Incident Rows Container -->
			<div
				bind:clientHeight={listContainerHeight}
				data-lenis-prevent
				class="divide-y divide-white/5 flex flex-col flex-1 min-h-0 overflow-y-auto"
			>
				{#each paginatedIncidents as incident (incident.ID)}
					<IncidentListItem {incident} orgId={currentOrgId} timezone={tz} />
				{/each}

				<!-- Empty State when filter/search yields 0 results -->
				{#if paginatedIncidents.length === 0}
					<div class="flex-1 flex flex-col items-center justify-center p-12 text-center space-y-3">
						{#if activeFilter === 'active' && activeCount === 0}
							<CheckCircle2 class="size-10 text-emerald-400 mx-auto opacity-90" />
							<h3 class="text-base font-semibold text-foreground">All Systems Operational</h3>
							<p class="text-xs text-muted-foreground max-w-110 mx-auto">
								No active incidents recorded at this time. All microservices are operating within
								healthy metrics.
							</p>
						{:else}
							<AlertTriangle class="size-10 text-muted-foreground mx-auto opacity-50" />
							<h3 class="text-base font-semibold text-foreground">No Incidents Found</h3>
							<p class="text-xs text-muted-foreground max-w-110 mx-auto">
								No incident records match your current search query or filter criteria.
							</p>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Pagination Footer Bar -->
			{#if filteredIncidents.length > 0}
				<div class="shrink-0 mt-auto bg-white/3 border-t border-border px-5 py-3">
					<Pagination
						bind:currentPage
						{totalPages}
						totalItems={filteredIncidents.length}
						itemName="incidents"
						class="pt-0 border-t-0"
					/>
				</div>
			{/if}
		</div>
	</div>
{/if}
