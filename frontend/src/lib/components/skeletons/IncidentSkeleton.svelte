<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
	<!-- Active Incidents Skeleton -->
	<div class="flex flex-col gap-4">
		<div class="flex items-center justify-between pb-3 border-b border-white/10 mb-2">
			<div class="flex items-center gap-3">
				<div class="w-8 h-8 rounded-lg bg-white/5 animate-pulse"></div>
				<div class="h-6 w-40 bg-white/10 rounded-md animate-pulse"></div>
			</div>
			<div class="w-16 h-6 rounded-full bg-white/5 animate-pulse"></div>
		</div>

		{#each Array(3) as _}
			<article class="glass-card p-5 relative overflow-hidden">
				<div class="absolute top-0 left-0 w-1 h-full bg-white/10 animate-pulse"></div>
				<div class="flex justify-between items-start mb-4 pl-1">
					<div class="flex flex-col gap-2 w-2/3">
						<div class="h-4 w-1/4 bg-white/10 rounded animate-pulse"></div>
						<div class="h-5 w-3/4 bg-white/10 rounded animate-pulse"></div>
					</div>
					<div class="h-6 w-16 bg-white/5 rounded animate-pulse"></div>
				</div>
				<div class="h-4 w-full bg-white/5 rounded animate-pulse mb-2"></div>
				<div class="h-4 w-4/5 bg-white/5 rounded animate-pulse mb-6"></div>
				<div class="flex justify-between items-end">
					<div class="flex items-center gap-4">
						<div class="h-4 w-20 bg-white/5 rounded animate-pulse"></div>
						<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
					</div>
				</div>
				<div class="mt-4 flex gap-3">
					<div class="h-9 w-32 bg-white/10 rounded-lg animate-pulse"></div>
				</div>
			</article>
		{/each}
	</div>

	<!-- Resolved Incidents Skeleton -->
	<div class="flex flex-col gap-4 opacity-80">
		<div class="flex items-center justify-between pb-3 border-b border-white/10 mb-2">
			<div class="flex items-center gap-3">
				<div class="w-8 h-8 rounded-lg bg-white/5 animate-pulse"></div>
				<div class="h-6 w-48 bg-white/10 rounded-md animate-pulse"></div>
			</div>
			<div class="w-16 h-6 rounded-full bg-white/5 animate-pulse"></div>
		</div>

		{#each Array(4) as _}
			<article class="glass-card p-5 relative border-l-2 border-l-white/10 bg-black/40">
				<div class="flex justify-between items-start mb-2">
					<div class="h-5 w-2/3 bg-white/10 rounded animate-pulse"></div>
					<div class="h-4 w-16 bg-white/5 rounded animate-pulse"></div>
				</div>
				<div class="h-4 w-full bg-white/5 rounded animate-pulse mb-4"></div>
				<div class="flex justify-between items-end border-t border-white/5 pt-3">
					<div class="flex flex-col gap-2 w-1/2">
						<div class="h-3 w-3/4 bg-white/5 rounded animate-pulse"></div>
						<div class="h-3 w-3/4 bg-white/5 rounded animate-pulse"></div>
					</div>
					<div class="h-4 w-16 bg-white/5 rounded animate-pulse"></div>
				</div>
			</article>
		{/each}
	</div>
</div>
