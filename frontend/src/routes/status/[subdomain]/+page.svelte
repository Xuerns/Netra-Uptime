<script lang="ts">
	import type { PageData } from './$types';
	import { CheckCircle2, AlertCircle } from 'lucide-svelte';
	import logo from '$lib/assets/Logo.svg';

	let { data } = $props<{ data: PageData }>();
	let statusPage = $derived(data.statusPage);

	import Sparkline from '$lib/components/Sparkline.svelte';
</script>

<svelte:head>
	<title>{statusPage.title} • Status Page</title>
</svelte:head>

<div
	class="min-h-screen relative overflow-x-hidden text-foreground bg-background pattern-bg flex flex-col"
>
	<!-- Top Navigation Bar -->
	<nav class="sticky top-0 w-full z-50 bg-transparent border-b border-border backdrop-blur-sm">
		<div class="flex justify-between items-center px-6 py-4 w-full">
			<div class="flex items-center gap-3">
				<img class="h-8 w-8 object-contain" src={logo} />
				<div class="flex flex-col">
					<span class="font-headline-md text-headline-md font-bold text-foreground"
						>{statusPage.title}</span
					>
					{#if statusPage.description}
						<span class="text-sm text-muted-foreground font-body-md mt-0.5"
							>{statusPage.description}</span
						>
					{/if}
				</div>
			</div>
			<div class="hidden md:flex items-center gap-6">
				<!-- Additional links could go here -->
			</div>
		</div>
	</nav>

	<!-- Main Content -->
	<main class="flex-1 pb-16 px-6 max-w-full mx-auto w-full space-y-8 mt-5">
		<!-- Detailed Status Layout -->
		<div class="space-y-6">
			<!-- Services -->
			<div class="space-y-4">
				{#if statusPage.monitors && statusPage.monitors.length > 0}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
						{#each statusPage.monitors as monitor}
							{@const isUp = monitor.status === 'up'}

							<!-- Monitor Card -->
							<div
								class="relative bg-white/3 border border-border rounded-lg p-5 flex flex-col gap-3.5 backdrop-blur-md"
							>
								<!-- Card Top -->
								<div class="flex items-start justify-between gap-2">
									<div class="min-w-0">
										<p class="text-body-md font-semibold text-foreground truncate m-0">
											{monitor.name}
										</p>
									</div>
									<div class="flex items-center gap-2 shrink-0">
										<span
											class="text-[11px] font-mono px-1.5 py-0.5 bg-surface-dim text-muted-foreground uppercase border border-outline-variant rounded-sm"
										>
											{monitor.type}
										</span>
									</div>
								</div>

								<!-- Status Row -->
								<div class="flex items-center gap-1.5 text-[12.5px] text-muted-foreground mb-1">
									{#if isUp}
										<span
											class="size-2 rounded-full bg-primary shrink-0 animate-[pulseGlow_2.4s_ease-in-out_infinite]"
										></span>
										<span class="text-primary">Operational</span>
									{:else}
										<span class="size-2 rounded-full bg-destructive shrink-0"></span>
										<span class="text-destructive">Down</span>
									{/if}
								</div>

								<!-- Graphics -->
								<div class="mt-auto pt-2 border-t border-border">
									<div
										class="flex justify-between font-mono-data text-[10px] text-muted-foreground mb-1.5"
									>
										<span class={isUp ? 'text-primary' : 'text-destructive'}
											>{isUp ? 'Operational' : 'Degraded'}</span
										>
									</div>
									<div class="flex h-8 w-full items-end">
										{#if monitor.recent_checks && monitor.recent_checks.length > 0}
											<Sparkline
												data={monitor.recent_checks.slice().reverse()}
												color={isUp ? '#4edea3' : '#f87171'}
												height={32}
											/>
										{:else}
											<div class="text-[10px] text-muted-foreground w-full text-center pb-1">
												No data available
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div
						class="bg-surface-container border-dashed border border-border rounded p-8 text-center text-muted-foreground font-body-md"
					>
						No services are currently being monitored for this status page.
					</div>
				{/if}
			</div>
		</div>
	</main>

	<!-- Footer -->
	<footer class="w-full mt-auto border-t border-border bg-transparent">
		<div class="flex flex-col md:flex-row justify-center items-center px-6 py-6 w-full">
			<span class="font-body-md text-sm text-muted-foreground mb-4 md:mb-0"
				>© 2026 {statusPage.title} Monitoring. All systems operational.</span
			>
		</div>
	</footer>
</div>

<style>
	.pattern-bg {
		background-color: #f8fafc;
		background-image: radial-gradient(#e2e8f0 0.75px, transparent 0.75px);
		background-size: 20px 20px;
	}
	:global(.dark) .pattern-bg {
		background-color: #131314;
		background-image:
			conic-gradient(
				from 60deg at 56.25% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 60deg at 56.25% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 180deg at 43.75% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from 180deg at 43.75% calc(425% / 6),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from -60deg at 50% calc(175% / 12),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			),
			conic-gradient(
				from -60deg at 50% calc(175% / 12),
				transparent 0deg,
				transparent 0.5deg,
				rgba(255, 255, 255, 0.015) 0.5deg,
				rgba(255, 255, 255, 0.015) 119.5deg,
				transparent 119.5deg,
				transparent 120deg
			);
		background-position:
			0 0,
			6px 10px,
			0 0,
			6px 10px,
			6px 0,
			0 10px;
		background-size: 12px 20px;
	}
</style>
