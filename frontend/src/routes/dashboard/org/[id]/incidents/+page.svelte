<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { IncidentAPI, type Incident, type IncidentUpdateResponse } from '$lib/api/incident';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateTime, formatTimeOnly, formatIncidentTimeAgo } from '$lib/utils/datetime';
	import { OrgAPI } from '$lib/api/organization';
	import IncidentSkeleton from '$lib/components/skeletons/IncidentSkeleton.svelte';
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
		Radio
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	let allIncidents: Incident[] = $state([]);
	let isLoading = $state(true);
	let error = $state('');
	let currentOrgId = $state('');

	let activeFilter = $state<'active' | 'resolved'>('active');
	let searchQuery = $state('');
	let severityFilter = $state<'all' | 'sev1' | 'sev2' | 'sev3'>('all');
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
			if (searchQuery.trim()) {
				const q = searchQuery.toLowerCase();
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
		searchQuery;
		severityFilter;
		currentPage = 1;
	});

	onMount(async () => {
		try {
			currentOrgId = ($page.params.id as string) || '';
			if (currentOrgId) {
				await loadIncidents(currentOrgId);
			}
		} catch (err: any) {
			error = err.message;
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
	<title>Incident Command Board - PulseWatch</title>
</svelte:head>

{#if isLoading}
	<IncidentSkeleton />
{:else if error}
	<div class="p-4 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg">
		{error}
	</div>
{:else}
	<!-- Search Control Bar Above Table -->
	<div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 mb-4">
		<!-- Search Input -->
		<div class="relative flex-1">
			<Search class="absolute left-3.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search incidents by summary, monitor, or ID..."
				class="w-full bg-card border border-border rounded-lg pl-10 pr-3.5 py-2 text-xs text-foreground placeholder:text-foreground/30 focus:outline-none focus:border-primary transition-colors"
			/>
		</div>

		<!-- Severity Filter Select Dropdown -->
		<select
			bind:value={severityFilter}
			class="bg-card border border-border rounded-lg px-3.5 py-2 text-xs text-foreground focus:outline-none focus:border-primary transition-colors cursor-pointer capitalize shrink-0"
		>
			<option value="all">All Severities</option>
			<option value="sev1">Sev 1 (Critical)</option>
			<option value="sev2">Sev 2 (Major)</option>
			<option value="sev3">Sev 3 (Minor)</option>
		</select>
	</div>

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
						class="size-3.5 {activeFilter === 'active'
							? 'text-rose-400'
							: 'text-foreground/40'}"
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

			<!-- Right: Stats & Ordering -->
			<div class="flex items-center gap-3 text-[11.5px] font-mono text-foreground/40">
				<span
					>Showing {filteredIncidents.length} incident{filteredIncidents.length === 1
						? ''
						: 's'}</span
				>
				{#if searchQuery || severityFilter !== 'all'}
					<span class="text-primary font-sans">(filtered)</span>
				{/if}
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
													: 'bg-sky-500/10 text-sky-400 border-sky-500/20'}"
										>
											{incident.Severity || incident.SeverityCode}
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
											<span>{incident.Assignees.length} PIC</span>
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
			<div
				class="bg-white/3 border-t border-border px-5 py-3 flex items-center justify-between text-xs text-muted-foreground flex-wrap gap-3"
			>
				<div class="font-mono">
					Showing <span class="text-foreground">{(currentPage - 1) * itemsPerPage + 1}</span>
					to{' '}
					<span class="text-foreground"
						>{Math.min(currentPage * itemsPerPage, filteredIncidents.length)}</span
					>
					of <span class="text-foreground">{filteredIncidents.length}</span> incidents
				</div>

				<div class="flex items-center gap-2">
					<button
						class="p-1.5 rounded-lg border border-border bg-white/5 hover:bg-white/10 hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer"
						disabled={currentPage <= 1}
						onclick={() => (currentPage = Math.max(1, currentPage - 1))}
						aria-label="Previous page"
					>
						<ChevronLeft class="size-4" />
					</button>

					<span class="font-mono text-xs text-foreground px-2">
						Page {currentPage} of {totalPages}
					</span>

					<button
						class="p-1.5 rounded-lg border border-border bg-white/5 hover:bg-white/10 hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-all cursor-pointer"
						disabled={currentPage >= totalPages}
						onclick={() => (currentPage = Math.min(totalPages, currentPage + 1))}
						aria-label="Next page"
					>
						<ChevronRight class="size-4" />
					</button>
				</div>
			</div>
		{/if}
	</div>
{/if}
