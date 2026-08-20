<script lang="ts">
	import { onMount } from 'svelte';
	import { SubscriptionAPI } from '$lib/api/subscription';
	import { Building2, CreditCard } from 'lucide-svelte';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateOnly } from '$lib/utils/datetime';

	let subscriptions: any[] = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			const res = await SubscriptionAPI.getByUserId();
			subscriptions = res.data || [];
		} catch (err: any) {
			error = err.message || 'Failed to load subscriptions';
		} finally {
			loading = false;
		}
	});

	// Format date menggunakan timezone dari store
	let tz = $derived($userTimezone);
	const formatDate = (d: string | null) => formatDateOnly(d ?? undefined, tz);

	function getPlanName(plan: string) {
		if (!plan) return 'Free';
		return plan.charAt(0).toUpperCase() + plan.slice(1);
	}
</script>

<svelte:head>
	<title>Billing & Subscriptions • PulseWatch</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Hero Section: Clear, left aligned -->
	<div class="flex flex-col items-start gap-2 border-b border-[#2A2A2B] pb-6">
		<h1 class="text-3xl font-bold tracking-tight text-white">Billing & Subscriptions</h1>
		<p class="text-[#A1A1AA] text-base">
			Manage your organization subscriptions and billing details.
		</p>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="flex items-center justify-center p-12 text-[#A1A1AA]">
			Loading subscriptions...
		</div>
	{:else if error}
		<div class="p-4 bg-[#450a0a] border border-[#7f1d1d] text-[#fca5a5] rounded-lg">
			{error}
		</div>
	{:else if subscriptions.length === 0}
		<div class="flex flex-col items-center justify-center p-12 border border-[#2A2A2B] rounded-lg bg-[#18181B] text-[#A1A1AA]">
			<CreditCard class="w-12 h-12 mb-4 text-[#52525B]" />
			<p>No subscriptions found.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each subscriptions as sub}
				<div class="flex flex-col border border-[#2A2A2B] rounded-lg bg-[#18181B] overflow-hidden">
					<div class="flex items-center gap-3 p-6 border-b border-[#2A2A2B]">
						<div class="flex items-center justify-center w-10 h-10 rounded-lg bg-[#27272A]">
							<Building2 class="w-5 h-5 text-white" />
						</div>
						<div class="flex flex-col">
							<span class="font-semibold text-white text-lg">{sub.Organization?.Name || 'Unknown Organization'}</span>
							<span class="text-sm text-[#A1A1AA]">Status: {sub.Status || 'Active'}</span>
						</div>
					</div>
					<div class="flex flex-col p-6 gap-4">
						<div class="flex justify-between items-center">
							<span class="text-[#A1A1AA]">Plan</span>
							<span class="text-white font-medium">{getPlanName(sub.Plan)}</span>
						</div>
						<div class="flex justify-between items-center">
							<span class="text-[#A1A1AA]">Next Billing</span>
							<span class="text-white font-medium">{formatDate(sub.CurrentPeriodEnd)}</span>
						</div>
						<div class="mt-4 pt-4 border-t border-[#2A2A2B]">
							<a 
								href={`/dashboard/organization/${sub.OrganizationID}/subscription`}
								class="flex items-center justify-center w-full py-2 px-4 bg-[#27272A] hover:bg-[#3F3F46] text-white rounded-lg transition-colors font-medium"
							>
								Manage Plan
							</a>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
