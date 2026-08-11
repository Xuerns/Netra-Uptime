<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

<div class="flex flex-wrap gap-2 mb-5">
	{#each [80, 64, 72, 68] as w}
		<div class="h-7 rounded-full bg-white/6 animate-pulse" style="width: {w}px"></div>
	{/each}
</div>
<div class="h-9 w-full max-w-[320px] rounded-lg bg-white/6 animate-pulse mb-6"></div>

{#if viewMode === 'grid'}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(350px,1fr))] gap-4">
		{#each Array(6) as _}
			<div class="h-[148px] rounded-xl bg-white/4 border border-white/6 animate-pulse"></div>
		{/each}
	</div>
{:else}
	<div class="flex flex-col border border-white/8 rounded-xl overflow-hidden bg-white/2">
		<div class="h-[33px] bg-white/5 animate-pulse border-b border-white/10"></div>
		{#each Array(5) as _}
			<div class="h-[53px] border-b border-white/6 bg-white/4 animate-pulse"></div>
		{/each}
	</div>
{/if}
