<script lang="ts">
	let { variant = 'channels' }: { variant?: 'channels' | 'history' } = $props();
</script>

{#if variant === 'channels'}
	<div class="flex flex-col gap-3">
		{#each Array(4) as _}
			<div
				class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4.5 rounded-xl bg-white/2 border border-border animate-pulse"
			>
				<div class="flex items-start sm:items-center gap-3.5 min-w-0 flex-1">
					<div
						class="size-10 rounded-xl bg-white/5 border border-border shrink-0"
					></div>
					<div class="min-w-0 flex-1 space-y-1.5">
						<div class="flex items-center gap-2 flex-wrap">
							<div class="h-4 w-36 bg-white/10 rounded"></div>
							<div class="h-5 w-16 bg-white/5 rounded-full"></div>
						</div>
						<div class="h-3 w-48 bg-white/5 rounded font-mono"></div>
					</div>
				</div>

				<div class="flex items-center gap-3 shrink-0 self-end sm:self-auto">
					<div class="h-6 w-11 rounded-full bg-white/10"></div>
					<div class="size-8 rounded-lg bg-white/5 border border-border"></div>
				</div>
			</div>
		{/each}
	</div>
{:else}
	<div class="flex flex-col gap-2.5">
		{#each Array(5) as _}
			<div
				class="flex items-start gap-3 p-4 rounded-lg bg-white/3 border border-border animate-pulse"
			>
				<div class="size-4 rounded-full bg-white/10 shrink-0 mt-0.5"></div>
				<div class="flex flex-col min-w-0 flex-1 space-y-2">
					<div class="flex items-center gap-2 flex-wrap">
						<div class="h-4 w-1/2 bg-white/10 rounded"></div>
						<div class="h-5 w-20 bg-white/5 rounded-full"></div>
					</div>
					<div class="h-3 w-36 bg-white/5 rounded font-mono"></div>
				</div>
			</div>
		{/each}
	</div>
{/if}
