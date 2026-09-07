<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		ScrollText,
		Building2,
		Search,
		ChevronLeft,
		ChevronRight,
		ChevronsUpDown,
		Activity,
		Users,
		Globe,
		ShieldAlert,
		CheckCircle2,
		AlertCircle,
		Info,
		ArrowRight,
		Plus,
		ListFilter,
		X
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { OrgAPI } from '$lib/api/organization';
	import { AuditAPI, type AuditLogItem } from '$lib/api/audit';
	import { debounce } from '$lib/helper/utils';
	import LogsSkeleton from '$lib/components/skeletons/LogsSkeleton.svelte';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatTimeAgo, formatDateTime } from '$lib/utils/datetime';

	// ── Types ──────────────────────────────────────────────────────────────────
	interface Organization {
		ID: string;
		Name: string;
		Slug: string;
	}

	// ── State ──────────────────────────────────────────────────────────────────
	let orgs = $state<Organization[]>([]);
	let selectedOrgId = $state<string>('');
	let selectedOrg = $derived(orgs.find((o) => o.ID === selectedOrgId));

	let orgSearchQuery = $state('');
	let filteredOrgs = $derived(
		orgs.filter(
			(o) =>
				o.Name.toLowerCase().includes(orgSearchQuery.toLowerCase()) ||
				o.Slug.toLowerCase().includes(orgSearchQuery.toLowerCase())
		)
	);

	let isLoadingOrgs = $state(true);
	let isLoadingLogs = $state(false);
	let errorMsg = $state('');

	let logs = $state<AuditLogItem[]>([]);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalCount = $state(0);

	let searchQuery = $state('');
	let selectedCategory = $state<'ALL' | 'MONITOR' | 'MEMBER' | 'STATUSPAGE'>('ALL');
	let selectedAction = $state('');

	// ── Category Filters Definition ────────────────────────────────────────────
	const categories = [
		{ id: 'ALL', label: 'All Actions', icon: ScrollText },
		{ id: 'MONITOR', label: 'Monitors', icon: Activity },
		{ id: 'MEMBER', label: 'Members', icon: Users },
		{ id: 'STATUSPAGE', label: 'Status Pages', icon: Globe }
	];

	// ── Action Filter Definitions (grouped by category) ────────────────────────
	const actionFilterGroups = [
		{
			label: 'Monitor',
			actions: [
				{ value: 'MONITOR_CREATE', label: 'Monitor Create' },
				{ value: 'MONITOR_UPDATE', label: 'Monitor Update' },
				{ value: 'MONITOR_DELETE', label: 'Monitor Delete' }
			]
		},
		{
			label: 'Organization',
			actions: [
				{ value: 'ORG_UPDATE', label: 'Org Update' },
				{ value: 'ORG_DISBAND', label: 'Org Disband' },
				{ value: 'ORG_TRANSFER_OWNERSHIP', label: 'Org Transfer' }
			]
		},
		{
			label: 'Members',
			actions: [
				{ value: 'ORG_MEMBER_JOIN', label: 'Member Join' },
				{ value: 'ORG_MEMBER_INVITE', label: 'Member Invite' },
				{ value: 'ORG_MEMBER_PROMOTE', label: 'Member Promote' },
				{ value: 'ORG_MEMBER_DEMOTE', label: 'Member Demote' },
				{ value: 'ORG_MEMBER_KICK', label: 'Member Kick' },
				{ value: 'ORG_MEMBER_LEAVE', label: 'Member Leave' },
				{ value: 'ORG_INVITE_REJECT', label: 'Invite Reject' }
			]
		},
		{
			label: 'Status Page',
			actions: [
				{ value: 'STATUSPAGE_CREATE', label: 'Page Create' },
				{ value: 'STATUSPAGE_UPDATE', label: 'Page Update' },
				{ value: 'STATUSPAGE_DELETE', label: 'Page Delete' }
			]
		},
		{
			label: 'API Key',
			actions: [
				{ value: 'API_KEY_CREATE', label: 'Key Create' },
				{ value: 'API_KEY_REVOKE', label: 'Key Revoke' }
			]
		}
	];

	// Map category to which action groups to show
	const categoryToGroupLabels: Record<string, string[]> = {
		ALL: ['Monitor', 'Organization', 'Members', 'Status Page', 'API Key'],
		MONITOR: ['Monitor'],
		MEMBER: ['Organization', 'Members'],
		STATUSPAGE: ['Status Page']
	};

	let visibleActionGroups = $derived(
		actionFilterGroups.filter((g) =>
			(categoryToGroupLabels[selectedCategory] ?? []).includes(g.label)
		)
	);

	// ── Helper Badge Styling ───────────────────────────────────────────────────
	function getActionStyle(action: string) {
		const act = action.toUpperCase();
		if (
			act.includes('CREATE') ||
			act.includes('JOIN') ||
			act.includes('UPGRADE') ||
			act.includes('ACCEPT')
		) {
			return {
				bg: 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400',
				dot: 'bg-emerald-400',
				textColor: 'text-emerald-400'
			};
		}
		if (
			act.includes('UPDATE') ||
			act.includes('INVITE') ||
			act.includes('PROMOTE') ||
			act.includes('DEMOTE')
		) {
			return {
				bg: 'bg-sky-500/10 border-sky-500/20 text-sky-400',
				dot: 'bg-sky-400',
				textColor: 'text-sky-400'
			};
		}
		if (
			act.includes('DELETE') ||
			act.includes('KICK') ||
			act.includes('DISBAND') ||
			act.includes('TRANSFER') ||
			act.includes('REJECT') ||
			act.includes('LEAVE')
		) {
			return {
				bg: 'bg-rose-500/10 border-rose-500/20 text-rose-400',
				dot: 'bg-rose-400',
				textColor: 'text-rose-400'
			};
		}
		return {
			bg: 'bg-white/10 border-border text-muted-foreground',
			dot: 'bg-slate-400',
			textColor: 'text-muted-foreground'
		};
	}

	function initials(name?: string): string {
		if (!name) return 'US';
		const parts = name.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	// Format functions dari $lib/utils/datetime + timezone dari store
	let tz = $derived($userTimezone);
	const fmtTimeAgo = formatTimeAgo;
	const fmtFullDate = (iso: string) => formatDateTime(iso, tz);

	// ── Data Fetching ──────────────────────────────────────────────────────────
	async function loadOrganizations() {
		isLoadingOrgs = true;
		errorMsg = '';
		try {
			const res = await OrgAPI.getAll();
			orgs = res.data ?? [];

			// Check URL query param for pre-selected orgId
			const queryOrgId = $page.url.searchParams.get('orgId');
			if (queryOrgId && orgs.some((o) => o.ID === queryOrgId)) {
				selectOrganization(queryOrgId, false);
			} else if (orgs.length === 1) {
				// If user only belongs to 1 organization, auto select it for convenience
				selectOrganization(orgs[0].ID, false);
			}
		} catch (err: any) {
			errorMsg = err.message || 'Failed to load organizations.';
		} finally {
			isLoadingOrgs = false;
		}
	}

	async function fetchAuditLogs() {
		if (!selectedOrgId) return;
		isLoadingLogs = true;
		errorMsg = '';

		try {
			let entityTypeFilter = '';
			if (selectedCategory !== 'ALL') {
				switch (selectedCategory) {
					case 'MONITOR':
						entityTypeFilter = 'monitor';
						break;
					case 'MEMBER':
						entityTypeFilter = 'organization_member';
						break;
					case 'STATUSPAGE':
						entityTypeFilter = 'status_page';
						break;
				}
			}

			const res = await AuditAPI.getLogs(selectedOrgId, {
				page: currentPage,
				limit: 15,
				search: searchQuery.trim(),
				entity_type: entityTypeFilter,
				action: selectedAction || undefined
			});

			logs = res.logs || [];
			totalCount = res.total_count || 0;
			totalPages = res.total_pages || 1;
		} catch (err: any) {
			errorMsg = err.message || 'Failed to load audit logs.';
			logs = [];
		} finally {
			isLoadingLogs = false;
		}
	}

	const debouncedFetchLogs = debounce(() => {
		currentPage = 1;
		fetchAuditLogs();
	}, 350);

	function selectOrganization(orgId: string, updateUrl = true) {
		selectedOrgId = orgId;
		currentPage = 1;
		searchQuery = '';
		selectedCategory = 'ALL';
		selectedAction = '';

		if (updateUrl) {
			const newUrl = new URL(window.location.href);
			newUrl.searchParams.set('orgId', orgId);
			goto(newUrl.pathname + newUrl.search, { replaceState: true, keepFocus: true });
		}

		fetchAuditLogs();
	}

	function handleCategoryChange(cat: 'ALL' | 'MONITOR' | 'MEMBER' | 'STATUSPAGE') {
		selectedCategory = cat;
		currentPage = 1;

		// Reset action filter if it doesn't belong to the new category's visible groups
		if (selectedAction) {
			const allowedLabels = categoryToGroupLabels[cat] ?? [];
			const allowedGroups = actionFilterGroups.filter((g) => allowedLabels.includes(g.label));
			const allAllowedActions = allowedGroups.flatMap((g) => g.actions.map((a) => a.value));
			if (!allAllowedActions.includes(selectedAction)) {
				selectedAction = '';
			}
		}

		fetchAuditLogs();
	}

	function handleSearchInput() {
		debouncedFetchLogs();
	}

	function handleActionFilterChange(action: string) {
		selectedAction = action;
		currentPage = 1;
		fetchAuditLogs();
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages || p === currentPage) return;
		currentPage = p;
		fetchAuditLogs();
	}

	onMount(() => {
		loadOrganizations();
	});
</script>

<svelte:head>
	<title>Audit Logs • PulseWatch</title>
</svelte:head>

<div class="text-foreground antialiased mx-auto w-full flex-1 flex flex-col min-h-full">
	<!-- Page Header (Only shown when organization is selected) -->
	{#if selectedOrgId}
		<header class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
			<div>
				<div class="flex items-center gap-2 mb-1">
					<h1 class="text-[28px] font-semibold leading-tight tracking-tight text-foreground">
						Audit Logs
					</h1>
				</div>
				<p class="text-sm text-muted-foreground max-w-[54ch]">
					Traceable, real-time audit trail of all actions and security events across your
					organization.
				</p>
			</div>

			<!-- Active Organization Selector Dropdown (Shown when org is selected) -->
			{#if orgs.length > 0}
				<div class="flex items-center gap-2">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="focus:outline-none">
							<button
								type="button"
								class="flex items-center gap-2.5 px-3 py-2 bg-white/5 border border-border hover:border-border rounded-lg transition-all text-left text-sm"
							>
								<div
									class="size-6 rounded-md bg-primary/20 text-primary font-bold text-xs flex items-center justify-center"
								>
									{initials(selectedOrg?.Name)}
								</div>
								<span class="font-medium text-foreground max-w-[160px] truncate"
									>{selectedOrg?.Name}</span
								>
								<ChevronsUpDown class="size-4 text-muted-foreground shrink-0" />
							</button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							class="w-60 bg-secondary border-border text-foreground shadow-lg rounded-lg p-1"
							align="end"
						>
							<DropdownMenu.Label
								class="text-xs text-muted-foreground font-medium px-2 py-1.5 uppercase tracking-wider"
							>
								Select Organization
							</DropdownMenu.Label>
							<DropdownMenu.Separator class="bg-white/8 my-1" />
							{#each orgs as org}
								<DropdownMenu.Item
									class="cursor-pointer flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-white/8 text-sm font-medium {org.ID ===
									selectedOrgId
										? 'text-primary bg-white/5'
										: 'text-slate-200'}"
									onclick={() => selectOrganization(org.ID)}
								>
									<span class="truncate">{org.Name}</span>
									{#if org.ID === selectedOrgId}
										<CheckCircle2 class="size-4 text-primary shrink-0" />
									{/if}
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			{/if}
		</header>
	{/if}

	<!-- STATE 1: Organization Selection Required Prompt -->
	{#if isLoadingOrgs}
		<LogsSkeleton />
	{:else if !selectedOrgId}
		<div class="flex-1 min-h-[calc(100vh-10rem)] flex flex-col items-center justify-center py-6">
			<div
				class="w-full max-w-[420px] bg-card border border-border rounded-lg p-6 shadow-2xl backdrop-blur-xl flex flex-col items-center text-center"
			>
				<!-- Header Icon Badge -->
				<div
					class="size-11 rounded-lg bg-white/5 border border-border flex items-center justify-center text-primary mb-4 shadow-sm"
				>
					<ScrollText class="size-5" />
				</div>

				<h2 class="text-lg font-semibold text-foreground tracking-tight mb-1">Continue to Logs</h2>
				<p class="text-xs text-muted-foreground mb-5">Choose an organization to view audit logs</p>

				<!-- Search Organization Input -->
				<div class="relative w-full mb-3">
					<Search class="absolute left-3 top-2.5 size-4 text-muted-foreground/50" />
					<Input
						type="text"
						placeholder="Find organization..."
						bind:value={orgSearchQuery}
						class="pl-9 h-9 text-xs bg-black/40 border-border focus-visible:border-primary text-foreground rounded-lg placeholder:text-muted-foreground/40"
					/>
				</div>

				<!-- Organization List -->
				{#if filteredOrgs.length === 0}
					<div
						data-lenis-prevent
						class="py-6 px-4 w-full text-center text-xs text-muted-foreground/60 bg-black/20 rounded-lg border border-border"
					>
						{orgSearchQuery
							? `No organization matching "${orgSearchQuery}"`
							: 'No organizations found.'}
					</div>
				{:else}
					<div
						data-lenis-prevent
						class="flex flex-col gap-1.5 w-full max-h-[260px] overflow-y-auto pr-0.5"
					>
						{#each filteredOrgs as org}
							<button
								type="button"
								class="w-full flex items-center justify-between p-2.5 bg-white/4 hover:bg-white/8 border border-border hover:border-primary/40 rounded-lg transition-all text-left group"
								onclick={() => selectOrganization(org.ID)}
							>
								<div class="flex items-center gap-3 min-w-0">
									<div
										class="size-8 rounded-lg bg-primary/15 border border-primary/20 text-primary font-bold text-xs flex items-center justify-center shrink-0"
									>
										{initials(org.Name)}
									</div>
									<div class="min-w-0 text-left">
										<p
											class="text-xs font-semibold text-foreground truncate group-hover:text-primary transition-colors"
										>
											{org.Name}
										</p>
										<p class="text-[11px] text-muted-foreground/70 font-mono truncate">
											pulsewatch.app/o/{org.Slug}
										</p>
									</div>
								</div>
								<ChevronRight
									class="size-4 text-muted-foreground/50 group-hover:text-primary group-hover:translate-x-0.5 transition-all shrink-0 ml-2"
								/>
							</button>
						{/each}
					</div>
				{/if}

				<!-- Divider & Create Action -->
				<div class="w-full h-px bg-white/8 my-4"></div>

				<a
					href="/dashboard/organization"
					class="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-lg text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 border border-transparent hover:border-border transition-all"
				>
					<Plus class="size-3.5" />
					Create Organization
				</a>
			</div>
		</div>

		<!-- STATE 2: Organization Selected — Display Audit Logs -->
	{:else}
		<!-- Controls: Search & Category Chips -->
		<div class="flex flex-col gap-4 mb-6">
			<div
				class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-3 bg-white/3 border border-border rounded-lg backdrop-blur-md"
			>
				<!-- Search Input -->
				<div class="relative flex-1 min-w-[240px]">
					<Search class="absolute left-3 top-2.5 size-4 text-muted-foreground/50" />
					<Input
						type="text"
						placeholder="Search logs by action or description..."
						bind:value={searchQuery}
						oninput={handleSearchInput}
						class="pl-9 h-9 text-sm bg-black/30 border-border focus-visible:border-primary text-foreground rounded-lg placeholder:text-muted-foreground/40"
					/>
				</div>

				<!-- Category Filter Buttons (GitHub Style) -->
				<div class="flex items-center gap-1 overflow-x-auto pb-1 sm:pb-0">
					{#each categories as cat}
						{@const isSelected = selectedCategory === cat.id}
						<button
							type="button"
							class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs transition-all whitespace-nowrap border cursor-pointer {isSelected
								? 'font-bold text-foreground bg-white/10 border-border shadow-sm'
								: 'font-medium text-foreground/40 border-transparent bg-transparent hover:text-foreground/70 hover:bg-white/5'}"
							onclick={() => handleCategoryChange(cat.id as any)}
						>
							<cat.icon class="size-3.5 {isSelected ? 'text-primary' : 'text-foreground/40'}" />
							{cat.label}
						</button>
					{/each}
				</div>
			</div>
		</div>

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
		{#if isLoadingLogs}
			<LogsSkeleton />
		{:else if logs.length === 0}
			<div
				class="w-full border border-border rounded-lg bg-white/3 p-10 md:p-14 text-center flex flex-col items-center justify-center min-h-[260px]"
			>
				<div
					class="size-12 rounded-lg bg-white/5 border border-border flex items-center justify-center text-muted-foreground mb-3 font-semibold"
				>
					<ScrollText class="size-6 opacity-60" />
				</div>
				<h3 class="text-base font-semibold text-foreground mb-1.5">No Audit Logs Found</h3>
				<p class="text-sm text-muted-foreground max-w-[440px] w-full leading-relaxed mx-auto">
					{searchQuery
						? `No activity matching "${searchQuery}" in this organization.`
						: 'No recorded activity events for this organization yet.'}
				</p>
			</div>
		{:else}
			<div
				class="border border-border rounded-lg bg-white/3 backdrop-blur-md overflow-hidden flex flex-col max-h-[calc(100vh-16rem)]"
			>
				<!-- Table Column Headers -->
				<div class="log-table-header sticky top-0 bg-card z-10">
					<div class="log-col-actor">Actor</div>
					<div class="log-col-action">
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="focus:outline-none w-full">
								<span
									class="flex items-center gap-1.5 transition-colors {selectedAction
										? 'text-primary'
										: 'text-inherit hover:text-foreground'}"
								>
									<span>{selectedAction || 'Action'}</span>
									{#if selectedAction}
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<span
											role="button"
											tabindex="-1"
											class="size-3.5 rounded-full bg-primary/20 hover:bg-primary/40 flex items-center justify-center transition-colors"
											onclick={(e) => {
												e.stopPropagation();
												handleActionFilterChange('');
											}}
											onkeydown={(e) => {
												if (e.key === 'Enter') {
													e.stopPropagation();
													handleActionFilterChange('');
												}
											}}
										>
											<X class="size-2.5 text-primary" />
										</span>
									{:else}
										<ListFilter class="size-3 opacity-50" />
									{/if}
								</span>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content
								data-lenis-prevent
								class="w-56 max-h-[360px] overflow-y-auto bg-secondary border-border text-foreground shadow-xl rounded-lg p-1"
								align="start"
								sideOffset={6}
							>
								<!-- All Actions option -->
								<DropdownMenu.Item
									class="cursor-pointer flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-white/8 text-sm font-medium {selectedAction ===
									''
										? 'text-primary bg-white/5'
										: 'text-slate-200'}"
									onclick={() => handleActionFilterChange('')}
								>
									<span>All Actions</span>
									{#if selectedAction === ''}
										<CheckCircle2 class="size-3.5 text-primary shrink-0" />
									{/if}
								</DropdownMenu.Item>

								{#each visibleActionGroups as group}
									<DropdownMenu.Separator class="bg-white/8 my-1" />
									<DropdownMenu.Label
										class="text-[10px] text-muted-foreground font-medium px-2.5 py-1 uppercase tracking-wider"
									>
										{group.label}
									</DropdownMenu.Label>
									{#each group.actions as action}
										{@const style = getActionStyle(action.value)}
										<DropdownMenu.Item
											class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs {selectedAction ===
											action.value
												? 'bg-white/5'
												: ''}"
											onclick={() => handleActionFilterChange(action.value)}
										>
											<span
												class="text-[11px] font-mono font-medium {style.textColor} whitespace-nowrap"
											>
												{action.value}
											</span>
											{#if selectedAction === action.value}
												<CheckCircle2 class="size-3.5 text-primary shrink-0" />
											{/if}
										</DropdownMenu.Item>
									{/each}
								{/each}
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>
					<div class="log-col-description">Description</div>
					<div class="log-col-ip">IP Address</div>
					<div class="log-col-timestamp">Timestamp</div>
				</div>

				<!-- Scrollable Rows Container -->
				<div
					data-lenis-prevent
					class="log-scroll-container overflow-y-auto flex-1 divide-y divide-border"
				>
					{#each logs as log (log.id)}
						{@const badgeStyle = getActionStyle(log.action)}
						<div class="log-table-row group">
							<!-- Actor -->
							<div class="log-col-actor flex items-center gap-2.5">
								{#if log.user_avatar_url}
									<img
										src={log.user_avatar_url}
										alt={log.user_name || log.user_email || 'User'}
										class="size-7 rounded-md object-cover border border-border shrink-0"
										onerror={(e) => {
											const el = e.currentTarget as HTMLImageElement;
											el.style.display = 'none';
											const fallback = el.nextElementSibling as HTMLElement | null;
											if (fallback) fallback.style.display = 'flex';
										}}
									/>
									<div
										class="size-7 rounded-md bg-emerald-500/15 border border-emerald-500/20 text-emerald-400 font-bold text-[10px] items-center justify-center shrink-0"
										style="display:none"
									>
										{initials(log.user_name || log.user_email)}
									</div>
								{:else}
									<div
										class="size-7 rounded-md bg-emerald-500/15 border border-emerald-500/20 text-emerald-400 font-bold text-[10px] flex items-center justify-center shrink-0"
									>
										{initials(log.user_name || log.user_email)}
									</div>
								{/if}
								<span class="text-[13px] font-medium text-foreground truncate">
									{log.user_name || log.user_email || 'Unknown'}
								</span>
							</div>

							<!-- Action -->
							<div class="log-col-action">
								<span
									class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-mono font-semibold border {badgeStyle.bg} whitespace-nowrap"
								>
									{log.action}
								</span>
							</div>

							<!-- Description -->
							<div class="log-col-description">
								<p class="text-[13px] text-muted-foreground leading-snug line-clamp-2">
									{log.description}
								</p>
							</div>

							<!-- IP Address -->
							<div class="log-col-ip">
								<span
									class="font-mono text-[11px] px-2 py-0.5 rounded-md bg-white/5 border border-border text-muted-foreground whitespace-nowrap"
								>
									{log.ip_address || 'Internal'}
								</span>
							</div>

							<!-- Timestamp -->
							<div class="log-col-timestamp" title={fmtFullDate(log.created_at)}>
								<span class="text-[13px] text-muted-foreground font-medium whitespace-nowrap">
									{fmtTimeAgo(log.created_at)}
								</span>
								<span class="text-[10px] text-muted-foreground/55 font-mono block mt-0.5">
									{fmtFullDate(log.created_at)}
								</span>
							</div>
						</div>
					{/each}
				</div>

				<!-- Pagination Footer -->
				{#if totalPages > 1}
					<div
						class="px-4 py-3 bg-white/4 border-t border-border flex items-center justify-between text-xs text-muted-foreground sticky bottom-0 bg-card"
					>
						<span>Page {currentPage} of {totalPages} ({totalCount} items)</span>
						<div class="flex items-center gap-1">
							<Button
								variant="ghost"
								class="h-8 w-8 p-0 border border-border rounded-md text-foreground hover:bg-white/10 disabled:opacity-30"
								disabled={currentPage <= 1}
								onclick={() => goToPage(currentPage - 1)}
							>
								<ChevronLeft class="size-4" />
							</Button>
							<span class="px-2 font-mono text-foreground">{currentPage}</span>
							<Button
								variant="ghost"
								class="h-8 w-8 p-0 border border-border rounded-md text-foreground hover:bg-white/10 disabled:opacity-30"
								disabled={currentPage >= totalPages}
								onclick={() => goToPage(currentPage + 1)}
							>
								<ChevronRight class="size-4" />
							</Button>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	/* ── Log Table Grid Layout ───────────────────────────────────────────── */
	/* columns: Actor | Action | Description | IP Address | Timestamp */
	.log-table-header,
	.log-table-row {
		display: grid;
		grid-template-columns: 180px 200px 1fr 130px 160px;
		align-items: center;
		gap: 12px;
	}

	.log-table-header {
		padding: 8px 16px;
		background: var(--color-muted);
		border-bottom: 1px solid var(--color-border);
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.07em;
		color: var(--color-muted-foreground);
	}

	.log-table-row {
		padding: 12px 16px;
		transition: background 140ms ease;
	}

	.log-table-row:hover {
		background: var(--color-accent);
	}

	/* Column cells are just referenced by class name, alignment handled by grid */
	.log-col-actor,
	.log-col-action,
	.log-col-description,
	.log-col-ip,
	.log-col-timestamp {
		min-width: 0;
		overflow: hidden;
	}

	.log-col-ip {
		display: flex;
		align-items: center;
	}

	.log-col-timestamp {
		text-align: right;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}

	/* ── Scrollable Log Container ──────────────────────────────────────── */
	.log-scroll-container {
		scrollbar-width: thin;
		scrollbar-color: var(--color-muted-foreground) transparent;
	}

	.log-scroll-container::-webkit-scrollbar {
		width: 6px;
	}

	.log-scroll-container::-webkit-scrollbar-track {
		background: transparent;
	}

	.log-scroll-container::-webkit-scrollbar-thumb {
		background: var(--color-muted-foreground);
		opacity: 0.3;
		border-radius: 3px;
	}

	.log-scroll-container::-webkit-scrollbar-thumb:hover {
		opacity: 0.5;
	}
</style>
