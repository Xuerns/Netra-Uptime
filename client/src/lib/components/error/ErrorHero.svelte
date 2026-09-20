<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Radio } from 'lucide-svelte';

	interface Props {
		statusCode?: number;
		errorMessage?: string;
	}

	let { statusCode = 404, errorMessage = 'Endpoint does not exist.' }: Props = $props();

	function handleGoBack(): void {
		if (typeof window !== 'undefined') {
			window.history.back();
		}
	}
</script>

<div class="flex flex-col justify-between h-full space-y-10">
	<div>	
		<!-- Giant Architectural 404 Headline -->
		<div class="relative">
			<div
				class="text-[120px] sm:text-[180px] lg:text-[220px] font-mono font-bold leading-none tracking-tighter text-[#1f2024] select-none"
			>
				{statusCode}
			</div>
			<div class="absolute inset-0 flex flex-col justify-center">
				<h1 class="text-4xl sm:text-6xl lg:text-7xl font-bold tracking-tight text-white leading-tight">
					Route Severed.<br />
					<span class="text-muted-foreground font-normal">{errorMessage}</span>
				</h1>
			</div>
		</div>

		<!-- Explanatory Technical Statement -->
		<p class="mt-8 text-base sm:text-lg lg:text-xl text-muted-foreground leading-relaxed font-normal max-w-full">
			The requested URL path was not recognized by our edge router table. The resource has either
			been decommissioned, relocated to an internal cluster, or never provisioned.
		</p>
	</div>

	<!-- Action & Re-Routing Matrix -->
	<div class="border-t border-border pt-10">
		<div class="flex flex-wrap items-center gap-5 font-mono text-sm sm:text-base">
			<Button
				href="/dashboard"
				class="rounded bg-primary text-[#0e0e0f] font-semibold px-6 py-3.5 hover:bg-[#12c98d] transition-colors duration-150 inline-flex items-center gap-2"
			>
				<span>←</span>
				<span>Back to Home</span>
			</Button>

			<Button
				variant="outline"
				class="rounded border border-border bg-card text-white px-6 py-3.5 hover:bg-surface-panel transition-colors duration-150 inline-flex items-center gap-2"
				onclick={handleGoBack}
			>
				Previous Page
			</Button>
		</div>
	</div>
</div>
