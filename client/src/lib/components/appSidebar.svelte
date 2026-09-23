<script lang="ts">
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';

	// Icons
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import Building2 from 'lucide-svelte/icons/building-2';
	import LogOut from 'lucide-svelte/icons/log-out';
	import Bell from 'lucide-svelte/icons/bell';
	import BadgeCheck from 'lucide-svelte/icons/badge-check';
	import BookOpen from 'lucide-svelte/icons/book-open';
	import ExternalLink from 'lucide-svelte/icons/external-link';
	import Settings2 from 'lucide-svelte/icons/settings-2';
	import { env } from '$env/dynamic/public';

	import Logo from '$lib/assets/Logo.svg';
	import { apiFetch } from '$lib/helper/api';
	import UserSettingsModal from '$lib/components/UserSettingsModal.svelte';

	let isSettingsOpen = $state(false);

	let activeTeam = $state<{ name: string; logo: any }>({
		name: 'PulseWatch Workspace',
		logo: Logo
	});

	const navWorkspace = [{ title: 'Organizations', url: '/dashboard/org', icon: Building2 }];

	const navResources = [
		{
			title: 'Documentation',
			url: 'https://pulsewatch-dev.dmmaulana06.workers.dev/docs',
			icon: BookOpen,
			isExternal: true
		}
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
				window.location.href = '/auth/login';
			}
		} catch (err: any) {
			console.error(err.message);
		}
	}
</script>

<Sidebar.Root
	collapsible="icon"
	class="bg-sidebar! !border-r !border-sidebar-border text-sidebar-foreground"
>
	<!-- Header: Team Switcher -->
	<Sidebar.Header class="p-3 border-b border-sidebar-border">
		<div
			class="flex items-center gap-3 p-2 rounded-lg transition-all duration-300 text-left w-full border border-transparent group group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
		>
			{#if typeof activeTeam.logo === 'string'}
				<img src={activeTeam.logo} alt="Logo" class="size-8 object-contain" />
			{:else}
				<activeTeam.logo class="size-4" />
			{/if}

			<div class="grid flex-1 text-left text-lg leading-tight group-data-[collapsible=icon]:hidden">
				<span class="truncate font-semibold text-sidebar-foreground tracking-tight"
					>{activeTeam.name}</span
				>
			</div>
		</div>
	</Sidebar.Header>

	<!-- Content: Main Navigation -->
	<Sidebar.Content class="py-2">
		<!-- Workspace Group -->
		<Sidebar.Group class="py-1">
			<Sidebar.GroupContent>
				<Sidebar.Menu class="space-y-1">
					{#each navWorkspace as item}
						{@const isActive =
							$page.url.pathname === item.url || $page.url.pathname.startsWith(item.url + '/')}
						<Sidebar.MenuItem
							class="group/item relative flex items-center justify-between rounded-lg border border-transparent hover:border-sidebar-border transition-all duration-300 {isActive
								? 'bg-sidebar-accent'
								: 'hover:bg-sidebar-accent/60'}"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 {isActive
									? 'text-sidebar-foreground!'
									: 'text-muted-foreground! hover:text-sidebar-foreground!'}"
							>
								{#snippet child({ props })}
									<a
										href={item.url}
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<item.icon
											class="size-4.5 shrink-0 transition-colors {isActive
												? 'text-sidebar-foreground'
												: 'text-muted-foreground group-hover/item:text-sidebar-foreground'}"
										/>
										<span class="text-md font-medium flex-1 group-data-[collapsible=icon]:hidden"
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

		<Sidebar.Separator class="my-1.5" />

		<!-- Resources Group -->
		<Sidebar.Group class="py-1">
			<Sidebar.GroupContent>
				<Sidebar.Menu class="space-y-1">
					{#each navResources as item}
						<Sidebar.MenuItem
							class="group/item relative flex items-center justify-between rounded-lg border border-transparent hover:border-sidebar-border transition-all duration-300 hover:bg-sidebar-accent/60"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 text-muted-foreground! hover:text-sidebar-foreground!"
							>
								{#snippet child({ props })}
									<a
										href={item.url}
										target="_blank"
										rel="noopener noreferrer"
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<item.icon
											class="size-4.5 shrink-0 transition-colors text-muted-foreground group-hover/item:text-sidebar-foreground"
										/>
										<span class="text-md font-medium flex-1 group-data-[collapsible=icon]:hidden"
											>{item.title}</span
										>
										<ExternalLink
											class="size-3.5 text-muted-foreground group-hover/item:text-sidebar-foreground transition-colors group-data-[collapsible=icon]:hidden"
										/>
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
	<Sidebar.Footer class="p-3 border-t border-sidebar-border">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="w-full focus:outline-none">
				<div
					class="flex items-center gap-3 p-2 hover:bg-sidebar-accent/60 rounded-lg transition-all duration-300 text-left w-full border border-transparent hover:border-sidebar-border group group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
				>
					<div
						class="relative size-9 rounded-lg overflow-hidden shrink-0 border border-primary/20 flex items-center justify-center bg-primary text-primary-foreground font-bold"
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
						<span
							class="truncate font-semibold text-sidebar-foreground group-hover:text-sidebar-foreground transition-colors"
							>{user.name}</span
						>
						<span
							class="truncate text-xs text-muted-foreground group-hover:text-muted-foreground transition-colors"
							>{user.email}</span
						>
					</div>
					<ChevronsUpDown
						class="ml-auto size-4 text-muted-foreground shrink-0 group-hover:text-sidebar-foreground transition-colors group-data-[collapsible=icon]:hidden"
					/>
				</div>
			</DropdownMenu.Trigger>

			<DropdownMenu.Content
				class="w-56 !bg-popover backdrop-blur-[20px] !border-border text-popover-foreground"
				side={sidebar.state === 'collapsed' ? 'right' : 'top'}
				align="end"
				sideOffset={4}
			>
				<DropdownMenu.Label class="font-normal">
					<div class="flex items-center gap-3 px-1.5 py-2 text-left text-sm">
						<div
							class="relative size-9 rounded-lg overflow-hidden shrink-0 border border-primary/20 flex items-center justify-center bg-primary text-primary-foreground font-bold"
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
							<span class="truncate font-semibold text-popover-foreground">{user.name}</span>
							<span class="truncate text-xs text-muted-foreground">{user.email}</span>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator class="bg-border" />
				<DropdownMenu.Group>
					<a href="/dashboard/profile">
						<DropdownMenu.Item
							class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:!bg-accent hover:!text-accent-foreground"
							><BadgeCheck class="size-4" />Profile</DropdownMenu.Item
						>
					</a>
					<DropdownMenu.Item
						class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:!bg-accent hover:!text-accent-foreground"
						onclick={() => (isSettingsOpen = true)}
						><Settings2 class="size-4" /> User Settings</DropdownMenu.Item
					>
					<a class="" href="/dashboard/notification">
						<DropdownMenu.Item
							class="cursor-pointer rounded-lg p-2 flex items-center gap-2.5 hover:!bg-accent hover:!text-accent-foreground"
							><Bell class="size-4" /> Notifications
						</DropdownMenu.Item></a
					>
				</DropdownMenu.Group>
				<DropdownMenu.Separator class="bg-border" />
				<DropdownMenu.Item
					onclick={handleLogout}
					class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg text-destructive hover:!text-destructive hover:!bg-destructive/10"
					><LogOut class="size-4" /> Log out</DropdownMenu.Item
				>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Footer>
</Sidebar.Root>

<UserSettingsModal bind:open={isSettingsOpen} />
