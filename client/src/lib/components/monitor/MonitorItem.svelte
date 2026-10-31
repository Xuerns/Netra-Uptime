<script lang="ts">
	import {
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
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import type { Monitor } from '$lib/api/monitor';
	import {
		typeIcon,
		formatInterval,
		formatLastChecked,
		formatResponseTime,
		uptimeColor,
		trimTarget,
		formatSSLBadge,
		formatDomainBadge
	} from '$lib/utils/monitor';
	import MonitorStatusBadge from './MonitorStatusBadge.svelte';

	let {
		monitor,
		viewMode = 'grid',
		handleDeleteMonitor,
		handleEditMonitor,
		handleTogglePause
	}: {
		monitor: Monitor;
		viewMode?: 'grid' | 'list';
		handleDeleteMonitor?: (id: string) => Promise<void>;
		handleEditMonitor?: (monitor: Monitor) => void;
		handleTogglePause?: (monitor: Monitor) => Promise<void>;
	} = $props();

	let isConfirmDeleteOpen = $state(false);
	let isMenuOpen = $state(false);

	async function confirmDelete() {
		if (handleDeleteMonitor) {
			await handleDeleteMonitor(monitor.ID);
		}
	}

	let TypeIcon = $derived(typeIcon(monitor.Type));
	let sparkColor = $derived(
		monitor.Status === 'up'
			? '#10b981'
			: monitor.Status === 'degraded'
				? '#f59e0b'
				: monitor.Status === 'down'
					? '#ef4444'
					: '#71717a'
	);

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

	let hasCheckHistory = $derived(
		Boolean(monitor.recent_checks && monitor.recent_checks.length > 0)
	);
	let sparkData = $derived(
		hasCheckHistory ? [...monitor.recent_checks!].reverse().map((c) => c.ResponseTimeMs) : []
	);
</script>

{#if viewMode === 'grid'}
	<!-- Grid / Card View Mode -->
	<article class="monitor-card group w-full" data-status={monitor.Status}>
		<div class="p-3.5 flex flex-col gap-2.5">
			<!-- Row 1: type icon + name + status badge -->
			<div class="flex items-start gap-3">
				<div class="type-icon shrink-0" aria-hidden="true">
					<TypeIcon class="size-4" />
				</div>

				<div class="min-w-0 flex-1">
					<h3 class="text-sm font-medium text-foreground leading-tight truncate mb-0.5">
						<a
							href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
							class="hover:text-primary transition-colors"
						>
							{monitor.Name}
						</a>
					</h3>
					<p class="text-[12px] font-mono text-muted-foreground/70 truncate" title={monitor.Target}>
						{trimTarget(monitor.Target)}
					</p>
				</div>

				<!-- Status & Actions -->
				<div class="flex items-center shrink-0">
					<MonitorStatusBadge status={monitor.Status} class="shrink-0" />

					<DropdownMenu.Root bind:open={isMenuOpen}>
						<div
							class="overflow-hidden transition-all duration-200 ease-in-out flex items-center justify-end {isMenuOpen
								? 'w-7 ml-2 opacity-100 pointer-events-auto'
								: 'w-0 ml-0 opacity-0 pointer-events-none group-hover:w-7 group-hover:ml-2 group-hover:opacity-100 group-hover:pointer-events-auto focus-within:w-7 focus-within:ml-2 focus-within:opacity-100 focus-within:pointer-events-auto'}"
						>
							<DropdownMenu.Trigger
								class="flex items-center justify-center size-7 shrink-0 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors cursor-pointer focus:outline-none focus-visible:ring-1 focus-visible:ring-primary"
								aria-label="Monitor options"
							>
								<MoreVertical class="size-4" />
							</DropdownMenu.Trigger>
						</div>
						<DropdownMenu.Content
							align="end"
							class="w-40 bg-card border-border text-foreground rounded-lg p-1 shadow-xl z-50"
						>
							{#if handleTogglePause}
								<DropdownMenu.Item
									class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs hover:bg-white/10 transition-colors"
									onclick={() => handleTogglePause(monitor)}
								>
									{#if monitor.Status === 'paused'}
										<Play class="size-3.5 text-emerald-400" />
										<span>Resume check</span>
									{:else}
										<Pause class="size-3.5 text-amber-400" />
										<span>Pause check</span>
									{/if}
								</DropdownMenu.Item>
							{/if}
							{#if handleEditMonitor}
								<DropdownMenu.Item
									class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs hover:bg-white/10 transition-colors"
									onclick={() => handleEditMonitor(monitor)}
								>
									<Edit2 class="size-3.5" />
									<span>Edit monitor</span>
								</DropdownMenu.Item>
							{/if}
							{#if handleDeleteMonitor}
								<DropdownMenu.Separator class="bg-white/10 my-1" />
								<DropdownMenu.Item
									class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs text-destructive hover:bg-destructive/10 transition-colors"
									onclick={() => {
										isConfirmDeleteOpen = true;
									}}
								>
									<Trash2 class="size-3.5" />
									<span>Delete monitor</span>
								</DropdownMenu.Item>
							{/if}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			</div>

			<!-- Divider -->
			<div class="h-px bg-white/6"></div>

			<!-- Sparkline -->
			{#if sparkData.length > 0}
				<div class="w-full opacity-60 group-hover:opacity-100 transition-opacity -mt-1 -mb-1">
					<Sparkline data={sparkData} color={sparkColor} height={32} />
				</div>
			{:else}
				<div
					class="h-8 flex items-center justify-center text-[11px] text-muted-foreground/40 font-mono border-b border-dashed border-white/5"
				>
					Awaiting probe...
				</div>
			{/if}

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

			<!-- Domain & SSL Badges (HTTP only) -->
			{#if monitor.Type === 'http' && (monitor.EnableSSLDomainCheck ?? true) && ((monitor.SSLStatus && monitor.SSLStatus !== 'none') || (monitor.DomainStatus && monitor.DomainStatus !== 'none'))}
				<div class="flex items-center gap-2 pt-1 border-t border-white/4 flex-wrap">
					{#if monitor.SSLStatus && monitor.SSLStatus !== 'none'}
						<div
							class="flex items-center gap-1 text-[11px] font-mono px-1.5 py-px rounded-sm border {sslBadge.badgeClass}"
							title={sslBadge.tooltip}
						>
							<ShieldCheck class="size-3" />
							<span>{sslBadge.label}</span>
						</div>
					{/if}

					{#if monitor.DomainStatus && monitor.DomainStatus !== 'none'}
						<div
							class="flex items-center gap-1 text-[11px] font-mono px-1.5 py-px rounded-sm border {domainBadge.badgeClass}"
							title={domainBadge.tooltip}
						>
							<Globe class="size-3" />
							<span>{domainBadge.label}</span>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Row 3: footer (last checked + view details arrow) -->
			<div
				class="flex items-center justify-between pt-1 border-t border-white/4 text-[12px] text-muted-foreground/60"
			>
				<span class="flex items-center gap-1.5">
					<Clock class="size-3" aria-hidden="true" />
					{formatLastChecked(monitor.LastCheckedAt)}
				</span>

				<a
					href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
					class="flex items-center gap-1 hover:text-primary transition-colors text-[12px] font-medium"
					aria-label="View {monitor.Name} details"
				>
					Details
					<ArrowUpRight class="size-3.5" aria-hidden="true" />
				</a>
			</div>
		</div>
	</article>
{:else}
	<!-- List / Row View Mode -->
	<div
		class="grid grid-cols-[minmax(220px,1.6fr)_minmax(130px,1fr)_90px_85px_85px_75px_100px_70px] gap-3 px-4 py-2 items-center hover:bg-white/3 transition-colors group relative"
		data-status={monitor.Status}
	>
		<!-- 1. Name + target -->
		<div class="flex items-center gap-2.5 min-w-0">
			<div class="type-icon shrink-0" aria-hidden="true">
				<TypeIcon class="size-3.5" />
			</div>
			<div class="min-w-0 flex-1">
				<a
					href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
					class="text-[13px] font-medium text-foreground hover:text-primary transition-colors truncate block"
				>
					{monitor.Name}
				</a>
				<div class="flex items-center gap-1.5 truncate">
					<p
						class="text-[11.5px] font-mono text-muted-foreground/60 truncate"
						title={monitor.Target}
					>
						{trimTarget(monitor.Target, 36)}
					</p>
					{#if monitor.Type === 'http' && monitor.SSLStatus && monitor.SSLStatus !== 'none'}
						<span
							class="text-[10px] font-mono px-1 rounded-sm border {sslBadge.badgeClass}"
							title={sslBadge.tooltip}
						>
							{sslBadge.label}
						</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- 2. Activity / Latency Sparkline Graph -->
		<div class="flex items-center min-w-0 pr-2">
			<div
				class="w-full max-w-[160px] h-6 flex items-center opacity-70 group-hover:opacity-100 transition-opacity"
				title={sparkData.length > 0 ? 'Recent latency trend' : 'No recent latency checks'}
			>
				{#if sparkData.length > 0}
					<Sparkline data={sparkData} color={sparkColor} height={22} width={100} />
				{:else}
					<span class="text-[11px] text-muted-foreground/40 font-mono">Awaiting probe...</span>
				{/if}
			</div>
		</div>

		<!-- 3. Status Badge -->
		<div>
			<MonitorStatusBadge status={monitor.Status} />
		</div>

		<!-- 4. Response Time -->
		<div class="font-mono text-[13px] text-foreground">
			{formatResponseTime(monitor.ResponseTimeMs)}
		</div>

		<!-- 5. Uptime 30d -->
		<div class="font-mono text-[13px] {uptimeColor(monitor.UptimePercent)}">
			{monitor.UptimePercent != null ? monitor.UptimePercent.toFixed(2) + '%' : '—'}
		</div>

		<!-- 6. Interval -->
		<div class="font-mono text-[13px] text-muted-foreground">
			{formatInterval(monitor.IntervalSeconds)}
		</div>

		<!-- 7. Last Checked -->
		<div class="text-[12px] text-muted-foreground/60 flex items-center gap-1.5 whitespace-nowrap">
			<Clock class="size-3 shrink-0" aria-hidden="true" />
			{formatLastChecked(monitor.LastCheckedAt)}
		</div>

		<!-- 8. Actions Dropdown & View Link -->
		<div class="flex items-center justify-end gap-1 shrink-0">
			<a
				href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
				class="flex items-center justify-center size-7 rounded-sm text-muted-foreground hover:bg-white/10 hover:text-primary transition-colors cursor-pointer"
				title="View details"
				aria-label="View {monitor.Name} details"
			>
				<ArrowUpRight class="size-4" />
			</a>

			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors cursor-pointer focus:outline-none"
					aria-label="Monitor options"
				>
					<MoreVertical class="size-4" />
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					align="end"
					class="w-40 bg-card border-border text-foreground rounded-lg p-1 shadow-xl z-50"
				>
					{#if handleTogglePause}
						<DropdownMenu.Item
							class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs hover:bg-white/10 transition-colors"
							onclick={() => handleTogglePause(monitor)}
						>
							{#if monitor.Status === 'paused'}
								<Play class="size-3.5 text-emerald-400" />
								<span>Resume check</span>
							{:else}
								<Pause class="size-3.5 text-amber-400" />
								<span>Pause check</span>
							{/if}
						</DropdownMenu.Item>
					{/if}
					{#if handleEditMonitor}
						<DropdownMenu.Item
							class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs hover:bg-white/10 transition-colors"
							onclick={() => handleEditMonitor(monitor)}
						>
							<Edit2 class="size-3.5" />
							<span>Edit monitor</span>
						</DropdownMenu.Item>
					{/if}
					{#if handleDeleteMonitor}
						<DropdownMenu.Separator class="bg-white/10 my-1" />
						<DropdownMenu.Item
							class="cursor-pointer flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs text-destructive hover:bg-destructive/10 transition-colors"
							onclick={() => {
								isConfirmDeleteOpen = true;
							}}
						>
							<Trash2 class="size-3.5" />
							<span>Delete monitor</span>
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
	</div>
{/if}

<ConfirmDialog
	bind:open={isConfirmDeleteOpen}
	title="Delete Monitor"
	description="Are you sure you want to permanently delete '{monitor.Name}'? All historical latency records, incidents, and uptime stats will be purged immediately."
	confirmText="Delete"
	destructive={true}
	onConfirm={confirmDelete}
/>

<style>
	/* ── Shared Card Styles ────────────────────────────────────────────── */
	.monitor-card {
		background: var(--card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		transition: border-color 0.2s ease;
	}

	.monitor-card:hover {
		border-color: color-mix(in srgb, var(--primary) 35%, var(--color-border));
	}

	.monitor-card[data-status='up']:hover {
		border-color: color-mix(in srgb, #10b981 40%, var(--color-border));
	}

	.monitor-card[data-status='degraded']:hover {
		border-color: color-mix(in srgb, #f59e0b 40%, var(--color-border));
	}

	.monitor-card[data-status='down']:hover {
		border-color: color-mix(in srgb, #ef4444 40%, var(--color-border));
	}

	.monitor-card[data-status='paused']:hover {
		border-color: color-mix(in srgb, #71717a 40%, var(--color-border));
	}

	.type-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border-radius: 4px;
		background: var(--surface-sidebar);
		color: var(--color-muted-foreground);
		flex-shrink: 0;
	}

	.metric-label {
		font-size: 10.5px;
		color: var(--color-muted-foreground);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-family: var(--font-mono, monospace);
	}

	.metric-value {
		font-size: 13.5px;
		font-weight: 600;
		font-family: var(--font-mono, monospace);
		color: var(--color-foreground);
	}
</style>
