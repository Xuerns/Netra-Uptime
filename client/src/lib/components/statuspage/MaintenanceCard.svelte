<script lang="ts">
	import type { PublicMaintenance } from '$lib/api/statuspage';
	import { Wrench, Calendar, Clock } from 'lucide-svelte';
	import { formatDateTimeShort } from '$lib/utils/datetime';

	let { maintenance, isOngoing = false } = $props<{
		maintenance: PublicMaintenance;
		isOngoing?: boolean;
	}>();

	function formatDateTimeRange(startStr: string, endStr: string): string {
		if (!startStr) return '';
		const startFmt = formatDateTimeShort(startStr);
		if (!endStr) return startFmt;
		const endFmt = formatDateTimeShort(endStr);
		return `${startFmt} – ${endFmt}`;
	}

	function timeUntil(startStr: string): string {
		if (!startStr) return '';
		const start = new Date(startStr).getTime();
		const now = new Date().getTime();
		const diffMs = start - now;
		if (diffMs <= 0) return 'starting now';
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		if (diffHours < 1) {
			const diffMins = Math.floor(diffMs / (1000 * 60));
			return `in ${diffMins}m`;
		}
		if (diffHours < 24) {
			return `in ${diffHours}h`;
		}
		const diffDays = Math.floor(diffHours / 24);
		return `in ${diffDays} day${diffDays > 1 ? 's' : ''}`;
	}
</script>

<div
	class="rounded-xl border transition-all duration-200 backdrop-blur-md overflow-hidden p-5 sm:p-6 space-y-3 {isOngoing
		? 'bg-blue-500/10 border-blue-500/30 text-blue-300'
		: 'bg-indigo-500/5 border-indigo-500/20 text-indigo-300'}"
>
	<!-- Card Top -->
	<div class="flex flex-wrap items-center justify-between gap-2">
		<div class="flex items-center gap-2">
			{#if isOngoing}
				<span
					class="inline-flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-blue-500/20 text-blue-300 border border-blue-500/30"
				>
					<span class="size-1.5 rounded-full bg-blue-400"></span>
					In Progress
				</span>
			{:else}
				<span
					class="inline-flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
				>
					<Calendar class="size-3" />
					Scheduled ({timeUntil(maintenance.scheduled_start)})
				</span>
			{/if}
			<span class="text-xs text-muted-foreground font-mono flex items-center gap-1">
				<Clock class="size-3" />
				{formatDateTimeRange(maintenance.scheduled_start, maintenance.scheduled_end)}
			</span>
		</div>
	</div>

	<!-- Title & Description -->
	<div>
		<h3 class="text-base sm:text-lg font-semibold text-foreground flex items-center gap-2">
			<Wrench class="size-4 text-blue-400 shrink-0" />
			{maintenance.title}
		</h3>
		{#if maintenance.description}
			<p
				class="text-xs sm:text-sm text-muted-foreground mt-1.5 leading-relaxed whitespace-pre-line"
			>
				{maintenance.description}
			</p>
		{/if}
	</div>

	<!-- Affected Monitors -->
	{#if maintenance.affected_monitors && maintenance.affected_monitors.length > 0}
		<div class="pt-2 border-t border-border/40 flex flex-wrap items-center gap-2">
			<span class="text-[11px] text-muted-foreground font-medium">Affected Services:</span>
			{#each maintenance.affected_monitors as monitorName (monitorName)}
				<span
					class="text-[11px] font-mono px-2 py-0.5 rounded bg-white/5 border border-border/80 text-foreground/80"
				>
					{monitorName}
				</span>
			{/each}
		</div>
	{/if}
</div>
