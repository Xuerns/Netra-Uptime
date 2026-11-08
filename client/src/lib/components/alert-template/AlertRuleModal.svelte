<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { AlertTemplateAPI } from '$lib/api/alertTemplate';
	import type {
		AlertRule,
		CreateAlertRuleRequest,
		UpdateAlertRuleRequest
	} from '$lib/api/alertRule';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { ShieldAlert, AlertTriangle } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		orgId: string;
		templateId: string;
		ruleToEdit?: AlertRule | null;
		onSuccess?: () => void;
	}

	let {
		open = $bindable(false),
		orgId,
		templateId,
		ruleToEdit = null,
		onSuccess
	}: Props = $props();

	let isEditing = $derived(!!ruleToEdit);
	let ruleName = $state('5xx Server Error Alert');
	let metric = $state<'status_code' | 'response_time_ms' | 'timeout'>('status_code');
	let operator = $state<'==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in'>('==');
	let targetValue = $state('5xx');
	let consecutive = $state(2);
	let severity = $state<'critical' | 'warning'>('critical');
	let isRuleActive = $state(true);
	let isSaving = $state(false);

	const metricOptions = [
		{ value: 'status_code', label: 'Status Code' },
		{ value: 'response_time_ms', label: 'Latency (ms)' },
		{ value: 'timeout', label: 'Timeout' }
	] as const;

	const operatorOptions = [
		{ value: '==', label: '== (Equals)' },
		{ value: '!=', label: '!= (Not Equals)' },
		{ value: '>', label: '> (Greater than)' },
		{ value: '>=', label: '>= (Greater or equal)' },
		{ value: '<', label: '< (Less than)' },
		{ value: '<=', label: '<= (Less or equal)' }
	] as const;

	let selectedMetricLabel = $derived(
		metricOptions.find((o) => o.value === metric)?.label ?? 'Status Code'
	);
	let selectedOperatorLabel = $derived(
		operatorOptions.find((o) => o.value === operator)?.label ?? operator
	);

	$effect(() => {
		if (open) {
			if (ruleToEdit) {
				ruleName = ruleToEdit.name;
				metric = ruleToEdit.metric as 'status_code' | 'response_time_ms' | 'timeout';
				operator = ruleToEdit.operator as '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
				targetValue = ruleToEdit.target_value;
				consecutive = ruleToEdit.consecutive || 2;
				severity = ruleToEdit.severity as 'critical' | 'warning';
				isRuleActive = ruleToEdit.is_active ?? true;
			} else {
				ruleName = '5xx Server Error Alert';
				metric = 'status_code';
				operator = '==';
				targetValue = '5xx';
				consecutive = 2;
				severity = 'critical';
				isRuleActive = true;
			}
			isSaving = false;
		}
	});

	async function handleSaveRule(e: Event) {
		e.preventDefault();
		if (!ruleName.trim() || !targetValue.trim() || !orgId || !templateId) return;

		isSaving = true;
		try {
			if (isEditing && ruleToEdit) {
				const payload: UpdateAlertRuleRequest = {
					name: ruleName.trim(),
					metric,
					operator,
					target_value: targetValue.trim(),
					consecutive: Number(consecutive) || 2,
					severity,
					is_active: isRuleActive
				};
				await AlertTemplateAPI.updateRule(orgId, templateId, ruleToEdit.id, payload);
				toast.success('Alert rule updated');
			} else {
				const payload: CreateAlertRuleRequest = {
					name: ruleName.trim(),
					metric,
					operator,
					target_value: targetValue.trim(),
					consecutive: Number(consecutive) || 2,
					severity,
					is_active: isRuleActive
				};
				await AlertTemplateAPI.createRule(orgId, templateId, payload);
				toast.success('Alert rule added to template');
			}
			open = false;
			onSuccess?.();
		} catch (err: unknown) {
			toast.error(parseApiError(err, 'Failed to save alert rule'));
		} finally {
			isSaving = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="w-[calc(100%-2rem)] sm:max-w-155! bg-card border border-border shadow-2xl p-6 rounded-lg text-foreground"
	>
		<Dialog.Header>
			<div class="flex items-center gap-2.5">
				<Dialog.Title class="text-lg font-semibold tracking-tight text-foreground">
					{isEditing ? 'Edit Alert Rule' : 'Add Alert Rule'}
				</Dialog.Title>
			</div>
			<Dialog.Description class="text-xs text-muted-foreground mt-1">
				Configure the threshold, comparison operator, and consecutive breach condition for this
				rule.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-2"></div>

		<form onsubmit={handleSaveRule} class="space-y-4">
			<div class="space-y-1.5">
				<Label for="rule-name-input" class="text-xs font-medium">Rule Name</Label>
				<Input
					id="rule-name-input"
					bind:value={ruleName}
					placeholder="e.g. 5xx Server Error Alert"
					required
					class="bg-background border-border text-foreground h-10"
				/>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
				<div class="space-y-1.5">
					<Label class="text-xs font-medium">Metric</Label>
					<Select.Root
						type="single"
						value={metric}
						onValueChange={(val) => {
							if (val) metric = val as 'status_code' | 'response_time_ms' | 'timeout';
						}}
					>
						<Select.Trigger
							data-testid="rule-metric"
							class="w-full h-10 rounded-md border border-border bg-background px-3 text-xs text-foreground focus:ring-1 focus:ring-primary justify-between"
						>
							<span>{selectedMetricLabel}</span>
						</Select.Trigger>
						<Select.Content class="z-[60] bg-card border-border text-foreground">
							{#each metricOptions as opt (opt.value)}
								<Select.Item
									value={opt.value}
									label={opt.label}
									class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
								>
									{opt.label}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="space-y-1.5">
					<Label class="text-xs font-medium">Operator</Label>
					<Select.Root
						type="single"
						value={operator}
						onValueChange={(val) => {
							if (val) operator = val as '==' | '!=' | '>' | '>=' | '<' | '<=' | 'in' | 'not_in';
						}}
					>
						<Select.Trigger
							data-testid="rule-operator"
							class="w-full h-10 rounded-md border border-border bg-background px-3 text-xs text-foreground focus:ring-1 focus:ring-primary justify-between"
						>
							<span>{selectedOperatorLabel}</span>
						</Select.Trigger>
						<Select.Content class="z-[60] bg-card border-border text-foreground">
							{#each operatorOptions as opt (opt.value)}
								<Select.Item
									value={opt.value}
									label={opt.label}
									class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
								>
									{opt.label}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
				<div class="space-y-1.5">
					<Label for="rule-target" class="text-xs font-medium">Target Value</Label>
					<Input
						id="rule-target"
						bind:value={targetValue}
						placeholder={metric === 'status_code' ? '5xx or 500' : '2000'}
						required
						class="bg-background border-border text-foreground font-mono h-10"
					/>
				</div>

				<div class="space-y-1.5">
					<Label for="rule-consecutive" class="text-xs font-medium">Consecutive Breaches</Label>
					<Input
						id="rule-consecutive"
						type="number"
						min={1}
						max={20}
						bind:value={consecutive}
						required
						class="bg-background border-border text-foreground font-mono h-10"
					/>
				</div>
			</div>

			<div class="space-y-1.5">
				<Label class="text-xs font-medium">Severity Level</Label>
				<div class="grid grid-cols-2 gap-3">
					<button
						type="button"
						onclick={() => (severity = 'critical')}
						class="flex items-center justify-center gap-2 h-10 rounded-lg border text-xs font-medium transition-colors cursor-pointer {severity ===
						'critical'
							? 'bg-rose-500/10 border-rose-500/40 text-rose-400 font-semibold shadow-xs'
							: 'bg-background border-border text-muted-foreground hover:bg-white/5'}"
					>
						<ShieldAlert class="size-4" />
						<span>Critical</span>
					</button>

					<button
						type="button"
						onclick={() => (severity = 'warning')}
						class="flex items-center justify-center gap-2 h-10 rounded-lg border text-xs font-medium transition-colors cursor-pointer {severity ===
						'warning'
							? 'bg-amber-500/10 border-amber-500/40 text-amber-400 font-semibold shadow-xs'
							: 'bg-background border-border text-muted-foreground hover:bg-white/5'}"
					>
						<AlertTriangle class="size-4" />
						<span>Warning</span>
					</button>
				</div>
			</div>

			<label
				for="rule-is-active"
				class="flex items-center gap-3 p-3 rounded-lg border border-border bg-background cursor-pointer select-none"
			>
				<Checkbox id="rule-is-active" bind:checked={isRuleActive} />
				<div class="space-y-0.5 text-xs">
					<p class="font-medium text-foreground">Rule is active</p>
					<p class="text-muted-foreground text-[11px] leading-relaxed">
						When enabled, monitor checks that breach these thresholds will trigger incident alerts.
					</p>
				</div>
			</label>

			<Dialog.Footer class="gap-2 sm:gap-0 pt-3">
				<button
					type="button"
					onclick={() => (open = false)}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg border border-border text-xs font-medium text-muted-foreground hover:text-foreground cursor-pointer"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSaving || !ruleName.trim() || !targetValue.trim()}
					class="inline-flex items-center justify-center h-9 px-4 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 disabled:opacity-50 cursor-pointer shadow-sm"
				>
					{isSaving ? 'Saving...' : isEditing ? 'Update Rule' : 'Add Rule'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
