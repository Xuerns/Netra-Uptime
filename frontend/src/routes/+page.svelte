<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/auth/login');
	});
</script>

<svelte:head>
	<title>PulseWatch</title>
</svelte:head>
