<script lang="ts">
	import { Wrench, Calendar } from 'lucide-svelte';
	import MaintenanceCard from '$lib/components/statuspage/MaintenanceCard.svelte';

	let {
		ongoingMaintenances = [],
		upcomingMaintenances = []
	}: {
		ongoingMaintenances?: any[];
		upcomingMaintenances?: any[];
	} = $props();
</script>

{#if ongoingMaintenances.length > 0}
	<section class="space-y-4">
		<div class="flex items-center gap-2">
			<Wrench class="size-4 text-blue-400" />
			<h2 class="text-lg font-semibold tracking-tight text-foreground">Ongoing Maintenance</h2>
		</div>
		<div class="space-y-3">
			{#each ongoingMaintenances as maintenance (maintenance.id)}
				<MaintenanceCard {maintenance} isOngoing={true} />
			{/each}
		</div>
	</section>
{/if}

{#if upcomingMaintenances.length > 0}
	<section class="space-y-4">
		<div class="flex items-center gap-2">
			<Calendar class="size-4 text-indigo-400" />
			<h2 class="text-lg font-semibold tracking-tight text-foreground">Scheduled Maintenance</h2>
		</div>
		<div class="space-y-3">
			{#each upcomingMaintenances as maintenance (maintenance.id)}
				<MaintenanceCard {maintenance} isOngoing={false} />
			{/each}
		</div>
	</section>
{/if}
