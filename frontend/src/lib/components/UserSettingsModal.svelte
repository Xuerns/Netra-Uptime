<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';

	let { open = $bindable(false) } = $props();

	let theme = $state('system');
	let language = $state('en');
	let timezone = $state('UTC');

</script>

<Dialog.Root bind:open>
	<Dialog.Content class="bg-[#131314] border-[#2A2A2B] text-white sm:max-w-[425px]">
		<Dialog.Header>
			<Dialog.Title class="text-xl">User Settings</Dialog.Title>
			<Dialog.Description class="text-slate-400">
				Manage your app preferences like theme and language.
			</Dialog.Description>
		</Dialog.Header>

		<div class="grid gap-6 py-4">
			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Theme</Label>
				<Select.Root type="single" bind:value={theme}>
					<Select.Trigger class="w-full bg-[#1A1A1B] border-[#2A2A2B] text-white h-11 focus:ring-emerald-500">
						{theme === 'light' ? 'Light' : theme === 'dark' ? 'Dark' : 'System (Default)'}
					</Select.Trigger>
					<Select.Content class="bg-[#1A1A1B] border-[#2A2A2B] text-white">
						<Select.Item value="light">Light</Select.Item>
						<Select.Item value="dark">Dark</Select.Item>
						<Select.Item value="system">System (Default)</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>

			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Language</Label>
				<Select.Root type="single" bind:value={language}>
					<Select.Trigger class="w-full bg-[#1A1A1B] border-[#2A2A2B] text-white h-11 focus:ring-emerald-500">
						{language === 'en' ? 'English (US)' : 'Bahasa Indonesia'}
					</Select.Trigger>
					<Select.Content class="bg-[#1A1A1B] border-[#2A2A2B] text-white">
						<Select.Item value="en">English (US)</Select.Item>
						<Select.Item value="id">Bahasa Indonesia</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>

			<div class="space-y-3">
				<Label class="text-slate-300 font-medium">Timezone</Label>
				<Select.Root type="single" bind:value={timezone}>
					<Select.Trigger class="w-full bg-[#1A1A1B] border-[#2A2A2B] text-white h-11 focus:ring-emerald-500">
						{timezone}
					</Select.Trigger>
					<Select.Content class="bg-[#1A1A1B] border-[#2A2A2B] text-white max-h-60 overflow-y-auto">
						<Select.Item value="UTC">UTC (Default)</Select.Item>
						<Select.Item value="Asia/Jakarta">Asia/Jakarta (WIB)</Select.Item>
						<Select.Item value="America/New_York">America/New_York (EST)</Select.Item>
						<Select.Item value="Europe/London">Europe/London (GMT)</Select.Item>
					</Select.Content>
				</Select.Root>
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
				onclick={() => {
					// In a real app, save to localStorage or backend here
					open = false;
				}}
				class="bg-[#1A1A1B] h-10 border border-[#2A2A2B] text-emerald-500 hover:bg-emerald-600 hover:border-emerald-600 hover:text-white active:scale-[0.98] transition-all duration-200"
			>
				Save Changes
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
