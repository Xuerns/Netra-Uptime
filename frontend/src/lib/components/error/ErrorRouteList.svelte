<script lang="ts">
	import ErrorRouteCard from './ErrorRouteCard.svelte';

	const ingressRoutes = [
		{
			index: '01',
			title: 'Monitors',
			href: '/dashboard/monitor',
			description: 'Global telemetry, latency curves & active monitors overview'
		},
		{
			index: '02',
			title: 'Incidents',
			href: '/dashboard/incidents',
			description: 'Real-time triage dispatch, active alarms & audit logs'
		},
		{
			index: '03',
			title: 'Status Page',
			href: '/dashboard/statuspage',
			description: 'Public endpoint health definitions & uptime SLA reporting'
		},
		{
			index: '04',
			title: 'Organizations',
			href: '/dashboard/organization',
			description: 'Workspace access controls, tenant isolation & team members'
		},
		{
			index: '05',
			title: 'Logs',
			href: '/dashboard/logs',
			description: 'High-throughput stream ingestion & distributed event traces'
		}
	];
</script>

<aside class="flex flex-col border border-border bg-card rounded-lg p-6 sm:p-7 space-y-5">
	<div class="flex items-center justify-between pb-4 border-b border-border">
		<div class="flex items-center gap-2.5">
			<span class="size-2 bg-primary rounded-none inline-block"></span>
			<span class="text-xs sm:text-sm font-mono font-semibold tracking-wider text-foreground uppercase">
				Available Ingress Routes
			</span>
		</div>
		<span class="text-xs font-mono text-muted-foreground">EDGE_CLUSTER // OK</span>
	</div>

	<div class="flex flex-col space-y-3.5">
		{#each ingressRoutes as route (route.index)}
			<ErrorRouteCard
				index={route.index}
				title={route.title}
				href={route.href}
				description={route.description}
			/>
		{/each}
	</div>
</aside>
