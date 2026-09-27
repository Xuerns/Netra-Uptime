<script lang="ts">
	let { viewMode = 'grid' }: { viewMode?: 'grid' | 'list' } = $props();
</script>

{#if viewMode === 'grid'}
	<div class="grid grid-cols-[repeat(auto-fill,minmax(430px,1fr))] gap-5">
		{#each Array(6) as _}
			<article
				class="relative flex flex-col gap-2 p-4 bg-white/3 border border-border rounded-lg animate-pulse justify-between min-h-[190px]"
			>
				<!-- Top row: Avatar + Name + Slug -->
				<div class="flex items-start gap-3">
					<div class="size-11 rounded-md bg-white/10 shrink-0"></div>
					<div class="min-w-0 flex-1 space-y-1.5 mt-0.5">
						<div class="h-4.5 w-40 bg-white/10 rounded"></div>
						<div class="h-3.5 w-48 bg-white/5 rounded font-mono"></div>
					</div>
				</div>

				<div class="h-px bg-white/8 my-1"></div>

				<!-- Stats: Members & Monitors -->
				<div class="flex gap-6">
					<div class="flex items-center gap-2">
						<div class="size-3.5 rounded bg-white/10"></div>
						<div class="h-3.5 w-16 bg-white/5 rounded"></div>
					</div>
					<div class="flex items-center gap-2">
						<div class="size-3.5 rounded bg-white/10"></div>
						<div class="h-3.5 w-16 bg-white/5 rounded"></div>
					</div>
				</div>

				<!-- Status Monitor Bar -->
				<div class="mt-2 flex flex-col gap-2">
					<div class="flex items-center justify-between text-[13px]">
						<div class="h-3.5 w-24 bg-white/5 rounded"></div>
						<div class="h-3.5 w-20 bg-white/10 rounded font-mono"></div>
					</div>
					<div class="flex items-center gap-1 h-[14px]">
						{#each Array(6) as _}
							<div class="flex-1 h-full rounded-[2px] bg-white/8"></div>
						{/each}
					</div>
				</div>
			</article>
		{/each}
	</div>
{:else}
	<div class="flex flex-col gap-2">
		{#each Array(5) as _}
			<article
				class="flex flex-col lg:flex-row lg:items-center gap-3 lg:gap-5 p-3 lg:py-2.5 lg:px-5 bg-white/3 border border-border rounded-lg animate-pulse"
			>
				<!-- 1. Avatar + Name + Slug -->
				<div class="flex items-start lg:items-center gap-3 flex-1 min-w-[240px]">
					<div class="size-11 rounded-md bg-white/10 shrink-0"></div>
					<div class="min-w-0 flex-1 space-y-1.5 mt-0.5 lg:mt-0">
						<div class="h-4 w-40 bg-white/10 rounded"></div>
						<div class="h-3 w-44 bg-white/5 rounded font-mono"></div>
					</div>
				</div>

				<div class="h-px bg-white/8 w-full lg:hidden"></div>

				<!-- 2. Members & Monitors Count -->
				<div class="flex gap-5 lg:w-[180px] lg:pl-5 lg:border-l lg:border-border shrink-0 py-1">
					<div class="flex items-center gap-1.5">
						<div class="size-3.5 rounded bg-white/10"></div>
						<div class="h-3.5 w-14 bg-white/5 rounded"></div>
					</div>
					<div class="flex items-center gap-1.5">
						<div class="size-3.5 rounded bg-white/10"></div>
						<div class="h-3.5 w-14 bg-white/5 rounded"></div>
					</div>
				</div>

				<div class="h-px bg-white/8 w-full lg:hidden my-0.5"></div>

				<!-- 3. Status Monitor Bar -->
				<div
					class="flex items-center justify-between lg:w-[260px] lg:pl-5 lg:border-l lg:border-border shrink-0"
				>
					<div class="flex items-center gap-1 h-[10px] w-[140px]">
						{#each Array(5) as _}
							<div class="flex-1 h-full rounded-[2px] bg-white/8"></div>
						{/each}
					</div>
					<div class="h-3.5 w-20 bg-white/10 rounded font-mono"></div>
				</div>
			</article>
		{/each}
	</div>
{/if}
