<div class="flex flex-col gap-2">
	{#each Array(5) as _}
		<div class="flex items-center gap-2.5 p-2 rounded-lg animate-pulse">
			<div class="size-8 rounded-full bg-white/10 shrink-0"></div>
			<div class="flex-1 flex flex-col gap-1.5 mt-0.5">
				<div class="h-2.5 bg-white/10 rounded w-3/4"></div>
				<div class="h-2 bg-white/6 rounded w-1/2"></div>
			</div>
			<div class="h-4 w-12 bg-white/5 rounded-full shrink-0"></div>
		</div>
	{/each}
</div>
