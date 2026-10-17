<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import {
		Layers,
		Plus,
		Trash2,
		Sparkles,
		AlertTriangle,
		ShieldAlert,
		X,
		Pencil,
		Monitor
	} from 'lucide-svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
	import {
		AlertTemplateAPI,
		type AlertTemplate,
		type CreateAlertTemplateRequest
	} from '$lib/api/alertTemplate';
	import type { CreateAlertRuleRequest } from '$lib/api/alertRule';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { DataSkeleton } from '$lib/components/common';

	const orgId = $derived($page.params.id as string);

	let templates = $state<AlertTemplate[]>([]);
	let isLoading = $state(true);
	let errorMsg = $state('');

	// Create/Edit Template Modal
	let isCreateModalOpen = $state(false);
	let isEditing = $state(false);
	let selectedTemplateId = $state<string | null>(null);
	let templateName = $state('');
	let templateDesc = $state('');
	let isDefault = $state(false);
	let isSaving = $state(false);

	// Add Rule to Template Modal
	let isAddRuleModalOpen = $state(false);
	let ruleTemplateId = $state<string | null>(null);
	let ruleName = $state('');
	let metric = $state<'status_code' | 'response_time_ms' | 'timeout'>('status_code');
	let operator = $state<'==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in'>('==');
	let targetValue = $state('5xx');
	let consecutive = $state(2);
	let severity = $state<'critical' | 'warning'>('critical');
	let isSavingRule = $state(false);

	// Delete Confirmation
	let isDeleteConfirmOpen = $state(false);
	let templateToDelete = $state<AlertTemplate | null>(null);
	let isDeleting = $state(false);

	async function loadTemplates() {
		isLoading = true;
		errorMsg = '';
		try {
			const res = await AlertTemplateAPI.list(orgId);
			templates = res.data || [];
		} catch (err: unknown) {
			errorMsg = parseApiError(err, 'Failed to load alert templates');
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		if (orgId) {
			loadTemplates();
			if (typeof window !== 'undefined') {
				const params = new URLSearchParams(window.location.search);
				if (params.get('create') === 'true') {
					openCreateModal();
				}
			}
		}
	});

	// Synchronize top navigation header
	$effect(() => {
		headerStore.set({
			title: 'Alert Templates',
			showSearch: false,
			actionType: 'button',
			actionLabel: 'Create Template',
			onAction: () => {
				openCreateModal();
			},
			badge: {
				text: `${templates.length} Template${templates.length !== 1 ? 's' : ''}`,
				variant: 'slate'
			}
		});
	});

	function openCreateModal() {
		isEditing = false;
		selectedTemplateId = null;
		templateName = '';
		templateDesc = '';
		isDefault = false;
		isCreateModalOpen = true;
	}

	function openEditModal(tmpl: AlertTemplate) {
		isEditing = true;
		selectedTemplateId = tmpl.id;
		templateName = tmpl.name;
		templateDesc = tmpl.description;
		isDefault = tmpl.is_default;
		isCreateModalOpen = true;
	}

	function applyPreset(preset: 'web' | 'api' | 'lenient') {
		if (preset === 'web') {
			templateName = 'Standard Web Service';
			templateDesc = 'Alerts on 5xx errors, high response latency, and connection timeouts';
		} else if (preset === 'api') {
			templateName = 'Strict API Gateway';
			templateDesc = 'Strict SLA rules alerting quickly on unexpected HTTP statuses';
		} else if (preset === 'lenient') {
			templateName = 'Lenient Background Tasks';
			templateDesc = 'Allows longer latency and intermittent breaches for heavy endpoints';
		}
	}

	async function handleSaveTemplate(e: Event) {
		e.preventDefault();
		if (!templateName.trim() || !orgId) return;

		isSaving = true;
		try {
			if (isEditing && selectedTemplateId) {
				await AlertTemplateAPI.update(orgId, selectedTemplateId, {
					name: templateName.trim(),
					description: templateDesc.trim(),
					is_default: isDefault
				});
				toast.success('Alert template updated');
			} else {
				let initialRules: CreateAlertRuleRequest[] = [];
				if (templateName === 'Standard Web Service') {
					initialRules = [
						{
							name: '5xx Server Error Alert',
							metric: 'status_code',
							operator: '==',
							target_value: '5xx',
							consecutive: 2,
							severity: 'critical',
							is_active: true
						},
						{
							name: 'High Latency Warning',
							metric: 'response_time_ms',
							operator: '>',
							target_value: '2000',
							consecutive: 3,
							severity: 'warning',
							is_active: true
						},
						{
							name: 'Connection Timeout Warning',
							metric: 'timeout',
							operator: '>',
							target_value: '5000',
							consecutive: 3,
							severity: 'warning',
							is_active: true
						}
					];
				} else if (templateName === 'Strict API Gateway') {
					initialRules = [
						{
							name: 'Status Code Not 200',
							metric: 'status_code',
							operator: '!=',
							target_value: '200',
							consecutive: 1,
							severity: 'critical',
							is_active: true
						},
						{
							name: 'Strict Latency Limit (>800ms)',
							metric: 'response_time_ms',
							operator: '>',
							target_value: '800',
							consecutive: 2,
							severity: 'warning',
							is_active: true
						}
					];
				}

				const payload: CreateAlertTemplateRequest = {
					name: templateName.trim(),
					description: templateDesc.trim(),
					is_default: isDefault,
					rules: initialRules
				};
				await AlertTemplateAPI.create(orgId, payload);
				toast.success('Alert template created');
			}

			isCreateModalOpen = false;
			await loadTemplates();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to save template'));
		} finally {
			isSaving = false;
		}
	}

	function openAddRuleModal(tmplId: string) {
		ruleTemplateId = tmplId;
		ruleName = '5xx Server Error Alert';
		metric = 'status_code';
		operator = '==';
		targetValue = '5xx';
		consecutive = 2;
		severity = 'critical';
		isAddRuleModalOpen = true;
	}

	async function handleCreateRule(e: Event) {
		e.preventDefault();
		if (!ruleTemplateId || !ruleName.trim() || !targetValue.trim()) return;

		isSavingRule = true;
		try {
			await AlertTemplateAPI.createRule(orgId, ruleTemplateId, {
				name: ruleName.trim(),
				metric,
				operator,
				target_value: targetValue.trim(),
				consecutive: Number(consecutive) || 2,
				severity,
				is_active: true
			});
			toast.success('Rule added to template');
			isAddRuleModalOpen = false;
			await loadTemplates();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to add rule'));
		} finally {
			isSavingRule = false;
		}
	}

	async function handleDeleteRule(tmplId: string, ruleId: string) {
		if (!confirm('Are you sure you want to remove this rule from the template?')) return;
		try {
			await AlertTemplateAPI.deleteRule(orgId, tmplId, ruleId);
			toast.success('Rule removed from template');
			await loadTemplates();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to remove rule'));
		}
	}

	function confirmDelete(tmpl: AlertTemplate) {
		templateToDelete = tmpl;
		isDeleteConfirmOpen = true;
	}

	async function handleDeleteTemplate() {
		if (!templateToDelete) return;
		isDeleting = true;
		try {
			await AlertTemplateAPI.delete(orgId, templateToDelete.id);
			toast.success('Alert template deleted');
			isDeleteConfirmOpen = false;
			templateToDelete = null;
			await loadTemplates();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to delete template'));
		} finally {
			isDeleting = false;
		}
	}
</script>

<svelte:head>
	<title>Alert Templates | Netra Uptime</title>
</svelte:head>

<div class="text-foreground antialiased space-y-6">
	<!-- ── Page Header ──────────────────────────────────────────────────────────── -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[32px] font-semibold leading-tight tracking-tight text-foreground mb-1.5">
				Alert Templates
			</h1>
			<p class="text-base leading-relaxed text-muted-foreground max-w-[60ch]">
				Define reusable alert rule profiles and link them to monitors. Changes to a template
				automatically apply to all linked monitors.
			</p>
		</div>
	</header>

	<!-- ── Templates Grid ─────────────────────────────────────────────────────── -->
	{#if isLoading}
		<DataSkeleton count={3} />
	{:else if errorMsg}
		<div
			class="rounded-xl border border-destructive/20 bg-destructive/10 p-6 text-center text-destructive"
		>
			<AlertTriangle class="h-8 w-8 mx-auto mb-2 opacity-80" />
			<p class="font-medium">{errorMsg}</p>
			<button
				onclick={loadTemplates}
				class="mt-4 inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground h-9 px-4 rounded-lg font-medium text-xs cursor-pointer"
			>
				Try Again
			</button>
		</div>
	{:else if templates.length === 0}
		<!-- Standard Empty State matching monitors/notifications -->
		<div
			class="flex flex-col items-center text-center py-20 px-4 border border-border rounded-xl bg-card"
		>
			<div
				class="flex items-center justify-center size-14 rounded-xl bg-primary/10 border border-primary/20 mb-4 text-primary"
			>
				<Layers class="size-6" />
			</div>
			<h3 class="text-xl font-semibold text-foreground mb-1.5">No alert templates yet</h3>
			<p class="text-sm text-muted-foreground max-w-[44ch] mb-6 leading-relaxed">
				Create reusable templates with custom rules (status codes, response time thresholds, timeout
				conditions) to easily manage alerts across your monitors.
			</p>
			<button
				onclick={openCreateModal}
				class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 rounded-lg font-medium text-xs cursor-pointer focus:outline-none"
			>
				<Plus class="size-3.5 stroke-[2.5]" />
				<span>Create Template</span>
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
			{#each templates as tmpl (tmpl.id)}
				<div
					class="bg-white/3 border border-border rounded-xl p-5 flex flex-col justify-between gap-4 hover:border-primary/30 transition-colors"
				>
					<div>
						<!-- Card Header -->
						<div class="flex items-start justify-between gap-3 mb-3">
							<div class="min-w-0">
								<div class="flex items-center gap-2 flex-wrap">
									<h3 class="text-base font-semibold text-foreground truncate">
										{tmpl.name}
									</h3>
									{#if tmpl.is_default}
										<span
											class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-primary/10 border border-primary/20 text-primary"
										>
											<Sparkles class="size-3" />
											Default
										</span>
									{/if}
								</div>
								{#if tmpl.description}
									<p class="text-xs text-muted-foreground mt-1 line-clamp-2">
										{tmpl.description}
									</p>
								{/if}
							</div>

							<!-- Actions -->
							<div class="flex items-center gap-1 shrink-0">
								<button
									onclick={() => openEditModal(tmpl)}
									class="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors cursor-pointer"
									title="Edit template info"
								>
									<Pencil class="size-4" />
								</button>
								<button
									onclick={() => confirmDelete(tmpl)}
									class="p-1.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
									title="Delete template"
								>
									<Trash2 class="size-4" />
								</button>
							</div>
						</div>

						<!-- Monitors Linked Badge -->
						<div class="flex items-center gap-1.5 mb-4 text-xs text-muted-foreground">
							<Monitor class="size-3.5 text-muted-foreground/70" />
							<span>
								Linked to <b class="text-foreground">{tmpl.monitors_count}</b>
								{tmpl.monitors_count === 1 ? 'monitor' : 'monitors'}
							</span>
						</div>

						<!-- Rules Section -->
						<div class="space-y-2 pt-3 border-t border-border">
							<div
								class="flex items-center justify-between text-xs font-medium text-muted-foreground pb-1"
							>
								<span>Configured Rules ({tmpl.rules?.length || 0})</span>
								<button
									onclick={() => openAddRuleModal(tmpl.id)}
									class="inline-flex items-center gap-1 text-primary hover:underline font-semibold cursor-pointer"
								>
									<Plus class="size-3.5" />
									Add Rule
								</button>
							</div>

							{#if !tmpl.rules || tmpl.rules.length === 0}
								<div
									class="rounded-lg border border-dashed border-border bg-white/2 p-3 text-center text-xs text-muted-foreground"
								>
									No rules defined in this template yet.
								</div>
							{:else}
								<div class="space-y-2">
									{#each tmpl.rules as r (r.id)}
										<div
											class="flex items-center justify-between gap-3 p-2.5 rounded-lg border border-border bg-background text-xs"
										>
											<div class="flex items-center gap-2.5 min-w-0">
												{#if r.severity === 'critical'}
													<ShieldAlert class="size-4 text-rose-500 shrink-0" />
												{:else}
													<AlertTriangle class="size-4 text-amber-500 shrink-0" />
												{/if}
												<div class="truncate">
													<div class="font-medium text-foreground truncate">{r.name}</div>
													<div class="text-[11px] text-muted-foreground font-mono">
														{r.metric}
														{r.operator}
														{r.target_value}
														<span class="text-muted-foreground/60 ml-1"
															>({r.consecutive}x breaches)</span
														>
													</div>
												</div>
											</div>

											<div class="flex items-center gap-2 shrink-0">
												<span
													class={`px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider ${
														r.severity === 'critical'
															? 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
															: 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
													}`}
												>
													{r.severity}
												</span>
												<button
													onclick={() => handleDeleteRule(tmpl.id, r.id)}
													class="p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
													title="Delete rule"
												>
													<X class="size-3.5" />
												</button>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- ── Create / Edit Template Dialog ───────────────────────────────────────── -->
<Dialog.Root bind:open={isCreateModalOpen}>
	<Dialog.Content
		class="sm:max-w-130 bg-card border border-border text-foreground p-6 rounded-xl shadow-2xl"
	>
		<Dialog.Header>
			<Dialog.Title
				class="text-xl font-semibold tracking-tight text-foreground flex items-center gap-2"
			>
				<Layers class="size-5 text-primary" />
				{isEditing ? 'Edit Alert Template' : 'Create Alert Template'}
			</Dialog.Title>
			<Dialog.Description class="text-muted-foreground text-sm mt-1">
				{isEditing
					? 'Update template information. All linked monitors will continue using this template.'
					: 'Create a reusable rule profile to avoid repeating rule definitions on every monitor.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-4"></div>

		{#if !isEditing}
			<!-- Quick Presets -->
			<div class="space-y-2 mb-4">
				<Label class="text-xs text-muted-foreground flex items-center gap-1 font-medium">
					<Sparkles class="size-3.5 text-primary" />
					Or start with a preset:
				</Label>
				<div class="grid grid-cols-3 gap-2">
					<button
						type="button"
						onclick={() => applyPreset('web')}
						class="p-2.5 rounded-lg border text-left transition-all border-border bg-background hover:bg-white/5 hover:border-primary/40 cursor-pointer"
					>
						<div class="text-xs font-semibold text-foreground">Standard Web</div>
						<div class="text-[11px] text-muted-foreground mt-0.5">5xx & Latency</div>
					</button>
					<button
						type="button"
						onclick={() => applyPreset('api')}
						class="p-2.5 rounded-lg border text-left transition-all border-border bg-background hover:bg-white/5 hover:border-primary/40 cursor-pointer"
					>
						<div class="text-xs font-semibold text-foreground">Strict API</div>
						<div class="text-[11px] text-muted-foreground mt-0.5">Fast Alerting</div>
					</button>
					<button
						type="button"
						onclick={() => applyPreset('lenient')}
						class="p-2.5 rounded-lg border text-left transition-all border-border bg-background hover:bg-white/5 hover:border-primary/40 cursor-pointer"
					>
						<div class="text-xs font-semibold text-foreground">Lenient</div>
						<div class="text-[11px] text-muted-foreground mt-0.5">Heavy Jobs</div>
					</button>
				</div>
			</div>
		{/if}

		<form onsubmit={handleSaveTemplate} class="flex flex-col gap-4">
			<div class="space-y-1.5">
				<Label for="template-name" class="text-sm font-medium text-foreground">
					Template Name <span class="text-destructive font-bold">*</span>
				</Label>
				<Input
					id="template-name"
					bind:value={templateName}
					placeholder="e.g. Standard API Profile"
					required
					class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
				/>
			</div>

			<div class="space-y-1.5">
				<Label for="template-desc" class="text-sm font-medium text-foreground">Description</Label>
				<Input
					id="template-desc"
					bind:value={templateDesc}
					placeholder="Brief note about when to use this profile"
					class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
				/>
			</div>

			<div class="flex items-center gap-2 pt-1">
				<input
					type="checkbox"
					id="template-default"
					bind:checked={isDefault}
					class="rounded border-border text-primary focus:ring-primary/20 h-4 w-4 cursor-pointer"
				/>
				<Label
					for="template-default"
					class="text-xs cursor-pointer text-muted-foreground hover:text-foreground"
				>
					Set as default template for new monitors in this organization
				</Label>
			</div>

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-4 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (isCreateModalOpen = false)}
					disabled={isSaving}
					class="px-4 py-2 rounded-lg text-xs font-medium text-foreground hover:bg-accent transition-colors cursor-pointer"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSaving || !templateName.trim()}
					class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 rounded-lg font-medium text-xs cursor-pointer disabled:opacity-50"
				>
					{isSaving ? 'Saving…' : isEditing ? 'Update Template' : 'Create Template'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- ── Add Rule to Template Dialog ─────────────────────────────────────────── -->
<Dialog.Root bind:open={isAddRuleModalOpen}>
	<Dialog.Content
		class="sm:max-w-130 bg-card border border-border text-foreground p-6 rounded-xl shadow-2xl"
	>
		<Dialog.Header>
			<Dialog.Title
				class="text-xl font-semibold tracking-tight text-foreground flex items-center gap-2"
			>
				<Plus class="size-5 text-primary" />
				Add Rule to Template
			</Dialog.Title>
			<Dialog.Description class="text-muted-foreground text-sm mt-1">
				Configure metric conditions that will automatically evaluate for any monitor linked to this
				template.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-4"></div>

		<form onsubmit={handleCreateRule} class="flex flex-col gap-4">
			<div class="space-y-1.5">
				<Label class="text-sm font-medium text-foreground">
					Rule Name <span class="text-destructive font-bold">*</span>
				</Label>
				<Input
					bind:value={ruleName}
					placeholder="e.g. 5xx Server Error Alert"
					required
					class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
				/>
			</div>

			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1.5">
					<Label class="text-sm font-medium text-foreground">Metric</Label>
					<select
						bind:value={metric}
						class="w-full h-10 px-3 rounded-lg border border-border bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer"
					>
						<option value="status_code">Status Code</option>
						<option value="response_time_ms">Response Time (ms)</option>
						<option value="timeout">Timeout / Connection</option>
					</select>
				</div>

				<div class="space-y-1.5">
					<Label class="text-sm font-medium text-foreground">Operator</Label>
					<select
						bind:value={operator}
						class="w-full h-10 px-3 rounded-lg border border-border bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer"
					>
						<option value="==">Equals (==)</option>
						<option value="!=">Not Equals (!=)</option>
						<option value=">">Greater Than (&gt;)</option>
						<option value=">=">Greater or Equal (&gt;=)</option>
						<option value="<">Less Than (&lt;)</option>
						<option value="<=">Less or Equal (&lt;=)</option>
					</select>
				</div>
			</div>

			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1.5">
					<Label class="text-sm font-medium text-foreground">Target Value</Label>
					<Input
						bind:value={targetValue}
						placeholder="e.g. 5xx, 2000, 10000"
						required
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
					/>
				</div>

				<div class="space-y-1.5">
					<Label class="text-sm font-medium text-foreground">Consecutive Breaches</Label>
					<Input
						type="number"
						bind:value={consecutive}
						min="1"
						max="20"
						required
						class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 text-sm"
					/>
				</div>
			</div>

			<div class="space-y-1.5">
				<Label class="text-sm font-medium text-foreground">Severity</Label>
				<div class="grid grid-cols-2 gap-2">
					<label
						class={`flex items-center gap-2 p-2.5 rounded-lg border cursor-pointer transition-all ${
							severity === 'critical'
								? 'border-rose-500/40 bg-rose-500/10 text-rose-400'
								: 'border-border bg-background hover:bg-white/5 text-muted-foreground'
						}`}
					>
						<input
							type="radio"
							name="rule-severity"
							value="critical"
							bind:group={severity}
							class="hidden"
						/>
						<ShieldAlert class="size-4" />
						<div class="text-xs font-semibold">Critical (Down)</div>
					</label>

					<label
						class={`flex items-center gap-2 p-2.5 rounded-lg border cursor-pointer transition-all ${
							severity === 'warning'
								? 'border-amber-500/40 bg-amber-500/10 text-amber-400'
								: 'border-border bg-background hover:bg-white/5 text-muted-foreground'
						}`}
					>
						<input
							type="radio"
							name="rule-severity"
							value="warning"
							bind:group={severity}
							class="hidden"
						/>
						<AlertTriangle class="size-4" />
						<div class="text-xs font-semibold">Warning (Amber)</div>
					</label>
				</div>
			</div>

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-4 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (isAddRuleModalOpen = false)}
					disabled={isSavingRule}
					class="px-4 py-2 rounded-lg text-xs font-medium text-foreground hover:bg-accent transition-colors cursor-pointer"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSavingRule || !ruleName.trim()}
					class="inline-flex items-center gap-1.5 bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 rounded-lg font-medium text-xs cursor-pointer disabled:opacity-50"
				>
					{isSavingRule ? 'Adding…' : 'Add Rule'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- ── Delete Confirmation Dialog ─────────────────────────────────────────── -->
<ConfirmDialog
	bind:open={isDeleteConfirmOpen}
	title="Delete Alert Template"
	description={`Are you sure you want to delete template "${templateToDelete?.name}"? Linked monitors will remain active and fall back to system default checks.`}
	confirmText={isDeleting ? 'Deleting...' : 'Delete Template'}
	onConfirm={handleDeleteTemplate}
	destructive
/>
