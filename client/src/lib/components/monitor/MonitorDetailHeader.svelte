<script lang="ts">
	import { type Monitor } from '$lib/api/monitor';
	import { formatSSLBadge, formatDomainBadge } from '$lib/utils/monitor';
	import {
		Globe,
		ShieldCheck,
		ExternalLink,
		Clock,
		RefreshCw,
		Pause,
		Play,
		Pencil,
		Trash2
	} from 'lucide-svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import MonitorStatusBadge from './MonitorStatusBadge.svelte';

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
</script>

<!-- Monitor Header (flat, no card container) -->
<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
	<div class="flex items-start gap-4 min-w-0">
		<!-- Monitor Icon -->
		<div class="size-11 rounded-md bg-primary/10 flex items-center justify-center shrink-0">
			<Globe class="size-5 text-primary" />
		</div>

		<!-- Title & Target Info -->
		<div class="space-y-1.5 min-w-0">
			<div class="flex items-center gap-3 flex-wrap">
				<h1 class="text-2xl md:text-3xl font-bold text-foreground tracking-tight break-all">
					{monitor.Name}
				</h1>

				<MonitorStatusBadge status={monitor.Status} iconClass="size-4" class="text-xs" />
			</div>

			<!-- Target URL & Metadata Badges -->
			<div class="flex items-center gap-x-4 gap-y-2 text-xs text-muted-foreground flex-wrap">
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

		{#if onEdit}
			<Button variant="outline" size="sm" onclick={onEdit} class="gap-1.5">
				<Pencil class="size-3.5" />
				<span>Edit</span>
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
