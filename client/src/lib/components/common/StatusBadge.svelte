<script lang="ts">
	import { cn } from '$lib/helper/utils';

	type StatusType =
		| 'up'
		| 'down'
		| 'paused'
		| 'degraded'
		| 'warning'
		| 'investigating'
		| 'identified'
		| 'monitoring'
		| 'resolved'
		| 'active'
		| 'inactive'
		| 'public'
		| 'private'
		| 'scheduled'
		| 'in_progress'
		| 'completed'
		| string;

	let {
		status,
		label,
		variant = 'badge',
		withDot = true,
		pulse = false,
		class: className = ''
	} = $props<{
		status: StatusType;
		label?: string;
		variant?: 'badge' | 'pill' | 'dot';
		withDot?: boolean;
		pulse?: boolean;
		class?: string;
	}>();

	const normalized = $derived(status?.toLowerCase?.() || '');

	const theme = $derived.by(() => {
		switch (normalized) {
			case 'up':
			case 'resolved':
			case 'completed':
			case 'active':
			case 'public':
			case 'operational':
				return {
					text: 'text-emerald-400',
					bg: 'bg-emerald-500/10',
					border: 'border-emerald-500/20',
					dot: 'bg-emerald-500',
					defaultLabel: label || status.toUpperCase()
				};
			case 'down':
			case 'critical':
			case 'error':
				return {
					text: 'text-rose-400',
					bg: 'bg-rose-500/10',
					border: 'border-rose-500/20',
					dot: 'bg-rose-500',
					defaultLabel: label || status.toUpperCase()
				};
			case 'degraded':
			case 'warning':
			case 'investigating':
			case 'scheduled':
				return {
					text: 'text-amber-400',
					bg: 'bg-amber-500/10',
					border: 'border-amber-500/20',
					dot: 'bg-amber-500',
					defaultLabel: label || status.toUpperCase()
				};
			case 'identified':
			case 'monitoring':
			case 'in_progress':
				return {
					text: 'text-sky-400',
					bg: 'bg-sky-500/10',
					border: 'border-sky-500/20',
					dot: 'bg-sky-500',
					defaultLabel: label || status.toUpperCase()
				};
			case 'paused':
			case 'inactive':
			case 'private':
			default:
				return {
					text: 'text-muted-foreground',
					bg: 'bg-white/5',
					border: 'border-white/10',
					dot: 'bg-muted-foreground/60',
					defaultLabel: label || status.toUpperCase()
				};
		}
	});
</script>

{#if variant === 'dot'}
	<span class={cn('relative flex size-2 shrink-0', className)} title={theme.defaultLabel}>
		{#if pulse}
			<span
				class={cn(
					'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
					theme.dot
				)}
			></span>
		{/if}
		<span class={cn('relative inline-flex rounded-full size-2', theme.dot)}></span>
	</span>
{:else}
	<span
		class={cn(
			'inline-flex items-center gap-1.5 font-mono text-[11px] font-medium border uppercase tracking-wider',
			variant === 'pill' ? 'rounded-full px-2.5 py-0.5' : 'rounded-md px-2 py-0.5',
			theme.bg,
			theme.text,
			theme.border,
			className
		)}
	>
		{#if withDot}
			<span class={cn('size-1.5 rounded-full shrink-0', theme.dot)}></span>
		{/if}
		<span>{theme.defaultLabel}</span>
	</span>
{/if}
