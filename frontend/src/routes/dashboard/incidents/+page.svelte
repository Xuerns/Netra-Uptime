<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { IncidentAPI, type Incident, type IncidentUpdateResponse } from '$lib/api/incident';
	import { OrgAPI } from '$lib/api/organization';
	import {
		AlertTriangle,
		AlertCircle,
		Info,
		Clock,
		Server,
		CheckCircle,
		Activity,
		Users,
		ChevronDown,
		ChevronUp
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	let activeIncidents: Incident[] = $state([]);
	let resolvedIncidents: Incident[] = $state([]);
	let isLoading = $state(true);
	let error = $state('');
	let currentOrgId = $state('');

	let currentUserId = $derived($page.data.user?.id || $page.data.user?.ID);

	async function loadIncidents(orgId: string) {
		const res = await IncidentAPI.getAll(orgId);
		const all = res.data || [];
		activeIncidents = all.filter((inc) => inc.Status === 'ongoing');
		resolvedIncidents = all.filter((inc) => inc.Status === 'resolved');
	}

	async function handleJoin(incidentId: string) {
		try {
			if (!currentOrgId) return;
			await IncidentAPI.joinAssignee(currentOrgId, incidentId);
			await loadIncidents(currentOrgId);
			toast.success('Successfully joined the incident bridge!');
		} catch (err: any) {
			toast.error(err.message || 'Failed to join incident');
		}
	}

	let updatingIncidentId: string | null = $state(null);
	let updateMessage: string = $state('');
	let isUpdating: boolean = $state(false);

	let expandedTimelines = $state<Record<string, boolean>>({});
	let incidentUpdates = $state<Record<string, IncidentUpdateResponse[]>>({});
	let loadingUpdates = $state<Record<string, boolean>>({});

	async function toggleTimeline(incidentId: string) {
		expandedTimelines[incidentId] = !expandedTimelines[incidentId];
		if (expandedTimelines[incidentId] && !incidentUpdates[incidentId]) {
			loadingUpdates[incidentId] = true;
			try {
				if (currentOrgId) {
					const res = await IncidentAPI.getUpdates(currentOrgId, incidentId);
					incidentUpdates[incidentId] = res.data || [];
				}
			} catch (err: any) {
				console.error(err);
			} finally {
				loadingUpdates[incidentId] = false;
			}
		}
	}

	async function handleUpdate(incidentId: string) {
		if (!updateMessage.trim()) return;
		try {
			if (!currentOrgId) return;
			isUpdating = true;
			await IncidentAPI.addUpdate(currentOrgId, incidentId, { message: updateMessage });
			await loadIncidents(currentOrgId);

			// Refresh timeline if it's currently expanded
			if (expandedTimelines[incidentId]) {
				const res = await IncidentAPI.getUpdates(currentOrgId, incidentId);
				incidentUpdates[incidentId] = res.data || [];
			}

			updatingIncidentId = null;
			updateMessage = '';
			toast.success('Update posted successfully');
		} catch (err: any) {
			toast.error(err.message || 'Failed to add update');
		} finally {
			isUpdating = false;
		}
	}

	onMount(async () => {
		try {
			// Fetch user's organizations first (same pattern as org page)
			const orgRes = await OrgAPI.getAll();
			const orgs: { ID: string }[] = orgRes.data ?? [];

			if (orgs.length === 0) {
				error = 'No organization found. Please create or join an organization first.';
				return;
			}

			// Use the first org (can be extended later with an org switcher)
			currentOrgId = orgs[0].ID;
			await loadIncidents(currentOrgId);
		} catch (err: any) {
			error = err.message;
		} finally {
			isLoading = false;
		}
	});

	function formatTimeAgo(dateString: string) {
		const diff = Date.now() - new Date(dateString).getTime();
		const hours = Math.floor(diff / (1000 * 60 * 60));
		const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
		return `T-${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:00`;
	}

	function formatDateTime(dateString: string) {
		const date = new Date(dateString);
		return date.toISOString().replace('T', ' ').substring(0, 16) + ' UTC';
	}

	function formatTimeOnly(dateString: string) {
		const date = new Date(dateString);
		return (
			date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) +
			' UTC'
		);
	}
</script>

<svelte:head>
	<title>Incident - PulseWatch</title>
</svelte:head>

<div class="mb-8">
	<h1 class="text-3xl font-semibold text-on-surface mb-2 tracking-tight">Incident Command Board</h1>
	<p class="text-[16px] text-on-surface-variant">
		Real-time triage and resolution tracking across all microservices.
	</p>
</div>

{#if isLoading}
	<div class="flex items-center justify-center p-12">
		<div
			class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"
		></div>
	</div>
{:else if error}
	<div class="p-4 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg">
		{error}
	</div>
{:else}
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
		<!-- Column 1: Active Incidents -->
		<div
			class="flex flex-col gap-4 h-[calc(100vh-14rem)] overflow-y-auto custom-scrollbar pr-2 pb-4"
		>
			<!-- Column Header -->
			<div class="border-b border-white/10 pb-4 mb-4 mt-2">
				<div class="border-l-[3px] border-red-500 pl-4">
					<div class="text-xs font-mono font-medium tracking-widest text-on-surface-variant uppercase mb-1">
						INC &middot; UNRESOLVED
					</div>
					<div class="flex items-center justify-between">
						<h2 class="text-[22px] font-semibold text-on-surface tracking-tight">Active incidents</h2>
						<span class="text-[13px] font-medium text-on-surface-variant border border-white/10 rounded-full px-3 py-1 bg-transparent">
							{activeIncidents.length} open
						</span>
					</div>
				</div>
			</div>

			{#each activeIncidents as incident}
				<article
					class="glass-card p-5 relative transition-colors"
					class:ring-1={incident.SeverityCode === 'sev1'}
					class:ring-red-500_20={incident.SeverityCode === 'sev1'}
					class:hover:border-white_20={incident.SeverityCode !== 'sev1'}
					class:opacity-75={incident.SeverityCode === 'sev3'}
				>
					<div
						class="absolute top-0 left-0 w-1 h-full"
						class:bg-red-500={incident.SeverityCode === 'sev1'}
						class:bg-orange-400={incident.SeverityCode === 'sev2'}
						class:bg-yellow-400={incident.SeverityCode === 'sev3'}
					></div>

					<div class="flex justify-between items-start mb-4">
						<div>
							{#if incident.Severity}
								<div class="flex items-center gap-2 mb-1">
									{#if incident.SeverityCode === 'sev1'}
										<AlertCircle size={14} class="text-red-400" />
									{:else if incident.SeverityCode === 'sev2'}
										<AlertTriangle size={14} class="text-orange-400" />
									{:else}
										<Info size={14} class="text-yellow-400" />
									{/if}
									<span
										class="text-xs font-semibold tracking-wider uppercase"
										class:text-red-400={incident.SeverityCode === 'sev1'}
										class:text-orange-400={incident.SeverityCode === 'sev2'}
										class:text-yellow-400={incident.SeverityCode === 'sev3'}
									>
										{incident.Severity}
									</span>
								</div>
							{/if}
							<h3 class="text-lg font-semibold text-on-surface">
								{incident.Monitor?.Name || 'Unknown Monitor'}
							</h3>
						</div>
						<span
							class="font-mono text-xs text-on-surface-variant bg-black/30 px-2 py-1 rounded border border-white/5"
							>{incident.ID.substring(0, 8)}</span
						>
					</div>

					<p class="text-[15px] text-on-surface-variant mb-4 line-clamp-2">{incident.Summary}</p>

					<div class="flex justify-between items-end">
						<div class="flex items-center gap-4 text-xs font-mono text-on-surface-variant">
							<div class="flex items-center gap-1.5">
								<Clock size={14} />
								<span>{formatTimeAgo(incident.StartedAt)}</span>
							</div>
							<div class="flex items-center gap-1.5">
								<Server size={14} />
								<span>{incident.MonitorID.substring(0, 8)}</span>
							</div>
						</div>

						{#if incident.Assignees && incident.Assignees.length > 0}
							<div class="flex items-center gap-1.5 text-xs text-on-surface-variant">
								<Users size={13} />
								<span>{incident.Assignees.length}</span>
							</div>
						{/if}
					</div>

					<!-- Timeline Toggle Button -->
					<div class="mt-4 flex justify-between items-center">
						<button
							class="text-xs font-medium text-on-surface-variant hover:text-white transition-colors flex items-center gap-1"
							onclick={() => toggleTimeline(incident.ID)}
						>
							{#if expandedTimelines[incident.ID]}
								<ChevronUp size={14} /> Hide Timeline
							{:else}
								<ChevronDown size={14} /> View Timeline
							{/if}
						</button>
					</div>

					<!-- Vertical Timeline -->
					{#if expandedTimelines[incident.ID]}
						<div
							class="mt-3 bg-black/20 rounded-lg p-4 border border-white/5 max-h-[300px] overflow-y-auto custom-scrollbar"
						>
							<h4 class="text-sm font-medium text-on-surface mb-3 flex items-center gap-2">
								<Activity size={15} />
								Incident Timeline
							</h4>
							{#if loadingUpdates[incident.ID]}
								<div class="text-xs text-on-surface-variant text-center py-4">
									Loading updates...
								</div>
							{:else if incidentUpdates[incident.ID] && incidentUpdates[incident.ID].length > 0}
								<div class="relative pl-4 border-l border-white/10 space-y-4">
									{#each incidentUpdates[incident.ID] as update, i}
										<div class="relative">
											<div
												class="absolute -left-[21px] top-1 w-2.5 h-2.5 rounded-full ring-4 ring-[#0A0A0B]"
												class:bg-primary-container={i === 0}
												class:bg-surface-variant={i > 0}
											></div>
											<div class="flex justify-between items-baseline mb-1 gap-2">
												<span class="text-sm font-medium text-on-surface"
													>{update.message.split('\n')[0]}</span
												>
												<span class="text-[11px] font-mono text-on-surface-variant flex-shrink-0"
													>{formatTimeOnly(update.createAt)}</span
												>
											</div>

											{#if update.message.includes('\n')}
												<div
													class="text-xs font-mono text-on-surface-variant mt-1 bg-black/40 p-2 rounded border border-white/5 overflow-hidden text-ellipsis whitespace-nowrap"
												>
													{update.message.split('\n').slice(1).join(' ').replace(/`/g, '')}
												</div>
											{/if}

											<div class="flex items-center gap-2 mt-2">
												{#if update.avatarUrl}
													<img
														alt="Avatar"
														class="w-5 h-5 rounded-full object-cover grayscale opacity-80"
														src={update.avatarUrl}
													/>
												{/if}
												<span class="text-xs font-medium text-on-surface-variant"
													>{update.username || 'System'}</span
												>
											</div>
										</div>
									{/each}

									<!-- Initial Alert -->
									<div class="relative pt-2">
										<div
											class="absolute -left-[21px] top-3 w-2.5 h-2.5 rounded-full bg-red-500 ring-4 ring-[#0A0A0B]"
										></div>
										<div class="flex justify-between items-baseline">
											<span class="text-sm font-medium text-red-400">Alert Triggered</span>
											<span class="text-[11px] font-mono text-on-surface-variant"
												>{formatTimeOnly(incident.StartedAt)}</span
											>
										</div>
									</div>
								</div>
							{:else}
								<div class="relative pl-4 border-l border-white/10 space-y-4">
									<!-- Initial Alert -->
									<div class="relative pt-2">
										<div
											class="absolute -left-[21px] top-3 w-2.5 h-2.5 rounded-full bg-red-500 ring-4 ring-[#0A0A0B]"
										></div>
										<div class="flex justify-between items-baseline">
											<span class="text-sm font-medium text-red-400">Alert Triggered</span>
											<span class="text-[11px] font-mono text-on-surface-variant"
												>{formatTimeOnly(incident.StartedAt)}</span
											>
										</div>
									</div>
								</div>
							{/if}
						</div>
					{/if}

					<!-- Action Buttons: always shown for active incidents -->
					<div class="mt-4 flex gap-3">
						{#if incident.Assignees?.some((a) => a.UserID === currentUserId)}
							<button
								class="bg-primary-container text-white px-4 py-2 rounded text-sm font-medium hover:bg-primary transition-colors"
								onclick={() => {
									updatingIncidentId = updatingIncidentId === incident.ID ? null : incident.ID;
									updateMessage = '';
								}}
							>
								Update Status
							</button>
							<span
								class="border border-white/5 bg-white/5 text-on-surface-variant px-4 py-2 rounded-lg text-sm font-medium flex items-center gap-2"
							>
								<CheckCircle size={14} /> Joined
							</span>
						{:else}
							<button
								class="border border-white/10 text-on-surface px-4 py-2 rounded-lg text-sm font-medium hover:bg-white/5 transition-colors flex items-center gap-2"
								onclick={() => handleJoin(incident.ID)}
							>
								<Users size={14} /> Join Bridge to Update
							</button>
						{/if}
					</div>

					<!-- Update Form Inline -->
					{#if updatingIncidentId === incident.ID}
						<div class="mt-4 bg-black/30 border border-white/10 rounded-lg p-4">
							<textarea
								bind:value={updateMessage}
								placeholder="What's the latest status?..."
								class="w-full bg-black/50 border border-white/10 rounded p-3 text-sm text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary transition-colors resize-none"
								rows="3"></textarea>
							<div class="flex justify-end gap-2 mt-3">
								<button
									class="px-4 py-1.5 text-sm font-medium text-on-surface-variant hover:text-on-surface transition-colors"
									onclick={() => {
										updatingIncidentId = null;
										updateMessage = '';
									}}
								>
									Cancel
								</button>
								<button
									class="bg-primary text-white px-4 py-1.5 rounded text-sm font-medium hover:bg-primary-container transition-colors flex items-center gap-2 disabled:opacity-50"
									disabled={!updateMessage.trim() || isUpdating}
									onclick={() => handleUpdate(incident.ID)}
								>
									{#if isUpdating}
										<div
											class="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin"
										></div>
										Posting...
									{:else}
										Post Update
									{/if}
								</button>
							</div>
						</div>
					{/if}
				</article>
			{/each}

			{#if activeIncidents.length === 0}
				<div
					class="text-center py-8.5 px-4 text-on-surface-variant border border-dashed border-white/10 rounded-lg"
				>
					<h3 class="text-base font-medium text-on-surface mb-1">All Systems Operational</h3>
					<p class="text-sm text-on-surface-variant">No active incidents at this time.</p>
				</div>
			{/if}
		</div>

		<!-- Column 2: Resolved Incidents -->
		<div
			class="flex flex-col gap-4 opacity-80 h-[calc(100vh-14rem)] overflow-y-auto custom-scrollbar pr-2 pb-4"
		>
			<!-- Column Header -->
			<div class="border-b border-white/10 pb-4 mb-4 mt-2">
				<div class="border-l-[3px] border-[#10B981] pl-4">
					<div class="text-xs font-mono font-medium tracking-widest text-on-surface-variant uppercase mb-1">
						INC &middot; RESOLVED
					</div>
					<div class="flex items-center justify-between">
						<h2 class="text-[22px] font-semibold text-on-surface tracking-tight">Resolved (24h)</h2>
						<span class="text-[13px] font-medium text-on-surface-variant border border-white/10 rounded-full px-3 py-1 bg-transparent">
							{resolvedIncidents.length} closed
						</span>
					</div>
				</div>
			</div>

			{#each resolvedIncidents as incident}
				<article class="glass-card p-5 relative border-l-2 border-l-[#10B981]/50 bg-black/40">
					<div class="flex justify-between items-start mb-2">
						<div>
							<h3 class="text-lg font-medium text-[#e5e2e3] line-through decoration-white/30">
								{incident.Monitor?.Name || 'Unknown Monitor'}
							</h3>
						</div>
						<div class="flex items-center gap-2">
							<CheckCircle size={14} class="text-[#10B981]" />
							<span class="text-xs font-mono text-[#bbcabf]">{incident.ID.substring(0, 8)}</span>
						</div>
					</div>
					<p class="text-[15px] text-[#bbcabf] mb-4">{incident.Summary}</p>
					<div class="flex justify-between items-end border-t border-white/5 pt-3">
						<div class="flex flex-col text-xs font-mono text-[#bbcabf] gap-1">
							<span>Opened: {formatDateTime(incident.StartedAt)}</span>
							{#if incident.ResolvedAt}
								<span class="text-[#10B981]">Resolved: {formatDateTime(incident.ResolvedAt)}</span>
							{/if}
						</div>
						<button class="text-xs text-secondary hover:text-white transition-colors"
							>View RCA</button
						>
					</div>
				</article>
			{/each}

			{#if resolvedIncidents.length === 0}
				<div
					class="text-center py-12 text-[#bbcabf] border border-dashed border-white/10 rounded-lg"
				>
					<p class="text-sm">No resolved incidents in the last 24 hours.</p>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* Add specific tailwind classes that were dynamically constructed in class: directives */
	.ring-red-500_20 {
		--tw-ring-color: rgba(239, 68, 68, 0.2);
	}
	.hover\:border-white_20:hover {
		border-color: rgba(255, 255, 255, 0.2);
	}

	/* Custom Scrollbar for Timeline */
	.custom-scrollbar::-webkit-scrollbar {
		width: 6px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background-color: rgba(255, 255, 255, 0.2);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background-color: rgba(255, 255, 255, 0.3);
	}
</style>
