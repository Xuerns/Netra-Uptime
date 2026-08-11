<script lang="ts">
	import './layout.css';
	import PulseWatchLogo from '$lib/assets/Logo.svg';
	import { onMount } from 'svelte';
	import Lenis from 'lenis';
	import 'lenis/dist/lenis.css';
	import { Toaster } from '$lib/components/ui/sonner/index.js';

	let { children } = $props();

	onMount(() => {
		const lenis = new Lenis({
			autoRaf: true
		});

		return () => {
			lenis.destroy();
		};
	});
</script>

<svelte:head><link rel="icon" href={PulseWatchLogo} /></svelte:head>
<Toaster position="bottom-right" richColors theme="dark" />
{@render children()}
