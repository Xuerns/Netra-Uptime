<script lang="ts">
	import { StatusPageAPI, type StatusPage, type StatusPageMonitor } from '$lib/api/statuspage';
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
		statusPage,
		onSuccess
	} = $props<{
		isOpen: boolean;
		orgId: string;
		monitors: Monitor[];
		statusPage: StatusPage | null;
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

	async function loadCurrentMonitors(spId: string) {
		try {
			const res = await StatusPageAPI.getMonitors(orgId, spId);
			const spMonitors = res.data ?? [];
			selectedMonitors = spMonitors.map((spm: any) => spm.MonitorID);
		} catch (err: any) {
			error = 'Failed to load assigned monitors.';
			console.error(err);
		}
	}

	$effect(() => {
		if (isOpen && statusPage) {
			title = statusPage.Title;
			subdomain = statusPage.Subdomain;
			description = statusPage.Description || '';
			isPublic = statusPage.IsPublic;
			error = '';
			loadCurrentMonitors(statusPage.ID);
		}
	});

	async function handleSubmit() {
		if (!title.trim() || !subdomain.trim() || !statusPage) {
			error = 'Title and Subdomain are required.';
			return;
		}

		const subdomainRegex = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
		if (!subdomainRegex.test(subdomain)) {
			error = 'Invalid subdomain format (letters, numbers, hyphens only).';
			return;
		}

		loading = true;
		error = '';

		try {
			await StatusPageAPI.update(orgId, statusPage.ID, {
				title: title.trim(),
				subdomain: subdomain.trim(),
				description: description.trim() || null,
				is_public: isPublic,
				monitor_id: selectedMonitors
			});
			toast.success('Status Page updated successfully');
			isOpen = false;
			onSuccess();
		} catch (err: any) {
			error = err.message || 'Failed to update status page';
			toast.error(error);
		} finally {
			loading = false;
		}
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[750px] bg-card border border-border shadow-lg text-foreground p-6 rounded-2xl"
	>
		<Dialog.Header>
			<Dialog.Title class="text-xl font-semibold tracking-tight">Edit Status Page</Dialog.Title>
			<Dialog.Description class="text-muted-foreground text-sm">
				Update your public status page details.
			</Dialog.Description>
		</Dialog.Header>

		<div class="h-px w-full bg-border my-4"></div>

		<form
			class="flex flex-col gap-5"
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
		>
			{#if error}
				<div
					class="p-3 text-[13px] text-destructive bg-destructive/10 border border-destructive/20 rounded-lg font-medium flex items-center gap-2"
				>
					<TriangleAlert class="size-4 shrink-0" />
					<p class="m-0">{error}</p>
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Left Column: Settings -->
				<div class="space-y-5">
					<div class="space-y-2">
						<Label for="edit-title" class="text-sm font-medium text-foreground"
							>Title <span class="text-destructive font-bold">*</span></Label
						>
						<Input
							id="edit-title"
							bind:value={title}
							placeholder="e.g. Acme Status"
							required
							class="bg-background border-border text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary rounded-lg h-10"
						/>
					</div>

					<div class="space-y-2">
						<Label for="edit-description" class="text-sm font-medium text-foreground"
							>Description <span class="text-foreground/40 text-xs font-normal">(Optional)</span
							></Label
						>
						<textarea
							id="edit-description"
							bind:value={description}
							placeholder="e.g. Current status of all our core systems."
							class="flex min-h-[80px] w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50"
						></textarea>
					</div>

					<div class="space-y-2">
						<Label for="edit-subdomain" class="text-sm font-medium text-foreground"
							>Subdomain <span class="text-destructive font-bold">*</span></Label
						>
						<div class="flex items-center">
							<Input
								id="edit-subdomain"
								bind:value={subdomain}
								placeholder="acme"
								required
								class="bg-background border-border border-r-0 rounded-r-none text-foreground placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-primary focus-visible:border-primary h-10 z-10"
							/>
							<div
								class="bg-background/80 border border-border border-l-0 px-3 h-10 flex items-center rounded-r-lg text-foreground/50 text-sm font-mono whitespace-nowrap"
							>
								.pulsewatch.com
							</div>
						</div>
					</div>

					<div
						class="flex items-center justify-between p-4 bg-background border border-border rounded-lg"
					>
						<div class="space-y-0.5">
							<Label class="text-sm font-medium text-foreground">Public Visibility</Label>
							<p class="text-xs text-foreground/40">Anyone with the link can view this page.</p>
						</div>
						<Switch bind:checked={isPublic} />
					</div>
				</div>

				<!-- Right Column: Select Monitors -->
				<div class="space-y-3 flex flex-col h-full">
					<div class="flex items-center justify-between">
						<Label class="text-sm font-medium text-foreground">Select Monitors</Label>
						<span class="text-xs text-foreground/40">Choose monitors to display</span>
					</div>

					<div
						class="flex-1 min-h-[250px] max-h-[300px] overflow-y-auto bg-background border border-border rounded-lg p-2 space-y-1"
					>
						{#if monitors.length === 0}
							<div class="p-6 text-center text-sm text-foreground/40">
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
										class="size-8 rounded-full flex items-center justify-center font-semibold text-[12.5px] shrink-0 border bg-white/5 text-foreground/80 border-border"
									>
										<Activity class="size-4" />
									</div>

									<div class="flex flex-col flex-1 min-w-0">
										<span class="text-[13px] font-medium text-foreground truncate"
											>{monitor.Name}</span
										>
										<span class="text-[11px] text-foreground/40 font-mono truncate"
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
							{/each}
						{/if}
					</div>
				</div>
			</div>

			<Dialog.Footer
				class="gap-2 sm:justify-end border-t border-border py-4 pb-0 bg-transparent! mt-4"
			>
				<Button
					type="button"
					variant="ghost"
					class="text-foreground h-11 px-6 hover:text-foreground hover:bg-accent"
					onclick={() => (isOpen = false)}
					disabled={loading}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={loading}
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-foreground active:scale-[0.98] transition-all duration-200 h-11 px-6 font-medium shadow-sm"
				>
					{loading ? 'Saving...' : 'Save Changes'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
