<script lang="ts">
	import { MonitorAPI } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Select from '$lib/components/ui/select';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		debounce,
		validateMonitorName,
		validateMonitorTarget,
		validateMonitorInterval,
		validateMonitorTimeout
	} from '$lib/utils/validation';

	interface Monitor {
		ID: string;
		Name: string;
		Type: string;
		Target: string;
		IntervalSeconds: number;
		TimeoutSeconds: number;
		ExpectedStatusCode?: number;
		IsActive: boolean;
		Criticality?: 'extreme' | 'high' | 'medium' | 'low';
	}

	let {
		isOpen = $bindable(false),
		monitor = null,
		orgId,
		onSuccess
	} = $props<{
		isOpen: boolean;
		monitor: Monitor | null;
		orgId: string;
		onSuccess: () => void;
	}>();

	let name = $state('');
	let target = $state('');
	let intervalSeconds = $state<number>(60);
	let timeoutSeconds = $state<number>(10);
	let expectedStatusCodeStr = $state<string>('200');
	let isActive = $state<boolean>(true);
	let criticality = $state<'extreme' | 'high' | 'medium' | 'low'>('medium');
	let enableSSLDomainCheck = $state<boolean>(true);
	let loading = $state(false);
	let error = $state('');

	let fieldErrors = $state({ name: '', target: '', intervalSeconds: '', timeoutSeconds: '' });
	let shake = $state({ name: false, target: false, intervalSeconds: false, timeoutSeconds: false });

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

	let members = $state<any[]>([]);
	let selectedPICs = $state<string[]>([]);
	import { OrgAPI } from '$lib/api/organization';

	let canSubmit = $derived(
		!loading &&
			!fieldErrors.name &&
			!fieldErrors.target &&
			!fieldErrors.intervalSeconds &&
			!fieldErrors.timeoutSeconds &&
			selectedPICs.length > 0
	);

	$effect(() => {
		if (isOpen && monitor) {
			name = monitor.Name;
			target = monitor.Target;
			intervalSeconds = monitor.IntervalSeconds;
			timeoutSeconds = monitor.TimeoutSeconds;
			expectedStatusCodeStr = monitor.ExpectedStatusCode
				? monitor.ExpectedStatusCode.toString()
				: '200';
			isActive = monitor.IsActive !== undefined ? monitor.IsActive : true;
			criticality = monitor.Criticality || 'medium';
			enableSSLDomainCheck =
				monitor.EnableSSLDomainCheck !== undefined ? monitor.EnableSSLDomainCheck : true;
			error = '';
			fieldErrors = { name: '', target: '', intervalSeconds: '', timeoutSeconds: '' };
			shake = { name: false, target: false, intervalSeconds: false, timeoutSeconds: false };

			// Extract existing PICs
			selectedPICs = monitor.MonitorPICs ? monitor.MonitorPICs.map((pic: any) => pic.UserID) : [];

			if (orgId) {
				OrgAPI.getMembers(orgId)
					.then((res) => {
						members = res.data ?? [];
					})
					.catch((err) => console.error('Failed to load members', err));
			}
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!canSubmit) return;
		loading = true;
		error = '';

		try {
			if (!monitor) throw new Error('No monitor selected for editing');

			await MonitorAPI.update(orgId, monitor.ID, {
				name,
				target,
				interval_seconds: intervalSeconds,
				timeout_seconds: timeoutSeconds,
				expectation_status_code:
					monitor.Type === 'http' ? parseInt(expectedStatusCodeStr) : undefined,
				is_active: isActive,
				criticality,
				MonitorPICs: selectedPICs,
				enable_ssl_domain_check: enableSSLDomainCheck
			});
			onSuccess();
			isOpen = false;
		} catch (err: any) {
			error = err.message || 'Failed to update monitor';
		} finally {
			loading = false;
		}
	}

	function avatarColor(name: string) {
		const colors = [
			'bg-red-500/10 text-red-500 border-red-500/20',
			'bg-blue-500/10 text-blue-500 border-blue-500/20',
			'bg-green-500/10 text-green-500 border-green-500/20',
			'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
			'bg-purple-500/10 text-purple-500 border-purple-500/20'
		];
		let hash = 0;
		if (!name) return colors[0];
		for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
		return colors[Math.abs(hash) % colors.length];
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[750px] bg-[rgba(255,255,255,0.06)]! backdrop-blur-[20px] bg-gradient-to-b from-transparent to-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)] shadow-[0_0_20px_rgba(78,222,163,0.05)] text-white p-6 rounded-2xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight">Edit Monitor</Dialog.Title>
			<Dialog.Description class="text-[#bbcabf] text-sm">
				Update your monitor settings below.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-white/10 my-4"></div>

		<form onsubmit={handleSubmit} class="flex flex-col gap-5">
			{#if error}
				<div
					class="p-3 text-[13px] text-[#690005] bg-[#ffb4ab] border border-[#93000a] rounded-lg font-medium"
				>
					{error}
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Left Column: Monitor Settings -->
				<div class="space-y-5">
					<div class="space-y-2">
						<div class="flex items-center gap-3">
							<Label for="edit-name" class="text-sm font-medium text-[#e5e2e3]"
								>Monitor Name <span class="text-error font-bold">*</span></Label
							>
							{#if fieldErrors.name}
								<span class="text-xs text-error font-medium">{fieldErrors.name}</span>
							{/if}
						</div>
						<Input
							id="edit-name"
							placeholder="e.g., Production API"
							bind:value={name}
							oninput={() => handleNameValidation(name)}
							required
							class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 {fieldErrors.name
								? 'border-error/70 focus-visible:border-error focus-visible:ring-error'
								: ''} {shake.name ? 'animate-shake' : ''}"
						/>
					</div>

					<div class="space-y-2">
						<div class="flex items-center gap-3">
							<Label for="edit-target" class="text-sm font-medium text-[#e5e2e3]"
								>Target / URL <span class="text-error font-bold">*</span></Label
							>
							{#if fieldErrors.target}
								<span class="text-xs text-error font-medium">{fieldErrors.target}</span>
							{/if}
						</div>
						<Input
							id="edit-target"
							bind:value={target}
							oninput={() => handleTargetValidation(target)}
							required
							class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 font-mono {fieldErrors.target
								? 'border-error/70 focus-visible:border-error focus-visible:ring-error'
								: ''} {shake.target ? 'animate-shake' : ''}"
						/>
					</div>

					{#if monitor?.Type === 'http'}
						<div class="space-y-2">
							<Label class="text-sm font-medium text-[#e5e2e3]">Expected Status Code</Label>
							<Select.Root type="single" bind:value={expectedStatusCodeStr}>
								<Select.Trigger
									class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white focus:ring-1 focus:ring-[#3B82F6] focus:border-[#3B82F6] rounded-lg h-10 w-full"
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
								<Select.Content class="bg-[#161618] border-white/10 text-white">
									<Select.Item
										value="200"
										class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
										>200 (OK)</Select.Item
									>
									<Select.Item
										value="201"
										class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
										>201 (Created)</Select.Item
									>
									<Select.Item
										value="301"
										class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
										>301 (Moved Permanently)</Select.Item
									>
									<Select.Item
										value="404"
										class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
										>404 (Not Found)</Select.Item
									>
								</Select.Content>
							</Select.Root>
						</div>
					{/if}

					<!-- Criticality -->
					<div class="space-y-2">
						<Label class="text-sm font-medium text-[#e5e2e3]">Criticality Level</Label>
						<Select.Root type="single" bind:value={criticality}>
							<Select.Trigger
								class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white focus:ring-1 focus:ring-[#3B82F6] focus:border-[#3B82F6] rounded-lg h-10 w-full capitalize"
							>
								{criticality}
							</Select.Trigger>
							<Select.Content class="bg-[#161618] border-white/10 text-white">
								<Select.Item
									value="extreme"
									class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
									>Extreme</Select.Item
								>
								<Select.Item
									value="high"
									class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
									>High</Select.Item
								>
								<Select.Item
									value="medium"
									class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
									>Medium</Select.Item
								>
								<Select.Item
									value="low"
									class="hover:bg-white/10 focus:bg-white/10 focus:text-white cursor-pointer"
									>Low</Select.Item
								>
							</Select.Content>
						</Select.Root>
					</div>

					<!-- SSL & Domain Expiry Toggle (HTTP monitors) -->
					{#if monitor && monitor.Type === 'http'}
						<div
							class="flex items-center justify-between p-3 rounded-lg bg-white/4 border border-white/8 mt-3"
						>
							<div class="flex flex-col gap-0.5">
								<span class="text-xs font-semibold text-white">SSL & Domain Expiry Check</span>
								<span class="text-[11px] text-on-surface-variant"
									>Monitors SSL cert & WHOIS expiration daily</span
								>
							</div>
							<input
								type="checkbox"
								bind:checked={enableSSLDomainCheck}
								class="size-4 accent-primary rounded cursor-pointer"
							/>
						</div>
					{/if}

					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-2">
							<div class="flex items-center gap-3">
								<Label for="edit-interval" class="text-sm font-medium text-[#e5e2e3]"
									>Check Interval (sec) <span class="text-error font-bold">*</span></Label
								>
								{#if fieldErrors.intervalSeconds}
									<span class="text-[10px] text-error font-medium"
										>{fieldErrors.intervalSeconds}</span
									>
								{/if}
							</div>
							<Input
								id="edit-interval"
								type="number"
								min="10"
								bind:value={intervalSeconds}
								oninput={() => handleIntervalValidation(intervalSeconds)}
								required
								class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 font-mono [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none {fieldErrors.intervalSeconds
									? 'border-error/70 focus-visible:border-error focus-visible:ring-error'
									: ''} {shake.intervalSeconds ? 'animate-shake' : ''}"
							/>
						</div>

						<div class="space-y-2">
							<div class="flex items-center gap-3">
								<Label for="edit-timeout" class="text-sm font-medium text-[#e5e2e3]"
									>Timeout (sec) <span class="text-error font-bold">*</span></Label
								>
								{#if fieldErrors.timeoutSeconds}
									<span class="text-[10px] text-error font-medium"
										>{fieldErrors.timeoutSeconds}</span
									>
								{/if}
							</div>
							<Input
								id="edit-timeout"
								type="number"
								min="1"
								max="60"
								bind:value={timeoutSeconds}
								oninput={() => handleTimeoutValidation(timeoutSeconds)}
								required
								class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10 font-mono [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none {fieldErrors.timeoutSeconds
									? 'border-error/70 focus-visible:border-error focus-visible:ring-error'
									: ''} {shake.timeoutSeconds ? 'animate-shake' : ''}"
							/>
						</div>
					</div>
				</div>

				<!-- Right Column: PIC Settings -->
				<div class="space-y-3 flex flex-col h-full">
					<div class="flex items-center justify-between">
						<Label class="text-sm font-medium text-[#e5e2e3]"
							>Person in Charge (PIC) <span class="text-error font-bold">*</span></Label
						>
						<span class="text-xs text-white/40">Select 1-2 PICs</span>
					</div>

					<div
						data-lenis-prevent
						class="flex-1 min-h-[250px] overflow-y-auto bg-[rgba(0,0,0,0.2)] border border-white/10 rounded-lg p-2 space-y-1"
					>
						{#each members as member}
							{@const isSelected = selectedPICs.includes(member.User.ID)}
							{@const isDisabled = !isSelected && selectedPICs.length >= 2}

							<label
								class="flex items-center justify-between p-2 hover:bg-white/5 rounded-md cursor-pointer transition-colors {isDisabled
									? 'opacity-50 cursor-not-allowed'
									: ''}"
							>
								<div class="flex items-center gap-3 overflow-hidden">
									<!-- Avatar -->
									{#if member.User.AvatarURL}
										<img
											src={member.User.AvatarURL}
											alt={member.User.Name}
											class="size-8 rounded-full object-cover shrink-0 border border-white/10"
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

									<span class="text-[13px] font-medium text-white truncate">{member.User.Name}</span
									>
								</div>

								<input
									type="checkbox"
									class="sr-only"
									checked={isSelected}
									disabled={isDisabled}
									onchange={(e) => {
										if (e.currentTarget.checked) {
											if (selectedPICs.length < 2) {
												selectedPICs = [...selectedPICs, member.User.ID];
											} else {
												e.currentTarget.checked = false; // prevent checking more than 2
											}
										} else {
											selectedPICs = selectedPICs.filter((id) => id !== member.User.ID);
										}
									}}
								/>
								<div
									class="w-5 h-5 rounded-full border flex items-center justify-center shrink-0 transition-all {isSelected
										? 'bg-primary border-primary'
										: 'bg-white/5 border-white/20'} {isDisabled ? 'opacity-50' : ''}"
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
								class="p-4 text-xs text-white/40 text-center flex flex-col items-center justify-center h-full"
							>
								Loading members...
							</div>
						{/each}
					</div>

					{#if selectedPICs.length === 0}
						<span class="text-[11px] text-error font-medium block"
							>At least 1 Person in Charge is required.</span
						>
					{:else if selectedPICs.length === 2}
						<span class="text-[11px] text-[#4edea3] font-medium block"
							>Maximum 2 PICs selected.</span
						>
					{/if}
				</div>
			</div>

			<Dialog.Footer
				class="bg-transparent! m-0! p-0! pt-6 mt-2 flex justify-end gap-3 sm:space-x-0 border-t border-white/5"
			>
				<button
					type="button"
					onclick={() => (isOpen = false)}
					class="px-5 py-2.5 rounded-lg text-sm font-medium text-white hover:bg-white/10 transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={loading || !canSubmit}
					class="h-11 px-6 rounded-lg text-sm font-medium bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 shadow-sm disabled:opacity-50"
				>
					{loading ? 'Saving...' : 'Save Changes'}
				</button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
