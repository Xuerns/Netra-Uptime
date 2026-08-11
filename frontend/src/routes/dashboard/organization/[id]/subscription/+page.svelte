<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { toast } from 'svelte-sonner';
	import { Loader2, Check, ArrowLeft, Zap } from 'lucide-svelte';
	import { SubscriptionAPI } from '$lib/api/subscription';

	const orgId = $page.params.id as string;
	const userEmail = $page.data.user?.email;

	let isLoading = $state(true);
	let isProcessing = $state(false);
	let currentSub = $state<any>(null);

	onMount(async () => {
		await loadSubscription();
	});

	async function loadSubscription() {
		try {
			const res = await SubscriptionAPI.getByUserId();
			const subs = res.data || [];
			// Cari subscription untuk organisasi saat ini
			currentSub = subs.find((s: any) => s.OrganizationID === orgId);
		} catch (error) {
			console.error('Failed to load subscription:', error);
			toast.error('Failed to load subscription data');
		} finally {
			isLoading = false;
		}
	}

	async function handleSubscribe(plan: string) {
		if (!userEmail) {
			toast.error('User email not found');
			return;
		}

		isProcessing = true;
		try {
			const res = await SubscriptionAPI.createTransaction(orgId, {
				plan: plan,
				email: userEmail
			});

			const token = res.data?.token || res.data?.Token || res.token || res.Token;
			if (!token) throw new Error('No payment token received');

			// @ts-ignore
			window.snap.pay(token, {
				onSuccess: function (result: any) {
					toast.success('Payment successful!');
					loadSubscription();
				},
				onPending: function (result: any) {
					toast.info('Waiting for your payment!');
				},
				onError: function (result: any) {
					toast.error('Payment failed!');
				},
				onClose: function () {
					// toast.info('You closed the popup without finishing the payment');
				}
			});
		} catch (error: any) {
			toast.error(error.message || 'Failed to initiate payment');
		} finally {
			isProcessing = false;
		}
	}

	const plans = [
		{
			id: 'free',
			name: 'Free',
			price: '0',
			features: ['5 Monitors', '5-minute Interval', '1 Status Page', '2 PIC/Monitor']
		},
		{
			id: 'pro',
			name: 'Pro',
			price: '199000',
			features: [
				'10 Monitors',
				'3-minute Interval',
				'3 Status Pages',
				'4 PIC/Monitor',
				'Priority Support'
			]
		},
		{
			id: 'max',
			name: 'Max',
			price: '299000',
			features: [
				'20 Monitors',
				'1-minute Interval',
				'5 Status Pages',
				'6 PIC/Monitor',
				'Advanced Analytics'
			]
		}
	];

	let activePlanId = $derived(currentSub?.Status === 'active' ? currentSub.Plan : 'free');
</script>

<svelte:head>
	<title>Billing & Subscription - PulseWatch</title>
</svelte:head>

<div class="p-6 max-w-7xl mx-auto space-y-6">
	<div class="flex items-center gap-4 mb-2">
		<a
			href={`/dashboard/organization/${orgId}`}
			class="flex items-center justify-center size-9 rounded-lg border bg-white/2 border-white/10 text-on-surface-variant hover:bg-white/7 hover:text-on-surface transition-colors"
			aria-label="Back to Organization"
		>
			<ArrowLeft class="size-4.5" />
		</a>
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-white">Billing & Subscription</h1>
			<p class="text-slate-400">
				Manage your organization's subscription plan and billing details.
			</p>
		</div>
	</div>

	{#if isLoading}
		<div class="flex items-center justify-center py-20">
			<Loader2 class="size-8 text-primary animate-spin" />
		</div>
	{:else}
		<!-- Pricing Cards -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 xl:gap-8 mt-8">
			{#each plans as plan}
				{@const isCurrent = activePlanId === plan.id}
				{@const isFree = plan.id === 'free'}
				{@const isPro = plan.id === 'pro'}
				{@const isMax = plan.id === 'max'}

				<div
					class="pricing-card group relative rounded-3xl p-8 flex flex-col h-full overflow-hidden transition-colors duration-500 z-10 border
					{isCurrent ? 'border-primary ring-1 ring-primary shadow-[0_0_30px_rgba(16,185,129,0.15)]' : ''}
					{!isCurrent && isFree ? 'bg-background border-[#2A2A2B] hover:border-[#3a3a3c]' : ''}
					{!isCurrent && (isPro || isMax) ? 'glass-card border-primary/20 hover:border-primary/50' : ''}
					{!isCurrent && isMax ? 'hover:shadow-[0_0_40px_rgba(78,222,163,0.15)]' : ''}
					{isCurrent && (isPro || isMax) ? 'glass-card' : ''}
					{isCurrent && isFree ? 'bg-background' : ''}"
				>
					{#if isFree}
						<!-- Subtle hover glow -->
						<div
							class="absolute inset-0 bg-linear-to-b from-white/2 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"
						></div>
					{/if}

					{#if isPro || isMax}
						<!-- Grid pattern overlay -->
						<div
							class="absolute inset-0 pointer-events-none"
							style="background-image: radial-gradient(rgba(255, 255, 255, 0.03) 1.5px, transparent 1.5px); background-size: 24px 24px;"
						></div>
					{/if}

					{#if isMax}
						<!-- Subtle premium background tint for Max Plan -->
						<div
							class="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none"
						></div>
						<!-- Slow shimmer sweep on hover -->
						<div
							class="absolute inset-0 overflow-hidden rounded-3xl pointer-events-none"
							aria-hidden="true"
						>
							<div
								class="w-[200%] h-full bg-linear-to-r from-transparent via-white/10 to-transparent absolute top-0 left-[-200%] group-hover:left-full transition-all duration-1000 ease-out skew-x-12"
							></div>
						</div>
					{/if}

					<div class="mb-8 relative z-20 flex justify-between items-start">
						<div>
							<h3 class="font-headline-md text-[28px] text-white font-medium tracking-tight mb-2">
								{plan.name}
							</h3>
							<p class="font-body-md text-[15px] text-on-surface-variant">
								{plan.id === 'free'
									? 'Perfect for trying out PulseWatch.'
									: plan.id === 'pro'
										? 'For professionals and growing teams.'
										: 'Advanced features for maximum power.'}
							</p>
						</div>
					</div>

					<!-- Price -->
					<div class="mb-10 pb-8 border-b border-[#2A2A2B] relative z-20">
						<div class="flex items-baseline gap-1">
							<span class="text-lg font-medium text-slate-400 mr-1">Rp</span>
							<span class="font-display-lg text-[48px] font-bold tracking-tighter text-white">
								{parseInt(plan.price).toLocaleString('id-ID')}
							</span>
							<span class="font-body-md text-[15px] text-on-surface-variant">/mo</span>
						</div>
					</div>

					<!-- Features -->
					<ul
						class="space-y-4 mb-10 flex-1 font-body-md text-[15px] text-on-surface-variant relative z-20"
					>
						{#each plan.features as feature}
							<li class="flex items-start gap-3 group/item">
								<div
									class="w-5 h-5 rounded-full {isFree
										? 'bg-white/5 border-white/10'
										: 'bg-primary/10 border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40'} border flex items-center justify-center shrink-0 mt-0.5 transition-colors"
								>
									<Check class="size-3 {isFree ? 'text-white/50' : 'text-primary'} font-bold" />
								</div>
								<span class="text-white">{feature}</span>
							</li>
						{/each}
					</ul>

					<!-- Button -->
					<div class="relative z-20 mt-auto">
						{#if isCurrent}
							<Button
								disabled
								class="w-full bg-white/5 text-slate-400 hover:bg-white/5 border border-[#2A2A2B] h-12 rounded-xl text-base"
							>
								Current Plan
							</Button>
						{:else if plan.id === 'free'}
							<Button
								disabled
								class="w-full bg-white/5 text-slate-400 hover:bg-white/5 border border-[#2A2A2B] h-12 rounded-xl text-base"
							>
								Included
							</Button>
						{:else}
							<Button
								onclick={() => handleSubscribe(plan.id)}
								disabled={isProcessing}
								class="w-full bg-primary hover:bg-primary/90 text-white shadow-lg shadow-primary/20 transition-all active:scale-[0.98] h-12 rounded-xl text-base"
							>
								{#if isProcessing}
									<Loader2 class="w-5 h-5 mr-2 animate-spin" />
									Processing...
								{:else}
									Upgrade to {plan.name}
								{/if}
							</Button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
