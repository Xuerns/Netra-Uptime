<script lang="ts">
	import { Clock } from 'lucide-svelte';

	interface Props {
		statusConfig: {
			title: string;
			description: string;
			bannerClass: string;
			bannerStyle: string;
			isThemed: boolean;
		};
		lastUpdated: Date;
	}

	let { statusConfig, lastUpdated }: Props = $props();
</script>

<section
	style={statusConfig.isThemed ? statusConfig.bannerStyle : ''}
	class="rounded-md border p-4 sm:p-5 backdrop-blur-md transition-all duration-300 {statusConfig.bannerClass}"
>
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4">
		<div class="flex items-center gap-3">
			<div class="relative flex items-center justify-center"></div>
			<div>
				<h1
					class="text-lg sm:text-xl font-bold tracking-tight text-foreground flex items-center gap-2"
				>
					{statusConfig.title}
				</h1>
				<p class="text-xs sm:text-sm text-muted-foreground mt-0.5">
					{statusConfig.description}
				</p>
			</div>
		</div>
		<div
			class="flex items-center gap-1.5 text-xs text-muted-foreground self-start sm:self-auto shrink-0 bg-black/20 sm:bg-transparent px-2.5 py-1 sm:p-0 rounded-md"
		>
			<Clock class="size-3 text-muted-foreground" />
			<span class="font-mono text-[11px] sm:text-xs">
				Last verified: {lastUpdated.toLocaleTimeString('en-US', {
					hour: '2-digit',
					minute: '2-digit',
					second: '2-digit',
					hour12: false
				})}
			</span>
		</div>
	</div>
</section>
