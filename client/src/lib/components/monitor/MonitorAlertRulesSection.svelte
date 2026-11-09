<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { AlertTemplateAPI, type AlertTemplate } from '$lib/api/alertTemplate';
	import { MonitorAPI } from '$lib/api/monitor';
	import * as Select from '$lib/components/ui/select/index.js';
	import { ExternalLink } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { logger } from '$lib/utils/logger';

	interface Props {
		orgId: string;
		monitorId: string;
		monitorType?: string;
		alertTemplateId?: string | null;
		canManage?: boolean;
	}

	let {
		orgId,
		monitorId,
		monitorType = 'http',
		alertTemplateId = null,
		canManage = true
	}: Props = $props();

	let isHttp = $derived(monitorType?.toLowerCase() === 'http');

	let availableTemplates = $state<AlertTemplate[]>([]);
	let currentTemplateId = $state<string | null>(untrack(() => alertTemplateId));
	let isApplyingTemplate = $state(false);
	let isLoading = $state(true);

	let activeTemplate = $derived(availableTemplates.find((t) => t.id === currentTemplateId) || null);

	let selectedTemplateLabel = $derived.by(() => {
		if (!currentTemplateId || currentTemplateId === 'none') {
			return 'System Default Checks';
		}
		const found = availableTemplates.find((t) => t.id === currentTemplateId);
		if (!found) return 'System Default Checks';
		return `${found.name} (${found.rules?.length || 0})`;
	});

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
					// Save to monitor in background if user can manage
					if (canManage) {
						MonitorAPI.applyTemplate(orgId, monitorId, defaultTmpl.id).catch(() => {});
					}
				}
			}
		} catch (err: unknown) {
			logger.error('Failed to load alert templates', err);
			toast.error(parseApiError(err, 'Failed to load alert templates'));
		} finally {
			isLoading = false;
		}
	}

	async function handleTemplateSelect(val: string) {
		const newTmplId = val && val !== 'none' ? val : null;
		if (newTmplId === currentTemplateId) return;

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

<div class="flex flex-col gap-4 min-w-0">
	<!-- Section Header -->
	<div class="space-y-0.5">
		<h3 class="text-sm font-semibold text-foreground tracking-tight">Alert Policy</h3>
		<p class="text-xs text-muted-foreground">Rules inherited from the selected template.</p>
	</div>

	<!-- Template Selector -->
	{#if canManage}
		<div class="flex items-center gap-2 w-full">
			<Select.Root
				type="single"
				value={currentTemplateId || 'none'}
				disabled={isApplyingTemplate || isLoading}
				onValueChange={(val) => {
					if (val) handleTemplateSelect(val);
				}}
			>
				<Select.Trigger
					class="h-8.5 px-3 rounded-lg border border-border bg-background text-xs font-medium text-foreground focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer flex-1 min-w-0 truncate"
				>
					<span class="truncate">{selectedTemplateLabel}</span>
				</Select.Trigger>
				<Select.Content class="bg-card border-border text-foreground max-h-60">
					<Select.Item
						value="none"
						label="System Default Checks"
						class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
					>
						System Default Checks
					</Select.Item>
					{#each availableTemplates as tmpl (tmpl.id)}
						{@const tmplLabel = `${tmpl.name} (${tmpl.rules?.length || 0})`}
						<Select.Item
							value={tmpl.id}
							label={tmplLabel}
							class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
						>
							{tmplLabel}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>

			<a
				href={activeTemplate
					? `/dashboard/org/${orgId}/alert-templates/${activeTemplate.id}`
					: `/dashboard/org/${orgId}/alert-templates`}
				class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-8.5 px-3 rounded-lg font-medium text-xs cursor-pointer shrink-0"
			>
				<span>{activeTemplate ? 'Edit' : 'Manage'}</span>
				<ExternalLink class="size-3" />
			</a>
		</div>
	{:else}
		<div
			class="flex items-center justify-between gap-2 p-2 rounded-lg border border-border bg-background text-xs min-h-8.5"
		>
			<span class="font-medium text-foreground truncate">{selectedTemplateLabel}</span>
			{#if activeTemplate}
				<a
					href="/dashboard/org/{orgId}/alert-templates/{activeTemplate.id}"
					class="inline-flex items-center gap-1 text-primary hover:underline shrink-0 text-xs font-medium"
				>
					<span>View</span>
					<ExternalLink class="size-3" />
				</a>
			{/if}
		</div>
	{/if}

	<!-- Rules Content Area -->
	{#if isLoading}
		<div class="h-24 flex items-center justify-center text-xs text-muted-foreground font-mono">
			Loading alert rules...
		</div>
	{:else if activeTemplate}
		{#if !activeTemplate.rules || activeTemplate.rules.length === 0}
			<div class="py-6 text-center text-xs text-muted-foreground space-y-1">
				<p class="font-medium text-foreground">No rules defined in this template yet.</p>
				<p class="text-[11px]">
					System default checks are used until rules are added to this template.
				</p>
			</div>
		{:else}
			<ul class="divide-y divide-border/60">
				{#each activeTemplate.rules as r (r.id)}
					<li class="flex items-center justify-between gap-3 py-3 text-xs">
						<div class="min-w-0">
							<div class="font-medium text-foreground text-xs flex items-center gap-2 min-w-0">
								<span class="truncate">{r.name}</span>
								{#if !r.is_active}
									<span
										class="text-[10px] px-1.5 py-0.2 rounded bg-muted text-muted-foreground font-mono shrink-0"
									>
										Paused
									</span>
								{/if}
							</div>
							<div class="text-[11px] text-muted-foreground font-mono mt-0.5 truncate">
								{r.metric}
								{r.operator}
								{r.target_value}
								<span class="text-muted-foreground/50">&bull;</span>
								{r.consecutive}x consecutive
							</div>
						</div>

						<span
							class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider font-mono shrink-0 border {r.severity ===
							'critical'
								? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
								: 'bg-amber-500/10 text-amber-400 border-amber-500/20'}"
						>
							{r.severity}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	{:else}
		<!-- Fallback: System Default Checks -->
		<ul class="divide-y divide-border/60">
			<li class="py-3 space-y-1">
				<div class="flex items-center justify-between gap-2">
					<span class="text-xs font-medium text-foreground">Timeout Warning</span>
					<span
						class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider font-mono bg-amber-500/10 text-amber-400 border border-amber-500/20"
					>
						warning
					</span>
				</div>
				<p class="text-[11px] text-muted-foreground leading-relaxed">
					Triggered when requests time out for <b>3 consecutive checks</b>.
				</p>
			</li>

			<li class="py-3 space-y-1">
				<div class="flex items-center justify-between gap-2">
					<span class="text-xs font-medium text-foreground">Down Incident</span>
					<span
						class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider font-mono bg-rose-500/10 text-rose-400 border border-rose-500/20"
					>
						critical
					</span>
				</div>
				<p class="text-[11px] text-muted-foreground leading-relaxed">
					{#if isHttp}
						Triggered when connection fails or status code mismatches for <b>3 consecutive checks</b
						>.
					{:else}
						Triggered when port is closed or target is unreachable for <b>3 consecutive checks</b>.
					{/if}
				</p>
			</li>
		</ul>
	{/if}
</div>
