<script lang="ts">
	import {
		CircleDot,
		Clock,
		ArrowUpRight,
		MoreVertical,
		Edit2,
		Trash2,
		ShieldCheck,
		Globe,
		Pause,
		Play
	} from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import type { Monitor } from '$lib/api/monitor';
	import {
		typeIcon,
		statusLabel,
		formatInterval,
		formatLastChecked,
		formatResponseTime,
		uptimeColor,
		trimTarget
	} from '$lib/utils/monitor';

	let {
		monitor,
		openDropdownId,
		toggleDropdown,
		handleDeleteMonitor,
		handleEditMonitor,
		handleTogglePause
	} = $props<{
		monitor: Monitor;
		openDropdownId?: string | null;
		toggleDropdown?: (id: string) => void;
		handleDeleteMonitor?: (id: string) => Promise<void>;
		handleEditMonitor?: (monitor: Monitor) => void;
		handleTogglePause?: (monitor: Monitor) => Promise<void>;
	}>();

	let isConfirmDeleteOpen = $state(false);

	async function confirmDelete() {
		if (handleDeleteMonitor) {
			await handleDeleteMonitor(monitor.ID);
		}
	}

	let TypeIcon = $derived(typeIcon(monitor.Type));
	let sparkColor = $derived(
		monitor.Status === 'up' ? '#4edea3' : monitor.Status === 'down' ? '#ff6b6b' : '#888899'
	);

	// Derive real sparkline data from RecentChecks.
	// Backend returns data DESC (newest first), so reverse it for left-to-right display.
	let sparkData = $derived(
		monitor.recent_checks && monitor.recent_checks.length > 0
			? [...monitor.recent_checks].reverse().map((c) => c.ResponseTimeMs)
			: monitor.Status === 'up'
				? [45, 52, 48, 42, 60, 55, 49, 44, 38, 41, 48, 42]
				: monitor.Status === 'down'
					? [45, 52, 98, 100, 100, 100]
					: [45, 50, 48, 48, 48, 48]
	);
</script>

<article class="monitor-card group w-full" data-status={monitor.Status}>
	<div class="p-4 flex flex-col gap-3">
		<!-- Row 1: type icon + name + status badge -->
		<div class="flex items-start gap-3">
			<div class="type-icon shrink-0" aria-hidden="true">
				<TypeIcon class="size-4" />
			</div>

			<div class="min-w-0 flex-1">
				<h3 class="text-[15px] font-semibold text-foreground leading-tight truncate mb-0.5">
					<a href="/dashboard/monitor/{monitor.ID}" class="hover:text-primary transition-colors">
						{monitor.Name}
					</a>
				</h3>
				<p class="text-[12px] font-mono text-muted-foreground/70 truncate" title={monitor.Target}>
					{trimTarget(monitor.Target)}
				</p>
			</div>

			<!-- Status & Actions -->
			<div class="flex items-center gap-2 shrink-0">
				<span
					class="status-badge status-badge--{monitor.Status}"
					role="status"
					aria-label="Status: {statusLabel(monitor.Status)}"
				>
					<CircleDot class="size-2.5" aria-hidden="true" />
					{statusLabel(monitor.Status)}
				</span>

				{#if toggleDropdown}
					<div class="relative">
						<button
							class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
							aria-label="Monitor options"
							onclick={(e) => {
								e.preventDefault();
								e.stopPropagation();
								toggleDropdown('monitor-' + monitor.ID);
							}}
						>
							<MoreVertical class="size-4" />
						</button>

						{#if openDropdownId === 'monitor-' + monitor.ID}
							<div
								class="absolute top-full right-0 mt-1 min-w-[140px] bg-glass-card-bg backdrop-blur-xl border border-border rounded-lg p-1.5 z-20 shadow-xl"
								role="presentation"
								onclick={(e) => e.stopPropagation()}
							>
								{#if handleTogglePause}
									<button
										class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full"
										onclick={() => {
											toggleDropdown('');
											handleTogglePause(monitor);
										}}
									>
										{#if monitor.Status === 'paused'}
											<Play class="size-3.5 text-emerald-400" />
											Resume check
										{:else}
											<Pause class="size-3.5 text-amber-400" />
											Pause check
										{/if}
									</button>
								{/if}
								{#if handleEditMonitor}
									<button
										class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full"
										onclick={() => {
											toggleDropdown('');
											handleEditMonitor(monitor);
										}}
									>
										<Edit2 class="size-3.5" />
										Edit monitor
									</button>
								{/if}
								{#if handleDeleteMonitor}
									<button
										class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-destructive hover:bg-white/10 transition-colors text-left w-full mt-0.5"
										onclick={(e) => {
											e.preventDefault();
											toggleDropdown('');
											setTimeout(() => {
												isConfirmDeleteOpen = true;
											}, 0);
										}}
									>
										<Trash2 class="size-3.5" />
										Delete monitor
									</button>
								{/if}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>

		<!-- Divider -->
		<div class="h-px bg-white/6"></div>

		<!-- Sparkline -->
		<div class="w-full opacity-60 group-hover:opacity-100 transition-opacity -mt-1 -mb-1">
			<Sparkline data={sparkData} color={sparkColor} height={32} />
		</div>

		<!-- Row 2: metrics -->
		<div class="flex items-center gap-4">
			<div class="flex flex-col gap-0.5">
				<span class="metric-label">Response</span>
				<span class="metric-value">{formatResponseTime(monitor.ResponseTimeMs)}</span>
			</div>

			<div class="flex flex-col gap-0.5">
				<span class="metric-label">Uptime 30d</span>
				<span class="metric-value {uptimeColor(monitor.UptimePercent)}">
					{monitor.UptimePercent != null ? monitor.UptimePercent.toFixed(2) + '%' : '—'}
				</span>
			</div>

			<div class="flex flex-col gap-0.5 ml-auto text-right">
				<span class="metric-label">Interval</span>
				<span class="metric-value text-muted-foreground">
					{formatInterval(monitor.IntervalSeconds)}
				</span>
			</div>
		</div>

		<!-- SSL & Domain Expiry Row -->
		{#if monitor.Type === 'http' && (monitor.EnableSSLDomainCheck ?? true)}
			<div class="flex items-center gap-2 text-[11px] pt-1.5 border-t border-border flex-wrap">
				{#if monitor.SSLStatus && monitor.SSLStatus !== 'none'}
					<span
						class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium {monitor.SSLStatus ===
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

				{#if monitor.DomainStatus && monitor.DomainStatus !== 'none'}
					<span
						class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium {monitor.DomainStatus ===
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
			</div>
		{/if}

		<!-- Row 3: footer — org + last checked + link -->
		<div class="flex items-center justify-between pt-0.5">
			<div class="flex items-center gap-1.5 min-w-0">
				{#if monitor.Organization}
					<span
						class="text-[11px] font-medium text-muted-foreground bg-white/5 border border-border px-2 py-0.5 rounded-full truncate max-w-[120px]"
						title={monitor.Organization.Name}
					>
						{monitor.Organization.Name}
					</span>
				{/if}
				<span class="text-[11px] text-muted-foreground/40 flex items-center gap-1">
					<Clock class="size-3" aria-hidden="true" />
					{formatLastChecked(monitor.LastCheckedAt)}
				</span>
			</div>

			<a
				href="/dashboard/monitor/{monitor.ID}"
				class="flex items-center gap-1 text-[12px] font-medium text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
				aria-label="View {monitor.Name} details"
			>
				View <ArrowUpRight class="size-3.5" />
			</a>
		</div>
	</div>
</article>

<ConfirmDialog
	bind:open={isConfirmDeleteOpen}
	title="Delete monitor?"
	description="Are you sure you want to delete {monitor.Name}? All checking history and incidents for this monitor will be permanently removed."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDelete}
/>

<style>
	.monitor-card {
		position: relative;
		background: var(--color-card);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		overflow: hidden;
		transition:
			border-color 180ms ease,
			box-shadow 180ms ease,
			transform 180ms ease;
	}

	.monitor-card:hover {
		transform: translateY(-1px);
		box-shadow: 0 12px 28px -12px rgba(0, 0, 0, 0.15);
	}

	.monitor-card[data-status='up']:hover {
		border-color: rgba(78, 222, 163, 0.35);
		box-shadow:
			0 12px 28px -12px rgba(0, 0, 0, 0.15),
			0 0 20px rgba(78, 222, 163, 0.08);
	}

	.monitor-card[data-status='down'] {
		border-color: rgba(255, 107, 107, 0.25);
	}

	.monitor-card[data-status='down']:hover {
		border-color: rgba(255, 107, 107, 0.4);
		box-shadow:
			0 12px 28px -12px rgba(0, 0, 0, 0.15),
			0 0 20px rgba(255, 107, 107, 0.1);
	}

	:global(.monitor-card[data-status='up'] .monitor-card__accent) {
		background: rgba(78, 222, 163, 0.65);
	}

	:global(.monitor-card[data-status='down'] .monitor-card__accent) {
		background: rgba(255, 80, 80, 0.65);
	}

	:global(.monitor-card[data-status='paused'] .monitor-card__accent) {
		background: rgba(120, 120, 140, 0.35);
	}

	:global(.monitor-card[data-status='pending'] .monitor-card__accent) {
		background: rgba(172, 198, 255, 0.4);
	}

	/* Type icon pill */
	.type-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: var(--color-secondary);
		border: 1px solid var(--color-border);
		color: var(--color-muted-foreground);
	}

	/* ── Metric labels / values ───────────────────────────────────────────── */
	.metric-label {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--color-muted-foreground);
		opacity: 0.6;
	}

	.metric-value {
		font-family: 'JetBrains Mono', monospace;
		font-size: 14px;
		font-weight: 500;
		color: var(--color-foreground);
	}

	/* ── Status badge ─────────────────────────────────────────────────────── */
	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 7px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.04em;
		border: 1px solid transparent;
		white-space: nowrap;
	}

	.status-badge--up {
		color: var(--color-primary);
		background: rgba(78, 222, 163, 0.1);
		border-color: rgba(78, 222, 163, 0.22);
	}

	.status-badge--down {
		color: var(--color-destructive);
		background: rgba(255, 107, 107, 0.1);
		border-color: rgba(255, 107, 107, 0.22);
	}

	.status-badge--paused {
		color: var(--color-muted-foreground);
		background: rgba(136, 136, 153, 0.1);
		border-color: rgba(136, 136, 153, 0.2);
	}

	.status-badge--pending {
		color: var(--color-muted-foreground);
		background: rgba(173, 198, 255, 0.1);
		border-color: rgba(173, 198, 255, 0.2);
	}

	@media (prefers-reduced-motion: reduce) {
		.monitor-card {
			transition: none;
		}
	}
</style>
