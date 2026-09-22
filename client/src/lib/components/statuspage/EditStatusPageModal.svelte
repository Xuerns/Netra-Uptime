<script lang="ts">
	import { StatusPageAPI, type StatusPage } from '$lib/api/statuspage';
	import type { Monitor } from '$lib/api/monitor';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Switch } from '$lib/components/ui/switch';
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
		statusPage,
		onSuccess
	} = $props<{
		isOpen: boolean;
		orgId: string;
		monitors: Monitor[];
		statusPage: StatusPage | null;
		onSuccess: () => void;
	}>();

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
			logoUrl = statusPage.LogoURL || '';
			faviconUrl = statusPage.FaviconURL || '';
			themeColor = statusPage.ThemeColor || '#4edea3';
			customCss = statusPage.CustomCSS || '';
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
				logo_url: logoUrl.trim() || '',
				favicon_url: faviconUrl.trim() || '',
				theme_color: themeColor.trim() || '#4edea3',
				custom_css: customCss.trim() || '',
				monitor_id: selectedMonitors
			});
			toast.success('Status Page updated successfully');
			isOpen = false;
			onSuccess();
		} catch (err: any) {
			error = parseApiError(err, 'Failed to update status page');
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
		<div class="p-6 pb-4 border-b border-border/60">
			<Dialog.Title class="text-xl font-semibold tracking-tight">Edit Status Page</Dialog.Title>
			<Dialog.Description class="text-muted-foreground text-xs mt-1">
				Update configuration, branding, and services for your public status page.
			</Dialog.Description>

			<!-- Tab Navigation -->
			<div class="flex items-center gap-4 mt-5 border-b border-border/40 -mb-4">
				<button
					type="button"
					class="text-xs font-semibold pb-3 border-b-2 transition-colors flex items-center gap-1.5 cursor-pointer {activeTab === 'general'
						? 'border-primary text-primary'
						: 'border-transparent text-muted-foreground hover:text-foreground'}"
					onclick={() => (activeTab = 'general')}
				>
					<Settings class="size-3.5" />
					General & Monitors
				</button>
				<button
					type="button"
					class="text-xs font-semibold pb-3 border-b-2 transition-colors flex items-center gap-1.5 cursor-pointer {activeTab === 'branding'
						? 'border-primary text-primary'
						: 'border-transparent text-muted-foreground hover:text-foreground'}"
					onclick={() => (activeTab = 'branding')}
				>
					<Palette class="size-3.5" />
					Custom Branding
				</button>
			</div>
		</div>

		<form
			class="flex-1 overflow-y-auto p-6 flex flex-col gap-5"
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

			{#if activeTab === 'general'}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<!-- Left Column: General Settings -->
					<div class="space-y-4">
						<div class="space-y-1.5">
							<Label for="edit-title" class="text-xs font-medium text-foreground">
								Title <span class="text-destructive font-bold">*</span>
							</Label>
							<Input
								id="edit-title"
								bind:value={title}
								placeholder="e.g. Acme Status"
								required
								class="bg-background border-border text-foreground text-xs h-9"
							/>
						</div>

						<div class="space-y-1.5">
							<Label for="edit-description" class="text-xs font-medium text-foreground">
								Description <span class="text-muted-foreground text-[11px] font-normal">(Optional)</span>
							</Label>
							<textarea
								id="edit-description"
								bind:value={description}
								placeholder="e.g. Real-time status for all Acme systems and APIs."
								class="flex min-h-[70px] w-full rounded-lg border border-border bg-background px-3 py-2 text-xs text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50"
							></textarea>
						</div>

						<div class="space-y-1.5">
							<Label for="edit-subdomain" class="text-xs font-medium text-foreground">
								Subdomain <span class="text-destructive font-bold">*</span>
							</Label>
							<div class="flex items-center">
								<Input
									id="edit-subdomain"
									bind:value={subdomain}
									placeholder="acme"
									required
									class="bg-background border-border border-r-0 rounded-r-none text-foreground text-xs h-9 z-10"
								/>
								<div
									class="bg-background/80 border border-border border-l-0 px-2.5 h-9 flex items-center rounded-r-lg text-muted-foreground text-xs font-mono whitespace-nowrap"
								>
									.pulsewatch.com
								</div>
							</div>
						</div>

						<div
							class="flex items-center justify-between p-3.5 bg-background border border-border rounded-lg"
						>
							<div class="space-y-0.5">
								<Label class="text-xs font-medium text-foreground">Public Visibility</Label>
								<p class="text-[11px] text-muted-foreground">Anyone with the link can view this page.</p>
							</div>
							<Switch bind:checked={isPublic} />
						</div>
					</div>

					<!-- Right Column: Select Monitors -->
					<div class="space-y-2 flex flex-col h-full">
						<div class="flex items-center justify-between">
							<Label class="text-xs font-medium text-foreground">Select Monitors</Label>
							<span class="text-[11px] text-muted-foreground">{selectedMonitors.length} selected</span>
						</div>

						<div
							class="flex-1 min-h-[220px] max-h-[280px] overflow-y-auto bg-background border border-border rounded-lg p-2 space-y-1"
						>
							{#if monitors.length === 0}
								<div class="p-6 text-center text-xs text-muted-foreground">
									No monitors available.
								</div>
							{:else}
								{#each monitors as monitor}
									{@const isSelected = selectedMonitors.includes(monitor.ID)}
									<label
										data-lenis-prevent
										class="flex items-center gap-2.5 p-2 hover:bg-white/5 rounded-md cursor-pointer transition-colors"
									>
										<div
											class="size-7 rounded-full flex items-center justify-center text-xs shrink-0 border bg-white/5 text-foreground/80 border-border"
										>
											<Activity class="size-3.5 text-primary" />
										</div>

										<div class="flex flex-col flex-1 min-w-0">
											<span class="text-xs font-medium text-foreground truncate">{monitor.Name}</span>
											<span class="text-[10px] text-muted-foreground font-mono truncate">{monitor.Target}</span>
										</div>

										<input
											type="checkbox"
											checked={isSelected}
											onchange={() => toggleMonitor(monitor.ID)}
											class="sr-only"
										/>
										<div
											class="w-4 h-4 rounded border flex items-center justify-center shrink-0 transition-all {isSelected
												? 'bg-primary border-primary'
												: 'bg-white/5 border-border'}"
										>
											{#if isSelected}
												<Check class="size-3 text-black stroke-[3]" />
											{/if}
										</div>
									</label>
								{/each}
							{/if}
						</div>
					</div>
				</div>
			{:else}
				<!-- Branding Tab -->
				<div class="space-y-5">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-5">
						<!-- Logo URL -->
						<div class="space-y-1.5">
							<Label for="edit-logo" class="text-xs font-medium text-foreground flex items-center gap-1.5">
								<Image class="size-3.5 text-primary" />
								Custom Logo URL
							</Label>
							<Input
								id="edit-logo"
								bind:value={logoUrl}
								placeholder="https://example.com/logo.png"
								class="bg-background border-border text-foreground text-xs h-9"
							/>
							<p class="text-[11px] text-muted-foreground">
								Direct URL to image (PNG or SVG with transparent background recommended).
							</p>
							{#if logoUrl.trim()}
								<div class="mt-2 p-2.5 bg-black/40 rounded-lg border border-border/50 flex items-center gap-3">
									<img
										src={logoUrl}
										alt="Logo preview"
										class="h-7 max-w-28 object-contain rounded"
										onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
									/>
									<span class="text-[11px] text-muted-foreground font-mono">Logo preview</span>
								</div>
							{/if}
						</div>

						<!-- Favicon URL -->
						<div class="space-y-1.5">
							<Label for="edit-favicon" class="text-xs font-medium text-foreground flex items-center gap-1.5">
								<Sparkles class="size-3.5 text-blue-400" />
								Custom Favicon URL
							</Label>
							<Input
								id="edit-favicon"
								bind:value={faviconUrl}
								placeholder="https://example.com/favicon.ico"
								class="bg-background border-border text-foreground text-xs h-9"
							/>
							<p class="text-[11px] text-muted-foreground">
								Direct URL to `.ico` or `.png` icon for browser tab.
							</p>
							{#if faviconUrl.trim()}
								<div class="mt-2 p-2.5 bg-black/40 rounded-lg border border-border/50 flex items-center gap-3">
									<img
										src={faviconUrl}
										alt="Favicon preview"
										class="size-5 object-contain"
										onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
									/>
									<span class="text-[11px] text-muted-foreground font-mono">Favicon preview</span>
								</div>
							{/if}
						</div>
					</div>

					<!-- Theme Color -->
					<div class="space-y-2 p-4 rounded-xl border border-border/80 bg-white/2">
						<Label class="text-xs font-semibold text-foreground flex items-center gap-1.5">
							<Palette class="size-3.5 text-emerald-400" />
							Theme Accent Color
						</Label>
						<p class="text-[11px] text-muted-foreground">
							Choose your brand color to accent status badges, buttons, and indicators on the status page.
						</p>

						<div class="flex flex-wrap items-center gap-3 pt-1">
							<!-- Native Color Picker & Hex Input -->
							<div class="flex items-center gap-2 bg-background border border-border rounded-lg px-2 py-1">
								<input
									type="color"
									bind:value={themeColor}
									class="size-6 rounded border-0 bg-transparent cursor-pointer"
								/>
								<Input
									bind:value={themeColor}
									placeholder="#4edea3"
									class="w-24 h-7 text-xs font-mono bg-transparent border-0 p-0 text-foreground focus-visible:ring-0"
								/>
							</div>

							<!-- Color Swatches -->
							<div class="flex items-center gap-1.5">
								{#each colorPresets as preset}
									<button
										type="button"
										class="size-6 rounded-full border transition-transform hover:scale-110 flex items-center justify-center {themeColor.toLowerCase() === preset.hex.toLowerCase()
											? 'ring-2 ring-foreground ring-offset-2 ring-offset-background'
											: 'border-white/10'}"
										style="background-color: {preset.hex};"
										title={preset.name}
										onclick={() => (themeColor = preset.hex)}
									>
										{#if themeColor.toLowerCase() === preset.hex.toLowerCase()}
											<Check class="size-3 text-black stroke-[3]" />
										{/if}
									</button>
								{/each}
							</div>
						</div>
					</div>

					<!-- Custom CSS -->
					<div class="space-y-1.5">
						<Label for="edit-css" class="text-xs font-medium text-foreground flex items-center gap-1.5">
							<Code2 class="size-3.5 text-purple-400" />
							Custom CSS
						</Label>
						<p class="text-[11px] text-muted-foreground">
							Inject custom styles directly into your status page (e.g. fonts, overrides, styling tweaks).
						</p>
						<textarea
							id="edit-css"
							bind:value={customCss}
							rows={4}
							placeholder="/* Example custom CSS */&#10;.custom-header &#123; font-weight: bold; &#125;"
							class="w-full rounded-lg border border-border bg-black/60 p-3 text-xs font-mono text-foreground/90 placeholder:text-muted-foreground/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary"
						></textarea>
					</div>
				</div>
			{/if}

			<!-- Footer -->
			<div class="border-t border-border pt-4 mt-auto flex items-center justify-end gap-2 bg-transparent">
				<Button
					type="button"
					variant="ghost"
					size="sm"
					class="text-xs h-9 px-4 hover:text-foreground hover:bg-accent"
					onclick={() => (isOpen = false)}
					disabled={loading}
				>
					Cancel
				</Button>
				<Button
					type="submit"
					size="sm"
					disabled={loading}
					class="text-xs h-9 px-5 bg-primary text-primary-foreground font-medium shadow-sm"
				>
					{loading ? 'Saving...' : 'Save Changes'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
