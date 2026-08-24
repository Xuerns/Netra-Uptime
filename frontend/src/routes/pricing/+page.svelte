<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { isSelfHosted } from '$lib/utils/appMode';
	import Navigation from '$lib/components/landing/Navigation.svelte';
	import Pricing from '$lib/components/landing/Pricing.svelte';
	import Footer from '$lib/components/landing/Footer.svelte';

	onMount(() => {
		if (isSelfHosted()) {
			goto('/auth/login');
		}
	});
</script>

<svelte:head>
	<title>Pricing • PulseWatch</title>
	<meta
		name="description"
		content="Choose a plan that fits your monitoring needs. Start for free, upgrade anytime."
	/>
</svelte:head>

<div class="bg-background text-on-surface">
	<div class="flex items-center justify-center">
		<Navigation />
	</div>
	<main >
		<Pricing />
		<Footer />
	</main>
</div>

<style>
	main {
		background-color: #0e0e0f;
		background-image: radial-gradient(rgba(1, 101, 53, 0.462) 1.5px, transparent 1.5px);
		background-size: 24px 24px;
	}
</style>
