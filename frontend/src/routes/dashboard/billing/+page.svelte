<script lang="ts">
	import { onMount } from 'svelte';
	import { SubscriptionAPI, type Subscription } from '$lib/api/subscription';
	import { Building2, CreditCard, Crown, ArrowUpRight, Calendar, Zap, Shield } from 'lucide-svelte';
	import { userTimezone } from '$lib/stores/timezone';
	import { formatDateOnly } from '$lib/utils/datetime';

	let subscriptions = $state<Subscription[]>([]);
	let loading = $state(true);
	let error = $state('');

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

	function getPlanAccent(plan: string) {
		switch (plan?.toLowerCase()) {
			case 'pro':
				return { bg: 'bg-primary/10', border: 'border-primary/20', text: 'text-primary', glow: 'shadow-primary/10', icon: 'text-primary' };
			case 'max':
				return { bg: 'bg-amber-500/10', border: 'border-amber-500/20', text: 'text-amber-400', glow: 'shadow-amber-500/10', icon: 'text-amber-400' };
			default:
				return { bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', text: 'text-emerald-400', glow: 'shadow-emerald-500/10', icon: 'text-emerald-400' };
		}
	}

	function getStatusColor(status: string) {
		switch (status?.toLowerCase()) {
			case 'active':
				return 'bg-emerald-500/15 text-emerald-400 border-emerald-500/20';
			case 'pending':
				return 'bg-amber-500/15 text-amber-400 border-amber-500/20';
			case 'expired':
			case 'cancelled':
				return 'bg-rose-500/15 text-rose-400 border-rose-500/20';
			default:
				return 'bg-slate-500/15 text-slate-400 border-slate-500/20';
		}
	}
</script>

<svelte:head>
	<title>Billing & Subscriptions • PulseWatch</title>
</svelte:head>

<div class="flex flex-col gap-8">
	<!-- Header -->
	<div class="flex flex-col items-start gap-1.5 border-b border-white/8 pb-6">
		<h1 class="text-2xl font-bold tracking-tight text-white">Billing & Subscriptions</h1>
		<p class="text-sm text-on-surface-variant/70">
			Manage your organization subscriptions and billing details.
		</p>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="flex flex-col items-center justify-center py-20 gap-3">
			<div class="size-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
			<span class="text-sm text-on-surface-variant/60">Loading subscriptions...</span>
		</div>
	{:else if error}
		<div class="p-4 text-sm text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg flex items-center gap-2">
			<Shield class="size-4 shrink-0" />
			{error}
		</div>
	{:else if subscriptions.length === 0}
		<div class="flex flex-col items-center justify-center py-20 px-6 gap-4 bg-white/2 border border-white/6 rounded-xl backdrop-blur-sm">
			<div class="size-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center">
				<CreditCard class="size-6 text-on-surface-variant/50" />
			</div>
			<div class="text-center">
				<h3 class="text-base font-semibold text-on-surface mb-1">No subscriptions found</h3>
				<p class="text-sm text-on-surface-variant/60 max-w-xs">
					Create an organization and subscribe to a plan to start monitoring your services.
				</p>
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
			{#each subscriptions as sub}
				{@const accent = getPlanAccent(sub.plan)}
				{@const isPaid = sub.plan?.toLowerCase() === 'pro' || sub.plan?.toLowerCase() === 'max'}

				<!-- Card -->
				<div
					class="group relative flex flex-col rounded-xl border border-white/8 bg-white/2 overflow-hidden transition-all duration-300 hover:border-white/14 hover:shadow-lg {accent.glow}"
				>
					<!-- Subtle dot pattern background -->
					{#if isPaid}
						<div
							class="absolute inset-0 pointer-events-none opacity-40"
							style="background-image: radial-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px); background-size: 20px 20px;"
						></div>
					{/if}

					<!-- Shimmer sweep on hover -->
					<div class="absolute inset-0 overflow-hidden rounded-xl pointer-events-none" aria-hidden="true">
						<div
							class="w-[200%] h-full bg-linear-to-r from-transparent via-white/4 to-transparent absolute top-0 left-[-200%] group-hover:left-full transition-all duration-1200 ease-out skew-x-12"
						></div>
					</div>

					<!-- Top accent line -->
					<div class="h-0.5 w-full {isPaid ? 'bg-linear-to-r from-transparent via-primary/60 to-transparent' : 'bg-linear-to-r from-transparent via-white/10 to-transparent'}"></div>

					<!-- Card Header -->
					<div class="relative z-10 flex items-start justify-between gap-3 p-5 pb-4">
						<div class="flex items-center gap-3 min-w-0">
							<div class="size-10 rounded-xl {accent.bg} border {accent.border} flex items-center justify-center shrink-0 transition-colors">
								{#if isPaid}
									<Crown class="size-4.5 {accent.icon}" />
								{:else}
									<Building2 class="size-4.5 {accent.icon}" />
								{/if}
							</div>
							<div class="flex flex-col min-w-0">
								<span class="text-body-md font-semibold text-on-surface truncate">
									{sub.organization?.name || 'Unknown Organization'}
								</span>
								<div class="flex items-center gap-2 mt-0.5">
									<span
										class="inline-flex items-center text-[10.5px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-md border {getStatusColor(sub.status)}"
									>
										{sub.status ? sub.status.charAt(0).toUpperCase() + sub.status.slice(1) : 'Active'}
									</span>
								</div>
							</div>
						</div>
					</div>

					<!-- Card Body -->
					<div class="relative z-10 flex flex-col gap-3 px-5 pb-5 flex-1">
						<!-- Plan Row -->
						<div class="flex items-center justify-between py-2.5 px-3.5 rounded-lg bg-white/3 border border-white/6">
							<div class="flex items-center gap-2">
								<Zap class="size-3.5 text-on-surface-variant/50" />
								<span class="text-xs text-on-surface-variant/70 font-medium">Plan</span>
							</div>
							<span class="text-sm font-semibold {accent.text}">{getPlanName(sub.plan)}</span>
						</div>

						<!-- Next Billing Row -->
						<div class="flex items-center justify-between py-2.5 px-3.5 rounded-lg bg-white/3 border border-white/6">
							<div class="flex items-center gap-2">
								<Calendar class="size-3.5 text-on-surface-variant/50" />
								<span class="text-xs text-on-surface-variant/70 font-medium">Next Billing</span>
							</div>
							<span class="text-sm font-medium text-on-surface">{formatDate(sub.current_period_end)}</span>
						</div>

						<!-- Spacer -->
						<div class="flex-1"></div>

						<!-- Action Button -->
						<a
							href={`/dashboard/organization/${sub.organization_id}/subscription`}
							class="flex items-center justify-center gap-2 w-full h-10 mt-1 rounded-lg text-sm font-medium
							bg-[#1A1A1B] border border-[#2A2A2B] {accent.text}
							hover:bg-primary hover:border-primary hover:text-white
							active:scale-[0.98] transition-all duration-200 shadow-sm group/btn"
						>
							<span>Manage Plan</span>
							<ArrowUpRight class="size-3.5 opacity-60 group-hover/btn:opacity-100 transition-opacity" />
						</a>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

