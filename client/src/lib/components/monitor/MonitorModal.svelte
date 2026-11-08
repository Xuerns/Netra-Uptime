<script lang="ts">
	import { untrack } from 'svelte';
	import { MonitorAPI, type Monitor } from '$lib/api/monitor';
	import { OrgAPI } from '$lib/api/organization';
	import { AlertTemplateAPI, type AlertTemplate } from '$lib/api/alertTemplate';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		validateMonitorName,
		validateMonitorTarget,
		validateMonitorInterval,
		validateMonitorTimeout
	} from '$lib/utils/validation';
	import { debounce } from '$lib/utils';
	import MonitorTypeGuide from '$lib/components/monitor/MonitorTypeGuide.svelte';
	import AssigneeGuide from '$lib/components/monitor/AssigneeGuide.svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';
	import { getAvatarColor } from '$lib/utils/string';
	import { logger } from '$lib/utils/logger';
	import { Plus, Trash2, ChevronDown, ChevronRight } from 'lucide-svelte';

	interface Props {
		isOpen?: boolean;
		monitorType?: 'http' | 'ping' | 'tcp' | null;
		monitor?: Monitor | null;
		orgId: string;
		onSuccess: () => void;
	}

	let {
		isOpen = $bindable(false),
		monitorType = null,
		monitor = null,
		orgId,
		onSuccess
	}: Props = $props();

	const isEdit = $derived(!!monitor);
	const effectiveType = $derived<string>(monitor?.Type || monitorType || 'http');

	const INTERVAL_OPTIONS = [
		{ value: '30', label: '30s' },
		{ value: '60', label: '1m' },
		{ value: '120', label: '2m' },
		{ value: '300', label: '5m' },
		{ value: '600', label: '10m' },
		{ value: '1800', label: '30m' },
		{ value: '3600', label: '1h' }
	];

	const TIMEOUT_OPTIONS = [
		{ value: '2', label: '2s' },
		{ value: '5', label: '5s' },
		{ value: '10', label: '10s' },
		{ value: '15', label: '15s' },
		{ value: '20', label: '20s' },
		{ value: '30', label: '30s' }
	];

	function formatIntervalLabel(seconds: number): string {
		const match = INTERVAL_OPTIONS.find((opt) => opt.value === seconds.toString());
		if (match) return match.label;
		if (seconds % 3600 === 0) return `${seconds / 3600}h`;
		if (seconds % 60 === 0) return `${seconds / 60}m`;
		return `${seconds}s`;
	}

	function formatTimeoutLabel(seconds: number): string {
		const match = TIMEOUT_OPTIONS.find((opt) => opt.value === seconds.toString());
		if (match) return match.label;
		return `${seconds}s`;
	}

	let name = $state('');
	let target = $state('');
	let intervalSeconds = $state<number>(60);
	let intervalSecondsStr = $state<string>('60');
	let timeoutSeconds = $state<number>(10);
	let timeoutSecondsStr = $state<string>('10');
	let alertTemplates = $state<AlertTemplate[]>([]);
	let selectedTemplateId = $state<string>('');
	let expectedStatusCodeStr = $state<string>('200');
	let httpMethod = $state<'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD'>('GET');
	let headerPairs = $state<{ key: string; value: string }[]>([]);
	let httpBody = $state<string>('');
	let showAdvancedHttp = $state<boolean>(false);
	let isActive = $state<boolean>(true);
	let criticality = $state<'extreme' | 'high' | 'medium' | 'low'>('medium');
	let enableSSLDomainCheck = $state<boolean>(true);
	let loading = $state(false);
	let error = $state('');

	const HTTP_METHODS: Array<'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD'> = [
		'GET',
		'POST',
		'PUT',
		'PATCH',
		'HEAD'
	];

	function addHeaderPair() {
		if (headerPairs.length < 25) {
			headerPairs = [...headerPairs, { key: '', value: '' }];
		}
	}

	function removeHeaderPair(index: number) {
		headerPairs = headerPairs.filter((_, idx) => idx !== index);
	}

	function serializeHeaders(): string {
		const map: Record<string, string> = {};
		for (const pair of headerPairs) {
			const k = pair.key.trim();
			if (k) {
				map[k] = pair.value.trim();
			}
		}
		return Object.keys(map).length > 0 ? JSON.stringify(map) : '';
	}

	let fieldErrors = $state({ name: '', target: '', intervalSeconds: '', timeoutSeconds: '' });
	let shake = $state({ name: false, target: false, intervalSeconds: false, timeoutSeconds: false });

	let members = $state<any[]>([]);
	let selectedAssignees = $state<string[]>([]);

	const handleNameValidation = debounce((val: string) => {
		const err = validateMonitorName(val);
		if (err) {
			fieldErrors.name = err;
			shake.name = false;
			setTimeout(() => (shake.name = true), 10);
		} else {
			fieldErrors.name = '';
			shake.name = false;
		}
	}, 500);

	const handleTargetValidation = debounce((val: string) => {
		const err = validateMonitorTarget(val);
		if (err) {
			fieldErrors.target = err;
			shake.target = false;
			setTimeout(() => (shake.target = true), 10);
		} else {
			fieldErrors.target = '';
			shake.target = false;
		}
	}, 500);

	const handleIntervalValidation = debounce((val: number) => {
		const err = validateMonitorInterval(val);
		if (err) {
			fieldErrors.intervalSeconds = err;
			shake.intervalSeconds = false;
			setTimeout(() => (shake.intervalSeconds = true), 10);
		} else {
			fieldErrors.intervalSeconds = '';
			shake.intervalSeconds = false;
		}

		const timeoutErr = validateMonitorTimeout(timeoutSeconds, val);
		if (timeoutErr) {
			fieldErrors.timeoutSeconds = timeoutErr;
			shake.timeoutSeconds = false;
			setTimeout(() => (shake.timeoutSeconds = true), 10);
		} else {
			fieldErrors.timeoutSeconds = '';
			shake.timeoutSeconds = false;
		}
	}, 500);

	const handleTimeoutValidation = debounce((val: number) => {
		const err = validateMonitorTimeout(val, intervalSeconds);
		if (err) {
			fieldErrors.timeoutSeconds = err;
			shake.timeoutSeconds = false;
			setTimeout(() => (shake.timeoutSeconds = true), 10);
		} else {
			fieldErrors.timeoutSeconds = '';
			shake.timeoutSeconds = false;
		}
	}, 500);

	function handleIntervalSelect(val: string) {
		intervalSecondsStr = val;
		const num = parseInt(val, 10) || 60;
		intervalSeconds = num;

		// When user selects a smaller interval such that current timeout is >= intervalSeconds,
		// automatically clamp/adjust timeoutSeconds to a valid timeout option (< intervalSeconds, max 30s)
		if (timeoutSeconds >= num || timeoutSeconds > 30) {
			const validPresets = TIMEOUT_OPTIONS.map((o) => parseInt(o.value, 10)).filter((v) => v < num);
			const newTimeout =
				validPresets.length > 0
					? Math.max(...validPresets)
					: Math.max(1, Math.min(timeoutSeconds, Math.min(num - 1, 30)));
			timeoutSeconds = newTimeout;
			timeoutSecondsStr = newTimeout.toString();
			handleTimeoutValidation(newTimeout);
		}

		handleIntervalValidation(num);
	}

	function handleTimeoutSelect(val: string) {
		timeoutSecondsStr = val;
		const num = parseInt(val, 10) || 10;
		timeoutSeconds = num;
		handleTimeoutValidation(num);
	}

	let canSubmit = $derived(
		!loading &&
			!fieldErrors.name &&
			!fieldErrors.target &&
			!fieldErrors.intervalSeconds &&
			!fieldErrors.timeoutSeconds
	);

	let lastOpenedKey = $state<string | null>(null);

	function initForm() {
		error = '';
		fieldErrors = { name: '', target: '', intervalSeconds: '', timeoutSeconds: '' };
		shake = { name: false, target: false, intervalSeconds: false, timeoutSeconds: false };

		if (monitor) {
			// Edit mode: prefill from existing monitor
			name = monitor.Name || '';
			target = monitor.Target || '';
			intervalSeconds = monitor.IntervalSeconds || 60;
			intervalSecondsStr = intervalSeconds.toString();
			let initialTimeout = monitor.TimeoutSeconds || 10;
			if (initialTimeout > 30 || initialTimeout >= intervalSeconds) {
				const validPresets = TIMEOUT_OPTIONS.map((o) => parseInt(o.value, 10)).filter(
					(v) => v < intervalSeconds
				);
				initialTimeout =
					validPresets.length > 0
						? Math.max(...validPresets)
						: Math.max(1, Math.min(initialTimeout, Math.min(intervalSeconds - 1, 30)));
			}
			timeoutSeconds = initialTimeout;
			timeoutSecondsStr = timeoutSeconds.toString();
			expectedStatusCodeStr = monitor.ExpectedStatusCode
				? monitor.ExpectedStatusCode.toString()
				: '200';
			httpMethod =
				(monitor.HTTPMethod?.toUpperCase() as 'GET' | 'POST' | 'PUT' | 'PATCH' | 'HEAD') || 'GET';
			httpBody = monitor.HTTPBody || '';
			let parsedPairs: { key: string; value: string }[] = [];
			if (monitor.HTTPHeaders) {
				try {
					const parsed = JSON.parse(monitor.HTTPHeaders);
					if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
						parsedPairs = Object.entries(parsed).map(([k, v]) => ({
							key: String(k),
							value: String(v ?? '')
						}));
					}
				} catch {
					parsedPairs = [];
				}
			}
			headerPairs = parsedPairs;
			showAdvancedHttp = parsedPairs.length > 0 || httpBody.trim() !== '' || httpMethod !== 'GET';
			isActive = monitor.IsActive !== undefined ? monitor.IsActive : true;
			criticality = (monitor.Criticality as any) || 'medium';
			enableSSLDomainCheck =
				(monitor as any).EnableSSLDomainCheck !== undefined
					? (monitor as any).EnableSSLDomainCheck
					: true;
			selectedTemplateId =
				(monitor as any).AlertTemplateID || (monitor as any).alert_template_id || '';

			const rawAssignees =
				(monitor as any).Assignees ||
				(monitor as any).assignees ||
				(monitor as any).MonitorPICs ||
				(monitor as any).monitor_pics ||
				[];
			selectedAssignees = Array.isArray(rawAssignees)
				? rawAssignees
						.map((pic: any) =>
							typeof pic === 'string' ? pic : pic?.UserID || pic?.user_id || pic?.id
						)
						.filter(Boolean)
						.slice(0, 50)
				: [];
		} else {
			// Create mode: defaults
			name = '';
			target = '';
			intervalSeconds = 60;
			intervalSecondsStr = '60';
			timeoutSeconds = 10;
			timeoutSecondsStr = '10';
			expectedStatusCodeStr = '200';
			httpMethod = 'GET';
			headerPairs = [];
			httpBody = '';
			showAdvancedHttp = false;
			isActive = true;
			criticality = 'medium';
			enableSSLDomainCheck = true;
			selectedAssignees = [];
			selectedTemplateId = '';
		}

		if (orgId) {
			OrgAPI.getMembers(orgId)
				.then((res) => {
					members = res.data ?? [];
				})
				.catch((err) => {
					logger.error('Failed to load members', err);
				});

			AlertTemplateAPI.list(orgId)
				.then((res) => {
					alertTemplates = res.data ?? [];
					if (!monitor) {
						const defaultTmpl = alertTemplates.find((t) => t.is_default);
						if (defaultTmpl && !selectedTemplateId) {
							selectedTemplateId = defaultTmpl.id;
						}
					}
				})
				.catch(() => {});
		}
	}

	// Sync form state whenever dialog opens or monitor prop changes
	$effect(() => {
		if (isOpen) {
			const currentKey = monitor ? `edit-${monitor.ID}` : `create-${effectiveType}`;
			if (lastOpenedKey !== currentKey) {
				lastOpenedKey = currentKey;
				untrack(() => {
					initForm();
				});
			}
		} else {
			lastOpenedKey = null;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!canSubmit) return;

		loading = true;
		error = '';

		try {
			const serializedHeaders = effectiveType === 'http' ? serializeHeaders() : undefined;
			const supportsBody = httpMethod === 'POST' || httpMethod === 'PUT' || httpMethod === 'PATCH';
			const effectiveBody = effectiveType === 'http' ? (supportsBody ? httpBody : '') : undefined;

			if (isEdit && monitor) {
				await MonitorAPI.update(orgId, monitor.ID, {
					name,
					target,
					interval_seconds: intervalSeconds,
					timeout_seconds: timeoutSeconds,
					expectation_status_code:
						effectiveType === 'http' ? parseInt(expectedStatusCodeStr) : undefined,
					http_method: effectiveType === 'http' ? httpMethod : undefined,
					http_headers: serializedHeaders,
					http_body: effectiveBody,
					is_active: isActive,
					criticality,
					assignees: selectedAssignees,
					MonitorPICs: selectedAssignees,
					enable_ssl_domain_check: enableSSLDomainCheck,
					alert_template_id: selectedTemplateId || null
				});
				toast.success('Monitor updated successfully');
			} else {
				if (!effectiveType) throw new Error('Monitor type is required');
				await MonitorAPI.create(orgId, {
					name,
					type: effectiveType as 'http' | 'tcp' | 'ping',
					target,
					interval_seconds: intervalSeconds,
					timeout_seconds: timeoutSeconds,
					expectation_status_code:
						effectiveType === 'http' ? parseInt(expectedStatusCodeStr) : undefined,
					http_method: effectiveType === 'http' ? httpMethod : undefined,
					http_headers: serializedHeaders,
					http_body: effectiveBody,
					criticality,
					assignees: selectedAssignees,
					MonitorPICs: selectedAssignees,
					enable_ssl_domain_check: enableSSLDomainCheck,
					alert_template_id: selectedTemplateId || null
				});
				toast.success('Monitor created successfully');
			}

			onSuccess();
			isOpen = false;
		} catch (err: any) {
			error = parseApiError(err, isEdit ? 'Failed to update monitor' : 'Failed to create monitor');
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	let modalTitle = $derived(
		isEdit
			? 'Edit Monitor'
			: effectiveType === 'http'
				? 'Add HTTP/API Monitor'
				: effectiveType === 'ping'
					? 'Add Ping Monitor'
					: effectiveType === 'tcp'
						? 'Add TCP Port Monitor'
						: 'Add Monitor'
	);

	let selectedTemplateLabel = $derived.by(() => {
		if (!selectedTemplateId || selectedTemplateId === 'none') {
			return 'None (Use System Defaults)';
		}
		const found = alertTemplates.find((t) => t.id === selectedTemplateId);
		if (!found) return 'None (Use System Defaults)';
		return `${found.name} (${found.rules?.length || 0} rules)${found.is_default ? ' (Default)' : ''}`;
	});

	const MONITOR_AVATAR_PALETTE = [
		'bg-red-500/10 text-red-500 border-red-500/20',
		'bg-blue-500/10 text-blue-500 border-blue-500/20',
		'bg-green-500/10 text-green-500 border-green-500/20',
		'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
		'bg-purple-500/10 text-purple-500 border-purple-500/20'
	];
	const avatarColor = (nameStr: string) => getAvatarColor(nameStr, MONITOR_AVATAR_PALETTE);
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		data-lenis-prevent
		class="sm:max-w-4xl max-h-[92vh] overflow-y-auto bg-card border border-border text-foreground p-6 rounded-lg"
	>
		<Dialog.Header>
			<div class="flex items-center gap-2">
				<Dialog.Title class="text-xl font-semibold tracking-tight">{modalTitle}</Dialog.Title>
				{#if !isEdit}
					<MonitorTypeGuide monitorType={effectiveType as any} />
				{/if}
			</div>
			<Dialog.Description class="text-muted-foreground text-sm mt-0.5">
				{isEdit
					? 'Update your monitor configurations and alert policies below.'
					: 'Configure your new monitor settings and notification assignees below.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-3"></div>

		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			{#if error}
				<div
					class="p-3 text-[13px] text-destructive bg-destructive/10 border border-destructive/20 rounded-lg font-medium"
				>
					{error}
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-12 gap-6">
				<!-- Left Column: Target & Check Settings (7 cols) -->
				<div class="md:col-span-7 space-y-4">
					<!-- Row 1: Monitor Name + Criticality Level -->
					<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
						<div class="sm:col-span-2 space-y-1.5">
							<div class="flex items-center gap-2">
								<Label for="monitor-name" class="text-sm font-medium text-foreground">
									Monitor Name <span class="text-destructive font-bold">*</span>
								</Label>
								{#if fieldErrors.name}
									<span class="text-xs text-destructive font-medium">{fieldErrors.name}</span>
								{/if}
							</div>
							<Input
								id="monitor-name"
								placeholder="e.g., Production API"
								bind:value={name}
								oninput={() => handleNameValidation(name)}
								required
								class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-9.5 {fieldErrors.name
									? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
									: ''} {shake.name ? 'animate-shake' : ''}"
							/>
						</div>

						<div class="space-y-1.5">
							<Label class="text-sm font-medium text-foreground">Criticality</Label>
							<Select.Root type="single" bind:value={criticality}>
								<Select.Trigger
									class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full capitalize"
								>
									{criticality}
								</Select.Trigger>
								<Select.Content class="bg-card border-border text-foreground">
									<Select.Item
										value="extreme"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>Extreme</Select.Item
									>
									<Select.Item
										value="high"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>High</Select.Item
									>
									<Select.Item
										value="medium"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>Medium</Select.Item
									>
									<Select.Item
										value="low"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>Low</Select.Item
									>
								</Select.Content>
							</Select.Root>
						</div>
					</div>

					<!-- Row 2: Target URL / Hostname -->
					<div class="space-y-1.5">
						<div class="flex items-center gap-3">
							<Label for="monitor-target" class="text-sm font-medium text-foreground">
								{effectiveType === 'http'
									? 'URL'
									: effectiveType === 'ping'
										? 'Hostname / IP'
										: 'Hostname / IP:Port'}
								<span class="text-destructive font-bold">*</span>
							</Label>
							{#if fieldErrors.target}
								<span class="text-xs text-destructive font-medium">{fieldErrors.target}</span>
							{/if}
						</div>
						<Input
							id="monitor-target"
							placeholder={effectiveType === 'http'
								? 'https://api.example.com'
								: effectiveType === 'ping'
									? '8.8.8.8'
									: 'db.example.com:5432'}
							bind:value={target}
							oninput={() => handleTargetValidation(target)}
							required
							class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-9.5 font-mono {fieldErrors.target
								? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
								: ''} {shake.target ? 'animate-shake' : ''}"
						/>
					</div>

					<!-- Row 3: Method, Expected Status, Check Interval, Timeout in a single compact grid -->
					<div
						class="grid grid-cols-2 {effectiveType === 'http'
							? 'sm:grid-cols-4'
							: 'sm:grid-cols-2'} gap-3"
					>
						{#if effectiveType === 'http'}
							<div class="space-y-1.5">
								<Label class="text-xs font-medium text-foreground">HTTP Method</Label>
								<Select.Root
									type="single"
									bind:value={httpMethod}
									onValueChange={(val) => {
										if (val && ['POST', 'PUT', 'PATCH'].includes(val)) {
											showAdvancedHttp = true;
										}
									}}
								>
									<Select.Trigger
										class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full font-mono text-xs"
									>
										{httpMethod}
									</Select.Trigger>
									<Select.Content class="bg-card border-border text-foreground">
										{#each HTTP_METHODS as methodOpt (methodOpt)}
											<Select.Item
												value={methodOpt}
												class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer font-mono text-xs"
											>
												{methodOpt}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>

							<div class="space-y-1.5">
								<Label class="text-xs font-medium text-foreground">Expected Status</Label>
								<Select.Root type="single" bind:value={expectedStatusCodeStr}>
									<Select.Trigger
										class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full text-xs"
									>
										{expectedStatusCodeStr === '200'
											? '200 (OK)'
											: expectedStatusCodeStr === '201'
												? '201 (Created)'
												: expectedStatusCodeStr === '301'
													? '301 (Redirect)'
													: expectedStatusCodeStr === '404'
														? '404 (Not Found)'
														: expectedStatusCodeStr}
									</Select.Trigger>
									<Select.Content class="bg-card border-border text-foreground">
										<Select.Item
											value="200"
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
											>200 (OK)</Select.Item
										>
										<Select.Item
											value="201"
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
											>201 (Created)</Select.Item
										>
										<Select.Item
											value="301"
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
											>301 (Moved Permanently)</Select.Item
										>
										<Select.Item
											value="404"
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
											>404 (Not Found)</Select.Item
										>
									</Select.Content>
								</Select.Root>
							</div>
						{/if}

						<!-- Check Interval Dropdown -->
						<div class="space-y-1.5">
							<div class="flex items-center gap-1.5">
								<Label class="text-xs font-medium text-foreground">
									Interval <span class="text-destructive font-bold">*</span>
								</Label>
								{#if fieldErrors.intervalSeconds}
									<span class="text-[10px] text-destructive font-medium"
										>{fieldErrors.intervalSeconds}</span
									>
								{/if}
							</div>
							<Select.Root
								type="single"
								bind:value={intervalSecondsStr}
								onValueChange={(val) => {
									if (val) handleIntervalSelect(val);
								}}
							>
								<Select.Trigger
									class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full font-mono text-xs {fieldErrors.intervalSeconds
										? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
										: ''} {shake.intervalSeconds ? 'animate-shake' : ''}"
								>
									{formatIntervalLabel(intervalSeconds)}
								</Select.Trigger>
								<Select.Content class="bg-card border-border text-foreground">
									{#each INTERVAL_OPTIONS as opt (opt.value)}
										<Select.Item
											value={opt.value}
											label={opt.label}
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer font-mono text-xs"
										>
											{opt.label}
										</Select.Item>
									{/each}
									{#if !INTERVAL_OPTIONS.some((opt) => opt.value === intervalSecondsStr)}
										<Select.Item
											value={intervalSecondsStr}
											label={formatIntervalLabel(intervalSeconds)}
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer font-mono text-xs"
										>
											{formatIntervalLabel(intervalSeconds)}
										</Select.Item>
									{/if}
								</Select.Content>
							</Select.Root>
						</div>

						<!-- Timeout Dropdown -->
						<div class="space-y-1.5">
							<div class="flex items-center gap-1.5">
								<Label class="text-xs font-medium text-foreground">
									Timeout <span class="text-destructive font-bold">*</span>
								</Label>
								{#if fieldErrors.timeoutSeconds}
									<span class="text-[10px] text-destructive font-medium"
										>{fieldErrors.timeoutSeconds}</span
									>
								{/if}
							</div>
							<Select.Root
								type="single"
								bind:value={timeoutSecondsStr}
								onValueChange={(val) => {
									if (val) handleTimeoutSelect(val);
								}}
							>
								<Select.Trigger
									class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full font-mono text-xs {fieldErrors.timeoutSeconds
										? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
										: ''} {shake.timeoutSeconds ? 'animate-shake' : ''}"
								>
									{formatTimeoutLabel(timeoutSeconds)}
								</Select.Trigger>
								<Select.Content class="bg-card border-border text-foreground">
									{#each TIMEOUT_OPTIONS as opt (opt.value)}
										{@const optNum = parseInt(opt.value, 10)}
										{@const isExceeding = optNum >= intervalSeconds}
										<Select.Item
											value={opt.value}
											label={opt.label}
											disabled={isExceeding}
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer font-mono text-xs {isExceeding
												? 'opacity-40 cursor-not-allowed pointer-events-none'
												: ''}"
										>
											{opt.label}
										</Select.Item>
									{/each}
									{#if !TIMEOUT_OPTIONS.some((opt) => opt.value === timeoutSecondsStr)}
										<Select.Item
											value={timeoutSecondsStr}
											label={formatTimeoutLabel(timeoutSeconds)}
											class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer font-mono text-xs"
										>
											{formatTimeoutLabel(timeoutSeconds)}
										</Select.Item>
									{/if}
								</Select.Content>
							</Select.Root>
						</div>
					</div>

					<!-- Row 4: Advanced HTTP Request (Custom Headers & Body) -->
					{#if effectiveType === 'http'}
						<div class="rounded-lg border border-border bg-background/60 p-3 space-y-2.5">
							<button
								type="button"
								onclick={() => (showAdvancedHttp = !showAdvancedHttp)}
								class="w-full flex items-center justify-between text-left text-xs font-semibold text-foreground hover:text-primary transition-colors cursor-pointer"
							>
								<span class="flex items-center gap-1.5">
									{#if showAdvancedHttp}
										<ChevronDown class="size-3.5 text-primary" />
									{:else}
										<ChevronRight class="size-3.5 text-muted-foreground" />
									{/if}
									<span>Advanced HTTP Request (Headers & Body)</span>
								</span>
								<span class="text-[11px] font-mono text-muted-foreground shrink-0">
									{headerPairs.filter((p) => p.key.trim() !== '').length} header(s)
								</span>
							</button>

							{#if showAdvancedHttp}
								<div class="pt-2 border-t border-border/60 space-y-3">
									<!-- Custom HTTP Headers -->
									<div class="space-y-2">
										<div class="flex items-center justify-between">
											<Label class="text-xs font-medium text-foreground">Custom HTTP Headers</Label>
											<button
												type="button"
												onclick={addHeaderPair}
												disabled={headerPairs.length >= 25}
												class="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline disabled:opacity-40 cursor-pointer"
											>
												<Plus class="size-3.5" />
												<span>Add Header</span>
											</button>
										</div>

										{#if headerPairs.length === 0}
											<p class="text-[11px] text-muted-foreground">
												No custom headers configured. Click "Add Header" to include Authorization or
												custom API headers.
											</p>
										{:else}
											<div data-lenis-prevent class="space-y-2 max-h-32 overflow-y-auto pr-0.5">
												{#each headerPairs as pair, idx (idx)}
													<div class="grid grid-cols-12 items-center gap-2">
														<Input
															placeholder="Header-Name (e.g. Authorization)"
															bind:value={pair.key}
															class="col-span-5 bg-background border-border text-foreground placeholder:text-muted-foreground rounded-md h-8 text-xs font-mono"
														/>
														<Input
															placeholder="Value (e.g. Bearer token)"
															bind:value={pair.value}
															class="col-span-6 bg-background border-border text-foreground placeholder:text-muted-foreground rounded-md h-8 text-xs font-mono"
														/>
														<button
															type="button"
															onclick={() => removeHeaderPair(idx)}
															aria-label="Remove header"
															class="col-span-1 flex items-center justify-center p-1.5 rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
														>
															<Trash2 class="size-3.5" />
														</button>
													</div>
												{/each}
											</div>
										{/if}
									</div>

									<!-- Request Body (for POST, PUT, PATCH) -->
									{#if httpMethod === 'POST' || httpMethod === 'PUT' || httpMethod === 'PATCH'}
										<div class="space-y-1.5">
											<div class="flex items-center justify-between">
												<Label for="monitor-http-body" class="text-xs font-medium text-foreground">
													Request Body ({httpMethod})
												</Label>
												<span class="text-[10px] font-mono text-muted-foreground"
													>JSON / Raw Payload</span
												>
											</div>
											<textarea
												id="monitor-http-body"
												data-lenis-prevent
												rows={3}
												placeholder={'{"ping": "pong"}'}
												bind:value={httpBody}
												class="w-full rounded-md bg-background border border-border p-2 text-xs font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary resize-y"
											></textarea>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/if}

					<!-- Row 5: Toggles (SSL & Domain Expiry Check + Active Toggle in Edit mode) -->
					{#if effectiveType === 'http' || isEdit}
						<div
							class="grid grid-cols-1 {effectiveType === 'http' && isEdit
								? 'sm:grid-cols-2'
								: 'sm:grid-cols-1'} gap-3"
						>
							{#if effectiveType === 'http'}
								<label
									for="monitor-ssl-check"
									class="flex items-center justify-between p-2.5 rounded-lg bg-background border border-border cursor-pointer select-none"
								>
									<div class="flex flex-col gap-0.5 pr-2">
										<span class="text-xs font-semibold text-foreground"
											>SSL & Domain Expiry Check</span
										>
										<span class="text-[11px] text-muted-foreground"
											>Monitors SSL cert & WHOIS daily</span
										>
									</div>
									<Checkbox id="monitor-ssl-check" bind:checked={enableSSLDomainCheck} />
								</label>
							{/if}

							{#if isEdit}
								<label
									for="monitor-is-active"
									class="flex items-center justify-between p-2.5 rounded-lg bg-background border border-border cursor-pointer select-none"
								>
									<div class="flex flex-col gap-0.5 pr-2">
										<span class="text-xs font-semibold text-foreground">Monitor is active</span>
										<span class="text-[11px] text-muted-foreground">
											{isActive ? 'Checks & alerts enabled' : 'Paused: checks & alerts paused'}
										</span>
									</div>
									<Checkbox id="monitor-is-active" bind:checked={isActive} />
								</label>
							{/if}
						</div>
					{/if}
				</div>

				<!-- Right Column: Alert Template & Assignees Settings (5 cols) -->
				<div class="md:col-span-5 flex flex-col gap-4 h-full">
					<!-- Alert Template Profile -->
					<div class="space-y-1.5">
						<Label
							for="monitor-template"
							class="text-sm font-medium text-foreground flex items-center justify-between"
						>
							<span>Alert Template</span>
							<span class="text-[11px] text-muted-foreground font-normal">Optional</span>
						</Label>
						<Select.Root
							type="single"
							value={selectedTemplateId || 'none'}
							onValueChange={(val) => {
								selectedTemplateId = val === 'none' ? '' : val;
							}}
						>
							<Select.Trigger
								id="monitor-template"
								class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-9.5 w-full text-xs truncate"
							>
								<span class="truncate">{selectedTemplateLabel}</span>
							</Select.Trigger>
							<Select.Content class="bg-card border-border text-foreground max-h-60">
								<Select.Item
									value="none"
									label="None (Use System Defaults)"
									class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer text-xs"
								>
									None (Use System Defaults)
								</Select.Item>
								{#each alertTemplates as tmpl (tmpl.id)}
									{@const tmplLabel = `${tmpl.name} (${tmpl.rules?.length || 0} rules)${tmpl.is_default ? ' (Default)' : ''}`}
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
					</div>

					<!-- Assignees List -->
					<div class="space-y-2 flex flex-col flex-1 min-h-0">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1.5">
								<Label class="text-sm font-medium text-foreground">Assignees</Label>
								<AssigneeGuide />
							</div>
							<span class="text-xs text-foreground/40 font-mono"
								>{selectedAssignees.length}/50 selected</span
							>
						</div>

						<div
							data-lenis-prevent
							class="flex-1 min-h-44 max-h-60 overflow-y-auto bg-background border border-border rounded-lg p-2 space-y-1"
						>
							{#each members as member (member.User?.ID || member.ID)}
								{@const memberId = member.User?.ID || member.ID}
								{@const isSelected = selectedAssignees.includes(memberId)}
								{@const isDisabled = !isSelected && selectedAssignees.length >= 50}

								<label
									class="flex items-center justify-between p-2 rounded-md transition-colors {isDisabled
										? 'opacity-40 cursor-not-allowed'
										: 'hover:bg-white/5 cursor-pointer'}"
								>
									<div class="flex items-center gap-3 overflow-hidden">
										{#if member.User?.AvatarURL}
											<img
												src={member.User.AvatarURL}
												alt={member.User.Name}
												class="size-7 rounded-full object-cover shrink-0 border border-border"
											/>
										{:else}
											<div
												class="size-7 rounded-full flex items-center justify-center font-semibold text-xs shrink-0 border {avatarColor(
													member.User?.Name || ''
												)}"
											>
												{member.User?.Name?.charAt(0).toUpperCase() || '?'}
											</div>
										{/if}

										<span class="text-[13px] font-medium text-foreground truncate"
											>{member.User?.Name || 'Unknown User'}</span
										>
									</div>

									<input
										type="checkbox"
										class="sr-only"
										checked={isSelected}
										disabled={isDisabled}
										onchange={(e) => {
											if (e.currentTarget.checked) {
												if (selectedAssignees.length < 50) {
													selectedAssignees = [...selectedAssignees, memberId];
												}
											} else {
												selectedAssignees = selectedAssignees.filter((id) => id !== memberId);
											}
										}}
									/>
									<div
										class="w-5 h-5 rounded-full border flex items-center justify-center shrink-0 transition-all {isSelected
											? 'bg-primary border-primary'
											: isDisabled
												? 'bg-white/5 border-border/40'
												: 'bg-white/5 border-border'}"
									>
										{#if isSelected}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												width="12"
												height="12"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="3"
												stroke-linecap="round"
												stroke-linejoin="round"
												class="text-black"><polyline points="20 6 9 17 4 12"></polyline></svg
											>
										{/if}
									</div>
								</label>
							{:else}
								<div
									class="p-4 text-xs text-foreground/40 text-center flex flex-col items-center justify-center h-full"
								>
									Loading members...
								</div>
							{/each}
						</div>

						{#if selectedAssignees.length === 50}
							<span class="text-[11px] text-amber-500 font-medium block"
								>Maximum 50 assignees reached for this monitor</span
							>
						{:else if selectedAssignees.length === 0}
							<span class="text-[11px] text-muted-foreground/60 block"
								>No assignees selected. Team notification channels will be used.</span
							>
						{/if}
					</div>
				</div>
			</div>

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-4 mt-1 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (isOpen = false)}
					class="px-5 py-2 rounded-lg text-sm font-medium text-foreground hover:bg-accent transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={loading || !canSubmit}
					class="h-10 px-6 rounded-lg text-sm font-medium bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 disabled:opacity-50 cursor-pointer"
				>
					{loading
						? isEdit
							? 'Saving...'
							: 'Creating...'
						: isEdit
							? 'Save Changes'
							: 'Create Monitor'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
