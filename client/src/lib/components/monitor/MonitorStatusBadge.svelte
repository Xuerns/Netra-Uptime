<script lang="ts">
	import type { Monitor } from '$lib/api/monitor';
	import { statusLabel } from '$lib/utils/monitor';
	import { ArrowUp, ArrowDown, ArrowDownRight, Pause, Clock } from 'lucide-svelte';

	interface Props {
		status: Monitor['Status'] | string;
		class?: string;
		iconClass?: string;
	}

	let { status, class: className = '', iconClass = 'size-3.5' }: Props = $props();

	let config = $derived.by(() => {
		switch (status) {
			case 'up':
				return {
					icon: ArrowUp,
					color: 'text-emerald-500 dark:text-emerald-400'
				};
			case 'down':
				return {
					icon: ArrowDown,
					color: 'text-rose-500 dark:text-rose-400'
				};
			case 'degraded':
				return {
					icon: ArrowDownRight,
					color: 'text-amber-500 dark:text-amber-400'
				};
			case 'paused':
				return {
					icon: Pause,
					color: 'text-zinc-500 dark:text-zinc-400'
				};
			default:
				return {
					icon: Clock,
					color: 'text-sky-500 dark:text-sky-400'
				};
		}
	});

	let StatusIcon = $derived(config.icon);
</script>

<span
	class="inline-flex items-center gap-1 font-mono text-xs font-semibold uppercase tracking-wider {config.color} {className}"
	role="status"
	aria-label="Status: {statusLabel(status as Monitor['Status'])}"
>
	<StatusIcon class="{iconClass} stroke-[2.5] shrink-0" aria-hidden="true" />
	<span>{statusLabel(status as Monitor['Status'])}</span>
</span>
