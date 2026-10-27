<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Search, Plus, Globe, Radio, Cpu, UserPlus } from 'lucide-svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	interface Props {
		orgName: string;
		pageTitle: string;
		showSearch?: boolean;
		searchQuery?: string;
		searchPlaceholder?: string;
		actionType?: 'monitor' | 'member' | 'button' | 'custom' | null;
		actionLabel?: string;
		onAction?: ((type?: any) => void) | null;
	}

	let {
		orgName,
		pageTitle,
		showSearch = false,
		searchQuery = $bindable(''),
		searchPlaceholder = 'Search...',
		actionType = null,
		actionLabel = '',
		onAction = null
	}: Props = $props();

	let searchInputRef = $state<HTMLInputElement | null>(null);

	onMount(() => {
		function handleKeyDown(e: KeyboardEvent) {
			if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
				if (showSearch && searchInputRef) {
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

<header
	class="h-14 border-b border-border bg-card px-4 flex items-center justify-between sticky top-0 z-20 shrink-0 gap-3"
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
						class="truncate max-w-45 sm:max-w-70"
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
		{#if showSearch}
			<div class="relative w-48 sm:w-60 md:w-72">
				<Search
					class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground/50 pointer-events-none"
				/>
				<input
					bind:this={searchInputRef}
					type="search"
					placeholder={searchPlaceholder}
					bind:value={searchQuery}
					class="w-full text-xs bg-foreground/5 border border-border rounded-lg pl-8 pr-12 py-1.5 text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-colors"
				/>
				<kbd
					class="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded border border-border pointer-events-none hidden sm:inline-block"
					>⌘K</kbd
				>
			</div>
		{/if}

		<!-- Action CTA: Monitor Dropdown -->
		{#if actionType === 'monitor'}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="inline-flex items-center gap-1.5 bg-primary/10 hover:bg-primary/15 text-primary border border-primary/25 dark:bg-transparent dark:hover:bg-white/5 dark:text-emerald-400 dark:border-border px-3 py-1.5 rounded-lg font-semibold text-xs transition-colors active:scale-95 cursor-pointer focus:outline-none"
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
						onclick={() => onAction?.('http')}
					>
						<Globe class="mr-2 size-3.5 text-emerald-400" /> HTTP/API
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
						onclick={() => onAction?.('ping')}
					>
						<Radio class="mr-2 size-3.5 text-emerald-400" /> Ping
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="hover:bg-white/10 cursor-pointer rounded-md p-2 text-xs"
						onclick={() => onAction?.('tcp')}
					>
						<Cpu class="mr-2 size-3.5 text-emerald-400" /> TCP (Port)
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{:else if actionType === 'member'}
			<Button
				class="bg-primary/10 hover:bg-primary/15 text-primary border border-primary/25 dark:bg-transparent dark:hover:bg-white/5 dark:text-emerald-400 dark:border-border h-8 px-3.5 font-semibold rounded-lg text-xs gap-1.5 active:scale-95 transition-all cursor-pointer"
				onclick={() => onAction?.()}
			>
				<UserPlus class="size-3.5 stroke-[2.5]" />
				<span class="hidden sm:inline">Invite member</span>
			</Button>
		{:else if (actionType === 'button' || !actionType) && actionLabel}
			<Button
				class="bg-primary/10 hover:bg-primary/15 text-primary border border-primary/25 dark:bg-transparent dark:hover:bg-white/5 dark:text-emerald-400 dark:border-border h-8 px-3.5 font-semibold rounded-lg text-xs gap-1.5 active:scale-95 transition-all cursor-pointer"
				onclick={() => onAction?.()}
			>
				<Plus class="size-3.5 stroke-[2.5]" />
				<span>{actionLabel}</span>
			</Button>
		{/if}
	</div>
</header>
