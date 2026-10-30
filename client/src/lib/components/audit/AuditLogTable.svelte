<script lang="ts">
	import { ListFilter, X, CheckCircle2 } from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import type { AuditLogItem } from '$lib/api/audit';
	import type { ComponentType } from 'svelte';

	export interface ActionFilterGroup {
		categoryId: string;
		label: string;
		icon: ComponentType;
		actions: Array<{ value: string; label: string }>;
	}

	interface Props {
		logs: AuditLogItem[];
		isFetching: boolean;
		selectedAction: string;
		selectedActionLabel: string;
		visibleActionGroups: ActionFilterGroup[];
		getActionStyle: (action: string) => string;
		onActionFilterChange: (actionVal: string) => void;
		fmtTimeAgo: (iso: string) => string;
		fmtFullDate: (iso: string) => string;
		initials: (name?: string) => string;
	}

	let {
		logs,
		isFetching,
		selectedAction,
		selectedActionLabel,
		visibleActionGroups,
		getActionStyle,
		onActionFilterChange,
		fmtTimeAgo,
		fmtFullDate,
		initials
	}: Props = $props();

	let tableContainer = $state<HTMLDivElement | null>(null);

	$effect(() => {
		if (logs && tableContainer) {
			tableContainer.scrollTop = 0;
		}
	});
</script>

<div
	class="border border-border rounded-md bg-card overflow-hidden flex flex-col max-h-[calc(100vh-16rem)] transition-opacity duration-200 {isFetching
		? 'opacity-50 pointer-events-none'
		: 'opacity-100'}"
>
	<!-- Table Column Headers -->
	<div class="log-table-header sticky top-0 bg-table-header z-10">
		<div class="log-col-actor">Actor</div>
		<div class="log-col-action">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger class="focus:outline-none w-full">
					<span
						class="flex items-center gap-1.5 transition-colors cursor-pointer {selectedAction
							? 'text-primary'
							: 'text-inherit hover:text-foreground'}"
					>
						<span class="truncate">{selectedActionLabel || 'Action'}</span>
						{#if selectedAction}
							<!-- svelte-ignore a11y_no_static_element_interactions -->
							<span
								role="button"
								tabindex="-1"
								class="size-3.5 rounded-full bg-primary/20 hover:bg-primary/40 flex items-center justify-center transition-colors shrink-0"
								onclick={(e) => {
									e.stopPropagation();
									onActionFilterChange('');
								}}
								onkeydown={(e) => {
									if (e.key === 'Enter') {
										e.stopPropagation();
										onActionFilterChange('');
									}
								}}
							>
								<X class="size-2.5 text-primary" />
							</span>
						{:else}
							<ListFilter class="size-3 opacity-50 shrink-0" />
						{/if}
					</span>
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					data-lenis-prevent
					class="w-72 max-h-90 overflow-y-auto bg-card border-border text-foreground rounded-lg p-1.5 shadow-xl z-50"
					align="start"
					sideOffset={6}
				>
					<!-- All Actions option -->
					<DropdownMenu.Item
						class="cursor-pointer flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-white/8 text-xs font-medium {selectedAction ===
						''
							? 'text-primary bg-white/5 font-semibold'
							: 'text-foreground'}"
						onclick={() => onActionFilterChange('')}
					>
						<span>All Actions</span>
						{#if selectedAction === ''}
							<CheckCircle2 class="size-3 text-primary shrink-0" />
						{/if}
					</DropdownMenu.Item>

					{#each visibleActionGroups as group (group.categoryId)}
						<DropdownMenu.Separator class="bg-white/8 my-1" />
						<DropdownMenu.Label
							class="text-[10px] text-muted-foreground font-medium px-2.5 py-1 uppercase tracking-wider flex items-center gap-1.5"
						>
							<group.icon class="size-3 opacity-70" />
							<span>{group.label}</span>
						</DropdownMenu.Label>
						{#each group.actions as action (action.value)}
							{@const isSelected = selectedAction === action.value}
							{@const style = getActionStyle(action.value)}
							<DropdownMenu.Item
								class="cursor-pointer flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md hover:bg-white/8 text-xs {isSelected
									? 'text-primary bg-white/5 font-semibold'
									: 'text-foreground'}"
								onclick={() => onActionFilterChange(action.value)}
							>
								<span class="font-mono text-[11px] {style}">{action.label}</span>
								{#if isSelected}
									<CheckCircle2 class="size-3 text-primary shrink-0" />
								{/if}
							</DropdownMenu.Item>
						{/each}
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
		<div class="log-col-desc">Event / Details</div>
		<div class="log-col-ip">IP Address</div>
		<div class="log-col-date text-right">Time</div>
	</div>

	<!-- Log Rows (Scrollable) -->
	<div
		bind:this={tableContainer}
		data-lenis-prevent
		class="overflow-y-auto flex-1 divide-y divide-white/5"
	>
		{#each logs as log (log.id)}
			{@const actionStyle = getActionStyle(log.action)}
			<div
				class="log-table-row hover:bg-white/3 transition-colors text-xs py-2 px-4 flex flex-col md:grid md:grid-cols-[200px_160px_1fr_130px_140px] items-start md:items-center gap-2 md:gap-4"
			>
				<!-- Col 1: Actor -->
				<div class="flex items-center gap-2.5 min-w-0 w-full md:w-auto">
					<div
						class="size-6 rounded-sm overflow-hidden bg-primary/10 text-primary flex items-center justify-center font-bold text-[11px] shrink-0"
					>
						{#if log.user_avatar_url}
							<img
								src={log.user_avatar_url}
								alt={log.user_name || 'User'}
								class="size-full object-cover"
							/>
						{:else}
							{initials(log.user_name)}
						{/if}
					</div>
					<div class="min-w-0 flex-1">
						<p class="text-foreground font-medium truncate text-xs">
							{log.user_name || 'Unknown'}
						</p>
						<p class="text-[10px] text-muted-foreground font-mono truncate">
							{log.user_email || 'System'}
						</p>
					</div>
				</div>

				<!-- Col 2: Action -->
				<div class="w-full md:w-auto">
					<span class="inline-flex items-center font-mono text-[11px] font-medium {actionStyle}">
						{log.action}
					</span>
				</div>

				<!-- Col 3: Description -->
				<div class="min-w-0 w-full md:w-auto">
					<p class="text-foreground/90 font-normal truncate leading-relaxed">
						{log.description}
					</p>
					{#if log.details}
						<p class="text-[11px] text-muted-foreground/60 font-mono truncate mt-0.5">
							{log.details}
						</p>
					{/if}
				</div>

				<!-- Col 4: IP Address -->
				<div class="text-[11px] font-mono text-muted-foreground/60 w-full md:w-auto">
					{log.ip_address || '—'}
				</div>

				<!-- Col 5: Timestamp -->
				<div class="text-right w-full md:w-auto" title={fmtFullDate(log.created_at)}>
					<span class="text-xs text-muted-foreground/80 font-medium">
						{fmtTimeAgo(log.created_at)}
					</span>
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.log-table-header {
		display: none;
		padding: 0.5rem 1rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--muted-foreground);
	}

	@media (min-width: 768px) {
		.log-table-header {
			display: grid;
			grid-template-columns: 200px 160px 1fr 130px 140px;
			gap: 1rem;
			align-items: center;
		}
	}
</style>
