<script lang="ts">
	import { ExternalLink } from 'lucide-svelte';

	const currentYear = new Date().getFullYear();
</script>

<footer
	class="border-t border-border px-6 py-4 flex flex-col sm:flex-row items-center justify-between` text-xs font-mono text-muted-foreground bg-card z-10 gap-3"
>
	<div>&copy; {currentYear} Netra Uptime. All rights reserved.</div>
</footer>
