<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { IncidentAPI, type Incident, type IncidentUpdateResponse } from '$lib/api/incident';
	import { userTimezone } from '$lib/stores/timezone';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { AlertTriangle, ArrowLeft } from 'lucide-svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	// Sub-components for page sections
	import IncidentDetailHeader from '$lib/components/incident/IncidentDetailHeader.svelte';
	import IncidentResponders from '$lib/components/incident/IncidentResponders.svelte';
	import IncidentTimeline from '$lib/components/incident/IncidentTimeline.svelte';
	import IncidentUpdateForm from '$lib/components/incident/IncidentUpdateForm.svelte';
	import { DataSkeleton } from '$lib/components/common';

	const orgId = $derived($page.params.id as string);
	const incidentId = $derived($page.params.incidentId as string);

	let incident = $state<Incident | null>(null);
	let updates = $state<IncidentUpdateResponse[]>([]);
	let isLoading = $state(true);
	let isLoadingUpdates = $state(true);
	let isJoining = $state(false);
	let isPosting = $state(false);
	let error = $state('');

	let currentUserId = $derived($page.data?.user?.id || $page.data?.user?.ID || '');
	let tz = $derived($userTimezone);

	async function loadIncidentData() {
		if (!orgId || !incidentId) return;
		try {
			// Fetch incident details & timeline updates concurrently
			const [incRes, updatesRes] = await Promise.all([
				IncidentAPI.getById(orgId, incidentId),
				IncidentAPI.getUpdates(orgId, incidentId).catch(() => ({ data: [] }))
			]);

			if (incRes.data) {
				incident = incRes.data;
			} else {
				error = 'Incident not found';
			}

			updates = updatesRes.data || [];
		} catch (err: any) {
			error = err.message || 'Failed to load incident details';
		} finally {
			isLoading = false;
			isLoadingUpdates = false;
		}
	}

	async function handleJoinBridge() {
		if (!orgId || !incidentId || isJoining) return;
		try {
			isJoining = true;
			await IncidentAPI.joinAssignee(orgId, incidentId);
			toast.success('Successfully joined the incident command bridge!');
			// Refresh incident data
			const res = await IncidentAPI.getById(orgId, incidentId);
			if (res.data) incident = res.data;
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to join incident bridge'));
		} finally {
			isJoining = false;
		}
	}

	async function handlePostUpdate(message: string) {
		if (!orgId || !incidentId || isPosting) return;
		try {
			isPosting = true;
			await IncidentAPI.addUpdate(orgId, incidentId, { message });
			toast.success('Incident status update posted successfully');

			// Refresh updates list
			const updatesRes = await IncidentAPI.getUpdates(orgId, incidentId);
			updates = updatesRes.data || [];
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to post update'));
		} finally {
			isPosting = false;
		}
	}

	onMount(() => {
		loadIncidentData();
	});
</script>

<svelte:head>
	<title
		>{incident
			? `Incident #${incident.ID.substring(0, 8)} | Netra Uptime`
			: 'Incident Detail | Netra Uptime'}</title
	>
</svelte:head>

<div class="max-w-7xl mx-auto text-foreground antialiased">
	<!-- Navigation back button -->
	<div class="mb-6">
		<a
			href="/dashboard/org/{orgId}/incidents"
			class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors group"
		>
			<ArrowLeft class="size-4 group-hover:-translate-x-1 transition-transform" />
			<span>Back to Incidents</span>
		</a>
	</div>

	{#if isLoading}
		<DataSkeleton variant="detail" />
	{:else if error}
		<div class="border border-border rounded-lg bg-card p-8 text-center space-y-3">
			<AlertTriangle class="size-10 text-primary mx-auto" />
			<h2 class="text-lg font-bold text-foreground">Failed to Load Incident</h2>
			<p class="text-xs text-muted-foreground">{error || 'Incident not found.'}</p>
			<Button
				href="/dashboard/org/{orgId}/incidents"
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 gap-1.5 font-medium"
			>
				Back to Incidents
			</Button>
		</div>
	{:else if incident}
		<!-- 1. Header Section Component -->
		<header class="pb-6 border-b border-border/60">
			<IncidentDetailHeader {incident} {currentUserId} {isJoining} {tz} onJoin={handleJoinBridge} />
		</header>

		<!-- 2. Timeline Feed (2/3) + Responders (1/3) -->
		<section class="grid grid-cols-1 lg:grid-cols-3" aria-label="Incident timeline and responders">
			<!-- Timeline Feed + Status Update Form (2/3) -->
			<div class="py-6 min-w-0 lg:col-span-2 lg:pr-6 space-y-6">
				<IncidentTimeline {updates} isLoading={isLoadingUpdates} {tz} />

				<!-- Post Status Update Form (Active Ongoing Incident) placed inside timeline column -->
				{#if incident.Status === 'ongoing'}
					<IncidentUpdateForm {isPosting} onPostUpdate={handlePostUpdate} />
				{/if}
			</div>

			<!-- Command Responders & Assignees (1/3) -->
			<div class="py-6 min-w-0 border-t border-border/60 lg:border-t-0 lg:border-l lg:pl-6">
				<IncidentResponders assignees={incident.Assignees} />
			</div>
		</section>
	{/if}
</div>
