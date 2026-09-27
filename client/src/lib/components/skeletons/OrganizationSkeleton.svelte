<script lang="ts">
	import DataSkeleton from '$lib/components/common/DataSkeleton.svelte';
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

<DataSkeleton variant="cards" {viewMode} count={6} />
