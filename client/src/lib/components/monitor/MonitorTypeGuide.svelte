<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { HelpCircle } from 'lucide-svelte';

	let { monitorType } = $props<{ monitorType: 'http' | 'ping' | 'tcp' | null }>();
</script>

<Tooltip.Provider delayDuration={150}>
	<Tooltip.Root>
		<Tooltip.Trigger>
			<button
				type="button"
				class="text-muted-foreground hover:text-foreground transition-colors p-1 flex items-center justify-center rounded-full hover:bg-white/10"
			>
				<HelpCircle size={18} strokeWidth={2} />
			</button>
		</Tooltip.Trigger>
		<Tooltip.Content
			side="right"
			sideOffset={12}
			class="max-w-[320px] bg-card border border-border p-4 shadow-xl z-[100] rounded-xl"
		>
			<div class="space-y-4">
				{#if monitorType === 'http' || !monitorType}
					<div class={!monitorType ? 'border-b border-border pb-3' : ''}>
						<h4 class="text-sm font-semibold text-foreground mb-1">HTTP/API Monitor</h4>
						<p class="text-[13px] leading-snug text-muted-foreground mb-2">
							Checks website or API endpoint availability. Expects a standard URL.
						</p>
						<code
							class="text-[11px] font-mono bg-foreground/20 border border-border px-2 py-1 rounded-md text-primary block w-fit"
							>https://api.example.com</code
						>
					</div>
				{/if}
				{#if monitorType === 'ping' || !monitorType}
					<div class={!monitorType ? 'border-b border-border pb-3' : ''}>
						<h4 class="text-sm font-semibold text-foreground mb-1">Ping Monitor</h4>
						<p class="text-[13px] leading-snug text-muted-foreground mb-2">
							Checks if a server is reachable via ICMP. Expects a hostname or IP address.
						</p>
						<code
							class="text-[11px] font-mono bg-foreground/20 border border-border px-2 py-1 rounded-md text-primary block w-fit"
							>8.8.8.8</code
						>
					</div>
				{/if}
				{#if monitorType === 'tcp' || !monitorType}
					<div>
						<h4 class="text-sm font-semibold text-foreground mb-1">TCP Port Monitor</h4>
						<p class="text-[13px] leading-snug text-muted-foreground mb-2">
							Checks if a specific port is open on a server. Expects host:port format.
						</p>
						<code
							class="text-[11px] font-mono bg-foreground/20 border border-border px-2 py-1 rounded-md text-primary block w-fit"
							>db.example.com:5432</code
						>
					</div>
				{/if}
			</div>
		</Tooltip.Content>
	</Tooltip.Root>
</Tooltip.Provider>
