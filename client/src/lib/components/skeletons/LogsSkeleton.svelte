<div class="flex flex-col gap-4">
	<!-- Filter bar skeleton -->
	<div
		class="flex flex-wrap items-center justify-between gap-3 p-4 bg-white/3 border border-border rounded-lg"
	>
		<div class="h-9 w-64 bg-white/5 border border-border rounded-lg"></div>
		<div class="flex items-center gap-2">
			<div class="h-9 w-28 bg-white/5 border border-border rounded-lg"></div>
			<div class="h-9 w-28 bg-white/5 border border-border rounded-lg"></div>
		</div>
	</div>

	<!-- Table skeleton rows -->
	<div class="border border-border rounded-lg bg-white/3 overflow-hidden">
		<div class="px-4 py-3 bg-white/5 border-b border-border flex items-center justify-between">
			<div class="h-4 w-32 bg-white/10 rounded-md"></div>
			<div class="h-4 w-20 bg-white/10 rounded-md"></div>
		</div>
		<div class="divide-y divide-border">
			{#each Array(5) as _}
				<div class="p-4 flex flex-col md:flex-row md:items-center justify-between gap-3">
					<div class="flex items-center gap-3 flex-1 min-w-0">
						<div class="size-8 rounded-lg bg-white/8 shrink-0"></div>
						<div class="flex flex-col gap-1.5 flex-1 min-w-0">
							<div class="h-4 w-48 bg-white/10 rounded-md"></div>
							<div class="h-3 w-32 bg-white/5 rounded-md"></div>
						</div>
					</div>
					<div class="flex items-center gap-3 shrink-0">
						<div class="h-6 w-24 bg-white/8 rounded-md"></div>
						<div class="h-4 w-28 bg-white/5 rounded-md"></div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>
