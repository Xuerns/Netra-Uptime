<script lang="ts">
	import type { PublicMonitor } from '$lib/api/statuspage';
	import { Activity } from 'lucide-svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import UptimeBar from '$lib/components/statuspage/UptimeBar.svelte';

	interface Props {
		monitor: PublicMonitor;
		themeColor?: string;
	}

	let { monitor, themeColor = '#10b981' }: Props = $props();

	function getMonitorStatusInfo(status: string) {
		switch (status) {
			case 'up':
				return {
					label: 'Operational',
					dotClass: '',
					dotStyle: `background-color: ${themeColor};`,
					textClass: '',
					textStyle: `color: ${themeColor};`,
					sparklineColor: themeColor
				};
			case 'degraded':
				return {
					label: 'Degraded',
					dotClass: 'bg-amber-500',
					dotStyle: '',
					textClass: 'text-amber-400',
					textStyle: '',
					sparklineColor: '#f59e0b'
				};
			case 'paused':
				return {
					label: 'Paused',
					dotClass: 'bg-blue-400',
					dotStyle: '',
					textClass: 'text-blue-400',
					textStyle: '',
					sparklineColor: '#60a5fa'
				};
			default:
				return {
					label: 'Down',
					dotClass: 'bg-red-500',
					dotStyle: '',
					textClass: 'text-red-400',
					textStyle: '',
					sparklineColor: '#ef4444'
				};
		}
	}

	let statusInfo = $derived(getMonitorStatusInfo(monitor.status));
</script>

<div
	class="group rounded-md border border-border/70 hover:border-border/90 bg-card p-3 sm:px-4 sm:py-3 transition-all duration-200 space-y-2"
>
	<!-- Monitor Top Info Row -->
	<div class="flex flex-wrap items-center justify-between gap-2.5">
		<div class="flex items-center gap-2 min-w-0">
			<p
				class="text-sm sm:text-[15px] font-medium text-foreground truncate group-hover:text-primary transition-colors"
			>
				{monitor.name}
			</p>
			<span
				class="text-[10px] font-mono px-1.5 py-0.5 bg-surface-dim text-muted-foreground uppercase border border-outline-variant rounded-sm shrink-0"
			>
				{monitor.type}
			</span>
		</div>

		<div class="flex items-center gap-2 sm:gap-2.5 shrink-0">
			{#if monitor.recent_checks && monitor.recent_checks.length > 0}
				<div
					class="hidden sm:inline-flex items-center gap-2 px-2 py-0.5 rounded-full bg-white/3 border border-border/50"
					title="Recent latency trend"
				>
					<div class="w-14 md:w-16 h-3.5 flex items-center">
						<Sparkline
							data={monitor.recent_checks.slice().reverse()}
							color={statusInfo.sparklineColor}
							height={14}
						/>
					</div>
					<span class="text-[11px] font-mono text-muted-foreground inline-flex items-center gap-1">
						<Activity class="size-3 text-muted-foreground" />
						{monitor.recent_checks[0]}ms
					</span>
				</div>
				<span
					class="text-[11px] font-mono text-muted-foreground inline-flex sm:hidden items-center gap-1"
				>
					<Activity class="size-3 text-muted-foreground" />
					{monitor.recent_checks[0]}ms
				</span>
			{/if}
			<div
				class="flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-full bg-white/5 border border-border/60"
			>
				<span style={statusInfo.dotStyle} class="size-1.5 rounded-full {statusInfo.dotClass}"
				></span>
				<span style={statusInfo.textStyle} class={statusInfo.textClass}>{statusInfo.label}</span>
			</div>
		</div>
	</div>

	<!-- 90-Day Uptime Bar Component -->
	<div>
		<UptimeBar
			dailyStats={monitor.daily_stats}
			overallUptime={monitor.uptime_percent}
			accentColor={themeColor}
		/>
	</div>
</div>
