<script lang="ts">
	import { Clock } from 'lucide-svelte';

	interface Props {
		statusConfig: {
			title: string;
			description: string;
			bannerClass: string;
			bannerStyle: string;
			dotClass: string;
			dotStyle: string;
			isThemed: boolean;
		};
		lastUpdated: Date;
	}

	let { statusConfig, lastUpdated }: Props = $props();
</script>

<section
	style={statusConfig.isThemed ? statusConfig.bannerStyle : ''}
	class="rounded-xl border p-6 sm:p-7 backdrop-blur-md transition-all duration-300 {statusConfig.bannerClass}"
>
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div class="flex items-center gap-4">
			<div class="relative flex items-center justify-center">
				<span
					style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
					class="size-3.5 rounded-full {statusConfig.dotClass} animate-ping absolute opacity-75"
				></span>
				<span
					style={statusConfig.isThemed ? statusConfig.dotStyle : ''}
					class="size-3.5 rounded-full {statusConfig.dotClass} relative"
				></span>
			</div>
			<div>
				<h1
					class="text-xl sm:text-2xl font-bold tracking-tight text-foreground flex items-center gap-2.5"
				>
					{statusConfig.title}
				</h1>
				<p class="text-xs sm:text-sm text-muted-foreground mt-1">
					{statusConfig.description}
				</p>
			</div>
		</div>
		<div
			class="flex items-center gap-1.5 text-xs text-muted-foreground self-start sm:self-auto shrink-0 bg-black/20 sm:bg-transparent px-3 py-1.5 sm:p-0 rounded-lg"
		>
			<Clock class="size-3.5 text-muted-foreground" />
			<span>
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
