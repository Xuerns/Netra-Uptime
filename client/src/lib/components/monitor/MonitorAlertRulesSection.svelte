<script lang="ts">
	import { onMount } from 'svelte';
	import { AlertRuleAPI, type AlertRule, type CreateAlertRuleRequest } from '$lib/api/alertRule';
	import Button from '$lib/components/ui/button/button.svelte';
	import {
		Bell,
		Plus,
		Trash2,
		CheckCircle,
		AlertTriangle,
		ShieldAlert,
		Sparkles,
		X,
		Power
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	interface Props {
		orgId: string;
		monitorId: string;
		monitorType?: string;
	}

	let { orgId, monitorId, monitorType = 'http' }: Props = $props();

	let rules = $state<AlertRule[]>([]);
	let isLoading = $state(true);
	let isSaving = $state(false);
	let errorMsg = $state('');
	let showAddModal = $state(false);

	// Form state
	let ruleName = $state('');
	let metric = $state<'status_code' | 'response_time_ms' | 'timeout'>('status_code');
	let operator = $state<'==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in'>('==');
	let targetValue = $state('5xx');
	let consecutive = $state(2);
	let severity = $state<'critical' | 'warning'>('critical');

	async function loadRules() {
		isLoading = true;
		errorMsg = '';
		try {
			const res = await AlertRuleAPI.getByMonitor(orgId, monitorId);
			rules = res.data || [];
		} catch (err: any) {
			errorMsg = parseApiError(err, 'Failed to load alert rules');
			rules = [];
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		if (orgId && monitorId) {
			loadRules();
		}
	});

	function applyPreset(preset: '5xx' | 'latency' | 'not200') {
		if (preset === '5xx') {
			ruleName = '5xx Server Error Alert';
			metric = 'status_code';
			operator = '==';
			targetValue = '5xx';
			consecutive = 2;
			severity = 'critical';
		} else if (preset === 'latency') {
			ruleName = 'Degraded Response Time';
			metric = 'response_time_ms';
			operator = '>';
			targetValue = '2000';
			consecutive = 3;
			severity = 'warning';
		} else if (preset === 'not200') {
			ruleName = 'Unexpected Status Code';
			metric = 'status_code';
			operator = '!=';
			targetValue = '200';
			consecutive = 2;
			severity = 'critical';
		}
	}

	async function handleCreateRule(e: Event) {
		e.preventDefault();
		if (!ruleName.trim() || !targetValue.trim()) return;

		isSaving = true;
		errorMsg = '';
		try {
			const payload: CreateAlertRuleRequest = {
				name: ruleName.trim(),
				metric,
				operator,
				target_value: targetValue.trim(),
				consecutive: Number(consecutive) || 2,
				severity,
				is_active: true
			};
			await AlertRuleAPI.create(orgId, monitorId, payload);
			showAddModal = false;
			ruleName = '';
			await loadRules();
		} catch (err: any) {
			errorMsg = parseApiError(err, 'Failed to create alert rule');
		} finally {
			isSaving = false;
		}
	}

	async function handleToggleRule(rule: AlertRule) {
		try {
			const updated = !rule.is_active;
			await AlertRuleAPI.update(orgId, rule.id, { is_active: updated });
			rule.is_active = updated;
			toast.success(updated ? 'Alert rule enabled' : 'Alert rule disabled');
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to update rule'));
		}
	}

	async function handleDeleteRule(ruleId: string) {
		if (!confirm('Are you sure you want to delete this alert rule?')) return;
		try {
			await AlertRuleAPI.delete(orgId, ruleId);
			rules = rules.filter((r) => r.id !== ruleId);
			toast.success('Alert rule deleted successfully');
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to delete alert rule'));
		}
	}
</script>

<div class="bg-card border border-border rounded-2xl p-5 md:p-6 shadow-xl space-y-5">
	<div class="flex items-center justify-between flex-wrap gap-3">
		<div class="space-y-1">
			<h3 class="text-sm font-semibold text-foreground tracking-tight flex items-center gap-2">
				<span>Alert Rules & Policies</span>
			</h3>
			<p class="text-xs text-muted-foreground">
				Configure conditions that trigger critical down incidents or degraded performance warnings.
			</p>
		</div>

		<Button
			class="bg-primary text-primary-foreground hover:bg-primary/90 h-8 px-3 text-xs gap-1.5 font-medium rounded-lg"
			onclick={() => {
				applyPreset('5xx');
				showAddModal = true;
			}}
		>
			<Plus class="size-3.5" />
			<span>Add Alert Rule</span>
		</Button>
	</div>

	{#if isLoading}
		<div class="h-24 flex items-center justify-center text-xs text-muted-foreground font-mono">
			Loading alert rules...
		</div>
	{:else if rules.length === 0}
		<!-- Default fallback notice -->
		<div class="bg-white/3 border border-border/80 rounded-xl p-5 text-center space-y-3">
			<ShieldAlert class="size-8 text-muted-foreground/60 mx-auto" />
			<div class="space-y-1">
				<p class="text-sm font-medium text-foreground">Using Default Fallback Rule</p>
				<p class="text-xs text-muted-foreground max-w-[560px] mx-auto leading-relaxed">
					No custom rules configured. Currently alerting if connection times out or status code
					doesn't match expected after 3 consecutive failures.
				</p>
			</div>
		</div>
	{:else}
		<div class="space-y-2.5">
			{#each rules as rule}
				<div
					class="flex items-center justify-between bg-white/3 border border-border/80 hover:border-border rounded-xl p-4 transition-all gap-3 flex-wrap sm:flex-nowrap {rule.is_active
						? ''
						: 'opacity-60'}"
				>
					<div class="space-y-1.5">
						<div class="flex items-center gap-2 flex-wrap">
							<span
								class="size-2 rounded-full {rule.is_active
									? rule.severity === 'critical'
										? 'bg-rose-500'
										: 'bg-amber-400'
									: 'bg-muted-foreground'}"
							></span>
							<span class="text-sm font-semibold text-foreground">{rule.name}</span>
							<span
								class="text-[10px] font-mono px-2 py-0.5 rounded-full font-bold uppercase {rule.severity ===
								'critical'
									? 'text-rose-400 bg-rose-500/10 border border-rose-500/20'
									: 'text-amber-400 bg-amber-500/10 border border-amber-500/20'}"
							>
								{rule.severity}
							</span>
							{#if !rule.is_active}
								<span
									class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-white/5 text-muted-foreground border border-border"
								>
									Disabled
								</span>
							{/if}
						</div>

						<div class="flex items-center gap-2 text-xs text-muted-foreground font-mono">
							<span>When:</span>
							<span
								class="px-1.5 py-0.5 rounded bg-white/5 border border-white/5 text-foreground font-medium"
							>
								{rule.metric}
								{rule.operator}
								{rule.target_value}
							</span>
							<span>•</span>
							<span>{rule.consecutive} check{rule.consecutive > 1 ? 's' : ''} in a row</span>
						</div>
					</div>

					<div class="flex items-center gap-2 shrink-0 ml-auto sm:ml-0">
						<button
							type="button"
							class="size-8 rounded-lg flex items-center justify-center border transition-colors {rule.is_active
								? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400 hover:bg-emerald-500/20'
								: 'bg-white/5 border-border text-muted-foreground hover:text-foreground'}"
							title={rule.is_active ? 'Disable rule' : 'Enable rule'}
							onclick={() => handleToggleRule(rule)}
						>
							<Power class="size-3.5" />
						</button>

						<button
							type="button"
							class="size-8 rounded-lg flex items-center justify-center bg-white/5 border border-border text-muted-foreground hover:text-rose-400 hover:border-rose-500/30 hover:bg-rose-500/10 transition-colors"
							title="Delete rule"
							onclick={() => handleDeleteRule(rule.id)}
						>
							<Trash2 class="size-3.5" />
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Modal Add Alert Rule -->
{#if showAddModal}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 overflow-y-auto animate-in fade-in duration-200"
	>
		<div
			class="bg-card border border-border rounded-2xl w-full sm:max-w-[560px] shadow-2xl p-6 space-y-5 text-foreground my-auto animate-in zoom-in-95 duration-200"
		>
			<div class="flex items-center justify-between border-b border-border pb-3">
				<h3 class="text-base font-bold flex items-center gap-2">
					<Sparkles class="size-4 text-primary" />
					<span>Add Alert Rule</span>
				</h3>
				<button
					type="button"
					class="p-1 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors"
					onclick={() => (showAddModal = false)}
				>
					<X class="size-4" />
				</button>
			</div>

			<!-- Quick Presets -->
			<div class="space-y-1.5">
				<div class="text-xs text-muted-foreground font-medium">Quick Presets</div>
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
					<button
						type="button"
						class="text-xs p-2 rounded-lg border border-border bg-white/3 hover:bg-white/8 text-left transition-colors"
						onclick={() => applyPreset('5xx')}
					>
						<span class="block font-bold text-rose-400">5xx Error</span>
						<span class="text-[10px] text-muted-foreground">HTTP 500/502/503</span>
					</button>
					<button
						type="button"
						class="text-xs p-2 rounded-lg border border-border bg-white/3 hover:bg-white/8 text-left transition-colors"
						onclick={() => applyPreset('latency')}
					>
						<span class="block font-bold text-amber-400">High Latency</span>
						<span class="text-[10px] text-muted-foreground">> 2000ms warning</span>
					</button>
					<button
						type="button"
						class="text-xs p-2 rounded-lg border border-border bg-white/3 hover:bg-white/8 text-left transition-colors"
						onclick={() => applyPreset('not200')}
					>
						<span class="block font-bold text-blue-400">Not 200 OK</span>
						<span class="text-[10px] text-muted-foreground">Status != 200</span>
					</button>
				</div>
			</div>

			<form onsubmit={handleCreateRule} class="space-y-4">
				{#if errorMsg}
					<div
						class="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 rounded-lg p-3"
					>
						{errorMsg}
					</div>
				{/if}

				<!-- Rule Name -->
				<div class="space-y-1">
					<label for="rule-name-input" class="text-xs font-medium text-foreground">Rule Name</label>
					<input
						id="rule-name-input"
						type="text"
						required
						bind:value={ruleName}
						placeholder="e.g. 5xx Server Error Alert"
						class="w-full h-9 rounded-lg bg-secondary border border-border px-3 text-xs text-foreground focus:outline-none focus:border-primary"
					/>
				</div>

				<!-- Metric & Operator & Value -->
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
					<div class="space-y-1">
						<label for="metric-select" class="text-xs font-medium text-foreground">Metric</label>
						<select
							id="metric-select"
							bind:value={metric}
							class="w-full h-9 rounded-lg bg-secondary border border-border px-2 text-xs text-foreground focus:outline-none focus:border-primary"
						>
							{#if monitorType === 'http'}
								<option value="status_code">Status Code</option>
							{/if}
							<option value="response_time_ms">Latency (ms)</option>
							<option value="timeout">Connection Timeout</option>
						</select>
					</div>

					<div class="space-y-1">
						<label for="operator-select" class="text-xs font-medium text-foreground">Operator</label
						>
						<select
							id="operator-select"
							bind:value={operator}
							class="w-full h-9 rounded-lg bg-secondary border border-border px-2 text-xs text-foreground focus:outline-none focus:border-primary"
						>
							<option value="==">== (Equals)</option>
							<option value="!=">!= (Not equals)</option>
							<option value=">">&gt; (Greater than)</option>
							<option value=">=">&gt;= (Greater or equal)</option>
							<option value="<">&lt; (Less than)</option>
							<option value="in">in (Matches list/range)</option>
						</select>
					</div>

					<div class="space-y-1">
						<label for="target-value-input" class="text-xs font-medium text-foreground"
							>Target Value</label
						>
						<input
							id="target-value-input"
							type="text"
							required
							bind:value={targetValue}
							placeholder="e.g. 5xx, 200, 2000"
							class="w-full h-9 rounded-lg bg-secondary border border-border px-3 text-xs text-foreground focus:outline-none focus:border-primary"
						/>
					</div>
				</div>

				<!-- Consecutive & Severity -->
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div class="space-y-1">
						<label for="consecutive-input" class="text-xs font-medium text-foreground"
							>Consecutive Breaches</label
						>
						<input
							id="consecutive-input"
							type="number"
							min="1"
							max="10"
							bind:value={consecutive}
							class="w-full h-9 rounded-lg bg-secondary border border-border px-3 text-xs text-foreground focus:outline-none focus:border-primary"
						/>
						<span class="text-[10px] text-muted-foreground">Times in a row before alert</span>
					</div>

					<div class="space-y-1">
						<label for="severity-select" class="text-xs font-medium text-foreground"
							>Severity Level</label
						>
						<select
							id="severity-select"
							bind:value={severity}
							class="w-full h-9 rounded-lg bg-secondary border border-border px-2 text-xs text-foreground focus:outline-none focus:border-primary"
						>
							<option value="critical">Critical (Down)</option>
							<option value="warning">Warning (Degraded)</option>
						</select>
						<span class="text-[10px] text-muted-foreground">Sets incident urgency</span>
					</div>
				</div>

				<div class="flex items-center justify-end gap-2 pt-3 border-t border-border">
					<Button
						type="button"
						variant="outline"
						class="h-9 px-4 text-xs"
						onclick={() => (showAddModal = false)}
					>
						Cancel
					</Button>
					<Button
						type="submit"
						disabled={isSaving}
						class="bg-primary text-primary-foreground h-9 px-4 text-xs font-medium"
					>
						{isSaving ? 'Saving...' : 'Create Alert Rule'}
					</Button>
				</div>
			</form>
		</div>
	</div>
{/if}
