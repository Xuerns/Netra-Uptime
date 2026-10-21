<script lang="ts">
	import type { AlertTemplate } from '$lib/api/alertTemplate';
	import { formatDateTime } from '$lib/utils/datetime';
	import { userTimezone } from '$lib/stores/timezone';
	import { ShieldAlert, Monitor, Clock, ChevronRight } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';

	interface Props {
		template: AlertTemplate;
		orgId: string;
	}

	let { template, orgId }: Props = $props();
	const tz = $derived($userTimezone);
</script>

<div
	class="group flex flex-col md:grid md:grid-cols-12 gap-3 md:gap-4 px-5 py-4 items-start md:items-center bg-card hover:bg-white/4 transition-colors"
>
	<!-- Column 1: Template Name & Description -->
	<div class="col-span-5 min-w-0 w-full space-y-1">
		<div class="flex items-center gap-2 flex-wrap">
			<a
				href={`/dashboard/org/${orgId}/alert-templates/${template.id}`}
				class="text-[14.5px] font-semibold text-foreground hover:text-primary transition-colors truncate"
			>
				{template.name}
			</a>
			{#if template.is_default}
				<Badge
					variant="outline"
					class="px-2 py-0.5 text-[10.5px] font-medium bg-primary/10 border-primary/20 text-primary"
				>
					Default
				</Badge>
			{/if}
		</div>
		{#if template.description}
			<p class="text-xs text-muted-foreground line-clamp-1">
				{template.description}
			</p>
		{/if}
	</div>

	<!-- Column 2: Rules Count -->
	<div class="col-span-2 flex items-center gap-1.5 text-xs text-muted-foreground">
		<ShieldAlert class="size-3.5 text-muted-foreground/70 shrink-0" />
		<span>
			<b class="text-foreground font-mono">{template.rules?.length || 0}</b>
			{template.rules?.length === 1 ? 'rule' : 'rules'}
		</span>
	</div>

	<!-- Column 3: Linked Monitors Count -->
	<div class="col-span-2 flex items-center gap-1.5 text-xs text-muted-foreground">
		<Monitor class="size-3.5 text-muted-foreground/70 shrink-0" />
		<span>
			<b class="text-foreground font-mono">{template.monitors_count}</b>
			{template.monitors_count === 1 ? 'monitor' : 'monitors'}
		</span>
	</div>

	<!-- Column 4: Last Updated -->
	<div class="col-span-2 flex items-center gap-1.5 text-xs font-mono text-muted-foreground/70">
		<Clock class="size-3 text-muted-foreground/50 shrink-0" />
		<span class="truncate">{formatDateTime(template.updated_at, tz)}</span>
	</div>

	<!-- Column 5: Forward Chevron -->
	<div class="col-span-1 hidden md:flex items-center justify-end">
		<a
			href={`/dashboard/org/${orgId}/alert-templates/${template.id}`}
			class="p-1 rounded text-muted-foreground/40 group-hover:text-primary transition-colors cursor-pointer"
			aria-label="View template details"
		>
			<ChevronRight class="size-4 group-hover:translate-x-0.5 transition-transform" />
		</a>
	</div>
</div>
