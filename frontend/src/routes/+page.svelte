<script lang="ts">
	import Navigation from '$lib/components/landing/Navigation.svelte';
	import Hero from '$lib/components/landing/Hero.svelte';
	import WhyPulsewatch from '$lib/components/landing/WhyPulsewatch.svelte';
	import FeaturesBento from '$lib/components/landing/FeaturesBento.svelte';
	import FeatureDeepDive from '$lib/components/landing/FeatureDeepDive.svelte';
	import Pricing from '$lib/components/landing/Pricing.svelte';
	import Footer from '$lib/components/landing/Footer.svelte';
</script>

<svelte:head>
	<title>PulseWatch</title>
</svelte:head>

<div class="bg-background text-on-surface">
	<div class="flex items-center justify-center">
		<Navigation />
	</div>
	<main>
		<Hero />
		<WhyPulsewatch />
		<FeaturesBento />
		<FeatureDeepDive />
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
