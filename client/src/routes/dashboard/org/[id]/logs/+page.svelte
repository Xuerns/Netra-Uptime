<script lang="ts">
	import { onMount, untrack, tick } from 'svelte';
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

	// Adaptive page size: fit as many rows as the available viewport height allows
	const ESTIMATED_ROW_HEIGHT = 38;
	const TABLE_HEADER_HEIGHT = 37;
	const PAGINATION_RESERVE_HEIGHT = 48;
	const CARD_BORDER_HEIGHT = 2;
	const MIN_PAGE_SIZE = 6;
	const MAX_PAGE_SIZE = 100;
	const DEFAULT_PAGE_SIZE = 20;
	const PAGE_SIZE_DEBOUNCE_MS = 300;

	let tableAreaHeight = $state(0);
	let actualRowHeight = $state(ESTIMATED_ROW_HEIGHT);
	let lastFetchedLimit = DEFAULT_PAGE_SIZE;

	let availableRowsHeight = $derived(
		tableAreaHeight > 0
			? Math.max(
					0,
					tableAreaHeight - TABLE_HEADER_HEIGHT - PAGINATION_RESERVE_HEIGHT - CARD_BORDER_HEIGHT
				)
			: 0
	);

	let pageSize = $derived(
		availableRowsHeight > 0
			? Math.min(
					MAX_PAGE_SIZE,
					Math.max(MIN_PAGE_SIZE, Math.ceil(availableRowsHeight / actualRowHeight))
				)
			: DEFAULT_PAGE_SIZE
	);
	let searchQuery = $derived(headerStore.searchQuery);
	let selectedCategories = $state<string[]>([]);
	let selectedActions = $state<string[]>([]);

	let activeCategories = $derived(
		AUDIT_CATEGORIES.filter((c) => c.id !== 'ALL' && selectedCategories.includes(c.id))
	);

	let selectedActionItems = $derived.by(() => {
		const items: Array<{ value: string; label: string }> = [];
		for (const val of selectedActions) {
			let label = val;
			for (const group of AUDIT_ACTION_GROUPS) {
				const found = group.actions.find((a) => a.value === val);
				if (found) {
					label = found.label;
					break;
				}
			}
			items.push({ value: val, label });
		}
		return items;
	});

	let selectedActionLabel = $derived(
		selectedActionItems.length === 0
			? ''
			: selectedActionItems.length === 1
				? selectedActionItems[0].label
				: `${selectedActionItems.length} Actions`
	);

	let visibleActionGroups = $derived(
		selectedCategories.length === 0
			? AUDIT_ACTION_GROUPS
			: AUDIT_ACTION_GROUPS.filter((g) => selectedCategories.includes(g.categoryId))
	);

	let hasActiveFilters = $derived(
		selectedCategories.length > 0 || selectedActions.length > 0 || searchQuery.trim() !== ''
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
		const limitToFetch = pageSize;

		try {
			const entityTypeFilter =
				activeCategories.length > 0
					? activeCategories
							.map((c) => c.entityTypes)
							.filter(Boolean)
							.join(',')
					: undefined;
			const actionFilter = selectedActions.length > 0 ? selectedActions.join(',') : undefined;

			const res = await AuditAPI.getLogs(orgId, {
				page: pageToFetch,
				limit: limitToFetch,
				search: searchQuery.trim() || undefined,
				action: actionFilter,
				entity_type: entityTypeFilter
			});

			logs = res.logs ?? [];
			totalCount = res.total_count ?? 0;
			totalPages = res.total_pages ?? 1;
			currentPage = pageToFetch;
			lastFetchedLimit = limitToFetch;
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

	function handleCategoryToggle(catId: string) {
		if (catId === 'ALL') {
			selectedCategories = [];
		} else if (selectedCategories.includes(catId)) {
			selectedCategories = selectedCategories.filter((id) => id !== catId);
		} else {
			selectedCategories = [...selectedCategories, catId];
		}

		if (selectedCategories.length > 0 && selectedActions.length > 0) {
			const allowedActions = new Set(
				AUDIT_ACTION_GROUPS.filter((g) => selectedCategories.includes(g.categoryId)).flatMap((g) =>
					g.actions.map((a) => a.value)
				)
			);
			selectedActions = selectedActions.filter((a) => allowedActions.has(a));
		}

		currentPage = 1;
		fetchAuditLogs(false);
	}

	function handleActionFilterChange(actionVal: string) {
		if (!actionVal) {
			selectedActions = [];
		} else if (selectedActions.includes(actionVal)) {
			selectedActions = selectedActions.filter((a) => a !== actionVal);
		} else {
			selectedActions = [...selectedActions, actionVal];
			const parentGroup = AUDIT_ACTION_GROUPS.find((g) =>
				g.actions.some((a) => a.value === actionVal)
			);
			if (
				parentGroup &&
				selectedCategories.length > 0 &&
				!selectedCategories.includes(parentGroup.categoryId)
			) {
				selectedCategories = [...selectedCategories, parentGroup.categoryId];
			}
		}

		currentPage = 1;
		fetchAuditLogs(false);
	}

	function resetFilters() {
		selectedCategories = [];
		selectedActions = [];
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

	onMount(async () => {
		await tick();
		fetchAuditLogs(true);
	});

	// Refetch from the first page when the viewport-derived page size changes
	$effect(() => {
		const size = pageSize;
		if (isInitialLoading || size === lastFetchedLimit) return;

		const timer = setTimeout(() => {
			untrack(() => {
				currentPage = 1;
				fetchAuditLogs(false, 1);
			});
		}, PAGE_SIZE_DEBOUNCE_MS);

		return () => clearTimeout(timer);
	});
</script>

<svelte:head>
	<title>Audit Logs | Netra Uptime</title>
</svelte:head>

<div
	class="text-foreground antialiased mx-auto w-full flex flex-col h-[calc(100vh-6rem)] min-h-120"
>
	<!-- Filter Controls Toolbar -->
	<div class="shrink-0">
		<AuditLogToolbar
			categories={AUDIT_CATEGORIES}
			{selectedCategories}
			{activeCategories}
			selectedActions={selectedActionItems}
			{searchQuery}
			{hasActiveFilters}
			{totalCount}
			onCategoryToggle={handleCategoryToggle}
			onActionRemove={(val) => handleActionFilterChange(val)}
			onResetFilters={resetFilters}
			onClearSearch={clearSearch}
		/>
	</div>

	<!-- Error alert -->
	{#if errorMsg}
		<div
			class="p-4 border border-rose-500/20 bg-rose-500/10 rounded-lg flex items-center gap-3 text-rose-400 text-sm mb-6 shrink-0"
		>
			<AlertCircle class="size-5 shrink-0" />
			<span>{errorMsg}</span>
		</div>
	{/if}

	<!-- Main Content Area with Adaptive Height Measurement -->
	<div bind:clientHeight={tableAreaHeight} class="flex-1 min-h-0 flex flex-col">
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
				{selectedActions}
				{selectedActionLabel}
				{visibleActionGroups}
				{getActionStyle}
				onActionFilterChange={handleActionFilterChange}
				{fmtTimeAgo}
				{fmtFullDate}
				{initials}
				bind:actualRowHeight
			/>

			{#if totalPages > 1}
				<div class="shrink-0 pt-3">
					<Pagination
						bind:currentPage
						{totalPages}
						totalItems={totalCount}
						itemName="events"
						onPageChange={(page) => fetchAuditLogs(false, page)}
					/>
				</div>
			{/if}
		{/if}
	</div>
</div>
