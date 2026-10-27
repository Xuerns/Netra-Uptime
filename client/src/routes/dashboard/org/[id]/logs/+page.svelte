<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { ScrollText, AlertCircle } from 'lucide-svelte';
	import { AuditAPI, type AuditLogItem } from '$lib/api/audit';
	import { debounce } from '$lib/utils';
	import { EmptyState, Pagination, DataSkeleton } from '$lib/components/common';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatTimeAgo, formatDateTime } from '$lib/utils/datetime';
	import { getInitials } from '$lib/utils/string';
	import AuditLogToolbar from '$lib/components/audit/AuditLogToolbar.svelte';
	import AuditLogTable from '$lib/components/audit/AuditLogTable.svelte';
	import {
		AUDIT_CATEGORIES,
		AUDIT_ACTION_GROUPS,
		getActionStyle
	} from '$lib/components/audit/audit-constants';

	let { data }: { data: PageData } = $props();

	const orgId = $derived(data.orgId || ($page.params.id as string) || '');

	let isInitialLoading = $state(true);
	let isFetching = $state(false);
	let errorMsg = $state('');

	let logs = $state<AuditLogItem[]>([]);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalCount = $state(0);

	let searchQuery = $derived(headerStore.searchQuery);
	let selectedCategory = $state<string>('ALL');
	let selectedAction = $state('');

	let activeCategory = $derived(
		AUDIT_CATEGORIES.find((c) => c.id === selectedCategory) || AUDIT_CATEGORIES[0]
	);

	let selectedActionLabel = $derived.by(() => {
		if (!selectedAction) return '';
		for (const group of AUDIT_ACTION_GROUPS) {
			const found = group.actions.find((a) => a.value === selectedAction);
			if (found) return found.label;
		}
		return selectedAction;
	});

	let visibleActionGroups = $derived(
		selectedCategory === 'ALL'
			? AUDIT_ACTION_GROUPS
			: AUDIT_ACTION_GROUPS.filter((g) => g.categoryId === selectedCategory)
	);

	let hasActiveFilters = $derived(
		selectedCategory !== 'ALL' || selectedAction !== '' || searchQuery.trim() !== ''
	);

	const initials = (name?: string) => getInitials(name);

	let tz = $derived($userTimezone);
	const fmtTimeAgo = formatTimeAgo;
	const fmtFullDate = (iso: string) => formatDateTime(iso, tz);

	async function fetchAuditLogs(showSkeleton = false, pageNum?: number) {
		if (!orgId) return;
		if (showSkeleton) {
			isInitialLoading = true;
		} else {
			isFetching = true;
		}
		errorMsg = '';

		const pageToFetch = pageNum ?? currentPage;

		try {
			const cat = AUDIT_CATEGORIES.find((c) => c.id === selectedCategory);
			const entityTypeFilter = cat && cat.id !== 'ALL' ? cat.entityTypes : undefined;

			const res = await AuditAPI.getLogs(orgId, {
				page: pageToFetch,
				limit: 20,
				search: searchQuery.trim() || undefined,
				action: selectedAction || undefined,
				entity_type: entityTypeFilter || undefined
			});

			logs = res.logs ?? [];
			totalCount = res.total_count ?? 0;
			totalPages = res.total_pages ?? 1;
			currentPage = pageToFetch;
		} catch (err: unknown) {
			errorMsg = err instanceof Error ? err.message : 'Failed to load audit logs.';
			logs = [];
		} finally {
			isInitialLoading = false;
			isFetching = false;
		}
	}

	const debouncedFetchLogs = debounce(() => {
		currentPage = 1;
		fetchAuditLogs(false);
	}, 350);

	function handleCategoryChange(catId: string) {
		selectedCategory = catId;
		currentPage = 1;

		if (selectedAction) {
			if (catId !== 'ALL') {
				const group = AUDIT_ACTION_GROUPS.find((g) => g.categoryId === catId);
				if (!group || !group.actions.some((a) => a.value === selectedAction)) {
					selectedAction = '';
				}
			}
		}

		fetchAuditLogs(false);
	}

	function handleActionFilterChange(actionVal: string) {
		selectedAction = actionVal;
		if (actionVal) {
			const parentGroup = AUDIT_ACTION_GROUPS.find((g) =>
				g.actions.some((a) => a.value === actionVal)
			);
			if (
				parentGroup &&
				selectedCategory !== 'ALL' &&
				selectedCategory !== parentGroup.categoryId
			) {
				selectedCategory = parentGroup.categoryId;
			}
		}
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function resetFilters() {
		selectedCategory = 'ALL';
		selectedAction = '';
		headerStore.searchQuery = '';
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function clearSearch() {
		headerStore.searchQuery = '';
		currentPage = 1;
		fetchAuditLogs(false);
	}

	// Sync with layout top bar
	$effect(() => {
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search audit logs...',
			badge: {
				text: `${totalCount} Event${totalCount !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	let prevSearch: string | null = null;

	// Only trigger search fetch when searchQuery actually changes after initial mount
	$effect(() => {
		const current = headerStore.searchQuery;
		if (prevSearch === null) {
			prevSearch = current;
			return;
		}
		if (current !== prevSearch) {
			prevSearch = current;
			debouncedFetchLogs();
		}
	});

	onMount(() => {
		fetchAuditLogs(true);
	});
</script>

<svelte:head>
	<title>Audit Logs | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased mx-auto w-full flex-1 flex flex-col min-h-full">
	<!-- Filter Controls Toolbar -->
	<AuditLogToolbar
		categories={AUDIT_CATEGORIES}
		{selectedCategory}
		{activeCategory}
		{selectedAction}
		{selectedActionLabel}
		{searchQuery}
		{hasActiveFilters}
		{totalCount}
		onCategoryChange={handleCategoryChange}
		onActionClear={() => handleActionFilterChange('')}
		onResetFilters={resetFilters}
		onClearSearch={clearSearch}
	/>

	<!-- Error alert -->
	{#if errorMsg}
		<div
			class="p-4 border border-rose-500/20 bg-rose-500/10 rounded-lg flex items-center gap-3 text-rose-400 text-sm mb-6"
		>
			<AlertCircle class="size-5 shrink-0" />
			<span>{errorMsg}</span>
		</div>
	{/if}

	<!-- Logs List / Table -->
	{#if isInitialLoading}
		<DataSkeleton variant="table" count={8} />
	{:else if logs.length === 0}
		<EmptyState
			icon={ScrollText}
			title="No Audit Logs Found"
			description={searchQuery
				? `No activity matching "${searchQuery}" in this organization.`
				: 'No recorded activity events for this organization yet.'}
		/>
	{:else}
		<AuditLogTable
			{logs}
			{isFetching}
			{selectedAction}
			{selectedActionLabel}
			{visibleActionGroups}
			{getActionStyle}
			onActionFilterChange={handleActionFilterChange}
			{fmtTimeAgo}
			{fmtFullDate}
			{initials}
		/>

		<!-- Pagination Footer -->
		<Pagination
			bind:currentPage
			{totalPages}
			totalItems={totalCount}
			itemName="events"
			onPageChange={(page) => fetchAuditLogs(false, page)}
			class="mt-4"
		/>
	{/if}
</div>
