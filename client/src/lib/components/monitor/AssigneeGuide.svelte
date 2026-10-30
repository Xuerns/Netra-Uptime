<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { HelpCircle, Info } from 'lucide-svelte';
</script>

<Tooltip.Provider delayDuration={150}>
	<Tooltip.Root>
		<Tooltip.Trigger>
			<button
				type="button"
				class="text-muted-foreground hover:text-foreground transition-colors p-0.5 flex items-center justify-center rounded-full hover:bg-white/10 cursor-pointer"
				aria-label="What is an Assignee?"
			>
				<HelpCircle size={15} strokeWidth={2} />
			</button>
		</Tooltip.Trigger>
		<Tooltip.Content
			side="top"
			align="start"
			sideOffset={8}
			class="max-w-[300px] bg-card border border-border p-3.5 shadow-2xl z-[100] rounded-md text-foreground text-left"
		>
			<div class="space-y-2">
				<div class="flex items-center gap-1.5 font-semibold text-xs text-primary">
					<Info class="size-3.5 shrink-0" />
					<span>Monitor Assignees</span>
				</div>
				<p class="text-[12px] leading-relaxed text-muted-foreground">
					Assignees are team members designated to respond to incidents on this monitor.
				</p>
				<div class="space-y-1 text-[11.5px] text-muted-foreground/90 border-t border-border pt-2">
					<div class="flex items-start gap-1.5">
						<span class="text-primary font-bold">&bull;</span>
						<span
							><strong class="text-foreground/90">Direct Alerts:</strong> Receive personal email notifications
							when this monitor goes down or recovers.</span
						>
					</div>
					<div class="flex items-start gap-1.5">
						<span class="text-primary font-bold">&bull;</span>
						<span
							><strong class="text-foreground/90">Optional:</strong> If left empty, alerts still route
							to your organization channels (Slack, Discord, Webhook).</span
						>
					</div>
				</div>
			</div>
		</Tooltip.Content>
	</Tooltip.Root>
</Tooltip.Provider>
