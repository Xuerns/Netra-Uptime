<script lang="ts">
	import { Globe, ExternalLink, Building2, Pencil, Trash2, MoreVertical, Wrench, FolderTree, Shield } from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

	let {
		sp,
		onEdit,
		onDelete,
		onMaintenance,
		onGroups,
		onEmbedBadge
	}: {
		sp: StatusPage;
		onEdit?: (sp: StatusPage) => void;
		onDelete?: (sp: StatusPage) => void;
		onMaintenance?: (sp: StatusPage) => void;
		onGroups?: (sp: StatusPage) => void;
		onEmbedBadge?: (sp: StatusPage) => void;
	} = $props();

	let menuOpen = $state(false);

	function toggleMenu(e: MouseEvent) {
		e.stopPropagation();
		menuOpen = !menuOpen;
	}

	function closeMenu() {
		menuOpen = false;
	}

	function handleEdit(e: MouseEvent) {
		e.stopPropagation();
		closeMenu();
		onEdit?.(sp);
	}

	function handleDelete(e: MouseEvent) {
		e.stopPropagation();
		closeMenu();
		onDelete?.(sp);
	}

	function handleMaintenance(e: MouseEvent) {
		e.stopPropagation();
		closeMenu();
		onMaintenance?.(sp);
	}

	function handleGroups(e: MouseEvent) {
		e.stopPropagation();
		closeMenu();
		onGroups?.(sp);
	}

	function handleEmbedBadge(e: MouseEvent) {
		e.stopPropagation();
		closeMenu();
		onEmbedBadge?.(sp);
	}
</script>

<svelte:window
	onclick={() => {
		if (menuOpen) closeMenu();
	}}
/>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="card-wrapper flex flex-col bg-card border border-border rounded-xl p-5 hover:border-border/80 hover:bg-card/80 transition-colors shadow-sm h-full w-full relative {menuOpen ? 'z-30' : ''}"
	onclick={closeMenu}
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
			<!-- Public/Private badge -->
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

			<!-- 3-dot action menu -->
			{#if onEdit || onDelete || onMaintenance || onGroups || onEmbedBadge}
				<div class="relative">
					<button
						onclick={toggleMenu}
						class="flex items-center justify-center size-8 rounded-lg bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors"
						aria-label="More options"
					>
						<MoreVertical class="size-4" />
					</button>

					{#if menuOpen}
						<div
							class="dropdown-menu absolute right-0 top-full mt-1.5 w-44 z-50 bg-card border border-border rounded-xl shadow-xl overflow-hidden"
						>
							{#if onEmbedBadge}
								<button
									onclick={handleEmbedBadge}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
								>
									<Shield class="size-3.5 text-emerald-400" />
									Embed Badge
								</button>
							{/if}
							{#if onGroups}
								<button
									onclick={handleGroups}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
								>
									<FolderTree class="size-3.5 text-emerald-400" />
									Groups
								</button>
							{/if}
							{#if onMaintenance}
								<button
									onclick={handleMaintenance}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
								>
									<Wrench class="size-3.5 text-emerald-400" />
									Maintenance
								</button>
							{/if}
							{#if onEdit}
								<button
									onclick={handleEdit}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
								>
									<Pencil class="size-3.5 text-emerald-400" />
									Edit
								</button>
							{/if}
							{#if onDelete}
								<button
									onclick={handleDelete}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-destructive hover:bg-destructive/10 transition-colors text-left"
								>
									<Trash2 class="size-3.5" />
									Delete
								</button>
							{/if}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Organization Name Display -->
	{#if sp.Organization}
		<div
			class="flex items-center gap-1.5 text-[12px] text-muted-foreground/80 mb-3 bg-white/5 w-fit px-2 py-1 rounded-md"
		>
			<Building2 class="size-3" />
			<span class="font-medium truncate max-w-50">{sp.Organization.Name}</span>
		</div>
	{/if}

	<p class="text-[13px] text-muted-foreground/80 line-clamp-2 flex-1 mb-4">
		{sp.Description || 'No description provided.'}
	</p>

	<div class="flex items-center justify-between mt-auto pt-4 border-t border-border">
		<div class="text-[12px] text-muted-foreground/60 flex items-center gap-1.5 font-medium">
			<span>{sp.StatusPageMonitors?.length || 0} monitors</span>
		</div>

		<a
			href={`/status/${sp.Subdomain}`}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center gap-1.5 text-[12px] font-medium text-primary hover:text-primary/80 transition-colors"
		>
			View Page <ExternalLink class="size-3" />
		</a>
	</div>
</div>

<style>
	.dropdown-menu {
		animation: dropdown-in 0.12s ease;
	}

	@keyframes dropdown-in {
		from {
			opacity: 0;
			transform: translateY(-6px) scale(0.97);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	.menu-item:first-child {
		border-radius: 0;
	}

	.menu-item:not(:last-child) {
		border-bottom: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
	}
</style>
