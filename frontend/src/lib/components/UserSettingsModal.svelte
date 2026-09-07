<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { userTimezone, POPULAR_TIMEZONES } from '$lib/stores/timezone';
	import { Search, Clock, ChevronDown } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { setMode, mode } from 'mode-watcher';

	let { open = $bindable(false) } = $props();

	let theme = $state(mode.current === 'light' ? 'light' : 'dark');
	let language = $state('en');
	let timezone = $state($userTimezone);

	// ── Timezone dropdown state ───────────────────────────────────────────────
	let tzDropdownOpen = $state(false);
	let tzSearch = $state('');
	let tzDropdownEl = $state<HTMLDivElement | null>(null);

	// Filter timezone berdasarkan search query
	const filteredTimezones = $derived(
		tzSearch.trim() === ''
			? POPULAR_TIMEZONES
			: POPULAR_TIMEZONES.filter(
					(tz) =>
						tz.label.toLowerCase().includes(tzSearch.toLowerCase()) ||
						tz.value.toLowerCase().includes(tzSearch.toLowerCase()) ||
						tz.region.toLowerCase().includes(tzSearch.toLowerCase())
				)
	);

	// Kelompokkan per region, hanya tampilkan region yang ada hasilnya
	const filteredRegions = $derived([...new Set(filteredTimezones.map((tz) => tz.region))]);
	const byRegionFiltered = (region: string) =>
		filteredTimezones.filter((tz) => tz.region === region);

	// Label trigger button
	const selectedLabel = $derived(
		POPULAR_TIMEZONES.find((tz) => tz.value === timezone)?.label ?? timezone
	);

	async function openTzDropdown() {
		tzDropdownOpen = true;
		tzSearch = '';
		await tick();
		const input = tzDropdownEl?.querySelector('input');
		input?.focus();
	}

	function selectTimezone(value: string) {
		timezone = value;
		tzDropdownOpen = false;
		tzSearch = '';
	}

	function closeTzDropdown() {
		tzDropdownOpen = false;
		tzSearch = '';
	}

	function saveSettings() {
		userTimezone.set(timezone);
		setMode(theme as 'light' | 'dark' | 'system');
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="bg-card border-border text-foreground sm:max-w-[480px]">
		<Dialog.Header>
			<Dialog.Title class="text-xl">User Settings</Dialog.Title>
			<Dialog.Description class="text-muted-foreground">
				Manage your app preferences like theme, language, and timezone.
			</Dialog.Description>
		</Dialog.Header>

		<div class="grid gap-6 py-4">
			<!-- Theme -->
			<div class="space-y-3">
				<Label class="text-muted-foreground font-medium">Theme</Label>
				<Select.Root type="single" bind:value={theme}>
					<Select.Trigger
						class="w-full bg-secondary border-border text-foreground h-11 focus:ring-primary"
					>
						{theme === 'light' ? 'Light' : theme === 'dark' ? 'Dark' : 'System (Default)'}
					</Select.Trigger>
					<Select.Content class="bg-popover border-border text-foreground">
						<Select.Item
							value="light"
							class="text-foreground data-highlighted:bg-accent data-selected:bg-primary/10 data-selected:text-primary"
							>Light</Select.Item
						>
						<Select.Item
							value="dark"
							class="text-foreground data-highlighted:bg-accent data-selected:bg-primary/10 data-selected:text-primary"
							>Dark</Select.Item
						>
						<Select.Item
							value="system"
							class="text-foreground data-highlighted:bg-accent data-selected:bg-primary/10 data-selected:text-primary"
							>System (Default)</Select.Item
						>
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Language -->
			<div class="space-y-3">
				<Label class="text-muted-foreground font-medium">Language</Label>
				<Select.Root type="single" bind:value={language}>
					<Select.Trigger
						class="w-full bg-secondary border-border text-foreground h-11 focus:ring-primary"
					>
						{language === 'en' ? 'English (US)' : 'Bahasa Indonesia'}
					</Select.Trigger>
					<Select.Content class="bg-popover border-border text-foreground">
						<Select.Item
							value="en"
							class="text-foreground data-highlighted:bg-accent data-selected:bg-primary/10 data-selected:text-primary"
							>English (US)</Select.Item
						>
						<Select.Item
							value="id"
							class="text-foreground data-highlighted:bg-accent data-selected:bg-primary/10 data-selected:text-primary"
							>Bahasa Indonesia</Select.Item
						>
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Timezone — Custom dropdown dengan search + scrollbar -->
			<div class="space-y-3">
				<Label class="text-muted-foreground font-medium">Timezone</Label>
				<div class="relative" bind:this={tzDropdownEl}>
					<!-- Trigger button -->
					<button
						type="button"
						id="tz-trigger"
						onclick={openTzDropdown}
						class="w-full flex items-center justify-between gap-2 h-11 px-3 rounded-md border border-border bg-secondary text-foreground text-sm text-left transition-colors hover:border-ring/40 focus:outline-none focus:ring-1 focus:ring-primary/60 {tzDropdownOpen
							? 'border-ring/40 ring-1 ring-primary/40'
							: ''}"
					>
						<span class="flex items-center gap-2 min-w-0">
							<Clock class="size-3.5 text-muted-foreground shrink-0" />
							<span class="truncate font-mono text-xs">{selectedLabel}</span>
						</span>
						<ChevronDown
							class="size-4 text-muted-foreground shrink-0 transition-transform duration-200 {tzDropdownOpen
								? 'rotate-180'
								: ''}"
						/>
					</button>

					<!-- Dropdown panel -->
					{#if tzDropdownOpen}
						<!-- Backdrop untuk close on outside click -->
						<button
							type="button"
							class="fixed inset-0 z-40 cursor-default bg-transparent"
							onclick={closeTzDropdown}
							tabindex="-1"
							aria-hidden="true"
						></button>

						<div
							class="tz-dropdown absolute left-0 right-0 top-[calc(100%+4px)] z-50 rounded-lg border border-border bg-popover shadow-2xl overflow-hidden"
						>
							<!-- Search bar -->
							<div class="p-2 border-b border-border">
								<div
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-md bg-secondary border border-border focus-within:border-ring/40 transition-colors"
								>
									<Search class="size-3.5 text-muted-foreground shrink-0" />
									<input
										type="text"
										bind:value={tzSearch}
										placeholder="Search timezone…"
										class="flex-1 bg-transparent text-sm text-foreground placeholder:text-muted-foreground outline-none"
									/>
									{#if tzSearch}
										<button
											type="button"
											onclick={() => (tzSearch = '')}
											class="text-muted-foreground hover:text-foreground transition-colors leading-none"
											aria-label="Clear search">✕</button
										>
									{/if}
								</div>
							</div>

							<!-- Scrollable list -->
							<div data-lenis-prevent class="tz-scroll-list max-h-52 overflow-y-auto py-1">
								{#if filteredTimezones.length === 0}
									<div class="px-4 py-6 text-center text-sm text-muted-foreground">
										No timezone found for "<span class="text-foreground">{tzSearch}</span>"
									</div>
								{:else}
									{#each filteredRegions as region}
										<div class="px-3 pt-2 pb-0.5">
											<span
												class="text-[9px] font-bold uppercase tracking-[0.12em] text-muted-foreground/60"
											>
												{region}
											</span>
										</div>
										{#each byRegionFiltered(region) as tz}
											<button
												type="button"
												onclick={() => selectTimezone(tz.value)}
												class="tz-item w-full text-left px-3 py-1.5 text-xs font-mono rounded-md transition-colors
												{timezone === tz.value
													? 'bg-primary/15 text-primary'
													: 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
											>
												<span class="flex items-center justify-between">
													<span>{tz.label}</span>
													{#if timezone === tz.value}
														<span class="text-primary text-[10px] shrink-0 ml-2">✓</span>
													{/if}
												</span>
											</button>
										{/each}
									{/each}
								{/if}
							</div>
						</div>
					{/if}
				</div>
				<p class="text-xs text-muted-foreground">
					Semua waktu di dashboard akan ditampilkan dalam timezone ini.
				</p>
			</div>
		</div>

		<Dialog.Footer class="bg-card border-border">
			<Button
				variant="outline"
				onclick={() => (open = false)}
				class="text-foreground h-10 hover:text-foreground border-transparent hover:bg-accent transition-colors"
			>
				Cancel
			</Button>
			<Button
				onclick={saveSettings}
				class="bg-secondary h-10 border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200"
			>
				Save Changes
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	/* Custom scrollbar */
	.tz-scroll-list {
		scrollbar-width: thin;
		scrollbar-color: var(--border) transparent;
	}
	.tz-scroll-list::-webkit-scrollbar {
		width: 5px;
	}
	.tz-scroll-list::-webkit-scrollbar-track {
		background: transparent;
	}
	.tz-scroll-list::-webkit-scrollbar-thumb {
		background-color: var(--border);
		border-radius: 999px;
	}
	.tz-scroll-list::-webkit-scrollbar-thumb:hover {
		background-color: var(--muted-foreground);
	}

	/* Timezone item full width inside container */
	.tz-item {
		display: block;
		width: calc(100% - 8px);
		margin-left: 4px;
		margin-right: 4px;
	}

	/* Dropdown slide-in animation */
	.tz-dropdown {
		animation: tz-slide-in 120ms cubic-bezier(0.16, 1, 0.3, 1);
	}
	@keyframes tz-slide-in {
		from {
			opacity: 0;
			transform: translateY(-6px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}
</style>
