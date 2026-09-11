<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';

	interface Props {
		index: string;
		title: string;
		href: string;
		description: string;
	}

	let { index, title, href, description }: Props = $props();
</script>

<a
	{href}
	class="p-4.5 sm:p-5 bg-card border border-border rounded-lg hover:border-primary transition-colors flex items-start justify-between group"
>
	<div class="space-y-1.5 pr-3">
		<div class="flex items-center gap-2.5">
			<span class="text-xs sm:text-sm font-mono text-primary font-semibold">{index}</span>
			<span class="text-xs sm:text-sm font-mono text-muted-foreground">/</span>
			<span class="text-base sm:text-lg font-semibold text-foreground group-hover:text-primary transition-colors">
				{title}
			</span>
		</div>
		<p class="text-xs sm:text-sm text-muted-foreground leading-relaxed">
			{description}
		</p>
	</div>
	<ArrowRight
		class="size-5 text-muted-foreground group-hover:text-primary transition-colors shrink-0 mt-1"
	/>
</a>
