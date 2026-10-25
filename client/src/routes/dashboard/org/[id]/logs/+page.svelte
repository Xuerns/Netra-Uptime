<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import {
		ScrollText,
		Activity,
		Users,
		Globe,
		Building2,
		Wrench,
		FolderTree,
		KeyRound,
		Bell,
		Send,
		AlertTriangle,
		AlertCircle
	} from 'lucide-svelte';
	import { AuditAPI, type AuditLogItem } from '$lib/api/audit';
	import { debounce } from '$lib/helper/utils';
	import { EmptyState, Pagination, DataSkeleton } from '$lib/components/common';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatTimeAgo, formatDateTime } from '$lib/utils/datetime';
	import { getInitials } from '$lib/utils/string';
	import AuditLogToolbar, {
		type AuditCategory
	} from '$lib/components/audit/AuditLogToolbar.svelte';
	import AuditLogTable, {
		type ActionFilterGroup
	} from '$lib/components/audit/AuditLogTable.svelte';

	const orgId = $derived($page.params.id);

	let isInitialLoading = $state(true);
	let isFetching = $state(false);
	let errorMsg = $state('');

	let logs = $state<AuditLogItem[]>([]);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalCount = $state(0);

	let searchQuery = $state('');
	let selectedCategory = $state<string>('ALL');
	let selectedAction = $state('');

	// ── Category Filters Definition ────────────────────────────────────────────
	const categories: AuditCategory[] = [
		{ id: 'ALL', label: 'All Categories', icon: ScrollText, entityTypes: '' },
		{ id: 'MONITOR', label: 'Monitors', icon: Activity, entityTypes: 'monitor' },
		{ id: 'ORGANIZATION', label: 'Organization', icon: Building2, entityTypes: 'organization' },
		{
			id: 'MEMBER',
			label: 'Members & Invites',
			icon: Users,
			entityTypes: 'organization_member,organization_invitation'
		},
		{ id: 'STATUSPAGE', label: 'Status Pages', icon: Globe, entityTypes: 'status_page' },
		{
			id: 'MAINTENANCE',
			label: 'Maintenance',
			icon: Wrench,
			entityTypes: 'scheduled_maintenance'
		},
		{
			id: 'COMPONENT_GROUP',
			label: 'Component Groups',
			icon: FolderTree,
			entityTypes: 'component_group'
		},
		{ id: 'API_KEY', label: 'API Keys', icon: KeyRound, entityTypes: 'api_key' },
		{
			id: 'ALERT_TEMPLATE',
			label: 'Alert Templates & Rules',
			icon: Bell,
			entityTypes: 'alert_template,alert_rule'
		},
		{
			id: 'NOTIFICATION',
			label: 'Notifications',
			icon: Send,
			entityTypes: 'notification_channel'
		},
		{ id: 'INCIDENT', label: 'Incidents', icon: AlertTriangle, entityTypes: 'incident' }
	];

	// ── Action Filter Definitions (grouped by category) ────────────────────────
	const actionFilterGroups: ActionFilterGroup[] = [
		{
			categoryId: 'MONITOR',
			label: 'Monitors',
			icon: Activity,
			actions: [
				{ value: 'MONITOR_CREATE', label: 'Monitor Created' },
				{ value: 'MONITOR_UPDATE', label: 'Monitor Updated' },
				{ value: 'MONITOR_DELETE', label: 'Monitor Deleted' }
			]
		},
		{
			categoryId: 'ORGANIZATION',
			label: 'Organization',
			icon: Building2,
			actions: [
				{ value: 'ORG_UPDATE', label: 'Organization Updated' },
				{ value: 'ORG_DISBAND', label: 'Organization Disbanded' },
				{ value: 'ORG_TRANSFER_OWNERSHIP', label: 'Ownership Transferred' }
			]
		},
		{
			categoryId: 'MEMBER',
			label: 'Members & Invites',
			icon: Users,
			actions: [
				{ value: 'ORG_MEMBER_INVITE', label: 'Member Invited' },
				{ value: 'ORG_MEMBER_JOIN', label: 'Member Joined' },
				{ value: 'ORG_MEMBER_ACCEPT', label: 'Invite Accepted' },
				{ value: 'ORG_INVITE_REJECT', label: 'Invite Rejected' },
				{ value: 'ORG_MEMBER_PROMOTE', label: 'Member Promoted' },
				{ value: 'ORG_MEMBER_DEMOTE', label: 'Member Demoted' },
				{ value: 'ORG_MEMBER_KICK', label: 'Member Kicked' },
				{ value: 'ORG_MEMBER_LEAVE', label: 'Member Left' }
			]
		},
		{
			categoryId: 'STATUSPAGE',
			label: 'Status Pages',
			icon: Globe,
			actions: [
				{ value: 'STATUSPAGE_CREATE', label: 'Status Page Created' },
				{ value: 'STATUSPAGE_UPDATE', label: 'Status Page Updated' },
				{ value: 'STATUSPAGE_DELETE', label: 'Status Page Deleted' }
			]
		},
		{
			categoryId: 'MAINTENANCE',
			label: 'Scheduled Maintenance',
			icon: Wrench,
			actions: [
				{ value: 'MAINTENANCE_CREATE', label: 'Maintenance Created' },
				{ value: 'MAINTENANCE_UPDATE', label: 'Maintenance Updated' },
				{ value: 'MAINTENANCE_DELETE', label: 'Maintenance Deleted' }
			]
		},
		{
			categoryId: 'COMPONENT_GROUP',
			label: 'Component Groups',
			icon: FolderTree,
			actions: [
				{ value: 'COMPONENT_GROUP_CREATE', label: 'Component Group Created' },
				{ value: 'COMPONENT_GROUP_UPDATE', label: 'Component Group Updated' },
				{ value: 'COMPONENT_GROUP_DELETE', label: 'Component Group Deleted' }
			]
		},
		{
			categoryId: 'API_KEY',
			label: 'API Keys',
			icon: KeyRound,
			actions: [
				{ value: 'API_KEY_CREATE', label: 'API Key Created' },
				{ value: 'API_KEY_REVOKE', label: 'API Key Revoked' }
			]
		},
		{
			categoryId: 'ALERT_TEMPLATE',
			label: 'Alert Templates & Rules',
			icon: Bell,
			actions: [
				{ value: 'ALERT_TEMPLATE_CREATE', label: 'Template Created' },
				{ value: 'ALERT_TEMPLATE_UPDATE', label: 'Template Updated' },
				{ value: 'ALERT_TEMPLATE_DELETE', label: 'Template Deleted' },
				{ value: 'ALERT_RULE_CREATE', label: 'Alert Rule Created' },
				{ value: 'ALERT_RULE_UPDATE', label: 'Alert Rule Updated' },
				{ value: 'ALERT_RULE_DELETE', label: 'Alert Rule Deleted' }
			]
		},
		{
			categoryId: 'NOTIFICATION',
			label: 'Notification Channels',
			icon: Send,
			actions: [
				{ value: 'NOTIFICATION_CHANNEL_CREATE', label: 'Channel Created' },
				{ value: 'NOTIFICATION_CHANNEL_UPDATE', label: 'Channel Updated' },
				{ value: 'NOTIFICATION_CHANNEL_DELETE', label: 'Channel Deleted' }
			]
		},
		{
			categoryId: 'INCIDENT',
			label: 'Incidents',
			icon: AlertTriangle,
			actions: [
				{ value: 'INCIDENT_UPDATE', label: 'Incident Updated' },
				{ value: 'INCIDENT_ASSIGNEE_JOIN', label: 'Assignee Joined' }
			]
		}
	];

	let activeCategory = $derived(categories.find((c) => c.id === selectedCategory) || categories[0]);

	let selectedActionLabel = $derived.by(() => {
		if (!selectedAction) return '';
		for (const group of actionFilterGroups) {
			const found = group.actions.find((a) => a.value === selectedAction);
			if (found) return found.label;
		}
		return selectedAction;
	});

	let visibleActionGroups = $derived(
		selectedCategory === 'ALL'
			? actionFilterGroups
			: actionFilterGroups.filter((g) => g.categoryId === selectedCategory)
	);

	let hasActiveFilters = $derived(
		selectedCategory !== 'ALL' || selectedAction !== '' || searchQuery.trim() !== ''
	);

	function getActionStyle(action: string) {
		const act = action.toUpperCase();
		if (
			act.includes('CREATE') ||
			act.includes('JOIN') ||
			act.includes('UPGRADE') ||
			act.includes('ACCEPT')
		) {
			return 'text-emerald-400';
		}
		if (
			act.includes('UPDATE') ||
			act.includes('INVITE') ||
			act.includes('PROMOTE') ||
			act.includes('DEMOTE')
		) {
			return 'text-sky-400';
		}
		if (
			act.includes('DELETE') ||
			act.includes('KICK') ||
			act.includes('DISBAND') ||
			act.includes('TRANSFER') ||
			act.includes('REJECT') ||
			act.includes('LEAVE') ||
			act.includes('REVOKE')
		) {
			return 'text-rose-400';
		}
		return 'text-muted-foreground';
	}

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
			const cat = categories.find((c) => c.id === selectedCategory);
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
				const group = actionFilterGroups.find((g) => g.categoryId === catId);
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
			const parentGroup = actionFilterGroups.find((g) =>
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
		searchQuery = '';
		currentPage = 1;
		fetchAuditLogs(false);
	}

	function clearSearch() {
		headerStore.searchQuery = '';
		searchQuery = '';
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

	let prevSearch = $state<string | null>(null);

	// Only trigger search fetch when searchQuery actually changes after initial mount
	$effect(() => {
		const current = headerStore.searchQuery;
		if (prevSearch === null) {
			prevSearch = current;
			searchQuery = current;
			return;
		}
		if (current !== prevSearch) {
			prevSearch = current;
			searchQuery = current;
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
		{categories}
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
