<script lang="ts">
	import type { Incident } from '$lib/api/incident';
	import { ArrowLeft, AlertTriangle, CheckCircle2, Radio, Users, Loader2 } from 'lucide-svelte';

	interface Props {
		incident: Incident;
		currentUserId: string;
		isJoining: boolean;
		onJoin: () => void;
	}

	let { incident, currentUserId, isJoining, onJoin }: Props = $props();

	let isOngoing = $derived(incident.Status === 'ongoing');
	let isAssigned = $derived(incident.Assignees?.some((a) => a.UserID === currentUserId));
</script>

<div class="space-y-4 mb-6">
	<!-- Breadcrumb -->
	<a
		href="/dashboard/incidents"
		class="inline-flex items-center gap-2 text-xs font-mono text-on-surface-variant hover:text-white transition-colors"
	>
		<ArrowLeft class="size-3.5" />
		<span>Back to Incidents</span>
	</a>

	<!-- Header Content Block -->
	<div class="bg-[#161618] border border-white/10 rounded-lg p-5 md:p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
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
						class="px-2.5 py-0.5 rounded-lg text-xs font-semibold uppercase font-mono border {incident.SeverityCode === 'sev1'
							? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
							: incident.SeverityCode === 'sev2'
								? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
								: 'bg-sky-500/10 text-sky-400 border-sky-500/20'}"
					>
						{incident.Severity || incident.SeverityCode}
					</span>
				{/if}

				<span class="text-xs text-white/40 font-mono">#{incident.ID.substring(0, 8)}</span>
			</div>

			<h1 class="text-xl md:text-2xl font-semibold text-white tracking-tight">
				{incident.Summary || 'Service Disruption'}
			</h1>
		</div>

		<!-- Action Button -->
		{#if isOngoing}
			<div class="shrink-0">
				{#if isAssigned}
					<div class="px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-lg text-xs font-semibold font-mono flex items-center gap-2">
						<Users class="size-3.5" />
						<span>You are on this bridge</span>
					</div>
				{:else}
					<button
						class="px-4 py-2 bg-primary text-on-primary rounded-lg text-xs font-semibold hover:bg-primary-container transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50"
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
</div>
