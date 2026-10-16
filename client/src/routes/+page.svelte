<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/auth/login');
	});
</script>

<svelte:head>
	<title>Netra Uptime</title>
</svelte:head>
