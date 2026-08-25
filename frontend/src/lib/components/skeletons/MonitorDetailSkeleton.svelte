<div class="space-y-6">
	<!-- 1. Header & Actions Banner Skeleton -->
	<div class="bg-[#161618] border border-white/8 rounded-2xl p-5 md:p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
		<div class="flex items-start gap-4">
			<div class="size-12 rounded-xl bg-white/5 border border-white/8 animate-pulse shrink-0"></div>
			<div class="space-y-2">
				<div class="flex items-center gap-3">
					<div class="h-7 w-48 bg-white/10 rounded-md animate-pulse"></div>
					<div class="h-6 w-20 rounded-full bg-white/6 animate-pulse"></div>
				</div>
				<div class="flex items-center gap-4 pt-1">
					<div class="h-4 w-44 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
					<div class="h-4 w-12 bg-white/5 rounded animate-pulse"></div>
				</div>
			</div>
		</div>
	</div>

	<!-- 2. Time Range Selector Control Bar Skeleton -->
	<div class="flex items-center justify-between bg-[#161618] border border-white/8 rounded-xl p-1.5">
		<div class="flex items-center gap-1.5">
			{#each [48, 40, 44, 44, 36] as w}
				<div class="h-7 rounded-lg bg-white/6 animate-pulse" style="width: {w}px"></div>
			{/each}
		</div>
		<div class="h-4 w-40 bg-white/5 rounded animate-pulse hidden sm:block"></div>
	</div>

	<!-- 3. Key Metrics Summary Grid Cards Skeleton -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
		{#each Array(4) as _}
			<div class="bg-[#161618] border border-white/8 rounded-2xl p-5 space-y-3 flex flex-col justify-between">
				<div class="flex justify-between items-center">
					<div class="h-3.5 w-24 bg-white/10 rounded animate-pulse"></div>
					<div class="size-7 rounded-lg bg-white/5 animate-pulse"></div>
				</div>
				<div class="space-y-2">
					<div class="h-8 w-28 bg-white/10 rounded animate-pulse"></div>
					<div class="h-4 w-20 bg-white/5 rounded animate-pulse"></div>
				</div>
			</div>
		{/each}
	</div>

	<!-- 4. Interactive Response Time Latency Line Chart Skeleton -->
	<div class="bg-[#161618] border border-white/8 rounded-2xl p-5 md:p-6 space-y-4">
		<div class="flex justify-between items-center">
			<div class="h-5 w-44 bg-white/10 rounded animate-pulse"></div>
			<div class="h-4 w-32 bg-white/5 rounded animate-pulse"></div>
		</div>
		<div class="h-48 w-full rounded-xl bg-white/4 border border-white/6 animate-pulse"></div>
	</div>

	<!-- 5. Daily Heatmap Bar Skeleton -->
	<div class="bg-[#161618] border border-white/8 rounded-2xl p-5 space-y-3">
		<div class="flex justify-between items-center">
			<div class="h-4 w-36 bg-white/10 rounded animate-pulse"></div>
			<div class="h-4 w-24 bg-white/5 rounded animate-pulse"></div>
		</div>
		<div class="h-8 w-full rounded-lg bg-white/5 animate-pulse"></div>
	</div>
</div>
