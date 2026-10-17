<script lang="ts">
	import {
		Globe,
		ExternalLink,
		Building2,
		Pencil,
		Trash2,
		MoreVertical,
		Wrench,
		FolderTree,
		Shield
	} from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

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

{#if viewMode === 'grid'}
	<!-- Card / Grid View Mode -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="card-wrapper flex flex-col bg-card border border-border rounded-xl p-5 hover:border-border/80 hover:bg-card/80 transition-colors shadow-sm h-full w-full relative {menuOpen
			? 'z-30'
			: ''}"
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
					<div class="relative">
						<button
							onclick={toggleMenu}
							class="flex items-center justify-center size-8 rounded-lg bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
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
										class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
									>
										<Shield class="size-3.5 text-emerald-400" />
										Embed Badge
									</button>
								{/if}
								{#if onGroups}
									<button
										onclick={handleGroups}
										class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
									>
										<FolderTree class="size-3.5 text-emerald-400" />
										Groups
									</button>
								{/if}
								{#if onMaintenance}
									<button
										onclick={handleMaintenance}
										class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
									>
										<Wrench class="size-3.5 text-emerald-400" />
										Maintenance
									</button>
								{/if}
								{#if onEdit}
									<button
										onclick={handleEdit}
										class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
									>
										<Pencil class="size-3.5 text-emerald-400" />
										Edit
									</button>
								{/if}
								{#if onDelete}
									<button
										onclick={handleDelete}
										class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-destructive hover:bg-destructive/10 transition-colors text-left cursor-pointer"
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
				class="flex items-center gap-1.5 text-[12px] font-medium text-primary hover:text-primary/80 transition-colors cursor-pointer"
			>
				View Page <ExternalLink class="size-3" />
			</a>
		</div>
	</div>
{:else}
	<!-- List / Row View Mode -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="grid grid-cols-1 md:grid-cols-[minmax(200px,2fr)_1.5fr_100px_110px_80px] gap-4 px-5 py-3.5 items-center hover:bg-white/3 transition-colors relative {menuOpen
			? 'z-30'
			: ''}"
		onclick={closeMenu}
	>
		<!-- 1. Status Page Title & Org -->
		<div class="flex flex-col min-w-0">
			<div class="flex items-center gap-2.5 min-w-0">
				<div
					class="size-7 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0"
				>
					<Globe class="size-3.5" />
				</div>
				<span class="text-sm font-medium text-foreground truncate">{sp.Title}</span>
			</div>
			{#if sp.Organization}
				<div class="flex items-center gap-1 text-[11px] text-muted-foreground/60 mt-1 pl-9.5">
					<Building2 class="size-3" />
					<span class="truncate max-w-50">{sp.Organization.Name}</span>
				</div>
			{/if}
		</div>

		<!-- 2. Subdomain -->
		<div class="min-w-0">
			<span class="text-xs font-mono text-muted-foreground truncate block">
				{sp.Subdomain}
			</span>
		</div>

		<!-- 3. Visibility -->
		<div>
			{#if sp.IsPublic}
				<span
					class="text-[10px] font-mono px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-full uppercase tracking-wider"
					>Public</span
				>
			{:else}
				<span
					class="text-[10px] font-mono px-2 py-0.5 bg-white/5 text-muted-foreground border border-border rounded-full uppercase tracking-wider"
					>Private</span
				>
			{/if}
		</div>

		<!-- 4. Monitors -->
		<div class="text-xs text-muted-foreground font-mono">
			{sp.StatusPageMonitors?.length || 0} monitors
		</div>

		<!-- 5. Actions: View + 3-dot menu -->
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
				<div class="relative">
					<button
						onclick={toggleMenu}
						class="flex items-center justify-center size-7 rounded bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
						aria-label="More options"
					>
						<MoreVertical class="size-3.5" />
					</button>

					{#if menuOpen}
						<div
							class="dropdown-menu absolute right-0 top-full mt-1.5 w-44 z-50 bg-card border border-border rounded-xl shadow-xl overflow-hidden"
						>
							{#if onEmbedBadge}
								<button
									onclick={handleEmbedBadge}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
								>
									<Shield class="size-3.5 text-emerald-400" />
									Embed Badge
								</button>
							{/if}
							{#if onGroups}
								<button
									onclick={handleGroups}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
								>
									<FolderTree class="size-3.5 text-emerald-400" />
									Groups
								</button>
							{/if}
							{#if onMaintenance}
								<button
									onclick={handleMaintenance}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
								>
									<Wrench class="size-3.5 text-emerald-400" />
									Maintenance
								</button>
							{/if}
							{#if onEdit}
								<button
									onclick={handleEdit}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left cursor-pointer"
								>
									<Pencil class="size-3.5 text-emerald-400" />
									Edit
								</button>
							{/if}
							{#if onDelete}
								<button
									onclick={handleDelete}
									class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-destructive hover:bg-destructive/10 transition-colors text-left cursor-pointer"
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
{/if}

<style>

	.menu-item {
		background: none;
		border: none;
		outline: none;
	}

	.dropdown-menu {
		animation: fadeIn 0.15s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
</style>
