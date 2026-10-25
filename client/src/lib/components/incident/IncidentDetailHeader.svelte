<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { formatDateTime, formatIncidentTimeAgo, formatDuration } from '$lib/utils/datetime';
	import { CheckCircle2, Radio, Users, Loader2, Server, Clock, ExternalLink } from 'lucide-svelte';

	interface Props {
		incident: Incident;
		currentUserId: string;
		isJoining: boolean;
		tz: string;
		onJoin: () => void;
	}

	let { incident, currentUserId, isJoining, tz, onJoin }: Props = $props();

	let isOngoing = $derived(incident.Status === 'ongoing');
	let isAssigned = $derived(incident.Assignees?.some((a) => a.UserID === currentUserId));
	let formattedStartedAt = $derived(formatDateTime(incident.StartedAt, tz));
	let formattedResolvedAt = $derived(
		incident.ResolvedAt ? formatDateTime(incident.ResolvedAt, tz) : null
	);
	let timeAgo = $derived(formatIncidentTimeAgo(incident.StartedAt));
	let impactDuration = $derived(formatDuration(incident.StartedAt, incident.ResolvedAt));
</script>

<!-- Header Content Block (Flat, no card container) -->
<div class="flex flex-col gap-6">
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
		<!-- Title & Badges -->
		<div class="space-y-2 min-w-0">
			<div class="flex items-center gap-2.5 flex-wrap">
				<!-- Status Chip -->
				<span
					class="px-2.5 py-0.5 rounded-full text-xs font-semibold font-mono flex items-center gap-1.5 {isOngoing
						? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
						: 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'}"
				>
					{#if isOngoing}
						<Radio class="size-3 text-rose-400" />
						<span>ONGOING</span>
					{:else}
						<CheckCircle2 class="size-3 text-emerald-400" />
						<span>RESOLVED</span>
					{/if}
				</span>

				<!-- Severity Tag -->
				{#if incident.SeverityCode}
					<span
						class="px-2.5 py-0.5 rounded-lg text-xs font-semibold uppercase font-mono border {incident.SeverityCode ===
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

				<span class="text-xs text-muted-foreground font-mono">#{incident.ID.substring(0, 8)}</span>
			</div>

			<h1 class="text-2xl md:text-3xl font-bold text-foreground tracking-tight break-words">
				{incident.Summary || 'Service Disruption'}
			</h1>
		</div>

		<!-- Action Button -->
		{#if isOngoing}
			<div class="shrink-0">
				{#if isAssigned}
					<div
						class="px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-lg text-xs font-semibold font-mono flex items-center gap-2"
					>
						<Users class="size-3.5" />
						<span>You are on this bridge</span>
					</div>
				{:else}
					<button
						class="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-xs font-semibold hover:bg-primary/90 transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
						disabled={isJoining}
						onclick={onJoin}
					>
						{#if isJoining}
							<Loader2 class="size-3.5 animate-spin" />
							<span>Joining...</span>
						{:else}
							<Users class="size-3.5" />
							<span>Join Incident Bridge</span>
						{/if}
					</button>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Metadata Strip (3 columns with vertical hairline dividers on sm+) -->
	<div class="grid grid-cols-1 sm:grid-cols-3 gap-y-4 pt-4 border-t border-border/60">
		<!-- 1. Target Monitor -->
		<div class="flex flex-col gap-1 pr-4 min-w-0">
			<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
				Target Monitor
			</span>
			{#if incident.Monitor}
				<a
					href="/dashboard/org/{incident.Monitor?.OrganizationID}/monitors/{incident.MonitorID}"
					class="inline-flex items-center gap-1.5 text-foreground hover:text-primary font-semibold text-sm transition-colors truncate mt-0.5"
				>
					<Server class="size-3.5 text-primary shrink-0" />
					<span class="truncate">{incident.Monitor.Name}</span>
					<ExternalLink class="size-3 text-muted-foreground shrink-0" />
				</a>
				<div class="text-xs font-mono text-muted-foreground truncate">
					{incident.Monitor.Target}
				</div>
			{:else}
				<span class="text-xs text-muted-foreground">Unknown Monitor ({incident.MonitorID})</span>
			{/if}
		</div>

		<!-- 2. Started -->
		<div class="flex flex-col gap-1 sm:border-l sm:border-border/60 sm:px-4 min-w-0">
			<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
				Started At
			</span>
			<div class="flex items-center gap-1.5 text-foreground font-semibold text-sm mt-0.5">
				<Clock class="size-3.5 text-amber-400 shrink-0" />
				<span class="truncate">{formattedStartedAt}</span>
			</div>
			<span class="text-xs text-muted-foreground font-mono"> Incident opened </span>
		</div>

		<!-- 3. Duration & Status -->
		<div class="flex flex-col gap-1 sm:border-l sm:border-border/60 sm:pl-4 min-w-0">
			<span class="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
				Duration & Status
			</span>
			<div class="text-sm font-semibold text-foreground mt-0.5">
				{#if formattedResolvedAt}
					<span class="text-emerald-400">Resolved</span>
				{:else}
					<span class="text-rose-400">Active since {timeAgo}</span>
				{/if}
			</div>
			<div class="text-xs text-muted-foreground">
				{#if formattedResolvedAt}
					<span>Resolved at {formattedResolvedAt} ({impactDuration})</span>
				{:else}
					<span>Impact duration: {impactDuration}</span>
				{/if}
			</div>
		</div>
	</div>
</div>
