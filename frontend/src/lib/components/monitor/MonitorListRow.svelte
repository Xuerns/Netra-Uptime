<script lang="ts">
	import { CircleDot, Clock, ArrowUpRight } from 'lucide-svelte';
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

	let { monitor } = $props<{ monitor: Monitor }>();

	let TypeIcon = $derived(typeIcon(monitor.Type));
</script>

<div class="list-row group" data-status={monitor.Status}>
	<!-- Name + target -->
	<div class="flex items-center gap-3 min-w-0">
		<div class="type-icon shrink-0" aria-hidden="true">
			<TypeIcon class="size-3.5" />
		</div>
		<div class="min-w-0">
			<a href="/dashboard/monitor/{monitor.ID}" class="text-[13.5px] font-semibold text-on-surface hover:text-primary transition-colors truncate block">
				{monitor.Name}
			</a>
			<div class="flex items-center gap-1.5 truncate">
				<p class="text-[11.5px] font-mono text-on-surface-variant/60 truncate" title={monitor.Target}>
					{trimTarget(monitor.Target, 36)}
				</p>
				{#if monitor.Type === 'http' && monitor.SSLStatus && monitor.SSLStatus !== 'none'}
					<span class="text-[10px] font-mono px-1.5 py-0.2 rounded border {monitor.SSLStatus === 'valid' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : monitor.SSLStatus === 'warning' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20'}" title={monitor.SSLExpiresAt ? `SSL Expires: ${new Date(monitor.SSLExpiresAt).toLocaleDateString()}` : ''}>
						SSL:{monitor.SSLDaysLeft != null ? monitor.SSLDaysLeft + 'd' : monitor.SSLStatus}
					</span>
				{/if}
			</div>
		</div>
	</div>

	<div>
		<span
			class="status-badge status-badge--{monitor.Status}"
			role="status"
			aria-label="Status: {statusLabel(monitor.Status)}"
		>
			<CircleDot class="size-2.5" aria-hidden="true" />
			{statusLabel(monitor.Status)}
		</span>
	</div>

	<div class="font-mono text-[13px] text-on-surface">
		{formatResponseTime(monitor.ResponseTimeMs)}
	</div>

	<div class="font-mono text-[13px] {uptimeColor(monitor.UptimePercent)}">
		{monitor.UptimePercent != null ? monitor.UptimePercent.toFixed(2) + '%' : '—'}
	</div>

	<div class="font-mono text-[13px] text-on-surface-variant">
		{formatInterval(monitor.IntervalSeconds)}
	</div>

	<div class="text-[12px] text-on-surface-variant/60 flex items-center gap-1">
		<Clock class="size-3 shrink-0" aria-hidden="true" />
		{formatLastChecked(monitor.LastCheckedAt)}
	</div>

	<div class="flex justify-end">
		<a
			href="/dashboard/monitor/{monitor.ID}"
			class="flex items-center gap-1 text-[12px] font-medium text-on-surface-variant hover:text-primary transition-colors opacity-0 group-hover:opacity-100 px-2 py-1 rounded-md hover:bg-white/5"
			aria-label="View {monitor.Name} details"
		>
			View <ArrowUpRight class="size-3.5" />
		</a>
	</div>
</div>

<style>
	.list-row {
		display: grid;
		grid-template-columns: 1fr 90px 76px 90px 64px 96px 56px;
		align-items: center;
		gap: 12px;
		padding: 10px 16px;
		font-size: 13px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: rgba(255, 255, 255, 0.025);
	}

	/* Down row has a subtle left border */
	.list-row[data-status='down'] {
		border-left: 2px solid rgba(255, 107, 107, 0.45);
		padding-left: 14px;
	}

	/* Type icon pill */
	.type-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		color: var(--color-on-surface-variant);
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
		color: #4edea3;
		background: rgba(78, 222, 163, 0.1);
		border-color: rgba(78, 222, 163, 0.22);
	}

	.status-badge--down {
		color: #ff6b6b;
		background: rgba(255, 107, 107, 0.1);
		border-color: rgba(255, 107, 107, 0.22);
	}

	.status-badge--paused {
		color: #888899;
		background: rgba(136, 136, 153, 0.1);
		border-color: rgba(136, 136, 153, 0.2);
	}

	.status-badge--pending {
		color: #adc6ff;
		background: rgba(173, 198, 255, 0.1);
		border-color: rgba(173, 198, 255, 0.2);
	}

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
