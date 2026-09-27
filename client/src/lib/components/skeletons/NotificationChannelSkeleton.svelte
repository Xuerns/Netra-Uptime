<script lang="ts">
	import DataSkeleton from '$lib/components/common/DataSkeleton.svelte';
	let { variant = 'channels' }: { variant?: 'channels' | 'history' } = $props();
</script>

<DataSkeleton variant={variant === 'history' ? 'list' : 'channel'} count={4} />
