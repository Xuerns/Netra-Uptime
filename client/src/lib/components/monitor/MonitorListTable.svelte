<script lang="ts">
	import type { Monitor } from '$lib/api/monitor';
	import MonitorItem from '$lib/components/monitor/MonitorItem.svelte';

	interface Props {
		monitors: Monitor[];
		canManage?: boolean;
		onDelete: (id: string) => void | Promise<void>;
		onEdit: (m: Monitor) => void;
		onTogglePause: (m: Monitor) => void | Promise<void>;
	}

	let { monitors, canManage = true, onDelete, onEdit, onTogglePause }: Props = $props();
</script>

<div class="monitors-scroll-area w-full">
	<div class="border border-border rounded-md overflow-hidden bg-card">
		<div class="overflow-x-auto">
			<div class="min-w-240">
				<!-- Header row -->
				<div
					class="grid grid-cols-[minmax(220px,1.6fr)_minmax(130px,1fr)_90px_85px_85px_75px_100px_70px] gap-3 px-4 py-2 border-b border-border bg-table-header text-[11px] font-semibold tracking-wider uppercase text-muted-foreground items-center"
					aria-hidden="true"
				>
					<div>Monitor</div>
					<div>Activity</div>
					<div>Status</div>
					<div>Response</div>
					<div>Uptime 30d</div>
					<div>Interval</div>
					<div>Last check</div>
					<div class="text-right">Actions</div>
				</div>

				<div data-lenis-prevent class="divide-y divide-border/60">
					{#each monitors as monitor (monitor.ID)}
						<MonitorItem
							{monitor}
							{canManage}
							viewMode="list"
							handleDeleteMonitor={async (id) => {
								await onDelete(id);
							}}
							handleEditMonitor={onEdit}
							handleTogglePause={async (m) => {
								await onTogglePause(m);
							}}
						/>
					{/each}
				</div>
			</div>
		</div>
	</div>
</div>
