<script lang="ts">
	import { onMount } from 'svelte';
	import { AlertTemplateAPI, type AlertTemplate } from '$lib/api/alertTemplate';
	import { MonitorAPI } from '$lib/api/monitor';
	import {
		ShieldAlert,
		AlertTriangle,
		Layers,
		Info,
		ExternalLink,
		CheckCircle2,
		Clock,
		ArrowRight
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		orgId: string;
		monitorId: string;
		monitorType?: string;
		alertTemplateId?: string | null;
	}

	let { orgId, monitorId, monitorType = 'http', alertTemplateId = null }: Props = $props();

	let isHttp = $derived(monitorType?.toLowerCase() === 'http');

	let availableTemplates = $state<AlertTemplate[]>([]);
	let currentTemplateId = $state<string | null>(alertTemplateId);
	let isApplyingTemplate = $state(false);
	let isLoading = $state(true);

	let activeTemplate = $derived(availableTemplates.find((t) => t.id === currentTemplateId) || null);

	async function loadTemplates() {
		isLoading = true;
		try {
			const res = await AlertTemplateAPI.list(orgId);
			availableTemplates = res.data || [];
			// If no template is linked yet, check if there is an organization default template
			if (!currentTemplateId) {
				const defaultTmpl = availableTemplates.find((t) => t.is_default);
				if (defaultTmpl) {
					currentTemplateId = defaultTmpl.id;
					// Save to monitor in background
					MonitorAPI.applyTemplate(orgId, monitorId, defaultTmpl.id).catch(() => {});
				}
			}
		} catch (err: unknown) {
			console.error('Failed to load alert templates', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleTemplateChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		const newTmplId = target.value ? target.value : null;

		isApplyingTemplate = true;
		try {
			await MonitorAPI.applyTemplate(orgId, monitorId, newTmplId);
			currentTemplateId = newTmplId;
			toast.success(
				newTmplId ? 'Alert template linked to monitor' : 'Switched to System Default Checks'
			);
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to update monitor template'));
		} finally {
			isApplyingTemplate = false;
		}
	}

	onMount(() => {
		if (orgId && monitorId) {
			loadTemplates();
		}
	});
</script>

<div class="bg-card border border-border rounded-lg p-5 md:p-6 space-y-5">
	<!-- Section Header -->
	<div class="flex items-center justify-between gap-4 flex-wrap">
		<div class="space-y-1">
			<div class="flex items-center gap-2">
				<h3 class="text-sm font-semibold text-foreground tracking-tight">
					Alert Policy & Template
				</h3>
				<span
					class="px-2 py-0.5 rounded-full text-[10px] font-medium bg-primary/10 text-primary border border-primary/20"
				>
					Template-First
				</span>
			</div>
			<p class="text-xs text-muted-foreground leading-relaxed">
				All incident thresholds and warning rules are inherited centrally from your selected Alert
				Template.
			</p>
		</div>

		<a
			href={`/dashboard/org/${orgId}/alert-templates`}
			class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-8 px-3 rounded-lg font-medium text-xs cursor-pointer shrink-0"
		>
			<span>Manage Templates</span>
			<ExternalLink class="size-3" />
		</a>
	</div>

	<!-- Template Selector Bar -->
	<div
		class="flex items-center justify-between p-4 rounded-xl border border-border/70 bg-muted/20 gap-3 flex-wrap sm:flex-nowrap"
	>
		<div class="flex items-center gap-3 min-w-0">
			<div
				class="size-9 rounded-lg bg-primary/10 border border-primary/20 text-primary flex items-center justify-center shrink-0"
			>
				<Layers class="size-4.5" />
			</div>
			<div class="min-w-0">
				<div class="text-xs font-semibold text-foreground flex items-center gap-1.5">
					<span>Assigned Template</span>
					{#if activeTemplate?.is_default}
						<span
							class="text-[10px] text-primary bg-primary/10 px-1.5 py-0.2 rounded border border-primary/20 font-medium"
						>
							Default
						</span>
					{/if}
				</div>
				<div class="text-[11px] text-muted-foreground truncate">
					{#if activeTemplate}
						{activeTemplate.name} &bull; {activeTemplate.rules?.length || 0} active rules
					{:else}
						No template attached &bull; Using Netra Uptime System Defaults
					{/if}
				</div>
			</div>
		</div>

		<div class="flex items-center gap-2 shrink-0 w-full sm:w-auto">
			<select
				value={currentTemplateId || ''}
				onchange={handleTemplateChange}
				disabled={isApplyingTemplate || isLoading}
				class="h-9 px-3 rounded-lg border border-border bg-background text-xs font-medium text-foreground focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer w-full sm:w-auto min-w-55"
			>
				<option value="">System Default Checks</option>
				{#each availableTemplates as tmpl (tmpl.id)}
					<option value={tmpl.id}>
						{tmpl.name} ({tmpl.rules?.length || 0} rules) {tmpl.is_default ? '★' : ''}
					</option>
				{/each}
			</select>
		</div>
	</div>

	<!-- Rules Content Area -->
	{#if isLoading}
		<div class="h-28 flex items-center justify-center text-xs text-muted-foreground font-mono">
			Loading alert rules & policies...
		</div>
	{:else if activeTemplate}
		<!-- Template Active & Rules Display -->
		<div class="space-y-3">
			<div
				class="flex items-center justify-between text-xs text-muted-foreground font-medium px-0.5"
			>
				<span class="flex items-center gap-1.5">
					<Info class="size-3.5 text-primary" />
					Rules defined in <b>{activeTemplate.name}</b>:
				</span>
				<a
					href={`/dashboard/org/${orgId}/alert-templates`}
					class="text-primary hover:underline text-[11px] flex items-center gap-1"
				>
					Edit rules in Templates <ExternalLink class="size-3" />
				</a>
			</div>

			{#if !activeTemplate.rules || activeTemplate.rules.length === 0}
				<div
					class="rounded-xl border border-dashed border-border/80 bg-muted/10 p-5 text-center text-xs text-muted-foreground space-y-1.5"
				>
					<p class="font-medium text-foreground">No rules defined in this template yet.</p>
					<p class="text-[11px] max-w-100 mx-auto">
						This monitor will safely use system default checks until you add rules to this template
						in the Alert Templates manager.
					</p>
				</div>
			{:else}
				<div class="grid grid-cols-1 gap-2.5">
					{#each activeTemplate.rules as r (r.id)}
						<div
							class="flex items-center justify-between bg-white/3 border border-border/80 hover:border-border rounded-xl p-3.5 text-xs gap-3 transition-colors flex-wrap sm:flex-nowrap"
						>
							<div class="flex items-center gap-3 min-w-0">
								{#if r.severity === 'critical'}
									<div
										class="size-8 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 flex items-center justify-center shrink-0"
									>
										<ShieldAlert class="size-4" />
									</div>
								{:else}
									<div
										class="size-8 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center shrink-0"
									>
										<AlertTriangle class="size-4" />
									</div>
								{/if}
								<div class="truncate">
									<div class="font-medium text-foreground text-xs truncate flex items-center gap-2">
										<span>{r.name}</span>
										{#if !r.is_active}
											<span
												class="text-[10px] px-1.5 py-0.2 rounded bg-muted text-muted-foreground font-mono"
											>
												Disabled
											</span>
										{/if}
									</div>
									<div
										class="text-[11px] text-muted-foreground font-mono flex items-center gap-1.5 mt-0.5"
									>
										<span class="text-foreground/90 font-medium">Condition:</span>
										<span>{r.metric} {r.operator} {r.target_value}</span>
										<span class="text-muted-foreground/50">&bull;</span>
										<span class="text-muted-foreground">{r.consecutive}x consecutive</span>
									</div>
								</div>
							</div>

							<div class="flex items-center gap-2 shrink-0">
								<span
									class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider bg-primary/10 text-primary border border-primary/20"
								>
									Template Rule
								</span>
								<span
									class={`px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider ${
										r.severity === 'critical'
											? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
											: 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
									}`}
								>
									{r.severity === 'critical' ? 'Critical (Down)' : 'Warning (Degraded)'}
								</span>
							</div>
						</div>
					{/each}
				</div>

				<!-- Centralized Governance Callout -->
				<div
					class="flex items-center justify-between p-3 rounded-xl bg-primary/5 border border-primary/15 text-xs text-muted-foreground"
				>
					<div class="flex items-center gap-2">
						<CheckCircle2 class="size-4 text-primary shrink-0" />
						<span>
							Centralized policy active: any update to <b>{activeTemplate.name}</b> applies immediately
							to this monitor.
						</span>
					</div>
					<a
						href={`/dashboard/org/${orgId}/alert-templates`}
						class="text-primary hover:underline font-medium text-xs flex items-center gap-1 shrink-0"
					>
						Configure Template <ArrowRight class="size-3" />
					</a>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Fallback: System Default Checks -->
		<div class="bg-white/2 border border-border/80 rounded-xl p-5 md:p-6 space-y-4">
			<div class="flex items-start gap-3.5">
				<div
					class="size-10 rounded-xl bg-primary/10 border border-primary/20 text-primary flex items-center justify-center shrink-0"
				>
					<ShieldAlert class="size-5" />
				</div>
				<div class="space-y-1.5 flex-1">
					<div class="flex items-center gap-2">
						<h4 class="text-sm font-semibold text-foreground">System Default Checks Active</h4>
						<span
							class="px-2 py-0.5 rounded-full text-[10px] font-medium bg-secondary text-muted-foreground border border-border"
						>
							Safe Defaults
						</span>
					</div>
					<p class="text-xs text-muted-foreground leading-relaxed">
						No custom alert template is attached. Checks run on standard Netra Uptime failover
						policies:
					</p>

					<div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 pt-2">
						<div class="p-3 rounded-lg bg-amber-500/5 border border-amber-500/15 space-y-1">
							<div class="flex items-center gap-1.5 text-xs font-medium text-amber-400">
								<Clock class="size-3.5" />
								<span>Unresponsive Warning (Amber)</span>
							</div>
							<p class="text-[11px] text-muted-foreground leading-relaxed">
								Triggered if pings or HTTP requests time out after <b>3 consecutive checks</b>.
								Creates a medium urgency incident with calm warning notices.
							</p>
						</div>

						<div class="p-3 rounded-lg bg-rose-500/5 border border-rose-500/15 space-y-1">
							<div class="flex items-center gap-1.5 text-xs font-medium text-rose-400">
								<ShieldAlert class="size-3.5" />
								<span>Hard Down Incident (Red)</span>
							</div>
							<p class="text-[11px] text-muted-foreground leading-relaxed">
								{#if isHttp}
									Triggered if connection is refused, DNS fails, or status code mismatches after <b
										>3 consecutive attempts</b
									>. Creates a critical incident.
								{:else}
									Triggered if port is closed or target is unreachable after <b
										>3 consecutive attempts</b
									>. Creates a critical incident.
								{/if}
							</p>
						</div>
					</div>
				</div>
			</div>

			<div class="flex items-center justify-between pt-2 border-t border-border/60 text-xs">
				<span class="text-muted-foreground">
					Need custom thresholds like 5xx errors only or latency alerts?
				</span>
				<a
					href={`/dashboard/org/${orgId}/alert-templates`}
					class="text-primary hover:underline font-medium flex items-center gap-1"
				>
					Create an Alert Template <ArrowRight class="size-3" />
				</a>
			</div>
		</div>
	{/if}
</div>
