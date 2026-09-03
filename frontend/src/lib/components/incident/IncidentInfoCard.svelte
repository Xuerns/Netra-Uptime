<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { formatDateTime, formatIncidentTimeAgo } from '$lib/utils/datetime';
	import { Server, Clock, Activity, ExternalLink } from 'lucide-svelte';

	interface Props {
		incident: Incident;
		tz: string;
	}

	let { incident, tz }: Props = $props();

	let formattedStartedAt = $derived(formatDateTime(incident.StartedAt, tz));
	let formattedResolvedAt = $derived(
		incident.ResolvedAt ? formatDateTime(incident.ResolvedAt, tz) : null
	);
	let timeAgo = $derived(formatIncidentTimeAgo(incident.StartedAt));
</script>

<div class="bg-[#161618] border border-white/10 rounded-lg p-5 space-y-4">
	<h2 class="text-sm font-semibold text-white tracking-tight flex items-center gap-2">
		<Activity class="size-4 text-primary" />
		<span>Incident Target & Context</span>
	</h2>

	<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs font-mono">
		<!-- Target Monitor -->
		<div class="bg-black/30 border border-white/5 rounded-lg p-3.5 space-y-1">
			<span class="text-white/40 block text-[11px] uppercase tracking-wider">Target Monitor</span>
			{#if incident.Monitor}
				<a
					href="/dashboard/monitor/{incident.MonitorID}"
					class="inline-flex items-center gap-1.5 text-white hover:text-primary font-semibold transition-colors mt-0.5"
				>
					<Server class="size-3.5 text-primary" />
					<span class="truncate">{incident.Monitor.Name}</span>
					<ExternalLink class="size-3 text-white/40" />
				</a>
				<div class="text-[11px] text-on-surface-variant truncate">
					{incident.Monitor.Target}
				</div>
			{:else}
				<span class="text-white/60">Unknown Monitor ({incident.MonitorID})</span>
			{/if}
		</div>

		<!-- Timing Details -->
		<div class="bg-black/30 border border-white/5 rounded-lg p-3.5 space-y-1">
			<span class="text-white/40 block text-[11px] uppercase tracking-wider">Timeline Status</span>
			<div class="flex items-center gap-1.5 text-white font-semibold">
				<Clock class="size-3.5 text-amber-400" />
				<span>Started: {formattedStartedAt}</span>
			</div>
			<div class="text-[11px] text-on-surface-variant">
				{#if formattedResolvedAt}
					<span class="text-emerald-400">Resolved at {formattedResolvedAt}</span>
				{:else}
					<span>Impact duration: {timeAgo}</span>
				{/if}
			</div>
		</div>
	</div>
</div>
