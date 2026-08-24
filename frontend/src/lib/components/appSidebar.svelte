<script lang="ts">
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Avatar from '$lib/components/ui/avatar';

	// Icons
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';

	// App Specific Icons
	import LayoutDashboard from 'lucide-svelte/icons/layout-dashboard';
	import Activity from 'lucide-svelte/icons/activity';
	import AlertTriangle from 'lucide-svelte/icons/alert-triangle';
	import Globe from 'lucide-svelte/icons/globe';

	import Users from 'lucide-svelte/icons/users';
	import CreditCard from 'lucide-svelte/icons/credit-card';
	import logs from 'lucide-svelte/icons/logs';
	import Settings2 from 'lucide-svelte/icons/settings-2';

	import LogOut from 'lucide-svelte/icons/log-out';
	import Bell from 'lucide-svelte/icons/bell';
	import BadgeCheck from 'lucide-svelte/icons/badge-check';
	import { env } from '$env/dynamic/public';

	import Logo from '$lib/assets/Logo.svg';
	import { apiFetch } from '$lib/helper/api';
	import UserSettingsModal from '$lib/components/UserSettingsModal.svelte';
	import { isSelfHosted } from '$lib/utils/appMode';

	let isSettingsOpen = $state(false);

	// State for dropdowns/selection
	let activeTeam = $state<{ name: string; logo: any }>({
		name: 'PulseWatch Workspace',
		logo: Logo
	});

	const navMonitoring = [
		{ title: 'Monitors', icon: Activity, url: '/dashboard/monitor', hasSub: false },
		{ title: 'Incidents', icon: AlertTriangle, url: '/dashboard/incidents', hasSub: false },
		{ title: 'Status Pages', icon: Globe, url: '/dashboard/statuspage', hasSub: false }
	];

	const navSettings = [
		{ title: 'Organization', url: '/dashboard/organization', icon: Users },
		{ title: 'Logs', url: '/dashboard/logs', icon: logs },
		...(!isSelfHosted() ? [{ title: 'Billing', url: '/dashboard/billing', icon: CreditCard }] : [])
	];

	let {
		user = {
			name: '',
			email: '',
			avatar: ''
		}
	} = $props();

	const API_Route = `${env.PUBLIC_API_URL}`;

	const sidebar = Sidebar.useSidebar();

	async function handleLogout(event: Event) {
		try {
			const res = await apiFetch(`${API_Route}/auth/logout`, { method: 'POST' });

			if (res.ok) {
				window.location.href = isSelfHosted() ? '/auth/login' : '/';
			}
		} catch (err: any) {
			console.error(err.message);
		}
	}
</script>

<!-- Glassmorphic Modernism styling according to DESIGN.md -->
<Sidebar.Root
	collapsible="icon"
	class="bg-[#161618]! !border-r !border-[rgba(255,255,255,0.08)] text-white shadow-[2px_0_30px_rgba(0,0,0,0.5)]"
>
	<!-- Header: Team Switcher -->
	<Sidebar.Header class="p-3 border-b border-[rgba(255,255,255,0.08)]">
		<div
			class="flex items-center gap-3 p-2 rounded-xl transition-all duration-300 text-left w-full border border-transparent group group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
		>
			<!-- Monitoring Blue accent (#3B82F6) -->
			{#if typeof activeTeam.logo === 'string'}
				<img src={activeTeam.logo} alt="Logo" class="size-8 object-contain drop-shadow-md" />
			{:else}
				<activeTeam.logo class="size-4" />
			{/if}

			<div class="grid flex-1 text-left text-lg leading-tight group-data-[collapsible=icon]:hidden">
				<span class="truncate font-semibold text-white tracking-tight">{activeTeam.name}</span>
			</div>
		</div>
	</Sidebar.Header>

	<!-- Content: Main Navigation & Projects -->
	<Sidebar.Content class="py-2">
		<!-- Monitoring Group -->
		<Sidebar.Group>
			<Sidebar.GroupLabel
				class="px-3 text-xs font-semibold text-[#94A3B8] tracking-widest uppercase group-data-[collapsible=icon]:hidden"
				>Monitoring</Sidebar.GroupLabel
			>
			<Sidebar.GroupContent class="mt-2">
				<Sidebar.Menu class="space-y-1">
					{#each navMonitoring as item}
						{@const isActive =
							$page.url.pathname === item.url || $page.url.pathname.startsWith(item.url + '/')}
						<Sidebar.MenuItem
							class="group/nav relative flex items-center justify-between rounded-xl border border-transparent hover:border-[rgba(255,255,255,0.08)] transition-all duration-300 {isActive
								? 'bg-[rgba(255,255,255,0.08)]'
								: 'hover:bg-[rgba(255,255,255,0.06)]'}"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 {isActive
									? 'text-white!'
									: 'text-[#94A3B8]! hover:text-white!'}"
							>
								{#snippet child({ props })}
									<a
										href={item.url}
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<!-- Use pure white on hover for high contrast -->
										<item.icon
											class="size-5 shrink-0 transition-colors {isActive
												? 'text-white'
												: 'text-[#94A3B8] group-hover/nav:text-white'}"
										/>
										<span class="font-medium flex-1 text-md group-data-[collapsible=icon]:hidden"
											>{item.title}</span
										>
										{#if item.hasSub}
											<ChevronRight
												class="size-4 shrink-0 transition-colors group-data-[collapsible=icon]:hidden {isActive
													? 'text-white'
													: 'text-[#94A3B8] group-hover/nav:text-white'}"
											/>
										{/if}
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<!-- Settings Group -->
		<Sidebar.Group class="mt-4">
			<Sidebar.GroupLabel
				class="px-3 text-xs font-semibold text-[#94A3B8] tracking-widest uppercase group-data-[collapsible=icon]:hidden"
				>Organization</Sidebar.GroupLabel
			>
			<Sidebar.GroupContent class="mt-2">
				<Sidebar.Menu class="space-y-1">
					{#each navSettings as item}
						{@const isActive =
							$page.url.pathname === item.url || $page.url.pathname.startsWith(item.url + '/')}
						<Sidebar.MenuItem
							class="group/item relative flex items-center justify-between rounded-xl border border-transparent hover:border-[rgba(255,255,255,0.08)] transition-all duration-300 {isActive
								? 'bg-[rgba(255,255,255,0.08)]'
								: 'hover:bg-[rgba(255,255,255,0.06)]'}"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 {isActive
									? 'text-white!'
									: 'text-[#94A3B8]! hover:text-white!'}"
							>
								{#snippet child({ props })}
									<a
										href={item.url}
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 pr-8 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<item.icon
											class="size-4.5 shrink-0 transition-colors {isActive
												? 'text-white'
												: 'text-[#94A3B8] group-hover/item:text-white'}"
										/>
										<span class="text-md font-medium group-data-[collapsible=icon]:hidden"
											>{item.title}</span
										>
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>

	<!-- Footer: User Switcher -->
	<Sidebar.Footer class="p-3 border-t border-[rgba(255,255,255,0.08)]">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="w-full focus:outline-none">
				<div
					class="flex items-center gap-3 p-2 hover:bg-[rgba(255,255,255,0.06)] rounded-xl transition-all duration-300 text-left w-full border border-transparent hover:border-[rgba(255,255,255,0.08)] group group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
				>
					<!-- Primary Uptime Green (#10B981) for Avatar fallback -->
					<div
						class="relative size-9 rounded-xl overflow-hidden shrink-0 border border-[rgba(255,255,255,0.1)] shadow-[0_0_10px_rgba(16,185,129,0.2)] flex items-center justify-center bg-[#10B981] text-[#003824] font-bold"
					>
						<span class="z-0">{user.name ? user.name.slice(0, 2).toUpperCase() : ''}</span>
						{#if user.avatar}
							<img
								src={user.avatar}
								alt={user.name}
								class="absolute inset-0 z-10 size-full object-cover bg-background"
								onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
							/>
						{/if}
					</div>
					<div
						class="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden"
					>
						<span class="truncate font-semibold text-white group-hover:text-white transition-colors"
							>{user.name}</span
						>
						<span
							class="truncate text-xs text-[#94A3B8] group-hover:text-[#bbcabf] transition-colors"
							>{user.email}</span
						>
					</div>
					<ChevronsUpDown
						class="ml-auto size-4 text-[#94A3B8] shrink-0 group-hover:text-white transition-colors group-data-[collapsible=icon]:hidden"
					/>
				</div>
			</DropdownMenu.Trigger>

			<DropdownMenu.Content
				class="w-56 !bg-[rgba(255,255,255,0.06)] backdrop-blur-[20px] !border-[rgba(255,255,255,0.08)] text-white shadow-[0_0_20px_rgba(16,185,129,0.05)]"
				side={sidebar.state === 'collapsed' ? 'right' : 'top'}
				align="end"
				sideOffset={4}
			>
				<DropdownMenu.Label class="font-normal">
					<div class="flex items-center gap-3 px-1.5 py-2 text-left text-sm">
						<div
							class="relative size-9 rounded-xl overflow-hidden shrink-0 border border-[rgba(255,255,255,0.1)] flex items-center justify-center bg-[#10B981] text-[#003824] font-bold"
						>
							<span class="z-0">{user.name ? user.name.slice(0, 2).toUpperCase() : ''}</span>
							{#if user.avatar}
								<img
									src={user.avatar}
									alt={user.name}
									class="absolute inset-0 z-10 size-full object-cover bg-background"
									onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
								/>
							{/if}
						</div>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-semibold text-white">{user.name}</span>
							<span class="truncate text-xs text-[#94A3B8]">{user.email}</span>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator class="bg-[rgba(255,255,255,0.08)]" />
				<DropdownMenu.Group>
					<a href="/dashboard/profile">
						<DropdownMenu.Item
							class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:!bg-[rgba(255,255,255,0.08)] hover:!text-white"
							><BadgeCheck class="size-4" />Profile</DropdownMenu.Item
						>
					</a>
					<DropdownMenu.Item
						class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:!bg-[rgba(255,255,255,0.08)] hover:!text-white"
						onclick={() => isSettingsOpen = true}
						><Settings2 class="size-4" /> User Settings</DropdownMenu.Item
					>
					<a class="" href="/dashboard/notification">
						<DropdownMenu.Item
							class="cursor-pointer rounded-lg p-2 flex items-center gap-2.5 hover:!bg-[rgba(255,255,255,0.08)] hover:!text-white"
							><Bell class="size-4" /> Notifications
						</DropdownMenu.Item></a
					>
				</DropdownMenu.Group>
				<DropdownMenu.Separator class="bg-[rgba(255,255,255,0.08)]" />
				<DropdownMenu.Item
					onclick={handleLogout}
					class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg text-[#ffb4ab] hover:!text-[#ffb4ab] hover:!bg-[#93000a]/30"
					><LogOut class="size-4" /> Log out</DropdownMenu.Item
				>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Footer>
</Sidebar.Root>

<UserSettingsModal bind:open={isSettingsOpen} />
