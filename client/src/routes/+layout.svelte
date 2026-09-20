<script lang="ts">
	import './layout.css';
	import PulseWatchLogo from '$lib/assets/Logo.svg';
	import { onMount } from 'svelte';
	import Lenis from 'lenis';
	import 'lenis/dist/lenis.css';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import { ModeWatcher, mode } from 'mode-watcher';

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

<svelte:head><link rel="icon" type="image/png" href="/favicon.png" /></svelte:head>
<ModeWatcher defaultMode="dark" track={true} />
<Toaster position="bottom-right" richColors theme={mode.current} />
{@render children()}
