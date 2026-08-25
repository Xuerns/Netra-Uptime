<script lang="ts">
	import { MoreVertical, Pencil, Trash2, ShieldCheck, Globe } from 'lucide-svelte';
	import type { Monitor } from '$lib/api/monitor';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import { formatResponseTime, formatInterval, uptimeColor } from '$lib/utils/monitor';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';

	let {
		monitor,
		typeBadgeClass,
		openDropdownId,
		toggleDropdown,
		handleDeleteMonitor,
		handleEditMonitor
	} = $props<{
		monitor: Monitor;
		typeBadgeClass: (type: string) => string;
		openDropdownId: string | null;
		toggleDropdown: (id: string) => void;
		handleDeleteMonitor: (id: string) => void;
		handleEditMonitor: (monitor: Monitor) => void;
	}>();

	let isConfirmDeleteOpen = $state(false);

	async function confirmDelete() {
		if (handleDeleteMonitor) {
			handleDeleteMonitor(monitor.ID);
		}
	}

	let sparkColor = $derived(
		monitor.Status === 'paused' || !monitor.IsActive
			? '#f59e0b'
			: monitor.Status === 'up'
				? '#4edea3'
				: monitor.Status === 'down'
					? '#ff6b6b'
					: '#888899'
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

<!-- Monitor Card -->
<div
	class="relative w-full bg-white/3 border border-white/8 rounded-lg p-5 flex flex-col gap-3.5 backdrop-blur-md transition-all duration-250 ease-in-out"
>
	<!-- Card Top -->
	<div class="flex items-start justify-between gaep-2">
		<div class="min-w-0">
			<p class="text-body-md font-semibold text-on-surface truncate m-0">
				<a href="/dashboard/monitor/{monitor.ID}" class="hover:text-primary transition-colors">
					{monitor.Name}
				</a>
			</p>
			<p class="font-mono text-[12px] text-on-surface-variant truncate mt-1 m-0">
				{monitor.Target}
			</p>
		</div>
		<div class="flex items-center gap-2 shrink-0">
			<span
				class="font-mono text-[10.5px] font-medium px-2.5 py-0.5 rounded-full uppercase tracking-wide whitespace-nowrap {typeBadgeClass(
					monitor.Type
				)}"
			>
				{monitor.Type}
			</span>

			<!-- Kebab Menu -->
			<div class="relative">
				<button
					class="flex items-center justify-center size-7 rounded-lg border border-transparent text-on-surface-variant hover:bg-white/7 hover:text-on-surface hover:border-white/10 transition-colors"
					aria-label="Monitor options"
					onclick={(e) => {
						e.stopPropagation();
						toggleDropdown(monitor.ID);
					}}
				>
					<MoreVertical class="size-4" />
				</button>

				{#if openDropdownId === monitor.ID}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="absolute top-full right-0 mt-1.5 min-w-[152px] bg-[#161618] shadow-xl border border-white/8 rounded-lg p-1.5 flex flex-col gap-0.5 z-20"
						onclick={(e) => e.stopPropagation()}
					>
						<button
							class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-on-surface hover:bg-white/7 transition-colors text-left w-full"
							onclick={(e) => {
								e.preventDefault();
								e.stopPropagation();
								toggleDropdown(monitor.ID);
								handleEditMonitor(monitor);
							}}
						>
							<Pencil class="size-[15px]" />
							Edit monitor
						</button>
						<button
							class="flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] text-error hover:bg-error/10 transition-colors text-left w-full"
							onclick={() => {
								toggleDropdown(monitor.ID);
								isConfirmDeleteOpen = true;
							}}
						>
							<Trash2 class="size-[15px]" />
							Delete
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Status Row -->
	<div class="flex items-center gap-1.5 text-[12.5px] text-on-surface-variant">
		{#if monitor.Status === 'paused' || !monitor.IsActive}
			<span class="size-2 rounded-full bg-amber-400 shrink-0"></span>
			<span class="text-amber-400 font-medium">Paused</span>
		{:else if monitor.Status === 'up'}
			<span class="size-2 rounded-full bg-primary shrink-0"></span>
			<span class="text-primary">Operational</span>
		{:else}
			<span class="size-2 rounded-full bg-error shrink-0"></span>
			<span class="text-error">Down</span>
		{/if}
	</div>

	<!-- Sparkline -->
	<div class="w-full opacity-60 group-hover:opacity-100 transition-opacity mt-2 -mb-2">
		<Sparkline data={sparkData} color={sparkColor} height={36} />
	</div>

	<!-- SSL & Domain Expiry Row (for HTTP monitors) -->
	{#if monitor.Type === 'http' && (monitor.EnableSSLDomainCheck ?? true)}
		<div class="flex items-center gap-2 text-[11px] pt-2 border-t border-white/8 flex-wrap">
			{#if monitor.SSLStatus && monitor.SSLStatus !== 'none'}
				<span
					class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium {monitor.SSLStatus === 'valid' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : monitor.SSLStatus === 'warning' ? 'bg-amber-500/10 border-amber-500/20 text-amber-400' : 'bg-rose-500/10 border-rose-500/20 text-rose-400'}"
					title={monitor.SSLError || (monitor.SSLExpiresAt ? `SSL Expires: ${new Date(monitor.SSLExpiresAt).toLocaleDateString()}` : '')}
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
					class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-mono font-medium {monitor.DomainStatus === 'valid' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : monitor.DomainStatus === 'warning' ? 'bg-amber-500/10 border-amber-500/20 text-amber-400' : 'bg-rose-500/10 border-rose-500/20 text-rose-400'}"
					title={monitor.DomainError || (monitor.DomainExpiresAt ? `Domain Expires: ${new Date(monitor.DomainExpiresAt).toLocaleDateString()}` : '')}
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

	<!-- Metrics Row -->
	<div class="flex justify-between gap-3 pt-3 border-t border-white/8 items-end">
		<div class="flex flex-col gap-0.5">
			<span
				class="font-mono text-[17px] font-medium {monitor.Status === 'down'
					? 'text-error'
					: 'text-on-surface'}"
			>
				{formatResponseTime(monitor.ResponseTimeMs)}
			</span>
			<span class="text-[10.5px] uppercase tracking-wide text-on-surface-variant">Response</span>
		</div>
		<div class="flex flex-col gap-0.5">
			<span class="font-mono text-[17px] font-medium {uptimeColor(monitor.UptimePercent)}">
				{monitor.UptimePercent != null ? monitor.UptimePercent.toFixed(2) + '%' : '—'}
			</span>
			<span class="text-[10.5px] uppercase tracking-wide text-on-surface-variant">Uptime 30d</span>
		</div>
		<div class="flex flex-col gap-0.5 text-right">
			<span class="font-mono text-[17px] font-medium text-on-surface">
				{formatInterval(monitor.IntervalSeconds)}
			</span>
			<span class="text-[10.5px] uppercase tracking-wide text-on-surface-variant">Interval</span>
		</div>
	</div>
</div>

<ConfirmDialog
	bind:open={isConfirmDeleteOpen}
	title="Delete monitor?"
	description="Are you sure you want to delete {monitor.Name}? All checking history and incidents for this monitor will be permanently removed."
	confirmText="Delete"
	cancelText="Cancel"
	destructive={true}
	onConfirm={confirmDelete}
/>
