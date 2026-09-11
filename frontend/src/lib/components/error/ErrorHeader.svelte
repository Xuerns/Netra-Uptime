<script lang="ts">
	import { Activity } from 'lucide-svelte';
	import Logo from '$lib/assets/Logo.svg';
	interface Props {
		statusCode?: number;
	}

	let { statusCode = 404 }: Props = $props();
</script>

<header class="border-b border-border bg-card px-6 py-3.5 flex items-center justify-between z-10">
	<div class="flex items-center gap-3">
		<a href="/" class="flex items-center gap-2 text-foreground">
			<img src={Logo} alt="Logo" class="size-5 text-primary" />
			<span class="font-bold tracking-tight text-sm">PulseWatch</span>
		</a>
		<span class="text-xs text-muted-foreground font-mono">/</span>
		<span class="text-xs font-mono text-muted-foreground uppercase tracking-wider">
			{statusCode === 404 ? 'Route Severed' : 'System Error'}
		</span>
	</div>
</header>
