<script lang="ts">
	import {
		CircleDot,
		Clock,
		ArrowUpRight,
		MoreVertical,
		Edit2,
		Trash2,
		Pause,
		Play
	} from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
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
</script>

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
						class="text-[10px] font-mono px-1.5 py-0.2 rounded border {monitor.SSLStatus === 'valid'
							? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
							: monitor.SSLStatus === 'warning'
								? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
								: 'bg-rose-500/10 text-rose-400 border-rose-500/20'}"
						title={monitor.SSLExpiresAt
							? `SSL Expires: ${new Date(monitor.SSLExpiresAt).toLocaleDateString()}`
							: ''}
					>
						SSL:{monitor.SSLDaysLeft != null ? monitor.SSLDaysLeft + 'd' : monitor.SSLStatus}
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
			class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-primary transition-colors"
			title="View details"
			aria-label="View {monitor.Name} details"
		>
			<ArrowUpRight class="size-4" />
		</a>

		{#if toggleDropdown}
			<div class="relative">
				<button
					class="flex items-center justify-center size-7 rounded-md text-muted-foreground hover:bg-white/10 hover:text-foreground transition-colors opacity-80 hover:opacity-100 focus-visible:opacity-100"
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
	.list-row {
		display: grid;
		grid-template-columns: 1fr 100px 90px 90px 80px 110px 80px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		font-size: 13px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: var(--color-accent);
	}

	/* Down row has a subtle left border */
	.list-row[data-status='down'] {
		border-left: 2px solid rgba(255, 107, 107, 0.45);
		padding-left: 18px;
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

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
	}
</style>
