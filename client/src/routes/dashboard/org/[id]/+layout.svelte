<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import OrgSidebar from '$lib/components/orgSidebar.svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { headerStore } from '$lib/stores/headerStore.svelte';
	import { Search, Plus, Globe, Radio, Cpu, UserPlus } from 'lucide-svelte';
	import { fade } from 'svelte/transition';
	import { page } from '$app/stores';
	import { beforeNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { OrgAPI } from '$lib/api/organization';

	let { data, children } = $props();

	let hasUnreadNotifications = $state(false);
	let searchInputRef = $state<HTMLInputElement | null>(null);

	const orgName = $derived(data?.org?.Name || data?.org?.name || 'Workspace');

	const pageTitle = $derived.by(() => {
		if (headerStore.title) return headerStore.title;
		const path = $page.url.pathname;
		if (path.includes('/monitors')) return 'Monitors';
		if (path.includes('/alert-templates')) return 'Alert Templates';
		if (path.includes('/members')) return 'Members';
		if (path.includes('/incidents')) return 'Incidents';
		if (path.includes('/statuspage')) return 'Status Pages';
		if (path.includes('/logs')) return 'Logs';
		if (path.includes('/notifications')) return 'Notifications';
		if (path.includes('/api-keys')) return 'API Keys';
		if (path.includes('/settings')) return 'Settings';
		return 'Dashboard';
	});

	// Reset header state before each navigation so the new page starts clean.
	// This fires synchronously before the old page unmounts, avoiding the
	// race condition where onDestroy reads the already-updated $page store.
	beforeNavigate(() => {
		headerStore.reset();
	});

	onMount(() => {
		async function checkInvitations() {
			try {
				const res = await OrgAPI.getInvitations();
				const invitations = res.data || res.message || res || [];
				hasUnreadNotifications = invitations.some((inv: any) => inv.Status === 'pending');
			} catch (e) {
				// ignore
			}
		}
		checkInvitations();

		function handleKeyDown(e: KeyboardEvent) {
			if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
				if (headerStore.showSearch && searchInputRef) {
					e.preventDefault();
					searchInputRef.focus();
				}
			}
		}

		window.addEventListener('keydown', handleKeyDown);
		return () => {
			window.removeEventListener('keydown', handleKeyDown);
		};
	});
</script>

<Sidebar.Provider>
	<OrgSidebar
		user={data?.user}
		org={data?.org}
		allOrgs={data?.allOrgs ?? []}
		{hasUnreadNotifications}
	/>
	<Sidebar.Inset class="flex flex-col min-h-screen">
		<!-- ── Top Dark Utility Bar ─────────────────────────────────────────── -->
		<header
			class="h-14 border-b border-border bg-card/60 backdrop-blur-md px-4 flex items-center justify-between sticky top-0 z-20 shrink-0 gap-3"
		>
			<!-- Left: Sidebar Trigger, Divider, Breadcrumb & Status Badge -->
			<div class="flex items-center gap-3 min-w-0">
				<Sidebar.Trigger class="-ml-1" />
				<Separator orientation="vertical" class="h-4" />

				<!-- Breadcrumb -->
				<Breadcrumb.Root>
					<Breadcrumb.List class="text-xs">
						<Breadcrumb.Item>
							<Breadcrumb.Link
								href="/dashboard/org/{$page.params.id}"
								class="truncate max-w-[180px] sm:max-w-[280px]"
							>
								{orgName}
							</Breadcrumb.Link>
						</Breadcrumb.Item>
						<Breadcrumb.Separator />
						<Breadcrumb.Item>
							<Breadcrumb.Page class="font-semibold text-foreground">
								{pageTitle}
							</Breadcrumb.Page>
						</Breadcrumb.Item>
					</Breadcrumb.List>
				</Breadcrumb.Root>
			</div>

			<!-- Right: Search Input & Action CTA -->
			<div class="flex items-center gap-3 shrink-0">
				<!-- Search Filter Input (If active) -->
				{#if headerStore.showSearch}
					<div class="relative w-48 sm:w-60 md:w-72">
						<Search
							class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/50 pointer-events-none"
						/>
						<input
							bind:this={searchInputRef}
							type="search"
							placeholder={headerStore.searchPlaceholder}
							bind:value={headerStore.searchQuery}
							class="w-full text-xs bg-foreground/5 border border-border rounded-lg pl-8 pr-12 py-1.5 text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-colors"
						/>
						<kbd
							class="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded border border-border pointer-events-none hidden sm:inline-block"
							>⌘K</kbd
						>
					</div>
				{/if}

				<!-- Action CTA: Monitor Dropdown -->
				{#if headerStore.actionType === 'monitor'}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="inline-flex items-center gap-1.5 bg-transparent hover:bg-white/5 text-emerald-400 border border-border px-3 py-1.5 rounded-lg font-semibold text-xs transition-colors active:scale-95 cursor-pointer focus:outline-none"
						>
							<Plus class="size-3.5 stroke-[2.5]" />
							<span class="hidden sm:inline">Add monitor</span>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content
							align="end"
							class="bg-card border-border text-foreground min-w-[160px] p-1.5 rounded-lg"
						>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
								onclick={() => headerStore.onAction?.('http')}
							>
								<Globe class="mr-2 size-3.5 text-emerald-400" /> HTTP/API
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
								onclick={() => headerStore.onAction?.('ping')}
							>
								<Radio class="mr-2 size-3.5 text-emerald-400" /> Ping
							</DropdownMenu.Item>
							<DropdownMenu.Item
								class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
								onclick={() => headerStore.onAction?.('tcp')}
							>
								<Cpu class="mr-2 size-3.5 text-emerald-400" /> TCP (Port)
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{:else if headerStore.actionType === 'member'}
					<Button
						class="bg-transparent hover:bg-white/5 text-emerald-400 border border-border h-8 px-3.5 font-semibold rounded-lg text-xs gap-1.5 active:scale-95 transition-all cursor-pointer"
						onclick={() => headerStore.onAction?.()}
					>
						<UserPlus class="size-3.5 stroke-[2.5]" />
						<span class="hidden sm:inline">Invite member</span>
					</Button>
				{:else if headerStore.actionType === 'button' && headerStore.actionLabel}
					<Button
						class="bg-transparent hover:bg-white/5 text-emerald-400 border border-border h-8 px-3.5 font-semibold rounded-lg text-xs gap-1.5 active:scale-95 transition-all cursor-pointer"
						onclick={() => headerStore.onAction?.()}
					>
						<Plus class="size-3.5 stroke-[2.5]" />
						<span>{headerStore.actionLabel}</span>
					</Button>
				{/if}
			</div>
		</header>

		<!-- ── Main View Area (Full width) ─────────────────────────────────── -->
		<div class="flex-1 relative flex min-h-0">
			<main class="flex-1 relative flex flex-col min-h-0 dashboard-bg">
				{#key $page.url.pathname}
					<div
						in:fade={{ duration: 150, delay: 150 }}
						out:fade={{ duration: 150 }}
						class="flex-1 p-6 overflow-y-auto"
					>
						{@render children()}
					</div>
				{/key}
			</main>
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
