<script lang="ts">
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { toast } from 'svelte-sonner';
	import {
		TriangleAlert,
		Activity,
		Palette,
		Settings,
		Image,
		Sparkles,
		Code2,
		Check
	} from 'lucide-svelte';
	import { parseApiError } from '$lib/utils/parseApiError';

	let {
		isOpen = $bindable(false),
		orgId,
		monitors,
		statusPage = null,
		onSuccess
	} = $props<{
		isOpen: boolean;
		orgId: string;
		monitors: Monitor[];
		statusPage?: StatusPage | null;
		onSuccess: () => void;
	}>();

	const isEdit = $derived(!!statusPage);

	// Active tab: 'general' | 'branding'
	let activeTab = $state<'general' | 'branding'>('general');

	// General form states
	let title = $state('');
	let subdomain = $state('');
	let description = $state('');
	let isPublic = $state(true);
	let selectedMonitors = $state<string[]>([]);

	// Branding form states
	let logoUrl = $state('');
	let faviconUrl = $state('');
	let themeColor = $state('#4edea3');
	let customCss = $state('');

	let loading = $state(false);
	let error = $state('');

	const colorPresets = [
		{ name: 'Emerald', hex: '#4edea3' },
		{ name: 'Sky Blue', hex: '#3b82f6' },
		{ name: 'Purple', hex: '#8b5cf6' },
		{ name: 'Indigo', hex: '#6366f1' },
		{ name: 'Rose', hex: '#ec4899' },
		{ name: 'Amber', hex: '#f59e0b' },
		{ name: 'Teal', hex: '#14b8a6' }
	];

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
			error = parseApiError(err, 'Failed to load assigned monitors.');
		}
	}

	$effect(() => {
		if (isOpen) {
			activeTab = 'general';
			error = '';

			if (statusPage) {
				title = statusPage.Title;
				subdomain = statusPage.Subdomain;
				description = statusPage.Description || '';
				isPublic = statusPage.IsPublic;
				logoUrl = statusPage.LogoURL || '';
				faviconUrl = statusPage.FaviconURL || '';
				themeColor = statusPage.ThemeColor || '#4edea3';
				customCss = statusPage.CustomCSS || '';
				loadCurrentMonitors(statusPage.ID);
			} else {
				title = '';
				subdomain = '';
				description = '';
				isPublic = true;
				selectedMonitors = [];
				logoUrl = '';
				faviconUrl = '';
				themeColor = '#4edea3';
				customCss = '';
			}
		}
	});

	async function handleSubmit() {
		if (!title.trim() || !subdomain.trim()) {
			error = 'Title and Subdomain are required.';
			return;
		}

		const subdomainRegex = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
		if (!subdomainRegex.test(subdomain)) {
			error = 'Invalid subdomain format (letters, numbers, hyphens only).';
			return;
		}

		if (!isEdit && selectedMonitors.length === 0) {
			error = 'Please select at least one monitor.';
			return;
		}

		loading = true;
		error = '';

		try {
			if (isEdit && statusPage) {
				await StatusPageAPI.update(orgId, statusPage.ID, {
					title: title.trim(),
					subdomain: subdomain.trim(),
					description: description.trim() || null,
					is_public: isPublic,
					logo_url: logoUrl.trim() || '',
					favicon_url: faviconUrl.trim() || '',
					theme_color: themeColor.trim() || '#4edea3',
					custom_css: customCss.trim() || '',
					monitor_id: selectedMonitors
				});
				toast.success('Status Page updated successfully');
			} else {
				await StatusPageAPI.create(orgId, {
					title: title.trim(),
					subdomain: subdomain.trim(),
					description: description.trim() || null,
					is_public: isPublic,
					logo_url: logoUrl.trim() || '',
					favicon_url: faviconUrl.trim() || '',
					theme_color: themeColor.trim() || '#4edea3',
					custom_css: customCss.trim() || '',
					monitor_id: selectedMonitors
				});
				toast.success('Status Page created successfully');
			}

			isOpen = false;
			onSuccess();
		} catch (err: any) {
			error = parseApiError(
				err,
				isEdit ? 'Failed to update status page' : 'Failed to create status page'
			);
			toast.error(error);
		} finally {
			loading = false;
		}
	}
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		class="sm:max-w-[780px] max-h-[90vh] flex flex-col bg-card border border-border shadow-2xl text-foreground p-0 rounded-2xl overflow-hidden"
	>
		<!-- Header -->
		<div class="p-6 pb-4 border-b border-border/70 shrink-0">
			<Dialog.Header>
				<Dialog.Title class="text-xl font-bold tracking-tight">
					{isEdit ? 'Edit Status Page' : 'Create Status Page'}
				</Dialog.Title>
				<Dialog.Description class="text-sm text-muted-foreground mt-1">
					{isEdit
						? 'Update your public-facing status page branding and monitor configurations.'
						: 'Publish a clean, transparent public status page to communicate system reliability.'}
				</Dialog.Description>
			</Dialog.Header>

			<!-- Tabs Navigation -->
			<div class="flex gap-2 mt-5 border-b border-border/50">
				<button
					type="button"
					onclick={() => (activeTab = 'general')}
					class="flex items-center gap-2 pb-2.5 px-3 text-xs font-semibold tracking-wide uppercase transition-colors relative cursor-pointer {activeTab ===
					'general'
						? 'text-primary border-b-2 border-primary'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<Settings class="size-3.5" />
					General & Monitors
				</button>

				<button
					type="button"
					onclick={() => (activeTab = 'branding')}
					class="flex items-center gap-2 pb-2.5 px-3 text-xs font-semibold tracking-wide uppercase transition-colors relative cursor-pointer {activeTab ===
					'branding'
						? 'text-primary border-b-2 border-primary'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<Palette class="size-3.5" />
					Branding & Styling
				</button>
			</div>
		</div>

		<!-- Form Body (Scrollable) -->
		<div class="p-6 flex-1 overflow-y-auto space-y-6">
			{#if error}
				<div
					class="flex items-center gap-3 p-3.5 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-xs font-medium"
				>
					<TriangleAlert class="size-4 shrink-0" />
					<span>{error}</span>
				</div>
			{/if}

			{#if activeTab === 'general'}
				<!-- Tab 1: General Info & Monitors -->
				<div class="space-y-5">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<!-- Title -->
						<div class="space-y-1.5">
							<Label for="sp-title" class="text-xs font-medium text-foreground">
								Page Title <span class="text-destructive">*</span>
							</Label>
							<Input
								id="sp-title"
								bind:value={title}
								placeholder="e.g. Acme Corp System Status"
								class="bg-background border-border text-foreground h-10 rounded-lg text-sm"
								required
							/>
						</div>

						<!-- Subdomain -->
						<div class="space-y-1.5">
							<Label for="sp-subdomain" class="text-xs font-medium text-foreground">
								Subdomain Slug <span class="text-destructive">*</span>
							</Label>
							<div class="flex items-center">
								<Input
									id="sp-subdomain"
									bind:value={subdomain}
									placeholder="acme"
									class="bg-background border-border text-foreground h-10 rounded-r-none rounded-l-lg text-sm font-mono lowercase"
									required
								/>
								<span
									class="h-10 px-3 bg-muted border border-l-0 border-border rounded-r-lg text-xs font-mono text-muted-foreground flex items-center shrink-0"
								>
									.netrauptime.com
								</span>
							</div>
						</div>
					</div>

					<!-- Description -->
					<div class="space-y-1.5">
						<Label for="sp-desc" class="text-xs font-medium text-foreground">Description</Label>
						<textarea
							id="sp-desc"
							bind:value={description}
							rows="2"
							placeholder="Optional short description about this status page or team..."
							class="w-full bg-background border border-border rounded-lg p-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary resize-none transition-all"
						></textarea>
					</div>

					<!-- Public Switch -->
					<div
						class="flex items-center justify-between p-4 rounded-xl bg-white/2 border border-border/80"
					>
						<div class="space-y-0.5">
							<div class="text-sm font-medium text-foreground">Public Availability</div>
							<div class="text-xs text-muted-foreground">
								If enabled, anyone with the subdomain link can view system operational status.
							</div>
						</div>
						<Switch bind:checked={isPublic} />
					</div>

					<!-- Monitors Selector -->
					<div class="space-y-2 pt-2">
						<div class="flex items-center justify-between">
							<Label class="text-xs font-medium text-foreground">
								Select Monitored Services <span class="text-destructive">*</span>
							</Label>
							<span class="text-[11px] font-mono text-muted-foreground">
								{selectedMonitors.length} of {monitors.length} selected
							</span>
						</div>

						<div
							class="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-[220px] overflow-y-auto p-1 border border-border/60 rounded-xl bg-background/50"
						>
							{#each monitors as m (m.ID)}
								{@const checked = selectedMonitors.includes(m.ID)}
								<button
									type="button"
									onclick={() => toggleMonitor(m.ID)}
									class="flex items-center gap-3 p-2.5 rounded-lg border text-left transition-all cursor-pointer {checked
										? 'bg-primary/5 border-primary/40 text-foreground'
										: 'bg-card/40 border-border/40 text-muted-foreground hover:bg-card/80 hover:text-foreground'}"
								>
									<div
										class="size-4 rounded border flex items-center justify-center shrink-0 transition-colors {checked
											? 'bg-primary border-primary text-primary-foreground'
											: 'border-border bg-background'}"
									>
										{#if checked}
											<Check class="size-3" />
										{/if}
									</div>
									<div class="min-w-0 flex-1">
										<div class="text-xs font-semibold truncate leading-tight">{m.Name}</div>
										<div class="text-[10px] font-mono text-muted-foreground truncate mt-0.5">
											{m.Target}
										</div>
									</div>
								</button>
							{:else}
								<div class="col-span-full py-8 text-center text-xs text-muted-foreground">
									No monitors found in this organization. Create a monitor first.
								</div>
							{/each}
						</div>
					</div>
				</div>
			{:else}
				<!-- Tab 2: Branding & Appearance -->
				<div class="space-y-5">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<!-- Logo URL -->
						<div class="space-y-1.5">
							<Label
								for="sp-logo"
								class="text-xs font-medium text-foreground flex items-center gap-1.5"
							>
								<Image class="size-3.5 text-muted-foreground" />
								Logo URL
							</Label>
							<Input
								id="sp-logo"
								bind:value={logoUrl}
								placeholder="https://example.com/logo.svg"
								class="bg-background border-border text-foreground h-10 rounded-lg text-sm font-mono text-xs"
							/>
						</div>

						<!-- Favicon URL -->
						<div class="space-y-1.5">
							<Label
								for="sp-favicon"
								class="text-xs font-medium text-foreground flex items-center gap-1.5"
							>
								<Sparkles class="size-3.5 text-muted-foreground" />
								Favicon URL
							</Label>
							<Input
								id="sp-favicon"
								bind:value={faviconUrl}
								placeholder="https://example.com/favicon.ico"
								class="bg-background border-border text-foreground h-10 rounded-lg text-sm font-mono text-xs"
							/>
						</div>
					</div>

					<!-- Theme Color -->
					<div class="space-y-3">
						<Label class="text-xs font-medium text-foreground">Theme Accent Color</Label>
						<div class="flex items-center gap-3">
							<input
								type="color"
								bind:value={themeColor}
								class="size-10 rounded-lg border border-border cursor-pointer bg-transparent p-0"
							/>
							<Input
								bind:value={themeColor}
								placeholder="#4edea3"
								class="bg-background border-border text-foreground h-10 rounded-lg text-sm font-mono uppercase max-w-[140px]"
							/>
							<div class="flex items-center gap-1.5 flex-wrap">
								{#each colorPresets as preset (preset.hex)}
									<button
										type="button"
										onclick={() => (themeColor = preset.hex)}
										class="size-6 rounded-full border border-white/20 transition-transform hover:scale-110 cursor-pointer shadow-sm relative"
										style="background-color: {preset.hex};"
										title={preset.name}
									>
										{#if themeColor.toLowerCase() === preset.hex.toLowerCase()}
											<Check class="size-3 text-black absolute inset-0 m-auto" />
										{/if}
									</button>
								{/each}
							</div>
						</div>
					</div>

					<!-- Custom CSS -->
					<div class="space-y-1.5">
						<Label
							for="sp-css"
							class="text-xs font-medium text-foreground flex items-center gap-1.5"
						>
							<Code2 class="size-3.5 text-muted-foreground" />
							Custom CSS Injection
						</Label>
						<textarea
							id="sp-css"
							bind:value={customCss}
							rows="4"
							placeholder="/* Add custom styles for the status page body, header, fonts */"
							class="w-full bg-background border border-border rounded-lg p-3 text-xs font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary resize-y transition-all"
						></textarea>
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div
			class="p-4 px-6 border-t border-border/70 bg-card/60 flex items-center justify-end gap-3 shrink-0"
		>
			<Button
				type="button"
				variant="ghost"
				onclick={() => (isOpen = false)}
				class="h-9 px-4 text-xs font-medium hover:bg-white/5 cursor-pointer text-foreground"
			>
				Cancel
			</Button>

			<Button
				type="button"
				onclick={handleSubmit}
				disabled={loading}
				class="h-9 px-5 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 active:scale-[0.98] transition-all cursor-pointer shadow-sm disabled:opacity-50"
			>
				{loading
					? isEdit
						? 'Saving Changes...'
						: 'Creating...'
					: isEdit
						? 'Save Changes'
						: 'Create Status Page'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
