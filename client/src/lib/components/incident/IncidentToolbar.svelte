<script lang="ts">
	import { Radio, CheckCircle2, ChevronDown, Check } from 'lucide-svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { headerStore } from '$lib/stores/headerStore.svelte';

	export type ActiveFilter = 'active' | 'resolved';
	export type SeverityFilter = 'all' | 'sev1' | 'sev2' | 'sev3' | 'sev4';

	interface Props {
		activeFilter: ActiveFilter;
		severityFilter: SeverityFilter;
		activeCount: number;
		resolvedCount: number;
		totalFiltered: number;
		onSelectActiveFilter: (v: ActiveFilter) => void;
		onSelectSeverityFilter: (v: SeverityFilter) => void;
	}

	let {
		activeFilter,
		severityFilter,
		activeCount,
		resolvedCount,
		totalFiltered,
		onSelectActiveFilter,
		onSelectSeverityFilter
	}: Props = $props();
</script>

<div
	class="bg-table-header border-b border-border px-4 py-2 flex items-center justify-between text-xs text-muted-foreground"
>
	<!-- Left: Status Filter Tabs (Active vs Resolved) -->
	<div class="flex items-center gap-5">
		<!-- Active / Open Tab -->
		<button
			type="button"
			class="flex items-center gap-1.5 transition-colors cursor-pointer {activeFilter === 'active'
				? 'font-bold text-foreground'
				: 'font-medium text-foreground/40 hover:text-foreground/70'}"
			onclick={() => onSelectActiveFilter('active')}
		>
			<Radio
				class="size-3.5 {activeFilter === 'active' ? 'text-rose-400' : 'text-foreground/40'}"
			/>
			<span>Active</span>
			<span
				class="px-1.5 py-0.5 text-[10.5px] rounded-full font-mono font-semibold transition-colors {activeFilter ===
				'active'
					? 'bg-rose-500/20 text-rose-300'
					: 'bg-white/5 text-foreground/40'}"
			>
				{activeCount}
			</span>
		</button>

		<!-- Resolved / Closed Tab -->
		<button
			type="button"
			class="flex items-center gap-1.5 transition-colors cursor-pointer {activeFilter === 'resolved'
				? 'font-bold text-foreground'
				: 'font-medium text-foreground/40 hover:text-foreground/70'}"
			onclick={() => onSelectActiveFilter('resolved')}
		>
			<CheckCircle2
				class="size-3.5 {activeFilter === 'resolved' ? 'text-emerald-400' : 'text-foreground/40'}"
			/>
			<span>Resolved</span>
			<span
				class="px-1.5 py-0.5 text-[10.5px] rounded-full font-mono font-semibold transition-colors {activeFilter ===
				'resolved'
					? 'bg-emerald-500/20 text-emerald-300'
					: 'bg-white/5 text-foreground/40'}"
			>
				{resolvedCount}
			</span>
		</button>
	</div>

	<!-- Right: Severity Filter Dropdown & Stats -->
	<div class="flex items-center gap-3 text-xs">
		<!-- Severity Dropdown Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="flex items-center gap-1.5 px-2 py-1 rounded-md border text-xs transition-colors cursor-pointer focus:outline-none {severityFilter !==
				'all'
					? 'bg-white/10 border-border text-foreground font-medium'
					: 'bg-transparent border-transparent hover:bg-white/5 text-muted-foreground hover:text-foreground'}"
			>
				<span class="flex items-center gap-1.5">
					{#if severityFilter === 'sev1'}
						<span class="size-2 rounded-full bg-rose-500"></span>
						<span>Critical (Sev 1)</span>
					{:else if severityFilter === 'sev2'}
						<span class="size-2 rounded-full bg-amber-500"></span>
						<span>High (Sev 2)</span>
					{:else if severityFilter === 'sev3'}
						<span class="size-2 rounded-full bg-sky-500"></span>
						<span>Medium (Sev 3)</span>
					{:else if severityFilter === 'sev4'}
						<span class="size-2 rounded-full bg-slate-400"></span>
						<span>Low (Sev 4)</span>
					{:else}
						<span>Severity: All</span>
					{/if}
				</span>
				<ChevronDown class="size-3 text-muted-foreground" />
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				align="end"
				class="w-48 p-1.5 bg-card border border-border text-foreground rounded-lg"
			>
				<DropdownMenu.Item
					class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
					'all'
						? 'text-primary font-medium'
						: 'text-foreground'}"
					onclick={() => onSelectSeverityFilter('all')}
				>
					<span>All Severities</span>
					{#if severityFilter === 'all'}
						<Check class="size-3.5 text-primary" />
					{/if}
				</DropdownMenu.Item>
				<DropdownMenu.Separator class="bg-border my-1" />
				<DropdownMenu.Item
					class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
					'sev1'
						? 'text-rose-400 font-medium'
						: 'text-foreground'}"
					onclick={() => onSelectSeverityFilter('sev1')}
				>
					<div class="flex items-center gap-2">
						<span class="size-2 rounded-full bg-rose-500"></span>
						<span>Sev 1 (Critical)</span>
					</div>
					{#if severityFilter === 'sev1'}
						<Check class="size-3.5 text-rose-400" />
					{/if}
				</DropdownMenu.Item>
				<DropdownMenu.Item
					class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
					'sev2'
						? 'text-amber-400 font-medium'
						: 'text-foreground'}"
					onclick={() => onSelectSeverityFilter('sev2')}
				>
					<div class="flex items-center gap-2">
						<span class="size-2 rounded-full bg-amber-500"></span>
						<span>Sev 2 (High)</span>
					</div>
					{#if severityFilter === 'sev2'}
						<Check class="size-3.5 text-amber-400" />
					{/if}
				</DropdownMenu.Item>
				<DropdownMenu.Item
					class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
					'sev3'
						? 'text-sky-400 font-medium'
						: 'text-foreground'}"
					onclick={() => onSelectSeverityFilter('sev3')}
				>
					<div class="flex items-center gap-2">
						<span class="size-2 rounded-full bg-sky-500"></span>
						<span>Sev 3 (Medium)</span>
					</div>
					{#if severityFilter === 'sev3'}
						<Check class="size-3.5 text-sky-400" />
					{/if}
				</DropdownMenu.Item>
				<DropdownMenu.Item
					class="flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md cursor-pointer hover:bg-white/10 {severityFilter ===
					'sev4'
						? 'text-slate-300 font-medium'
						: 'text-foreground'}"
					onclick={() => onSelectSeverityFilter('sev4')}
				>
					<div class="flex items-center gap-2">
						<span class="size-2 rounded-full bg-slate-400"></span>
						<span>Sev 4 (Low)</span>
					</div>
					{#if severityFilter === 'sev4'}
						<Check class="size-3.5 text-slate-300" />
					{/if}
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>

		<!-- Subtle Vertical Separator -->
		<div class="h-4 w-px bg-border/60"></div>

		<!-- Stats & Count -->
		<div class="text-[11.5px] font-mono text-foreground/40 whitespace-nowrap">
			<span>Showing {totalFiltered} incident{totalFiltered === 1 ? '' : 's'}</span>
			{#if headerStore.searchQuery || severityFilter !== 'all'}
				<span class="text-primary font-sans ml-1">(filtered)</span>
			{/if}
		</div>
	</div>
</div>
