<script lang="ts">
	import logo from '$lib/assets/Logo.svg';

	let {
		title,
		overallStatus
	}: {
		title: string;
		overallStatus: string;
	} = $props();
</script>

<footer class="w-full mt-auto border-t border-border/60 bg-background/50 backdrop-blur-sm">
	<div
		class="max-w-5xl mx-auto flex flex-col sm:flex-row justify-between items-center px-4 sm:px-6 py-6 w-full text-xs text-muted-foreground gap-3"
	>
		<span>
			{#if overallStatus === 'operational'}
				© 2026 {title}. All systems operational.
			{:else if overallStatus === 'major_outage'}
				© 2026 {title}. Systems currently experiencing a major outage.
			{:else}
				© 2026 {title}. Some systems are experiencing issues.
			{/if}
		</span>
		<div class="flex items-center gap-1.5 font-mono">
			<span>Powered by</span>
			<span class="text-foreground font-semibold flex items-center gap-1">
				<img class="h-3.5 w-3.5 object-contain" src={logo} alt="" />
				Netra Uptime
			</span>
		</div>
	</div>
</footer>
