<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { formatDateTime, formatIncidentTimeAgo } from '$lib/utils/datetime';
	import {
		AlertTriangle,
		CheckCircle2,
		Radio,
		Users,
		Loader2,
		Server,
		Clock,
		ExternalLink
	} from 'lucide-svelte';

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
</script>

<!-- Header Content Block -->
<div class="bg-card border border-border rounded-lg p-5 md:p-6 space-y-4">
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
		<!-- Title & Badges -->
		<div class="space-y-2">
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

				<span class="text-xs text-foreground/40 font-mono">#{incident.ID.substring(0, 8)}</span>
			</div>

			<h1 class="text-xl md:text-2xl font-semibold text-foreground tracking-tight">
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

	<!-- Target Monitor & Timeline -->
	<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs font-mono">
		<!-- Target Monitor -->
		<div>
			<span class="text-foreground/40 block text-[11px] uppercase tracking-wider"
				>Target Monitor</span
			>
			{#if incident.Monitor}
				<a
					href="/dashboard/org/{incident.Monitor?.OrganizationID}/monitors/{incident.MonitorID}"
					class="inline-flex items-center gap-1.5 text-foreground hover:text-primary font-semibold transition-colors mt-0.5"
				>
					<Server class="size-3.5 text-primary" />
					<span class="truncate">{incident.Monitor.Name}</span>
					<ExternalLink class="size-3 text-foreground/40" />
				</a>
				<div class="text-[11px] text-muted-foreground truncate">
					{incident.Monitor.Target}
				</div>
			{:else}
				<span class="text-foreground/60">Unknown Monitor ({incident.MonitorID})</span>
			{/if}
		</div>

		<!-- Timeline Status -->
		<div>
			<span class="text-foreground/40 block text-[11px] uppercase tracking-wider"
				>Timeline Status</span
			>
			<div class="flex items-center gap-1.5 text-foreground font-semibold">
				<Clock class="size-3.5 text-amber-400" />
				<span>Started: {formattedStartedAt}</span>
			</div>
			<div class="text-[11px] text-muted-foreground">
				{#if formattedResolvedAt}
					<span class="text-emerald-400">Resolved at {formattedResolvedAt}</span>
				{:else}
					<span>Impact duration: {timeAgo}</span>
				{/if}
			</div>
		</div>
	</div>
</div>
