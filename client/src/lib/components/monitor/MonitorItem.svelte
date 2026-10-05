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
		trimTarget,
		formatSSLBadge,
		formatDomainBadge
	} from '$lib/utils/monitor';

	let {
		monitor,
		viewMode = 'grid',
		openDropdownId,
		toggleDropdown,
		handleDeleteMonitor,
		handleEditMonitor,
		handleTogglePause
	} = $props<{
		monitor: Monitor;
		viewMode?: 'grid' | 'list';
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

	let sparkData = $derived(
		monitor.recent_checks && monitor.recent_checks.length > 0
			? [...monitor.recent_checks].reverse().map((c) => c.ResponseTimeMs)
			: monitor.Status === 'up'
				? [45, 52, 48, 42, 60, 55, 49, 44, 38, 41, 48, 42]
				: monitor.Status === 'down'
					? [45, 52, 98, 100, 100, 100]
					: [45, 50, 48, 48, 48, 48]
	);

	const dropdownId = $derived(
		viewMode === 'grid' ? 'monitor-' + monitor.ID : 'monitor-list-' + monitor.ID
	);
</script>

{#if viewMode === 'grid'}
	<!-- Grid / Card View Mode -->
	<article class="monitor-card group w-full" data-status={monitor.Status}>
		<div class="p-4 flex flex-col gap-3">
			<!-- Row 1: type icon + name + status badge -->
			<div class="flex items-start gap-3">
				<div class="type-icon shrink-0" aria-hidden="true">
					<TypeIcon class="size-4" />
				</div>

				<div class="min-w-0 flex-1">
					<h3 class="text-[15px] font-semibold text-foreground leading-tight truncate mb-0.5">
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
								class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors opacity-0 group-hover:opacity-100 focus-visible:opacity-100 cursor-pointer"
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
											class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full cursor-pointer"
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
											class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full cursor-pointer"
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
											class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-destructive hover:bg-white/10 transition-colors text-left w-full mt-0.5 cursor-pointer"
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

			<!-- Domain & SSL Badges (HTTP only) -->
			{#if monitor.Type === 'http' && (monitor.EnableSSLDomainCheck ?? true) && ((monitor.SSLStatus && monitor.SSLStatus !== 'none') || (monitor.DomainStatus && monitor.DomainStatus !== 'none'))}
				<div class="flex items-center gap-2 pt-1 border-t border-white/4 flex-wrap">
					{#if monitor.SSLStatus && monitor.SSLStatus !== 'none'}
						<div
							class="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded border {sslBadge.badgeClass}"
							title={sslBadge.tooltip}
						>
							<ShieldCheck class="size-3" />
							<span>{sslBadge.label}</span>
						</div>
					{/if}

					{#if monitor.DomainStatus && monitor.DomainStatus !== 'none'}
						<div
							class="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded border {domainBadge.badgeClass}"
							title={domainBadge.tooltip}
						>
							<Globe class="size-3" />
							<span>{domainBadge.label}</span>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Row 3: footer (last checked + view details arrow) -->
			<div class="flex items-center justify-between pt-1 border-t border-white/4 text-[12px] text-muted-foreground/60">
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
		class="list-row group relative {openDropdownId === 'monitor-list-' + monitor.ID ? 'z-30' : ''}"
		data-status={monitor.Status}
	>
		<!-- 1. Name + target -->
		<div class="flex items-center gap-3 min-w-0">
			<div class="type-icon shrink-0" aria-hidden="true">
				<TypeIcon class="size-3.5" />
			</div>
			<div class="min-w-0 flex-1">
				<a
					href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
					class="text-[13.5px] font-semibold text-foreground hover:text-primary transition-colors truncate block"
				>
					{monitor.Name}
				</a>
				<div class="flex items-center gap-1.5 truncate">
					<p class="text-[11.5px] font-mono text-muted-foreground/60 truncate" title={monitor.Target}>
						{trimTarget(monitor.Target, 36)}
					</p>
					{#if monitor.Type === 'http' && monitor.SSLStatus && monitor.SSLStatus !== 'none'}
						<span
							class="text-[10px] font-mono px-1.5 py-0.2 rounded border {sslBadge.badgeClass}"
							title={sslBadge.tooltip}
						>
							{sslBadge.label}
						</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- 2. Status Badge -->
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

		<!-- 3. Response Time -->
		<div class="font-mono text-[13px] text-foreground">
			{formatResponseTime(monitor.ResponseTimeMs)}
		</div>

		<!-- 4. Uptime 30d -->
		<div class="font-mono text-[13px] {uptimeColor(monitor.UptimePercent)}">
			{monitor.UptimePercent != null ? monitor.UptimePercent.toFixed(2) + '%' : '—'}
		</div>

		<!-- 5. Interval -->
		<div class="font-mono text-[13px] text-muted-foreground">
			{formatInterval(monitor.IntervalSeconds)}
		</div>

		<!-- 6. Last Checked -->
		<div class="text-[12px] text-muted-foreground/60 flex items-center gap-1">
			<Clock class="size-3 shrink-0" aria-hidden="true" />
			{formatLastChecked(monitor.LastCheckedAt)}
		</div>

		<!-- 7. Actions Dropdown & View Link -->
		<div class="flex items-center justify-end gap-1 shrink-0">
			<a
				href="/dashboard/org/{monitor.OrganizationID}/monitors/{monitor.ID}"
				class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-primary transition-colors cursor-pointer"
				title="View details"
				aria-label="View {monitor.Name} details"
			>
				<ArrowUpRight class="size-4" />
			</a>

			{#if toggleDropdown}
				<div class="relative">
					<button
						class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors opacity-80 hover:opacity-100 focus-visible:opacity-100 cursor-pointer"
						aria-label="Monitor options"
						onclick={(e) => {
							e.preventDefault();
							e.stopPropagation();
							toggleDropdown('monitor-list-' + monitor.ID);
						}}
					>
						<MoreVertical class="size-4" />
					</button>

					{#if openDropdownId === 'monitor-list-' + monitor.ID}
						<div
							class="absolute top-full right-0 mt-1 min-w-[140px] bg-glass-card-bg backdrop-blur-xl border border-border rounded-lg p-1.5 z-50 shadow-xl"
							role="presentation"
							onclick={(e) => e.stopPropagation()}
						>
							{#if handleTogglePause}
								<button
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full cursor-pointer"
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
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-foreground hover:bg-white/10 transition-colors text-left w-full cursor-pointer"
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
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[13px] text-destructive hover:bg-white/10 transition-colors text-left w-full mt-0.5 cursor-pointer"
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
		background: var(--color-glass-card-bg);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border: 1px solid var(--color-border);
		border-radius: 12px;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease,
			transform 0.15s ease;
		box-shadow: var(--shadow-sm);
	}

	.monitor-card:hover {
		border-color: var(--color-card-hover-border);
		box-shadow: var(--shadow-card-hover);
		transform: translateY(-1px);
	}

	.type-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid var(--color-border);
		color: var(--color-muted-foreground);
		flex-shrink: 0;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 3px 8px;
		border-radius: 6px;
		font-size: 11px;
		font-weight: 600;
		font-family: var(--font-mono, monospace);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		border-width: 1px;
		border-style: solid;
		flex-shrink: 0;
	}

	.status-badge--up {
		background: rgba(78, 222, 163, 0.1);
		color: #4edea3;
		border-color: rgba(78, 222, 163, 0.25);
	}

	.status-badge--down {
		background: rgba(255, 107, 107, 0.1);
		color: #ff6b6b;
		border-color: rgba(255, 107, 107, 0.25);
	}

	.status-badge--paused {
		background: rgba(255, 255, 255, 0.05);
		color: var(--color-muted-foreground);
		border-color: rgba(255, 255, 255, 0.1);
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

	/* ── Shared List Row Styles ────────────────────────────────────────── */
	.list-row {
		display: grid;
		grid-template-columns: minmax(200px, 1fr) 90px 110px 100px 90px 120px 80px;
		align-items: center;
		padding: 10px 16px;
		background: var(--color-glass-card-bg);
		backdrop-filter: blur(12px);
		border: 1px solid var(--color-border);
		border-radius: 10px;
		gap: 12px;
		transition:
			border-color 0.15s ease,
			background-color 0.15s ease;
	}

	.list-row:hover {
		border-color: var(--color-card-hover-border);
		background: rgba(255, 255, 255, 0.03);
	}

	@media (max-width: 900px) {
		.list-row {
			grid-template-columns: 1fr 100px 90px 80px;
		}
		.list-row > :nth-child(4),
		.list-row > :nth-child(5),
		.list-row > :nth-child(6) {
			display: none;
		}
	}
</style>
