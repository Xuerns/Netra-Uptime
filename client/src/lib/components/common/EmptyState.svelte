<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils';

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
		'p-12 text-center flex flex-col items-center justify-center gap-3 bg-card border border-border rounded-md transition-all',
		className
	)}
>
	{#if Icon}
		<div
			class="size-12 rounded-lg bg-white/5 border border-border flex items-center justify-center text-muted-foreground mb-1"
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
				variant={actionVariant === 'primary' ? 'default' : (actionVariant ?? 'default')}
				onclick={onAction}
				class="h-9 px-4 text-xs font-medium cursor-pointer"
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
