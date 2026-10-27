<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { userTimezone, POPULAR_TIMEZONES } from '$lib/stores/timezone';
	import { Clock, Check, ChevronsUpDown } from 'lucide-svelte';
	import { setMode, mode } from 'mode-watcher';

	let { open = $bindable(false) } = $props();

	let theme = $state(mode.current === 'light' ? 'light' : 'dark');
	let language = $state('en');
	let timezone = $state($userTimezone);

	// ── Timezone Combobox state ───────────────────────────────────────────────
	let tzDropdownOpen = $state(false);

	const regions = [...new Set(POPULAR_TIMEZONES.map((tz) => tz.region))];
	const byRegion = (region: string) => POPULAR_TIMEZONES.filter((tz) => tz.region === region);

	// Label trigger button
	const selectedLabel = $derived(
		POPULAR_TIMEZONES.find((tz) => tz.value === timezone)?.label ?? timezone
	);

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

			<!-- Timezone — Combobox (Popover + Command) -->
			<div class="space-y-3">
				<Label class="text-muted-foreground font-medium">Timezone</Label>
				<Popover.Root bind:open={tzDropdownOpen}>
					<Popover.Trigger
						class="w-full flex items-center justify-between gap-2 h-11 px-3 rounded-md border border-border bg-secondary text-foreground text-sm text-left transition-colors hover:border-ring/40 focus:outline-none focus:ring-1 focus:ring-primary/60 cursor-pointer"
					>
						<span class="flex items-center gap-2 min-w-0">
							<Clock class="size-3.5 text-muted-foreground shrink-0" />
							<span class="truncate font-mono text-xs">{selectedLabel}</span>
						</span>
						<ChevronsUpDown class="size-4 text-muted-foreground shrink-0 opacity-50" />
					</Popover.Trigger>
					<Popover.Content
						class="w-[var(--bits-popover-anchor-width)] min-w-[320px] p-0"
						align="start"
					>
						<Command.Root>
							<Command.Input placeholder="Search timezone..." />
							<Command.List class="max-h-56 overflow-y-auto">
								<Command.Empty>No timezone found.</Command.Empty>
								{#each regions as region (region)}
									<Command.Group heading={region}>
										{#each byRegion(region) as tz (tz.value)}
											<Command.Item
												value="{tz.label} {tz.value} {region}"
												onSelect={() => {
													timezone = tz.value;
													tzDropdownOpen = false;
												}}
												class="font-mono text-xs cursor-pointer flex items-center justify-between"
											>
												<span>{tz.label}</span>
												{#if timezone === tz.value}
													<Check class="size-3.5 text-primary shrink-0 ml-2" />
												{/if}
											</Command.Item>
										{/each}
									</Command.Group>
								{/each}
							</Command.List>
						</Command.Root>
					</Popover.Content>
				</Popover.Root>
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
