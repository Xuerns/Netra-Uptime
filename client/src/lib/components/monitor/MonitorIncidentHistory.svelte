<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { CheckCircle2 } from 'lucide-svelte';
	import { formatDateTime } from '$lib/utils/datetime';
	import { Pagination } from '$lib/components/common';

	interface Props {
		orgId: string;
		incidents: Incident[];
		tz?: string;
		itemsPerPage?: number;
	}

	let { orgId, incidents = [], tz = 'UTC', itemsPerPage = 6 }: Props = $props();

	let currentPage = $state(1);
	let totalPages = $derived(Math.max(1, Math.ceil(incidents.length / itemsPerPage)));
	let safeCurrentPage = $derived(Math.min(Math.max(1, currentPage), totalPages));
	let paginatedIncidents = $derived.by(() => {
		const start = (safeCurrentPage - 1) * itemsPerPage;
		return incidents.slice(start, start + itemsPerPage);
	});

	$effect(() => {
		if (currentPage > totalPages) {
			currentPage = totalPages;
		}
	});
</script>

<div class="flex flex-col gap-4 min-w-0">
	<div class="flex items-center justify-between gap-3 min-h-7">
		<h3 class="text-sm font-semibold text-foreground tracking-tight">Recent Incidents</h3>

		<span class="text-xs text-muted-foreground font-mono shrink-0">
			{incidents.length} recorded
		</span>
	</div>

	{#if incidents.length === 0}
		<div class="py-10 text-center space-y-2">
			<CheckCircle2 class="size-8 text-emerald-400 mx-auto" />
			<p class="text-sm font-medium text-foreground">All Systems Operational</p>
			<p class="text-xs text-muted-foreground">
				No downtime incidents recorded during this timeframe.
			</p>
		</div>
	{:else}
		<ul class="divide-y divide-border/60">
			{#each paginatedIncidents as inc (inc.ID)}
				<li>
					<a
						href="/dashboard/org/{orgId}/incidents/{inc.ID}"
						class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 sm:gap-4 py-3 px-2 -mx-2 rounded-md transition-colors hover:bg-muted/40 group"
					>
						<div class="flex items-start sm:items-center gap-2.5 min-w-0">
							<span
								class="size-2 rounded-full shrink-0 mt-1.5 sm:mt-0 {inc.Status === 'resolved'
									? 'bg-emerald-400'
									: 'bg-rose-500'}"
							></span>
							<span
								class="text-sm font-medium text-foreground group-hover:text-primary transition-colors min-w-0 sm:truncate"
							>
								{inc.Summary || 'Service Disruption'}
							</span>
							<span
								class="uppercase text-[10px] font-mono px-1.5 py-0.5 rounded shrink-0 {inc.Status ===
								'resolved'
									? 'bg-emerald-500/10 text-emerald-400'
									: 'bg-rose-500/10 text-rose-400'}"
							>
								{inc.Status}
							</span>
						</div>

						<div
							class="pl-4.5 sm:pl-0 sm:text-right text-[11px] text-muted-foreground font-mono shrink-0 leading-relaxed"
						>
							<div>Started: {formatDateTime(inc.StartedAt, tz)}</div>
							{#if inc.ResolvedAt}
								<div class="text-emerald-400">Resolved: {formatDateTime(inc.ResolvedAt, tz)}</div>
							{/if}
						</div>
					</a>
				</li>
			{/each}
		</ul>

		<Pagination bind:currentPage {totalPages} totalItems={incidents.length} itemName="incidents" />
	{/if}
</div>
