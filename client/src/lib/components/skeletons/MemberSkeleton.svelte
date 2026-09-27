<div class="border border-border rounded-lg bg-card divide-y divide-border">
	{#each Array(5) as _}
		<div class="flex items-center justify-between p-4 animate-pulse">
			<!-- User Info -->
			<div class="flex items-center gap-3.5 min-w-0">
				<div class="size-9 rounded-lg bg-white/10 shrink-0"></div>
				<div class="space-y-1.5 min-w-0">
					<div class="h-4 w-36 bg-white/10 rounded"></div>
					<div class="h-3 w-16 bg-white/5 rounded"></div>
				</div>
			</div>

			<!-- Role Pill & Actions -->
			<div class="flex items-center gap-3 shrink-0">
				<div class="h-6 w-20 rounded-md bg-white/10"></div>
				<div class="size-7 rounded-md bg-white/5"></div>
			</div>
		</div>
	{/each}
</div>

