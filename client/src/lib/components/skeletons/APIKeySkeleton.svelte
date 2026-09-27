<div class="flex flex-col gap-3">
	{#each Array(4) as _}
		<div
			class="p-4 rounded-lg bg-white/3 border border-border flex items-center justify-between gap-4 flex-wrap animate-pulse"
		>
			<div class="flex flex-col gap-2 min-w-0 flex-1">
				<div class="flex items-center gap-2.5 flex-wrap">
					<!-- Key Name -->
					<div class="h-4 w-36 bg-white/10 rounded"></div>
					<!-- Scope Badge -->
					<div class="h-5 w-24 bg-white/10 rounded-md font-mono"></div>
				</div>

				<div class="flex items-center gap-4 flex-wrap pt-0.5">
					<!-- Prefix Token -->
					<div class="h-5 w-32 bg-white/5 rounded border border-border/60"></div>
					<!-- Created Date -->
					<div class="h-3.5 w-36 bg-white/5 rounded"></div>
					<!-- Expiry -->
					<div class="h-3.5 w-24 bg-white/5 rounded"></div>
				</div>
			</div>

			<!-- Actions: Copy & Revoke -->
			<div class="flex items-center gap-2 shrink-0">
				<div class="size-8 rounded-lg bg-white/5 border border-border"></div>
				<div class="size-8 rounded-lg bg-white/5 border border-border"></div>
			</div>
		</div>
	{/each}
</div>
