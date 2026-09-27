<script lang="ts">
	import { onMount } from 'svelte';
	import { LayoutGrid, List } from 'lucide-svelte';
	import { cn } from '$lib/helper/utils';

	let {
		viewMode = $bindable('grid'),
		storageKey,
		class: className = '',
		onChange
	} = $props<{
		viewMode: 'grid' | 'list';
		storageKey?: string;
		class?: string;
		onChange?: (mode: 'grid' | 'list') => void;
	}>();

	onMount(() => {
		if (storageKey && typeof localStorage !== 'undefined') {
			const saved = localStorage.getItem(storageKey);
			if (saved === 'grid' || saved === 'list') {
				viewMode = saved;
			}
		}
	});

	function setMode(mode: 'grid' | 'list') {
		if (viewMode === mode) return;
		viewMode = mode;
		if (storageKey && typeof localStorage !== 'undefined') {
			localStorage.setItem(storageKey, mode);
		}
		onChange?.(mode);
	}
</script>

<div
	class={cn(
		'flex items-center gap-0.5 p-1 bg-white/5 border border-border rounded-lg shadow-sm',
		className
	)}
	role="group"
	aria-label="View mode toggle"
>
	<button
		type="button"
		class={cn(
			'p-1.5 rounded-md transition-all cursor-pointer',
			viewMode === 'grid'
				? 'bg-white/15 text-foreground shadow-sm'
				: 'text-muted-foreground hover:text-foreground hover:bg-white/5'
		)}
		onclick={() => setMode('grid')}
		aria-label="Grid view"
		aria-pressed={viewMode === 'grid'}
	>
		<LayoutGrid class="size-4" />
	</button>

	<button
		type="button"
		class={cn(
			'p-1.5 rounded-md transition-all cursor-pointer',
			viewMode === 'list'
				? 'bg-white/15 text-foreground shadow-sm'
				: 'text-muted-foreground hover:text-foreground hover:bg-white/5'
		)}
		onclick={() => setMode('list')}
		aria-label="List view"
		aria-pressed={viewMode === 'list'}
	>
		<List class="size-4" />
	</button>
</div>
