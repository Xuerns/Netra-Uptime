<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { CheckCircle2 } from 'lucide-svelte';
	import { formatDateTime } from '$lib/utils/datetime';

	interface Props {
		orgId: string;
		incidents: Incident[];
		tz?: string;
	}

	let { orgId, incidents = [], tz = 'UTC' }: Props = $props();
</script>

<div class="bg-card border border-border rounded-lg p-5 md:p-6 space-y-4">
	<div class="flex items-center justify-between">
		<h3 class="text-sm font-semibold text-foreground tracking-tight flex items-center gap-2">
			<span>Recent Outages & Incidents</span>
		</h3>

		<span class="text-xs text-muted-foreground font-mono">
			{incidents.length} incident{incidents.length === 1 ? '' : 's'} recorded
		</span>
	</div>

	{#if incidents.length === 0}
		<div class="bg-white/5 rounded-lg p-6 text-center space-y-2">
			<CheckCircle2 class="size-8 text-emerald-400 mx-auto" />
			<p class="text-sm font-medium text-foreground">All Systems Operational</p>
			<p class="text-xs text-muted-foreground">
				No downtime incidents recorded during this timeframe.
			</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each incidents as inc (inc.ID)}
				<a
					href="/dashboard/org/{orgId}/incidents/{inc.ID}"
					class="flex items-center justify-between bg-white/5 border border-border rounded-lg p-4 transition-colors hover:bg-white/8 hover:border-primary/30 group"
				>
					<div class="space-y-1">
						<div class="flex items-center gap-2">
							<span
								class="size-2 rounded-full {inc.Status === 'resolved'
									? 'bg-emerald-400'
									: 'bg-rose-500'}"
							></span>
							<span
								class="text-sm font-semibold text-foreground group-hover:text-primary transition-colors"
							>
								{inc.Summary || 'Service Disruption'}
							</span>
							<span
								class="uppercase text-[10px] font-mono px-2 py-0.5 rounded bg-white/5 border border-border text-muted-foreground"
							>
								{inc.Status}
							</span>
						</div>
					</div>

					<div class="text-right text-xs text-muted-foreground font-mono shrink-0">
						<div>Started: {formatDateTime(inc.StartedAt, tz)}</div>
						{#if inc.ResolvedAt}
							<div class="text-emerald-400">Resolved: {formatDateTime(inc.ResolvedAt, tz)}</div>
						{/if}
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
