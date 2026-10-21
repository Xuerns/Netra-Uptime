<script lang="ts">
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
	import { debounce } from '$lib/helper/utils';
	import MonitorTypeGuide from '$lib/components/monitor/MonitorTypeGuide.svelte';
	import AssigneeGuide from '$lib/components/monitor/AssigneeGuide.svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';
	import { getAvatarColor } from '$lib/utils/string';

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

	let name = $state('');
	let target = $state('');
	let intervalSeconds = $state<number>(60);
	let timeoutSeconds = $state<number>(10);
	let alertTemplates = $state<AlertTemplate[]>([]);
	let selectedTemplateId = $state<string>('');
	let expectedStatusCodeStr = $state<string>('200');
	let isActive = $state<boolean>(true);
	let criticality = $state<'extreme' | 'high' | 'medium' | 'low'>('medium');
	let enableSSLDomainCheck = $state<boolean>(true);
	let loading = $state(false);
	let error = $state('');

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
	}, 500);

	const handleTimeoutValidation = debounce((val: number) => {
		const err = validateMonitorTimeout(val);
		if (err) {
			fieldErrors.timeoutSeconds = err;
			shake.timeoutSeconds = false;
			setTimeout(() => (shake.timeoutSeconds = true), 10);
		} else {
			fieldErrors.timeoutSeconds = '';
			shake.timeoutSeconds = false;
		}
	}, 500);

	let canSubmit = $derived(
		!loading &&
			!fieldErrors.name &&
			!fieldErrors.target &&
			!fieldErrors.intervalSeconds &&
			!fieldErrors.timeoutSeconds
	);

	// Sync form state whenever dialog opens or monitor prop changes
	$effect(() => {
		if (isOpen) {
			error = '';
			fieldErrors = { name: '', target: '', intervalSeconds: '', timeoutSeconds: '' };
			shake = { name: false, target: false, intervalSeconds: false, timeoutSeconds: false };

			if (monitor) {
				// Edit mode: prefill from existing monitor
				name = monitor.Name || '';
				target = monitor.Target || '';
				intervalSeconds = monitor.IntervalSeconds || 60;
				timeoutSeconds = monitor.TimeoutSeconds || 10;
				expectedStatusCodeStr = monitor.ExpectedStatusCode
					? monitor.ExpectedStatusCode.toString()
					: '200';
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
					: [];
			} else {
				// Create mode: defaults
				name = '';
				target = '';
				intervalSeconds = 60;
				timeoutSeconds = 10;
				expectedStatusCodeStr = '200';
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
						if (import.meta.env.DEV) {
							console.error('Failed to load members', err);
						}
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
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!canSubmit) return;

		loading = true;
		error = '';

		try {
			if (isEdit && monitor) {
				await MonitorAPI.update(orgId, monitor.ID, {
					name,
					target,
					interval_seconds: intervalSeconds,
					timeout_seconds: timeoutSeconds,
					expectation_status_code:
						effectiveType === 'http' ? parseInt(expectedStatusCodeStr) : undefined,
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
		class="sm:max-w-187.5 bg-card border border-border shadow-lg text-foreground p-6 rounded-2xl"
	>
		<Dialog.Header>
			<div class="flex items-center gap-2">
				<Dialog.Title class="text-xl font-semibold tracking-tight">{modalTitle}</Dialog.Title>
				{#if !isEdit}
					<MonitorTypeGuide monitorType={effectiveType as any} />
				{/if}
			</div>
			<Dialog.Description class="text-muted-foreground text-sm mt-1">
				{isEdit
					? 'Update your monitor configurations and alert policies below.'
					: 'Configure your new monitor settings and notification assignees below.'}
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-4"></div>

		<form onsubmit={handleSubmit} class="flex flex-col gap-5">
			{#if error}
				<div
					class="p-3 text-[13px] text-destructive bg-destructive/10 border border-destructive/20 rounded-lg font-medium"
				>
					{error}
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Left Column: Monitor Settings -->
				<div class="space-y-5">
					<!-- Name -->
					<div class="space-y-2">
						<div class="flex items-center gap-3">
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
							class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 {fieldErrors.name
								? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
								: ''} {shake.name ? 'animate-shake' : ''}"
						/>
					</div>

					<!-- Target -->
					<div class="space-y-2">
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
							class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 font-mono {fieldErrors.target
								? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
								: ''} {shake.target ? 'animate-shake' : ''}"
						/>
					</div>

					<!-- Expected Status Code (HTTP only) -->
					{#if effectiveType === 'http'}
						<div class="space-y-2">
							<Label class="text-sm font-medium text-foreground">Expected Status Code</Label>
							<Select.Root type="single" bind:value={expectedStatusCodeStr}>
								<Select.Trigger
									class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-10 w-full"
								>
									{expectedStatusCodeStr === '200'
										? '200 (OK)'
										: expectedStatusCodeStr === '201'
											? '201 (Created)'
											: expectedStatusCodeStr === '301'
												? '301 (Moved Permanently)'
												: expectedStatusCodeStr === '404'
													? '404 (Not Found)'
													: expectedStatusCodeStr}
								</Select.Trigger>
								<Select.Content class="bg-card border-border text-foreground">
									<Select.Item
										value="200"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>200 (OK)</Select.Item
									>
									<Select.Item
										value="201"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>201 (Created)</Select.Item
									>
									<Select.Item
										value="301"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>301 (Moved Permanently)</Select.Item
									>
									<Select.Item
										value="404"
										class="hover:bg-accent focus:bg-accent focus:text-foreground cursor-pointer"
										>404 (Not Found)</Select.Item
									>
								</Select.Content>
							</Select.Root>
						</div>
					{/if}

					<!-- Criticality -->
					<div class="space-y-2">
						<Label class="text-sm font-medium text-foreground">Criticality Level</Label>
						<Select.Root type="single" bind:value={criticality}>
							<Select.Trigger
								class="bg-background border-border text-foreground focus:ring-1 focus:ring-primary focus:border-primary rounded-lg h-10 w-full capitalize"
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

					<!-- Status Toggle (in Edit Mode) -->
					{#if isEdit}
						<label
							for="monitor-is-active"
							class="flex items-center justify-between p-3 rounded-lg bg-background border border-border mt-3 cursor-pointer select-none"
						>
							<div class="flex flex-col gap-0.5">
								<span class="text-xs font-semibold text-foreground">Monitor is active</span>
								<span class="text-[11px] text-muted-foreground">
									{isActive
										? 'Automated uptime checks and alerts are enabled'
										: 'Paused: Automated checks & alert notifications are paused'}
								</span>
							</div>
							<Checkbox id="monitor-is-active" bind:checked={isActive} />
						</label>
					{/if}

					<!-- SSL & Domain Expiry Toggle (HTTP monitors) -->
					{#if effectiveType === 'http'}
						<label
							for="monitor-ssl-check"
							class="flex items-center justify-between p-3 rounded-lg bg-background border border-border mt-3 cursor-pointer select-none"
						>
							<div class="flex flex-col gap-0.5">
								<span class="text-xs font-semibold text-foreground">SSL & Domain Expiry Check</span>
								<span class="text-[11px] text-muted-foreground"
									>Monitors SSL cert & WHOIS expiration daily</span
								>
							</div>
							<Checkbox id="monitor-ssl-check" bind:checked={enableSSLDomainCheck} />
						</label>
					{/if}

					<div class="grid grid-cols-2 gap-4">
						<!-- Interval -->
						<div class="space-y-2">
							<div class="flex items-center gap-3">
								<Label for="monitor-interval" class="text-sm font-medium text-foreground">
									Check Interval (sec) <span class="text-destructive font-bold">*</span>
								</Label>
								{#if fieldErrors.intervalSeconds}
									<span class="text-[10px] text-destructive font-medium"
										>{fieldErrors.intervalSeconds}</span
									>
								{/if}
							</div>
							<Input
								id="monitor-interval"
								type="number"
								min="60"
								bind:value={intervalSeconds}
								oninput={() => handleIntervalValidation(intervalSeconds)}
								required
								class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 font-mono [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none {fieldErrors.intervalSeconds
									? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
									: ''} {shake.intervalSeconds ? 'animate-shake' : ''}"
							/>
						</div>

						<!-- Timeout -->
						<div class="space-y-2">
							<div class="flex items-center gap-3">
								<Label for="monitor-timeout" class="text-sm font-medium text-foreground">
									Timeout (sec) <span class="text-destructive font-bold">*</span>
								</Label>
								{#if fieldErrors.timeoutSeconds}
									<span class="text-[10px] text-destructive font-medium"
										>{fieldErrors.timeoutSeconds}</span
									>
								{/if}
							</div>
							<Input
								id="monitor-timeout"
								type="number"
								min="1"
								max="60"
								bind:value={timeoutSeconds}
								oninput={() => handleTimeoutValidation(timeoutSeconds)}
								required
								class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10 font-mono [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none {fieldErrors.timeoutSeconds
									? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
									: ''} {shake.timeoutSeconds ? 'animate-shake' : ''}"
							/>
						</div>
					</div>

					<!-- Alert Template Profile -->
					<div class="space-y-1.5 pt-1">
						<Label
							for="monitor-template"
							class="text-sm font-medium text-foreground flex items-center justify-between"
						>
							<span>Alert Template</span>
							<span class="text-[11px] text-muted-foreground font-normal">Optional</span>
						</Label>
						<select
							id="monitor-template"
							bind:value={selectedTemplateId}
							class="w-full bg-background border border-border text-foreground rounded-lg h-10 px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary cursor-pointer"
						>
							<option value="">None (Use System Defaults)</option>
							{#each alertTemplates as tmpl (tmpl.id)}
								<option value={tmpl.id}>
									{tmpl.name} ({tmpl.rules?.length || 0} rules) {tmpl.is_default ? '★ Default' : ''}
								</option>
							{/each}
						</select>
					</div>
				</div>

				<!-- Right Column: Assignees Settings -->
				<div class="space-y-3 flex flex-col h-full">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-1.5">
							<Label class="text-sm font-medium text-foreground">Assignees</Label>
							<AssigneeGuide />
						</div>
						<span class="text-xs text-foreground/40 font-mono"
							>{selectedAssignees.length}/5 selected</span
						>
					</div>

					<div
						data-lenis-prevent
						class="flex-1 min-h-62.5 overflow-y-auto bg-background border border-border rounded-lg p-2 space-y-1"
					>
						{#each members as member (member.User.ID || member.ID)}
							{@const isSelected = selectedAssignees.includes(member.User.ID)}
							{@const isDisabled = !isSelected && selectedAssignees.length >= 5}

							<label
								class="flex items-center justify-between p-2 hover:bg-white/5 rounded-md cursor-pointer transition-colors {isDisabled
									? 'opacity-50 cursor-not-allowed'
									: ''}"
							>
								<div class="flex items-center gap-3 overflow-hidden">
									{#if member.User.AvatarURL}
										<img
											src={member.User.AvatarURL}
											alt={member.User.Name}
											class="size-8 rounded-full object-cover shrink-0 border border-border"
										/>
									{:else}
										<div
											class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 border {avatarColor(
												member.User.Name
											)}"
										>
											{member.User.Name?.charAt(0).toUpperCase() || '?'}
										</div>
									{/if}

									<span class="text-[13px] font-medium text-foreground truncate"
										>{member.User.Name}</span
									>
								</div>

								<input
									type="checkbox"
									class="sr-only"
									checked={isSelected}
									disabled={isDisabled}
									onchange={(e) => {
										if (e.currentTarget.checked) {
											if (selectedAssignees.length < 5) {
												selectedAssignees = [...selectedAssignees, member.User.ID];
											} else {
												e.currentTarget.checked = false;
											}
										} else {
											selectedAssignees = selectedAssignees.filter((id) => id !== member.User.ID);
										}
									}}
								/>
								<div
									class="w-5 h-5 rounded-full border flex items-center justify-center shrink-0 transition-all {isSelected
										? 'bg-primary border-primary'
										: 'bg-white/5 border-border'} {isDisabled ? 'opacity-50' : ''}"
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

					{#if selectedAssignees.length === 5}
						<span class="text-[11px] text-primary font-medium block"
							>Maximum 5 assignees selected.</span
						>
					{:else if selectedAssignees.length === 0}
						<span class="text-[11px] text-muted-foreground/60 block"
							>No assignees selected. Team notification channels will be used.</span
						>
					{/if}
				</div>
			</div>

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-6 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-border"
			>
				<button
					type="button"
					onclick={() => (isOpen = false)}
					class="px-5 py-2.5 rounded-lg text-sm font-medium text-foreground hover:bg-accent transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={loading || !canSubmit}
					class="h-11 px-6 rounded-lg text-sm font-medium bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 shadow-sm disabled:opacity-50 cursor-pointer"
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
