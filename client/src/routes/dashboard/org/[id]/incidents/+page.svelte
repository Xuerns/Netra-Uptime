<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { IncidentAPI, type Incident, type IncidentUpdateResponse } from '$lib/api/incident';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateTime, formatTimeOnly, formatIncidentTimeAgo } from '$lib/utils/datetime';
	import { OrgAPI } from '$lib/api/organization';
	import { Pagination, DataSkeleton } from '$lib/components/common';
	import {
		AlertTriangle,
		Clock,
		Server,
		CheckCircle2,
		Activity,
		Users,
		ChevronDown,
		ChevronUp,
		Search,
		ChevronLeft,
		ChevronRight,
		Radio,
		Check
	} from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { toast } from 'svelte-sonner';

	let allIncidents: Incident[] = $state([]);
	let isLoading = $state(true);
	let error = $state('');
	let currentOrgId = $state('');

	let activeFilter = $state<'active' | 'resolved'>('active');
	let severityFilter = $state<'all' | 'sev1' | 'sev2' | 'sev3' | 'sev4'>('all');
	let currentPage = $state(1);
	let itemsPerPage = $state(10);

	let currentUserId = $derived($page.data.user?.id || $page.data.user?.ID);

	async function loadIncidents(orgId: string) {
		const res = await IncidentAPI.getAll(orgId);
		allIncidents = res.data || [];
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

	let paginatedIncidents = $derived.by(() => {
		const start = (currentPage - 1) * itemsPerPage;
		return filteredIncidents.slice(start, start + itemsPerPage);
	});

	// Reset page when filters change
	$effect(() => {
		activeFilter;
		headerStore.searchQuery;
		severityFilter;
		currentPage = 1;
	});

	// Sync with layout top bar
	$effect(() => {
		const activeCount = allIncidents.filter((i) => i.Status !== 'resolved').length;
		headerStore.set({
			showSearch: true,
			searchPlaceholder: 'Search incidents by summary, monitor...',
			badge: {
				text:
					activeCount === 0
						? 'All systems operational'
						: `${activeCount} Active incident${activeCount > 1 ? 's' : ''}`,
				variant: activeCount === 0 ? 'emerald' : 'rose'
			}
		});
	});

	onMount(async () => {
		try {
			currentOrgId = ($page.params.id as string) || '';
			if (currentOrgId) {
				await loadIncidents(currentOrgId);
			}
		} catch (e: any) {
			error = e.message || 'Failed to load incidents';
		} finally {
			isLoading = false;
		}
	});

	let tz = $derived($userTimezone);

	function fmtDateTime(d: string) {
		return formatDateTime(d, tz);
	}
	function fmtTimeOnly(d: string) {
		return formatTimeOnly(d, tz);
	}
	function fmtTimeAgo(d: string) {
		return formatIncidentTimeAgo(d);
	}
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
	<!-- GitHub Style Table List Box -->
	<div class="bg-card border border-border rounded-lg overflow-hidden">
		<!-- Table Header Bar with GitHub-Style Active/Resolved Toggles -->
		<div
			class="bg-white/3 border-b border-border px-5 py-3.5 flex items-center justify-between text-xs text-muted-foreground"
		>
			<!-- Left: GitHub-style Status Filter Tabs (Active vs Resolved) -->
			<div class="flex items-center gap-6">
				<!-- Active / Open Tab -->
				<button
					class="flex items-center gap-2 transition-colors cursor-pointer {activeFilter === 'active'
						? 'font-bold text-foreground'
						: 'font-medium text-foreground/40 hover:text-foreground/70'}"
					onclick={() => (activeFilter = 'active')}
				>
					<Radio
						class="size-3.5 {activeFilter === 'active' ? 'text-rose-400' : 'text-foreground/40'}"
					/>
					<span>Active</span>
					<span
						class="px-2 py-0.5 text-[11px] rounded-full font-mono font-semibold transition-colors {activeFilter ===
						'active'
							? 'bg-rose-500/20 text-rose-300'
							: 'bg-white/5 text-foreground/40'}"
					>
						{activeCount}
					</span>
				</button>

				<!-- Resolved / Closed Tab -->
				<button
					class="flex items-center gap-2 transition-colors cursor-pointer {activeFilter ===
					'resolved'
						? 'font-bold text-foreground'
						: 'font-medium text-foreground/40 hover:text-foreground/70'}"
					onclick={() => (activeFilter = 'resolved')}
				>
					<CheckCircle2
						class="size-3.5 {activeFilter === 'resolved'
							? 'text-emerald-400'
							: 'text-foreground/40'}"
					/>
					<span>Resolved</span>
					<span
						class="px-2 py-0.5 text-[11px] rounded-full font-mono font-semibold transition-colors {activeFilter ===
						'resolved'
							? 'bg-emerald-500/20 text-emerald-300'
							: 'bg-white/5 text-foreground/40'}"
					>
						{resolvedCount}
					</span>
				</button>
			</div>

			<!-- Right: Severity Filter Dropdown & Stats -->
			<div class="flex items-center gap-3.5 text-xs">
				<!-- Severity Dropdown Filter -->
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="flex items-center gap-2 px-2.5 py-1.5 rounded-md border text-xs transition-colors cursor-pointer focus:outline-none {severityFilter !==
						'all'
							? 'bg-white/10 border-border text-foreground font-medium'
							: 'bg-transparent border-transparent hover:bg-white/5 text-muted-foreground hover:text-foreground'}"
					>
						<span class="flex items-center gap-1.5">
							{#if severityFilter === 'sev1'}
								<span class="size-2 rounded-full bg-rose-500"></span>
								<span>Critical (Sev 1)</span>
							{:else if severityFilter === 'sev2'}
								<span class="size-2 rounded-full bg-amber-500"></span>
								<span>High (Sev 2)</span>
							{:else if severityFilter === 'sev3'}
								<span class="size-2 rounded-full bg-sky-500"></span>
								<span>Medium (Sev 3)</span>
							{:else if severityFilter === 'sev4'}
								<span class="size-2 rounded-full bg-slate-400"></span>
								<span>Low (Sev 4)</span>
							{:else}
								<span>Severity: All</span>
							{/if}
						</span>
						<ChevronDown class="size-3 text-muted-foreground" />
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						align="end"
						class="w-48 p-1.5 bg-card border-border text-foreground shadow-xl rounded-lg"
					>
						<DropdownMenu.Item
							class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
							'all'
								? 'text-primary font-medium'
								: 'text-foreground'}"
							onclick={() => (severityFilter = 'all')}
						>
							<span>All Severities</span>
							{#if severityFilter === 'all'}
								<Check class="size-3.5 text-primary" />
							{/if}
						</DropdownMenu.Item>
						<DropdownMenu.Separator class="bg-border my-1" />
						<DropdownMenu.Item
							class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
							'sev1'
								? 'text-rose-400 font-medium'
								: 'text-foreground'}"
							onclick={() => (severityFilter = 'sev1')}
						>
							<div class="flex items-center gap-2">
								<span class="size-2 rounded-full bg-rose-500"></span>
								<span>Sev 1 (Critical)</span>
							</div>
							{#if severityFilter === 'sev1'}
								<Check class="size-3.5 text-rose-400" />
							{/if}
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
							'sev2'
								? 'text-amber-400 font-medium'
								: 'text-foreground'}"
							onclick={() => (severityFilter = 'sev2')}
						>
							<div class="flex items-center gap-2">
								<span class="size-2 rounded-full bg-amber-500"></span>
								<span>Sev 2 (High)</span>
							</div>
							{#if severityFilter === 'sev2'}
								<Check class="size-3.5 text-amber-400" />
							{/if}
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
							'sev3'
								? 'text-sky-400 font-medium'
								: 'text-foreground'}"
							onclick={() => (severityFilter = 'sev3')}
						>
							<div class="flex items-center gap-2">
								<span class="size-2 rounded-full bg-sky-500"></span>
								<span>Sev 3 (Medium)</span>
							</div>
							{#if severityFilter === 'sev3'}
								<Check class="size-3.5 text-sky-400" />
							{/if}
						</DropdownMenu.Item>
						<DropdownMenu.Item
							class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
							'sev4'
								? 'text-slate-300 font-medium'
								: 'text-foreground'}"
							onclick={() => (severityFilter = 'sev4')}
						>
							<div class="flex items-center gap-2">
								<span class="size-2 rounded-full bg-slate-400"></span>
								<span>Sev 4 (Low)</span>
							</div>
							{#if severityFilter === 'sev4'}
								<Check class="size-3.5 text-slate-300" />
							{/if}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>

				<!-- Subtle Vertical Separator -->
				<div class="h-4 w-px bg-border/60"></div>

				<!-- Stats & Count -->
				<div class="text-[11.5px] font-mono text-foreground/40 whitespace-nowrap">
					<span
						>Showing {filteredIncidents.length} incident{filteredIncidents.length === 1
							? ''
							: 's'}</span
					>
					{#if headerStore.searchQuery || severityFilter !== 'all'}
						<span class="text-primary font-sans ml-1">(filtered)</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Incident Rows Container -->
		<div class="divide-y divide-white/5">
			{#each paginatedIncidents as incident}
				{@const isOngoing = incident.Status === 'ongoing'}

				<div class="transition-colors hover:bg-white/5">
					<!-- Main Row Content -->
					<div class="p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
						<!-- Left: Status Icon + Info -->
						<div class="flex items-start gap-3.5 min-w-0 flex-1">
							<!-- Status Icon Dot -->
							<div class="pt-1 shrink-0">
								{#if isOngoing}
									<span class="relative flex size-3">
										<span class="inline-flex rounded-full size-3 bg-rose-500"></span>
									</span>
								{:else}
									<CheckCircle2 class="size-4 text-emerald-400 shrink-0" />
								{/if}
							</div>

							<!-- Incident Title, Severity & Meta -->
							<div class="space-y-1.5 min-w-0 flex-1">
								<div class="flex items-center gap-2.5 flex-wrap">
									<!-- Severity Pill -->
									{#if incident.SeverityCode}
										<span
											class="px-2 py-0.5 rounded text-[10.5px] font-semibold uppercase tracking-wider font-mono border {incident.SeverityCode ===
											'sev1'
												? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
												: incident.SeverityCode === 'sev2'
													? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
													: incident.SeverityCode === 'sev3'
														? 'bg-sky-500/10 text-sky-400 border-sky-500/20'
														: 'bg-slate-500/10 text-slate-400 border-slate-500/20'}"
										>
											{incident.SeverityLabel || incident.Severity || incident.SeverityCode}
										</span>
									{/if}

									<!-- Title / Summary -->
									<a
										href="/dashboard/org/{currentOrgId}/incidents/{incident.ID}"
										class="text-sm font-semibold text-foreground hover:text-primary transition-colors truncate"
									>
										{incident.Summary || 'Service Disruption'}
									</a>

									<!-- Target Monitor Tag -->
									{#if incident.Monitor}
										<a
											href="/dashboard/org/{currentOrgId}/monitors/{incident.MonitorID}"
											class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-white/5 border border-border text-[11px] font-mono text-muted-foreground hover:text-foreground hover:border-border transition-all"
										>
											<Server class="size-3 text-primary" />
											<span class="truncate max-w-36">{incident.Monitor.Name}</span>
										</a>
									{/if}
								</div>

								<!-- Meta Details Row -->
								<div
									class="flex items-center gap-3 text-xs text-muted-foreground flex-wrap font-mono"
								>
									<span class="text-foreground/40">#{incident.ID.substring(0, 8)}</span>
									<span>&middot;</span>
									<span class="flex items-center gap-1">
										<Clock class="size-3" />
										<span
											>Started: {fmtDateTime(incident.StartedAt)} ({fmtTimeAgo(
												incident.StartedAt
											)})</span
										>
									</span>

									{#if !isOngoing && incident.ResolvedAt}
										<span>&middot;</span>
										<span class="text-emerald-400"
											>Resolved: {fmtDateTime(incident.ResolvedAt)}</span
										>
									{/if}

									{#if incident.Assignees && incident.Assignees.length > 0}
										<span>&middot;</span>
										<span class="flex items-center gap-1">
											<Users class="size-3 text-primary" />
											<span
												>{incident.Assignees.length}
												{incident.Assignees.length === 1 ? 'Assignee' : 'Assignees'}</span
											>
										</span>
									{/if}
								</div>
							</div>
						</div>

						<!-- Right Actions Group: Direct link to Incident Command Room -->
						<div
							class="flex items-center gap-2.5 shrink-0 self-end md:self-center group transition-colors"
						>
							<a
								href="/dashboard/org/{currentOrgId}/incidents/{incident.ID}"
								class="px-3.5 py-1.5 text-foreground/60 group-hover:text-foreground text-xs font-semibold transition-colors flex items-center gap-1.5 cursor-pointer"
							>
								<span>Command Room</span>
								<ChevronRight
									class="size-3.5 text-foreground/60 group-hover:text-foreground transition-colors"
								/>
							</a>
						</div>
					</div>
				</div>
			{/each}

			<!-- Empty State when filter/search yields 0 results -->
			{#if paginatedIncidents.length === 0}
				<div class="p-12 text-center space-y-3">
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
			<div class="bg-white/3 border-t border-border px-5 py-3">
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
{/if}
