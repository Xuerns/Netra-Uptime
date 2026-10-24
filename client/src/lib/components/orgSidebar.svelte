<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';

	// Icons
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import Check from 'lucide-svelte/icons/check';

	// Nav Icons
	import Activity from 'lucide-svelte/icons/activity';
	import AlertTriangle from 'lucide-svelte/icons/alert-triangle';
	import Globe from 'lucide-svelte/icons/globe';
	import LogsIcon from 'lucide-svelte/icons/logs';
	import Settings from 'lucide-svelte/icons/settings';
	import Key from 'lucide-svelte/icons/key';
	import Users from 'lucide-svelte/icons/users';
	import Layers from 'lucide-svelte/icons/layers';

	// Footer Icons
	import LogOut from 'lucide-svelte/icons/log-out';
	import Bell from 'lucide-svelte/icons/bell';
	import BadgeCheck from 'lucide-svelte/icons/badge-check';
	import BookOpen from 'lucide-svelte/icons/book-open';
	import ExternalLink from 'lucide-svelte/icons/external-link';
	import Settings2 from 'lucide-svelte/icons/settings-2';

	import Logo from '$lib/assets/Logo.svg';
	import { AuthAPI } from '$lib/api/auth';
	import UserSettingsModal from '$lib/components/UserSettingsModal.svelte';
	import { isSelfHosted } from '$lib/utils/appMode';
	import { activeOrgId } from '$lib/stores/organization';
	import { getInitials as initials } from '$lib/utils/string';
	import { toast } from 'svelte-sonner';
	import { parseApiError } from '$lib/utils/parseApiError';

	let isSettingsOpen = $state(false);

	interface OrgItem {
		ID: string;
		Name: string;
		Slug: string;
	}

	let {
		user = {
			name: '',
			email: '',
			avatar: ''
		},
		org,
		allOrgs = [],
		hasUnreadNotifications = false
	} = $props<{
		user: { name: string; email: string; avatar?: string };
		org: OrgItem;
		allOrgs: OrgItem[];
		hasUnreadNotifications?: boolean;
	}>();

	const orgId = $derived($page.params.id);

	const navMonitoring = [
		{ title: 'Monitors', icon: Activity, segment: 'monitors' },
		{ title: 'Alert Templates', icon: Layers, segment: 'alert-templates' },
		{ title: 'Incidents', icon: AlertTriangle, segment: 'incidents' },
		{ title: 'Status Pages', icon: Globe, segment: 'statuspage' }
	];

	const navOrganization = [
		{ title: 'Members', icon: Users, segment: 'members' },
		{ title: 'Logs', icon: LogsIcon, segment: 'logs' },
		{ title: 'Notifications', icon: Bell, segment: 'notifications' },
		{ title: 'API Keys', icon: Key, segment: 'api-keys' },
		{ title: 'Settings', icon: Settings, segment: 'settings' }
	];

	function navUrl(segment: string): string {
		return `/dashboard/org/${orgId}/${segment}`;
	}

	function isActive(segment: string): boolean {
		return $page.url.pathname.includes(`/org/${orgId}/${segment}`);
	}

	function switchOrg(newOrgId: string) {
		activeOrgId.set(newOrgId);
		if (typeof window !== 'undefined') {
			localStorage.setItem('netra_last_org', newOrgId);
		}
		goto(`/dashboard/org/${newOrgId}/monitors`);
	}

	function viewAllOrganizations() {
		if (typeof window !== 'undefined') {
			localStorage.removeItem('netra_last_org');
		}
		activeOrgId.set(null);
		goto('/dashboard/org');
	}

	const sidebar = Sidebar.useSidebar();

	async function handleLogout(event: Event) {
		event.preventDefault();
		try {
			await AuthAPI.logout();
			goto('/auth/login');
		} catch (err: any) {
			toast.error(parseApiError(err, 'Failed to log out. Please try again.'));
		}
	}
</script>

<Sidebar.Root
	collapsible="icon"
	class="bg-sidebar! border-r! border-sidebar-border! text-sidebar-foreground"
>
	<!-- Header: Org Switcher -->
	<Sidebar.Header class="p-3 border-b border-sidebar-border">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="w-full focus:outline-none">
				<div
					class="flex items-center gap-3 p-2 hover:bg-sidebar-accent/60 rounded-lg transition-all duration-300 text-left w-full border border-transparent hover:border-sidebar-border group group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
				>
					<div
						class="flex items-center justify-center size-8 rounded-md bg-primary/15 text-primary font-mono text-xs font-semibold shrink-0"
					>
						{initials(org?.Name ?? 'O')}
					</div>

					<div
						class="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden"
					>
						<span class="truncate font-semibold text-sidebar-foreground tracking-tight"
							>{org?.Name ?? 'Organization'}</span
						>
						<span class="truncate text-[11px] text-muted-foreground font-mono"
							>{org?.Slug ?? ''}</span
						>
					</div>
					<ChevronsUpDown
						class="ml-auto size-4 text-muted-foreground shrink-0 group-data-[collapsible=icon]:hidden"
					/>
				</div>
			</DropdownMenu.Trigger>

			<DropdownMenu.Content
				class="w-64 bg-popover! backdrop-blur-[20px] border-border! text-popover-foreground"
				side={sidebar.state === 'collapsed' ? 'right' : 'bottom'}
				align="start"
				sideOffset={4}
			>
				<DropdownMenu.Label
					class="text-[11px] text-muted-foreground/60 uppercase tracking-widest font-semibold px-2 py-1.5"
				>
					Switch Organization
				</DropdownMenu.Label>
				<DropdownMenu.Separator class="bg-border" />
				{#each allOrgs as orgItem (orgItem.ID)}
					<DropdownMenu.Item
						class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:bg-accent! hover:text-accent-foreground!"
						onSelect={() => switchOrg(orgItem.ID)}
						onclick={() => switchOrg(orgItem.ID)}
					>
						<div
							class="flex items-center justify-center size-7 rounded-md bg-primary/15 text-primary font-mono text-[10px] font-semibold shrink-0"
						>
							{initials(orgItem.Name)}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-[13px] font-medium text-foreground truncate">{orgItem.Name}</p>
							<p class="text-[11px] text-muted-foreground font-mono truncate">{orgItem.Slug}</p>
						</div>
						{#if orgItem.ID === orgId}
							<Check class="size-4 text-primary shrink-0" />
						{/if}
					</DropdownMenu.Item>
				{/each}
				<DropdownMenu.Separator class="bg-border" />
				<a href="/dashboard/org" onclick={viewAllOrganizations} class="w-full">
					<DropdownMenu.Item
						class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:bg-accent! hover:text-accent-foreground! text-muted-foreground"
						onSelect={viewAllOrganizations}
					>
						<span class="text-[13px]">View all organizations</span>
					</DropdownMenu.Item>
				</a>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Header>

	<!-- Content: Main Navigation -->
	<Sidebar.Content class="py-2">
		<!-- Monitoring Group -->
		<Sidebar.Group class="py-1">
			<Sidebar.GroupContent>
				<Sidebar.Menu class="space-y-1">
					{#each navMonitoring as item (item.title)}
						{@const active = isActive(item.segment)}
						<Sidebar.MenuItem
							class="group/nav relative flex items-center justify-between rounded-lg border border-transparent hover:border-sidebar-border transition-all duration-300 {active
								? 'bg-sidebar-accent'
								: 'hover:bg-sidebar-accent/60'}"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 {active
									? 'text-sidebar-foreground!'
									: 'text-muted-foreground! hover:text-sidebar-foreground!'}"
							>
								{#snippet child({ props })}
									<a
										href={navUrl(item.segment)}
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<item.icon
											class="size-5 shrink-0 transition-colors {active
												? 'text-sidebar-foreground'
												: 'text-muted-foreground group-hover/nav:text-sidebar-foreground'}"
										/>
										<span class="font-medium flex-1 text-md group-data-[collapsible=icon]:hidden"
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

		<!-- Organization Group -->
		<Sidebar.Group class="py-1">
			<Sidebar.GroupContent>
				<Sidebar.Menu class="space-y-1">
					{#each navOrganization as item (item.title)}
						{@const active = isActive(item.segment)}
						<Sidebar.MenuItem
							class="group/item relative flex items-center justify-between rounded-lg border border-transparent hover:border-sidebar-border transition-all duration-300 {active
								? 'bg-sidebar-accent'
								: 'hover:bg-sidebar-accent/60'}"
						>
							<Sidebar.MenuButton
								class="w-full flex-1 h-10 {active
									? 'text-sidebar-foreground!'
									: 'text-muted-foreground! hover:text-sidebar-foreground!'}"
							>
								{#snippet child({ props })}
									<a
										href={navUrl(item.segment)}
										{...props}
										class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
									>
										<item.icon
											class="size-4.5 shrink-0 transition-colors {active
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
					<!-- Docs (External Link) -->
					<Sidebar.MenuItem
						class="group/item relative flex items-center justify-between rounded-lg border border-transparent hover:border-sidebar-border transition-all duration-300 hover:bg-sidebar-accent/60"
					>
						<Sidebar.MenuButton
							class="w-full flex-1 h-10 text-muted-foreground! hover:text-sidebar-foreground!"
						>
							{#snippet child({ props })}
								<a
									href="https://netra-uptime.com/docs"
									target="_blank"
									rel="noopener noreferrer"
									{...props}
									class="flex items-center gap-3.5 w-full px-2.5 py-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0!"
								>
									<BookOpen
										class="size-4.5 shrink-0 transition-colors text-muted-foreground group-hover/item:text-sidebar-foreground"
									/>
									<span class="text-md font-medium flex-1 group-data-[collapsible=icon]:hidden"
										>Documentation</span
									>
									<ExternalLink
										class="size-3.5 text-muted-foreground group-hover/item:text-sidebar-foreground transition-colors group-data-[collapsible=icon]:hidden"
									/>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
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
						{#if hasUnreadNotifications}
							<span
								class="absolute top-1 right-1 size-2 rounded-full bg-primary border border-background z-20"
							></span>
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
				class="w-56 bg-popover! backdrop-blur-[20px] border-border! text-popover-foreground"
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
							class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:bg-accent! hover:text-accent-foreground!"
							><BadgeCheck class="size-4" />Profile</DropdownMenu.Item
						>
					</a>
					<DropdownMenu.Item
						class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg hover:bg-accent! hover:text-accent-foreground!"
						onclick={() => (isSettingsOpen = true)}
						><Settings2 class="size-4" /> User Settings</DropdownMenu.Item
					>
					<a class="" href="/dashboard/notification">
						<DropdownMenu.Item
							class="cursor-pointer rounded-lg p-2 flex items-center gap-2.5 hover:bg-accent! hover:text-accent-foreground!"
							><Bell class="size-4" /> Notifications
						</DropdownMenu.Item></a
					>
				</DropdownMenu.Group>
				<DropdownMenu.Separator class="bg-border" />
				<DropdownMenu.Item
					onclick={handleLogout}
					class="cursor-pointer flex items-center gap-2.5 p-2 rounded-lg text-destructive hover:text-destructive! hover:bg-destructive/10!"
					><LogOut class="size-4" /> Log out</DropdownMenu.Item
				>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Footer>
</Sidebar.Root>

<UserSettingsModal bind:open={isSettingsOpen} />
