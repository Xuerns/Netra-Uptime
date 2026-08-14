<script lang="ts">
	import { StatusPageAPI } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Switch } from '$lib/components/ui/switch';
	import { toast } from 'svelte-sonner';
	import { TriangleAlert, Activity } from 'lucide-svelte';

	let {
		isOpen = $bindable(false),
		orgId,
		monitors,
		onSuccess
	} = $props<{
		isOpen: boolean;
		orgId: string;
		monitors: Monitor[];
		onSuccess: () => void;
	}>();

	let title = $state('');
	let subdomain = $state('');
	let description = $state('');
	let isPublic = $state(true);
	let selectedMonitors = $state<string[]>([]);

	let loading = $state(false);
	let error = $state('');

	function toggleMonitor(monitorId: string) {
		if (selectedMonitors.includes(monitorId)) {
			selectedMonitors = selectedMonitors.filter((id) => id !== monitorId);
		} else {
			selectedMonitors = [...selectedMonitors, monitorId];
		}
	}

	$effect(() => {
		if (isOpen) {
			title = '';
			subdomain = '';
			description = '';
			isPublic = true;
			selectedMonitors = [];
			error = '';
		}
	});

	async function handleSubmit() {
		if (!title.trim() || !subdomain.trim()) {
			error = 'Title and Subdomain are required.';
			return;
		}

		// Simple regex for valid subdomains
		const subdomainRegex = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
		if (!subdomainRegex.test(subdomain)) {
			error = 'Invalid subdomain format (letters, numbers, hyphens only).';
			return;
		}

		loading = true;
		error = '';

		try {
			await StatusPageAPI.create(orgId, {
				title: title.trim(),
				subdomain: subdomain.trim(),
				description: description.trim() || null,
				is_public: isPublic,
				monitor_id: selectedMonitors
			});
			toast.success('Status Page created successfully');
			isOpen = false;
			onSuccess();
		} catch (err: any) {
			error = err.message || 'Failed to create status page';
			toast.error(error);
		} finally {
			loading = false;
		}
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[750px] bg-[rgba(255,255,255,0.06)]! backdrop-blur-[20px] bg-gradient-to-b from-transparent to-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.08)] shadow-[0_0_20px_rgba(78,222,163,0.05)] text-white p-6 rounded-2xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight">Create Status Page</Dialog.Title>
			<Dialog.Description class="text-[#bbcabf] text-sm">
				Configure a new public status page for your organization.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-white/10 my-4"></div>

		<form
			class="flex flex-col gap-5"
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
		>
			{#if error}
				<div
					class="p-3 text-[13px] text-[#690005] bg-[#ffb4ab] border border-[#93000a] rounded-lg font-medium flex items-center gap-2"
				>
					<TriangleAlert class="size-4 shrink-0" />
					<p class="m-0">{error}</p>
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Left Column: Settings -->
				<div class="space-y-5">
					<div class="space-y-2">
						<Label for="title" class="text-sm font-medium text-[#e5e2e3]"
							>Title <span class="text-error font-bold">*</span></Label
						>
						<Input
							id="title"
							bind:value={title}
							placeholder="e.g. Acme Status"
							required
							class="bg-[rgba(0,0,0,0.2)] border-white/10 text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] rounded-lg h-10"
						/>
					</div>

					<div class="space-y-2">
						<Label for="description" class="text-sm font-medium text-[#e5e2e3]"
							>Description <span class="text-white/40 text-xs font-normal">(Optional)</span></Label
						>
						<textarea
							id="description"
							bind:value={description}
							placeholder="e.g. Current status of all our core systems."
							class="flex min-h-[80px] w-full rounded-lg border border-white/10 bg-[rgba(0,0,0,0.2)] px-3 py-2 text-sm text-white placeholder:text-white/30 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] disabled:cursor-not-allowed disabled:opacity-50"
						></textarea>
					</div>

					<div class="space-y-2">
						<Label for="subdomain" class="text-sm font-medium text-[#e5e2e3]"
							>Subdomain <span class="text-error font-bold">*</span></Label
						>
						<div class="flex items-center">
							<Input
								id="subdomain"
								bind:value={subdomain}
								placeholder="acme"
								required
								class="bg-[rgba(0,0,0,0.2)] border-white/10 border-r-0 rounded-r-none text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-[#3B82F6] focus-visible:border-[#3B82F6] h-10 z-10"
							/>
							<div
								class="bg-[rgba(0,0,0,0.4)] border border-white/10 border-l-0 px-3 h-10 flex items-center rounded-r-lg text-white/50 text-sm font-mono whitespace-nowrap"
							>
								.pulsewatch.com
							</div>
						</div>
					</div>

					<div
						class="flex items-center justify-between p-4 bg-[rgba(0,0,0,0.2)] border border-white/10 rounded-lg"
					>
						<div class="space-y-0.5">
							<Label class="text-sm font-medium text-[#e5e2e3]">Public Visibility</Label>
							<p class="text-xs text-white/40">Anyone with the link can view this page.</p>
						</div>
						<Switch bind:checked={isPublic} />
					</div>
				</div>

				<!-- Right Column: Select Monitors -->
				<div class="space-y-3 flex flex-col h-full">
					<div class="flex items-center justify-between">
						<Label class="text-sm font-medium text-[#e5e2e3]">Select Monitors</Label>
						<span class="text-xs text-white/40">Choose monitors to display</span>
					</div>

					<div
						class="flex-1 min-h-[250px] max-h-[300px] overflow-y-auto bg-[rgba(0,0,0,0.2)] border border-white/10 rounded-lg p-2 space-y-1"
					>
						{#if monitors.length === 0}
							<div class="p-6 text-center text-sm text-white/40">
								No monitors available. Add monitors to your organization first.
							</div>
						{:else}
							{#each monitors as monitor}
								{@const isSelected = selectedMonitors.includes(monitor.ID)}
								<label
									data-lenis-prevent
									class="flex items-center gap-3 p-2 hover:bg-white/5 rounded-md cursor-pointer transition-colors"
								>
									<!-- Icon for Monitor -->
									<div
										class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 border bg-white/5 text-white/80 border-white/10"
									>
										<Activity class="size-4" />
									</div>

									<div class="flex flex-col flex-1 min-w-0">
										<span class="text-[13px] font-medium text-white truncate">{monitor.Name}</span>
										<span class="text-[11px] text-white/40 font-mono truncate"
											>{monitor.Target}</span
										>
									</div>

									<input
										type="checkbox"
										checked={isSelected}
										onchange={() => toggleMonitor(monitor.ID)}
										class="sr-only"
									/>
									<div
										class="w-5 h-5 rounded-full border flex items-center justify-center shrink-0 transition-all {isSelected
											? 'bg-primary border-primary'
											: 'bg-white/5 border-white/20'}"
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
							{/each}
						{/if}
					</div>
				</div>
			</div>

			<Dialog.Footer
				class="gap-2 sm:justify-end border-t border-white/10 py-4 pb-0 bg-transparent! mt-4"
			>
				<Button
					type="button"
					variant="ghost"
					class="text-white h-11 px-6 hover:text-white hover:bg-white/10"
					onclick={() => (isOpen = false)}
					disabled={loading}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={loading}
					class="bg-[#1A1A1B] border border-[#2A2A2B] text-primary hover:bg-primary hover:border-primary hover:text-white active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
				>
					{loading ? 'Creating...' : 'Create Status Page'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
