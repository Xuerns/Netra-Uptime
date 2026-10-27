<script lang="ts">
	import * as Tabs from '$lib/components/ui/tabs/index.js';

	interface Counts {
		total: number;
		owners: number;
		admins: number;
		members: number;
	}

	interface Props {
		value: string;
		counts: Counts;
		onChange?: (role: string) => void;
	}

	let { value = $bindable('all'), counts, onChange }: Props = $props();

	function handleValueChange(val: string) {
		value = val;
		if (onChange) {
			onChange(val);
		}
	}
</script>

<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-2">
	<Tabs.Root {value} onValueChange={handleValueChange}>
		<Tabs.List class="h-9">
			<Tabs.Trigger value="all" class="text-xs font-medium cursor-pointer">
				All <span class="ml-1 text-[11px] font-mono text-muted-foreground">({counts.total})</span>
			</Tabs.Trigger>
			<Tabs.Trigger value="owner" class="text-xs font-medium cursor-pointer">
				Owners <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({counts.owners})</span
				>
			</Tabs.Trigger>
			<Tabs.Trigger value="admin" class="text-xs font-medium cursor-pointer">
				Admins <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({counts.admins})</span
				>
			</Tabs.Trigger>
			<Tabs.Trigger value="member" class="text-xs font-medium cursor-pointer">
				Members <span class="ml-1 text-[11px] font-mono text-muted-foreground"
					>({counts.members})</span
				>
			</Tabs.Trigger>
		</Tabs.List>
	</Tabs.Root>
</div>
