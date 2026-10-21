<script lang="ts">
	import {
		Globe,
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
		viewMode = 'grid',
		onEdit,
		onDelete,
		onMaintenance,
		onGroups,
		onEmbedBadge
	}: {
		sp: StatusPage;
		viewMode?: 'grid' | 'list';
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

{#if viewMode === 'grid'}
	<!-- Card / Grid View Mode -->
	<div
		class="card-wrapper flex flex-col bg-card border border-border rounded-xl p-5 hover:border-primary/30 transition-colors shadow-sm h-full w-full relative"
	>
		<!-- Top row: icon + title + badge + 3-dot menu -->
		<div class="flex items-start justify-between gap-3 mb-3">
			<div class="flex items-center gap-3 min-w-0">
				<div
					class="size-9 rounded-lg bg-white/5 border border-border flex items-center justify-center shrink-0"
				>
					<Globe class="size-4 text-primary" />
				</div>
				<div class="min-w-0">
					<h3 class="text-base font-semibold text-foreground truncate leading-tight">{sp.Title}</h3>
					<span class="text-xs text-muted-foreground font-mono truncate block mt-0.5"
						>{sp.Subdomain}</span
					>
				</div>
			</div>

			<div class="flex items-center gap-2 shrink-0">
				{#if sp.IsPublic}
					<span
						class="text-[10px] font-mono px-2 py-0.5 bg-primary/10 text-primary border border-primary/20 rounded-full uppercase tracking-wider font-medium"
						>Public</span
					>
				{:else}
					<span
						class="text-[10px] font-mono px-2 py-0.5 bg-white/10 text-muted-foreground border border-border rounded-full uppercase tracking-wider font-medium"
						>Private</span
					>
				{/if}

				{#if onEdit || onDelete || onMaintenance || onGroups || onEmbedBadge}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="flex items-center justify-center size-8 rounded-lg bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
							aria-label="More options"
						>
							<MoreVertical class="size-4" />
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
							{#if onGroups}
								<DropdownMenu.Item
									onclick={handleGroups}
									class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
								>
									<FolderTree class="size-3.5 text-emerald-400" />
									Groups
								</DropdownMenu.Item>
							{/if}
							{#if onMaintenance}
								<DropdownMenu.Item
									onclick={handleMaintenance}
									class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
								>
									<Wrench class="size-3.5 text-emerald-400" />
									Maintenance
								</DropdownMenu.Item>
							{/if}
							{#if onEdit}
								<DropdownMenu.Item
									onclick={handleEdit}
									class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
								>
									<Pencil class="size-3.5 text-emerald-400" />
									Edit
								</DropdownMenu.Item>
							{/if}
							{#if onDelete}
								<DropdownMenu.Separator />
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

		<p class="text-[13px] text-muted-foreground/80 line-clamp-2 flex-1 mb-4">
			{sp.Description?.trim() ? sp.Description : '-'}
		</p>

		<div class="flex items-center justify-between mt-auto pt-4 border-t border-border">
			<div class="text-[12px] text-muted-foreground/60 flex items-center gap-1.5 font-medium">
				<span>{sp.StatusPageMonitors?.length || 0} monitors</span>
			</div>

			<a
				href={`/status/${sp.Subdomain}`}
				target="_blank"
				rel="noopener noreferrer"
				class="flex items-center gap-1.5 text-[12px] font-medium text-primary hover:text-primary/80 transition-colors cursor-pointer"
			>
				View Page <ExternalLink class="size-3" />
			</a>
		</div>
	</div>
{:else}
	<!-- List / Row View Mode -->
	<div
		class="grid grid-cols-1 md:grid-cols-[minmax(160px,1.5fr)_minmax(180px,2fr)_1.2fr_90px_100px_80px] gap-4 px-5 py-2.5 items-center hover:bg-white/3 transition-colors"
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
					class="text-[10px] font-mono px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full uppercase tracking-wider font-medium"
					>Public</span
				>
			{:else}
				<span
					class="text-[10px] font-mono px-2 py-0.5 bg-white/5 text-muted-foreground border border-border rounded-full uppercase tracking-wider font-medium"
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
				class="flex items-center justify-center size-7 rounded bg-white/5 border border-border hover:bg-white/10 hover:text-primary transition-colors text-muted-foreground cursor-pointer"
				aria-label="View status page"
			>
				<ExternalLink class="size-3.5" />
			</a>

			{#if onEdit || onDelete || onMaintenance || onGroups || onEmbedBadge}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="flex items-center justify-center size-7 rounded bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
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
						{#if onGroups}
							<DropdownMenu.Item
								onclick={handleGroups}
								class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
							>
								<FolderTree class="size-3.5 text-emerald-400" />
								Groups
							</DropdownMenu.Item>
						{/if}
						{#if onMaintenance}
							<DropdownMenu.Item
								onclick={handleMaintenance}
								class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
							>
								<Wrench class="size-3.5 text-emerald-400" />
								Maintenance
							</DropdownMenu.Item>
						{/if}
						{#if onEdit}
							<DropdownMenu.Item
								onclick={handleEdit}
								class="flex items-center gap-2.5 px-3 py-2 text-xs text-foreground hover:bg-white/10 cursor-pointer"
							>
								<Pencil class="size-3.5 text-emerald-400" />
								Edit
							</DropdownMenu.Item>
						{/if}
						{#if onDelete}
							<DropdownMenu.Separator />
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
{/if}
