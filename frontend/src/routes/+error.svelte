<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import gsap from 'gsap';

	import ErrorHeader from '$lib/components/error/ErrorHeader.svelte';
	import ErrorHero from '$lib/components/error/ErrorHero.svelte';
	import ErrorRouteList from '$lib/components/error/ErrorRouteList.svelte';
	import ErrorFooter from '$lib/components/error/ErrorFooter.svelte';

	let statusCode = $derived($page.status || 404);
	let errorMessage = $derived(
		$page.error?.message ||
			(statusCode === 404 ? 'Endpoint does not exist.' : 'Unexpected service disruption.')
	);

	onMount(() => {
		if (
			typeof window !== 'undefined' &&
			!window.matchMedia('(prefers-reduced-motion: reduce)').matches
		) {
			gsap.from('.gsap-fade-in', {
				opacity: 0,
				y: 12,
				duration: 0.35,
				stagger: 0.08,
				ease: 'power2.out'
			});
		}
	});
</script>

<svelte:head>
	<title>{statusCode} • PulseWatch</title>
</svelte:head>

<div
	class="min-h-screen flex flex-col justify-between text-foreground selection:bg-primary selection:text-primary-foreground error-bg"
>
	<div class="gsap-fade-in">
		<ErrorHeader {statusCode} />
	</div>

	<main class="flex-1 max-w-[1440px] w-full mx-auto px-6 sm:px-10 lg:px-16 py-12 lg:py-20 flex items-center">
		<div class="w-full flex flex-col lg:flex-row items-center justify-between gap-12 lg:gap-16">
			<div class="w-full lg:flex-1 gsap-fade-in">
				<ErrorHero {statusCode} {errorMessage} />
			</div>

			<div class="w-full lg:w-[480px] xl:w-[540px] shrink-0 gsap-fade-in">
				<ErrorRouteList />
			</div>
		</div>
	</main>

	<div class="gsap-fade-in">
		<ErrorFooter />
	</div>
</div>

<style>
	.error-bg {
		background-color: var(--background);
		background-image:
			linear-gradient(var(--border) 1px, transparent 1px),
			linear-gradient(to right, var(--border) 1px, var(--background) 1px);
		background-size: 20px 20px;
	}
</style>
