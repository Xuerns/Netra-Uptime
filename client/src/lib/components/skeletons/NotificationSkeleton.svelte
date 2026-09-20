<div class="flex flex-col gap-3">
	{#each Array(3) as _}
		<div
			class="flex items-center gap-4 p-5 rounded-xl bg-white/2 border border-border animate-pulse"
		>
			<div class="size-10 rounded-full bg-white/10 shrink-0"></div>
			<div class="flex-1 flex flex-col gap-2">
				<div class="h-4 bg-white/10 rounded w-1/3"></div>
				<div class="h-3 bg-white/5 rounded w-1/2"></div>
			</div>
			<div class="flex items-center gap-2 shrink-0">
				<div class="size-9 rounded-lg bg-white/5 shrink-0"></div>
				<div class="h-9 w-[100px] rounded-lg bg-white/5 shrink-0"></div>
			</div>
		</div>
	{/each}
</div>
