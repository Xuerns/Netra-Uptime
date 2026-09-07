<script lang="ts">
	import { type Monitor } from '$lib/api/monitor';
	import { statusLabel } from '$lib/utils/monitor';
	import { ArrowLeft, Globe, ShieldCheck, ExternalLink, Clock } from 'lucide-svelte';

	interface Props {
		monitor: Monitor;
		onPauseResume?: () => void;
		onEdit?: () => void;
		onDelete?: () => void;
	}

	let { monitor, onPauseResume, onEdit, onDelete }: Props = $props();

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

<div class="flex flex-col gap-4">
	<!-- Top Navigation Back Button -->
	<div class="flex items-center justify-between">
		<a
			href="/dashboard/monitor"
			class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors group"
		>
			<ArrowLeft class="size-4 group-hover:-translate-x-1 transition-transform" />
			<span>Back to Monitors</span>
		</a>
	</div>

	<!-- Monitor Header Banner -->
	<div
		class="bg-card border border-border rounded-2xl p-5 md:p-6 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4"
	>
		<div class="flex items-start gap-4">
			<!-- Monitor Icon Box -->
			<div
				class="size-12 rounded-xl bg-white/5 border border-border flex items-center justify-center shrink-0"
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
						{#if monitor.SSLStatus}
							<span
								class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium text-[11px] {monitor.SSLStatus ===
								'valid'
									? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
									: monitor.SSLStatus === 'warning'
										? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
										: 'bg-rose-500/10 border-rose-500/20 text-rose-400'}"
								title={monitor.SSLError ||
									(monitor.SSLExpiresAt
										? `SSL Expires: ${new Date(monitor.SSLExpiresAt).toLocaleDateString()}`
										: '')}
							>
								<ShieldCheck class="size-3 shrink-0" />
								{#if monitor.SSLStatus === 'valid'}
									SSL: {monitor.SSLDaysLeft}d
								{:else if monitor.SSLStatus === 'warning'}
									SSL: {monitor.SSLDaysLeft}d left
								{:else if monitor.SSLStatus === 'expired'}
									SSL: Expired
								{:else}
									SSL Error
								{/if}
							</span>
						{/if}

						{#if monitor.DomainStatus}
							<span
								class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium text-[11px] {monitor.DomainStatus ===
								'valid'
									? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
									: monitor.DomainStatus === 'warning'
										? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
										: 'bg-rose-500/10 border-rose-500/20 text-rose-400'}"
								title={monitor.DomainError ||
									(monitor.DomainExpiresAt
										? `Domain Expires: ${new Date(monitor.DomainExpiresAt).toLocaleDateString()}`
										: '')}
							>
								<Globe class="size-3 shrink-0" />
								{#if monitor.DomainStatus === 'valid'}
									Domain: {monitor.DomainDaysLeft}d
								{:else if monitor.DomainStatus === 'warning'}
									Domain: {monitor.DomainDaysLeft}d left
								{:else if monitor.DomainStatus === 'expired'}
									Domain: Expired
								{:else}
									Domain Error
								{/if}
							</span>
						{/if}
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>
