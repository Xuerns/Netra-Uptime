<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/helper/utils';

	interface Props {
		icon?: Component | any;
		title: string;
		description?: string;
		actionText?: string;
		onAction?: () => void;
		actionVariant?: 'default' | 'primary' | 'secondary' | 'outline' | 'ghost';
		class?: string;
		children?: Snippet;
	}

	let {
		icon: Icon,
		title,
		description = '',
		actionText = '',
		onAction = undefined,
		actionVariant = 'default',
		class: className = '',
		children = undefined
	}: Props = $props();
</script>

<div
	class={cn(
		'p-12 text-center flex flex-col items-center justify-center gap-3 bg-card dark:bg-white/2 border border-border shadow-xs rounded-xl transition-all',
		className
	)}
>
	{#if Icon}
		<div
			class="size-12 rounded-2xl bg-white/5 border border-border flex items-center justify-center text-muted-foreground mb-1 shadow-sm"
		>
			<Icon class="size-6 text-foreground/80" />
		</div>
	{/if}

	<div class="space-y-1 max-w-[46ch] mx-auto">
		<h3 class="text-base font-semibold text-foreground tracking-tight">{title}</h3>
		{#if description}
			<p class="text-xs text-muted-foreground leading-relaxed">{description}</p>
		{/if}
	</div>

	{#if actionText && onAction}
		<div class="pt-2">
			<Button
				onclick={onAction}
				class="h-9 px-4 text-xs font-medium bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 shadow-sm cursor-pointer"
			>
				{actionText}
			</Button>
		</div>
	{/if}

	{#if children}
		<div class="pt-2 flex items-center gap-2 flex-wrap justify-center">
			{@render children()}
		</div>
	{/if}
</div>
