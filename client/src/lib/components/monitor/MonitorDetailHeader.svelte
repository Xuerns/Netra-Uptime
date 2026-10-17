<script lang="ts">
	import { type Monitor } from '$lib/api/monitor';
	import { statusLabel, formatSSLBadge, formatDomainBadge } from '$lib/utils/monitor';
	import {
		Globe,
		ShieldCheck,
		ExternalLink,
		Clock,
		RefreshCw,
		Pause,
		Play,
		Trash2
	} from 'lucide-svelte';
	import Button from '$lib/components/ui/button/button.svelte';

	interface Props {
		monitor: Monitor;
		onPauseResume?: () => void;
		onEdit?: () => void;
		onDelete?: () => void;
		onCheckNow?: () => void;
		isChecking?: boolean;
	}

	let {
		monitor,
		onPauseResume,
		onEdit,
		onDelete,
		onCheckNow,
		isChecking = false
	}: Props = $props();

	let sslBadge = $derived(
		formatSSLBadge(monitor.SSLStatus, monitor.SSLDaysLeft, monitor.SSLExpiresAt, monitor.SSLError)
	);
	let domainBadge = $derived(
		formatDomainBadge(
			monitor.DomainStatus,
			monitor.DomainDaysLeft,
			monitor.DomainExpiresAt,
			monitor.DomainError
		)
	);

	let statusColor = $derived.by(() => {
		switch (monitor.Status) {
			case 'up':
				return {
					bg: 'bg-emerald-500/10',
					text: 'text-emerald-400',
					border: 'border-emerald-500/20',
					dot: 'bg-emerald-400'
				};
			case 'down':
				return {
					bg: 'bg-rose-500/10',
					text: 'text-rose-400',
					border: 'border-rose-500/20',
					dot: 'bg-rose-400'
				};
			case 'paused':
				return {
					bg: 'bg-amber-500/10',
					text: 'text-amber-400',
					border: 'border-amber-500/20',
					dot: 'bg-amber-400'
				};
			default:
				return {
					bg: 'bg-sky-500/10',
					text: 'text-sky-400',
					border: 'border-sky-500/20',
					dot: 'bg-sky-400'
				};
		}
	});
</script>

<!-- Monitor Header Banner -->
<div
	class="bg-card border border-border rounded-lg p-5 md:p-6 flex flex-col md:flex-row md:items-center justify-between gap-4"
>
	<div class="flex items-start gap-4">
		<!-- Monitor Icon Box -->
		<div
			class="size-12 rounded-lg bg-muted border border-border flex items-center justify-center shrink-0"
		>
			<Globe class="size-6 text-primary" />
		</div>

		<!-- Title & Target Info -->
		<div class="space-y-1">
			<div class="flex items-center gap-3 flex-wrap">
				<h1 class="text-xl md:text-2xl font-bold text-foreground tracking-tight">
					{monitor.Name}
				</h1>

				<!-- Status Badge -->
				<div
					class="inline-flex items-center gap-2 px-3 py-1 rounded-md text-xs font-semibold uppercase tracking-wider border {statusColor.bg} {statusColor.text} {statusColor.border}"
				>
					<span>{statusLabel(monitor.Status)}</span>
				</div>
			</div>

			<!-- Target URL & Metadata Badges -->
			<div class="flex items-center gap-4 text-xs text-muted-foreground flex-wrap pt-1">
				<a
					href={monitor.Target.startsWith('http') ? monitor.Target : `http://${monitor.Target}`}
					target="_blank"
					rel="noopener noreferrer"
					class="flex items-center gap-1 hover:text-primary transition-colors text-sm font-medium"
				>
					<span>{monitor.Target}</span>
					<ExternalLink class="size-3.5 opacity-70" />
				</a>

				<div class="flex items-center gap-1.5">
					<Clock class="size-3.5 text-primary" />
					<span>Interval: {monitor.IntervalSeconds}s</span>
				</div>

				<div
					class="uppercase px-2 py-0.5 rounded bg-white/5 border border-border font-mono text-[11px]"
				>
					{monitor.Type}
				</div>

				{#if monitor.Type === 'http' && (monitor.EnableSSLDomainCheck ?? true)}
					{#if monitor.SSLStatus && monitor.SSLStatus !== 'none'}
						<span
							class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium text-[11px] {sslBadge.badgeClass}"
							title={sslBadge.tooltip}
						>
							<ShieldCheck class="size-3 shrink-0" />
							<span>{sslBadge.label}</span>
						</span>
					{/if}

					{#if monitor.DomainStatus && monitor.DomainStatus !== 'none'}
						<span
							class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium text-[11px] {domainBadge.badgeClass}"
							title={domainBadge.tooltip}
						>
							<Globe class="size-3 shrink-0" />
							<span>{domainBadge.label}</span>
						</span>
					{/if}
				{/if}
			</div>
		</div>
	</div>

	<!-- Action Controls -->
	<div class="flex items-center gap-2 shrink-0 flex-wrap">
		{#if onCheckNow}
			<Button
				variant="outline"
				size="sm"
				disabled={isChecking || monitor.Status === 'paused'}
				onclick={onCheckNow}
				class="gap-1.5"
			>
				<RefreshCw class="size-3.5 {isChecking ? 'animate-spin' : ''}" />
				<span>{isChecking ? 'Checking...' : 'Check Now'}</span>
			</Button>
		{/if}

		{#if onPauseResume}
			<Button variant="outline" size="sm" onclick={onPauseResume} class="gap-1.5">
				{#if monitor.Status === 'paused'}
					<Play class="size-3.5" />
					<span>Resume</span>
				{:else}
					<Pause class="size-3.5" />
					<span>Pause</span>
				{/if}
			</Button>
		{/if}

		{#if onDelete}
			<Button
				variant="outline"
				size="sm"
				onclick={onDelete}
				class="gap-1.5 hover:border-destructive hover:text-destructive"
			>
				<Trash2 class="size-3.5" />
				<span>Delete</span>
			</Button>
		{/if}
	</div>
</div>
