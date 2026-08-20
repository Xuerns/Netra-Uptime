<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { userTimezone, POPULAR_TIMEZONES } from '$lib/stores/timezone';
	import { Search, Clock, ChevronDown } from 'lucide-svelte';
	import { tick } from 'svelte';

	let { open = $bindable(false) } = $props();

	let theme = $state('system');
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
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="bg-[#131314] border-[#2A2A2B] text-white sm:max-w-[480px]">
		<Dialog.Header>
			<Dialog.Title class="text-xl">User Settings</Dialog.Title>
			<Dialog.Description class="text-slate-400">
				Manage your app preferences like theme, language, and timezone.
			</Dialog.Description>
		</Dialog.Header>

		<div class="grid gap-6 py-4">
			<!-- Theme -->
			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Theme</Label>
				<Select.Root type="single" bind:value={theme}>
					<Select.Trigger
						class="w-full bg-[#1A1A1B] border-[#2A2A2B] text-white h-11 focus:ring-emerald-500"
					>
						{theme === 'light' ? 'Light' : theme === 'dark' ? 'Dark' : 'System (Default)'}
					</Select.Trigger>
					<Select.Content class="bg-[#1A1A1B] border-[#2A2A2B] text-white">
						<Select.Item
							value="light"
							class="text-white data-highlighted:bg-white/5 data-selected:bg-emerald-500/10 data-selected:text-emerald-400"
							>Light</Select.Item
						>
						<Select.Item
							value="dark"
							class="text-white data-highlighted:bg-white/5 data-selected:bg-emerald-500/10 data-selected:text-emerald-400"
							>Dark</Select.Item
						>
						<Select.Item
							value="system"
							class="text-white data-highlighted:bg-white/5 data-selected:bg-emerald-500/10 data-selected:text-emerald-400"
							>System (Default)</Select.Item
						>
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Language -->
			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Language</Label>
				<Select.Root type="single" bind:value={language}>
					<Select.Trigger
						class="w-full bg-[#1A1A1B] border-[#2A2A2B] text-white h-11 focus:ring-emerald-500"
					>
						{language === 'en' ? 'English (US)' : 'Bahasa Indonesia'}
					</Select.Trigger>
					<Select.Content class="bg-[#1A1A1B] border-[#2A2A2B] text-white">
						<Select.Item
							value="en"
							class="text-white data-highlighted:bg-white/5 data-selected:bg-emerald-500/10 data-selected:text-emerald-400"
							>English (US)</Select.Item
						>
						<Select.Item
							value="id"
							class="text-white data-highlighted:bg-white/5 data-selected:bg-emerald-500/10 data-selected:text-emerald-400"
							>Bahasa Indonesia</Select.Item
						>
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Timezone — Custom dropdown dengan search + scrollbar -->
			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Timezone</Label>
				<div class="relative" bind:this={tzDropdownEl}>
					<!-- Trigger button -->
					<button
						type="button"
						id="tz-trigger"
						onclick={openTzDropdown}
						class="w-full flex items-center justify-between gap-2 h-11 px-3 rounded-md border border-[#2A2A2B] bg-[#1A1A1B] text-white text-sm text-left transition-colors hover:border-white/20 focus:outline-none focus:ring-1 focus:ring-emerald-500/60 {tzDropdownOpen
							? 'border-white/20 ring-1 ring-emerald-500/40'
							: ''}"
					>
						<span class="flex items-center gap-2 min-w-0">
							<Clock class="size-3.5 text-slate-400 shrink-0" />
							<span class="truncate font-mono text-xs">{selectedLabel}</span>
						</span>
						<ChevronDown
							class="size-4 text-slate-400 shrink-0 transition-transform duration-200 {tzDropdownOpen
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
							class="tz-dropdown absolute left-0 right-0 top-[calc(100%+4px)] z-50 rounded-lg border border-[#2A2A2B] bg-[#131314] shadow-2xl shadow-black/70 overflow-hidden"
						>
							<!-- Search bar -->
							<div class="p-2 border-b border-[#2A2A2B]">
								<div
									class="flex items-center gap-2 px-2.5 py-1.5 rounded-md bg-[#1A1A1B] border border-[#2A2A2B] focus-within:border-white/20 transition-colors"
								>
									<Search class="size-3.5 text-slate-500 shrink-0" />
									<input
										type="text"
										bind:value={tzSearch}
										placeholder="Search timezone…"
										class="flex-1 bg-transparent text-sm text-white placeholder:text-slate-500 outline-none"
									/>
									{#if tzSearch}
										<button
											type="button"
											onclick={() => (tzSearch = '')}
											class="text-slate-500 hover:text-white transition-colors leading-none"
											aria-label="Clear search">✕</button
										>
									{/if}
								</div>
							</div>

							<!-- Scrollable list -->
							<div data-lenis-prevent class="tz-scroll-list max-h-52 overflow-y-auto py-1">
								{#if filteredTimezones.length === 0}
									<div class="px-4 py-6 text-center text-sm text-slate-500">
										No timezone found for "<span class="text-slate-300">{tzSearch}</span>"
									</div>
								{:else}
									{#each filteredRegions as region}
										<div class="px-3 pt-2 pb-0.5">
											<span class="text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600">
												{region}
											</span>
										</div>
										{#each byRegionFiltered(region) as tz}
											<button
												type="button"
												onclick={() => selectTimezone(tz.value)}
												class="tz-item w-full text-left px-3 py-1.5 text-xs font-mono rounded-md transition-colors
												{timezone === tz.value
													? 'bg-emerald-500/15 text-emerald-400'
													: 'text-slate-300 hover:bg-white/5 hover:text-white'}"
											>
												<span class="flex items-center justify-between">
													<span>{tz.label}</span>
													{#if timezone === tz.value}
														<span class="text-emerald-500 text-[10px] shrink-0 ml-2">✓</span>
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
				<p class="text-xs text-slate-500">
					Semua waktu di dashboard akan ditampilkan dalam timezone ini.
				</p>
			</div>
		</div>

		<Dialog.Footer class="bg-[#131314] border-[#2A2A2B]">
			<Button
				variant="outline"
				onclick={() => (open = false)}
				class="text-white h-10 hover:text-white border-transparent hover:bg-white/10 transition-colors"
			>
				Cancel
			</Button>
			<Button
				onclick={saveSettings}
				class="bg-[#1A1A1B] h-10 border border-[#2A2A2B] text-emerald-500 hover:bg-emerald-600 hover:border-emerald-600 hover:text-white active:scale-[0.98] transition-all duration-200"
			>
				Save Changes
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	/* Custom dark scrollbar */
	.tz-scroll-list {
		scrollbar-width: thin;
		scrollbar-color: #3a3a3c transparent;
	}
	.tz-scroll-list::-webkit-scrollbar {
		width: 5px;
	}
	.tz-scroll-list::-webkit-scrollbar-track {
		background: transparent;
	}
	.tz-scroll-list::-webkit-scrollbar-thumb {
		background-color: #3a3a3c;
		border-radius: 999px;
	}
	.tz-scroll-list::-webkit-scrollbar-thumb:hover {
		background-color: #555;
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
