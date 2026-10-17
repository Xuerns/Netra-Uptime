<script lang="ts">
	import type { AlertTemplate } from '$lib/api/alertTemplate';
	import type { AlertRule } from '$lib/api/alertRule';
	import { AlertTemplateAPI } from '$lib/api/alertTemplate';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { ShieldAlert, AlertTriangle, Plus, Pencil, Trash2, Layers } from 'lucide-svelte';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import AlertRuleModal from './AlertRuleModal.svelte';

	interface Props {
		template: AlertTemplate;
		orgId: string;
		onRefresh: () => Promise<void>;
	}

	let { template, orgId, onRefresh }: Props = $props();

	let isRuleModalOpen = $state(false);
	let ruleToEdit = $state<AlertRule | null>(null);

	let isDeleteRuleConfirmOpen = $state(false);
	let ruleToDelete = $state<AlertRule | null>(null);
	let isDeletingRule = $state(false);

	function openAddRule() {
		ruleToEdit = null;
		isRuleModalOpen = true;
	}

	function openEditRule(r: AlertRule) {
		ruleToEdit = r;
		isRuleModalOpen = true;
	}

	function confirmDeleteRule(r: AlertRule) {
		ruleToDelete = r;
		isDeleteRuleConfirmOpen = true;
	}

	async function handleDeleteRule() {
		if (!ruleToDelete || !orgId || !template.id) return;
		isDeletingRule = true;
		try {
			await AlertTemplateAPI.deleteRule(orgId, template.id, ruleToDelete.id);
			toast.success('Rule removed from template');
			isDeleteRuleConfirmOpen = false;
			ruleToDelete = null;
			await onRefresh();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to remove rule'));
		} finally {
			isDeletingRule = false;
		}
	}

	async function handleToggleRuleActive(r: AlertRule) {
		if (!orgId || !template.id) return;
		try {
			const newActive = !r.is_active;
			await AlertTemplateAPI.updateRule(orgId, template.id, r.id, {
				is_active: newActive
			});
			toast.success(newActive ? 'Rule activated' : 'Rule paused');
			await onRefresh();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to toggle rule status'));
		}
	}
</script>

<section class="bg-card border border-border rounded-xl p-5 md:p-6 space-y-4">
	<div class="flex items-center justify-between gap-3 flex-wrap">
		<div>
			<h2 class="text-base font-semibold text-foreground tracking-tight">Alert Rules</h2>
			<p class="text-xs text-muted-foreground mt-0.5">
				Evaluated on each monitor check. Breaches trigger incidents and notifications.
			</p>
		</div>
		<button
			onclick={openAddRule}
			class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-8.5 px-3 rounded-lg font-medium text-xs cursor-pointer shadow-xs"
		>
			<Plus class="size-3.5 stroke-[2.5]" />
			<span>Add Rule</span>
		</button>
	</div>

	{#if !template.rules || template.rules.length === 0}
		<div
			class="rounded-xl border border-dashed border-border bg-white/2 py-12 px-4 text-center space-y-3"
		>
			<div
				class="size-10 rounded-full bg-primary/10 border border-primary/20 text-primary flex items-center justify-center mx-auto"
			>
				<Layers class="size-5" />
			</div>
			<div>
				<h3 class="text-sm font-semibold text-foreground">No alert rules defined yet</h3>
				<p class="text-xs text-muted-foreground max-w-[44ch] mx-auto mt-1">
					Add threshold conditions (such as 5xx status codes, high latency, or timeouts) to activate
					automated alert triggers for linked monitors.
				</p>
			</div>
			<button
				onclick={openAddRule}
				class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-8 px-3 rounded-lg font-medium text-xs cursor-pointer"
			>
				<Plus class="size-3" />
				<span>Add First Rule</span>
			</button>
		</div>
	{:else}
		<div class="divide-y divide-border/60 border border-border/80 rounded-xl overflow-hidden">
			{#each template.rules as r (r.id)}
				<div
					class="flex items-center justify-between gap-4 p-4 bg-white/2 hover:bg-white/4 transition-colors flex-wrap sm:flex-nowrap"
				>
					<div class="flex items-center gap-3.5 min-w-0 flex-1">
						{#if r.severity === 'critical'}
							<div
								class="p-2 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-500 shrink-0"
								title="Critical Severity"
							>
								<ShieldAlert class="size-4" />
							</div>
						{:else}
							<div
								class="p-2 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-500 shrink-0"
								title="Warning Severity"
							>
								<AlertTriangle class="size-4" />
							</div>
						{/if}

						<div class="min-w-0 space-y-1">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="font-semibold text-sm text-foreground truncate">{r.name}</span>
								<span
									class="px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider font-mono border {r.severity ===
									'critical'
										? 'bg-rose-500/10 text-rose-400 border-rose-500/20'
										: 'bg-amber-500/10 text-amber-400 border-amber-500/20'}"
								>
									{r.severity}
								</span>
								{#if !r.is_active}
									<span
										class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-white/5 border border-border text-muted-foreground"
									>
										Paused
									</span>
								{/if}
							</div>

							<div class="text-xs text-muted-foreground font-mono flex items-center gap-2">
								<span class="text-foreground/90 font-medium">
									{r.metric}
									{r.operator}
									{r.target_value}
								</span>
								<span class="text-muted-foreground/60">
									• breaches: {r.consecutive}x consecutive
								</span>
							</div>
						</div>
					</div>

					<!-- Rule Actions -->
					<div class="flex items-center gap-1.5 shrink-0">
						<button
							onclick={() => handleToggleRuleActive(r)}
							class="px-2.5 py-1 text-xs rounded-md border border-border bg-secondary hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
							title={r.is_active ? 'Pause this rule' : 'Activate this rule'}
						>
							{r.is_active ? 'Pause' : 'Activate'}
						</button>
						<button
							onclick={() => openEditRule(r)}
							class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors cursor-pointer"
							title="Edit Rule"
						>
							<Pencil class="size-4" />
						</button>
						<button
							onclick={() => confirmDeleteRule(r)}
							class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
							title="Delete Rule"
						>
							<Trash2 class="size-4" />
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</section>

<!-- Add / Edit Rule Modal Component -->
<AlertRuleModal
	bind:open={isRuleModalOpen}
	{orgId}
	templateId={template.id}
	{ruleToEdit}
	onSuccess={onRefresh}
/>

<!-- Delete Rule Confirmation -->
<ConfirmDialog
	open={isDeleteRuleConfirmOpen}
	title="Remove Alert Rule"
	description={`Are you sure you want to remove rule "${ruleToDelete?.name || ''}" from this template? Linked monitors will immediately stop evaluating this rule.`}
	confirmText={isDeletingRule ? 'Removing...' : 'Remove Rule'}
	destructive={true}
	onConfirm={handleDeleteRule}
	onCancel={() => (isDeleteRuleConfirmOpen = false)}
/>
