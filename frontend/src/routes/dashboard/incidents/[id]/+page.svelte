<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { IncidentAPI, type Incident, type IncidentUpdateResponse } from '$lib/api/incident';
	import { OrgAPI } from '$lib/api/organization';
	import { userTimezone } from '$lib/stores/timezone';
	import { toast } from 'svelte-sonner';

	// Sub-components for page sections
	import IncidentDetailHeader from '$lib/components/incident/IncidentDetailHeader.svelte';
	import IncidentInfoCard from '$lib/components/incident/IncidentInfoCard.svelte';
	import IncidentResponders from '$lib/components/incident/IncidentResponders.svelte';
	import IncidentTimeline from '$lib/components/incident/IncidentTimeline.svelte';
	import IncidentUpdateForm from '$lib/components/incident/IncidentUpdateForm.svelte';
	import IncidentDetailSkeleton from '$lib/components/skeletons/IncidentDetailSkeleton.svelte';

	const incidentId = $page.params.id as string;
	let incident = $state<Incident | null>(null);
	let updates = $state<IncidentUpdateResponse[]>([]);
	let currentOrgId = $state('');
	let isLoading = $state(true);
	let isLoadingUpdates = $state(true);
	let isJoining = $state(false);
	let isPosting = $state(false);
	let error = $state('');

	let currentUserId = $derived($page.data.user?.id || $page.data.user?.ID || '');
	let tz = $derived($userTimezone);

	async function loadIncidentData() {
		try {
			const orgRes = await OrgAPI.getAll();
			const orgs: { ID: string }[] = orgRes.data ?? [];

			if (orgs.length === 0) {
				error = 'No organization found.';
				return;
			}

			currentOrgId = orgs[0].ID;

			// Fetch incident details & timeline updates concurrently
			const [incRes, updatesRes] = await Promise.all([
				IncidentAPI.getById(currentOrgId, incidentId),
				IncidentAPI.getUpdates(currentOrgId, incidentId).catch(() => ({ data: [] }))
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
		if (!currentOrgId || !incidentId || isJoining) return;
		try {
			isJoining = true;
			await IncidentAPI.joinAssignee(currentOrgId, incidentId);
			toast.success('Successfully joined the incident command bridge!');
			// Refresh incident data
			const res = await IncidentAPI.getById(currentOrgId, incidentId);
			if (res.data) incident = res.data;
		} catch (err: any) {
			toast.error(err.message || 'Failed to join incident bridge');
		} finally {
			isJoining = false;
		}
	}

	async function handlePostUpdate(message: string) {
		if (!currentOrgId || !incidentId || isPosting) return;
		try {
			isPosting = true;
			await IncidentAPI.addUpdate(currentOrgId, incidentId, { message });
			toast.success('Incident status update posted successfully');
			
			// Refresh updates list
			const updatesRes = await IncidentAPI.getUpdates(currentOrgId, incidentId);
			updates = updatesRes.data || [];
		} catch (err: any) {
			toast.error(err.message || 'Failed to post update');
		} finally {
			isPosting = false;
		}
	}

	onMount(() => {
		loadIncidentData();
	});
</script>

<svelte:head>
	<title>{incident ? `Incident #${incident.ID.substring(0, 8)} - PulseWatch` : 'Incident Detail - PulseWatch'}</title>
</svelte:head>

{#if isLoading}
	<IncidentDetailSkeleton />
{:else if error}
	<div class="p-4 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg font-mono text-xs">
		{error}
	</div>
{:else if incident}
	<div class="space-y-6">
		<!-- 1. Header Section Component -->
		<IncidentDetailHeader
			{incident}
			{currentUserId}
			{isJoining}
			onJoin={handleJoinBridge}
		/>

		<!-- 2. Incident Information & Target Monitor Card Component -->
		<IncidentInfoCard {incident} {tz} />

		<!-- 3. Command Responders Component -->
		<IncidentResponders assignees={incident.Assignees} />

		<!-- 4. Post Status Update Form Component (Active Ongoing Incident) -->
		{#if incident.Status === 'ongoing'}
			<IncidentUpdateForm {isPosting} onPostUpdate={handlePostUpdate} />
		{/if}

		<!-- 5. Timeline Feed Component -->
		<IncidentTimeline {updates} isLoading={isLoadingUpdates} {tz} />
	</div>
{/if}
