<script lang="ts">
	import {
		ExternalLink,
		Pencil,
		Trash2,
		MoreVertical,
		Wrench,
		FolderTree,
		Shield
	} from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';

	let {
		sp,
		canManage = true,
		onEdit,
		onDelete,
		onMaintenance,
		onGroups,
		onEmbedBadge
	}: {
		sp: StatusPage;
		canManage?: boolean;
		onEdit?: (sp: StatusPage) => void;
		onDelete?: (sp: StatusPage) => void;
		onMaintenance?: (sp: StatusPage) => void;
		onGroups?: (sp: StatusPage) => void;
		onEmbedBadge?: (sp: StatusPage) => void;
	} = $props();

	function handleEdit() {
		onEdit?.(sp);
	}

	function handleDelete() {
		onDelete?.(sp);
	}

	function handleMaintenance() {
		onMaintenance?.(sp);
	}

	function handleGroups() {
		onGroups?.(sp);
	}

	function handleEmbedBadge() {
		onEmbedBadge?.(sp);
	}
</script>

<div
	class="grid grid-cols-1 md:grid-cols-[minmax(160px,1.5fr)_minmax(180px,2fr)_1.2fr_90px_100px_80px] gap-4 px-4 py-2 items-center hover:bg-white/3 transition-colors"
>
	<!-- 1. Status Page Title -->
	<div class="flex items-center min-w-0">
		<span class="text-sm font-medium text-foreground truncate">{sp.Title}</span>
	</div>

	<!-- 2. Description -->
	<div class="min-w-0">
		<span class="text-xs text-muted-foreground truncate block" title={sp.Description || ''}>
			{sp.Description?.trim() ? sp.Description : '-'}
		</span>
	</div>

	<!-- 3. Subdomain -->
	<div class="min-w-0">
		<span class="text-xs font-mono text-muted-foreground truncate block">
			{sp.Subdomain}
		</span>
	</div>

	<!-- 4. Visibility -->
	<div>
		{#if sp.IsPublic}
			<span
				class="text-[10px] font-mono px-1.5 py-px bg-emerald-500/10 text-emerald-400 rounded-sm uppercase tracking-wider font-medium"
				>Public</span
			>
		{:else}
			<span
				class="text-[10px] font-mono px-1.5 py-px bg-white/5 text-muted-foreground rounded-sm uppercase tracking-wider font-medium"
				>Private</span
			>
		{/if}
	</div>

	<!-- 5. Monitors -->
	<div class="text-xs text-muted-foreground font-mono">
		{sp.StatusPageMonitors?.length || 0} monitors
	</div>

	<!-- 6. Actions: View + 3-dot DropdownMenu -->
	<div class="flex items-center justify-end gap-1.5">
		<a
			href={`/status/${sp.Subdomain}`}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center justify-center size-7 rounded-sm bg-white/5 hover:bg-white/10 hover:text-primary transition-colors text-muted-foreground cursor-pointer"
			aria-label="View status page"
		>
			<ExternalLink class="size-3.5" />
		</a>

		{#if (canManage && (onEdit || onDelete || onMaintenance || onGroups)) || onEmbedBadge}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center justify-center size-7 rounded-sm bg-white/5 hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
					aria-label="More options"
				>
					<MoreVertical class="size-3.5" />
				</DropdownMenu.Trigger>

				<DropdownMenu.Content align="end" class="w-44 bg-card border-border shadow-xl z-50">
					{#if onEmbedBadge}
						<DropdownMenu.Item
							onclick={handleEmbedBadge}
							class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
						>
							<Shield class="size-3.5 text-emerald-400" />
							Embed Badge
						</DropdownMenu.Item>
					{/if}
					{#if canManage && onGroups}
						<DropdownMenu.Item
							onclick={handleGroups}
							class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
						>
							<FolderTree class="size-3.5 text-emerald-400" />
							Groups
						</DropdownMenu.Item>
					{/if}
					{#if canManage && onMaintenance}
						<DropdownMenu.Item
							onclick={handleMaintenance}
							class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
						>
							<Wrench class="size-3.5 text-emerald-400" />
							Maintenance
						</DropdownMenu.Item>
					{/if}
					{#if canManage && onEdit}
						<DropdownMenu.Item
							onclick={handleEdit}
							class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
						>
							<Pencil class="size-3.5 text-emerald-400" />
							Edit
						</DropdownMenu.Item>
					{/if}
					{#if canManage && onDelete}
						{#if onEmbedBadge || onGroups || onMaintenance || onEdit}
							<DropdownMenu.Separator />
						{/if}
						<DropdownMenu.Item
							onclick={handleDelete}
							class="flex items-center gap-2.5 px-3 py-2 text-xs text-destructive hover:bg-destructive/10 cursor-pointer"
						>
							<Trash2 class="size-3.5" />
							Delete
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}
	</div>
</div>
