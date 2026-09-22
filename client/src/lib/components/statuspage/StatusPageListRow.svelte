<script lang="ts">
	import { Globe, ExternalLink, Building2, Pencil, Trash2, MoreVertical, Wrench, FolderTree } from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

	let {
		sp,
		onEdit,
		onDelete,
		onMaintenance,
		onGroups
	}: {
		sp: StatusPage;
		onEdit?: (sp: StatusPage) => void;
		onDelete?: (sp: StatusPage) => void;
		onMaintenance?: (sp: StatusPage) => void;
		onGroups?: (sp: StatusPage) => void;
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
</script>

<svelte:window
	onclick={() => {
		if (menuOpen) closeMenu();
	}}
/>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="list-row relative {menuOpen ? 'z-30' : ''}" onclick={closeMenu}>
	<div class="flex flex-col min-w-0">
		<div class="flex items-center gap-3 min-w-0">
			<Globe class="size-4 text-muted-foreground shrink-0" />
			<span class="text-sm font-medium text-foreground truncate">{sp.Title}</span>
		</div>
		{#if sp.Organization}
			<div class="flex items-center gap-1 text-[11px] text-muted-foreground/60 mt-1 pl-7">
				<Building2 class="size-3" />
				<span class="truncate max-w-37.5">{sp.Organization.Name}</span>
			</div>
		{/if}
	</div>

	<span class="text-[13px] text-muted-foreground truncate">{sp.Subdomain}</span>

	<div>
		{#if sp.IsPublic}
			<span
				class="text-[10px] font-mono px-2 py-0.5 bg-primary/10 text-primary border border-primary/20 rounded-full uppercase tracking-wider"
				>Public</span
			>
		{:else}
			<span
				class="text-[10px] font-mono px-2 py-0.5 bg-white/10 text-muted-foreground border border-border rounded-full uppercase tracking-wider"
				>Private</span
			>
		{/if}
	</div>

	<span class="text-[13px] text-muted-foreground"
		>{sp.StatusPageMonitors?.length || 0} monitors</span
	>

	<!-- Actions: View + 3-dot menu -->
	<div class="flex items-center gap-1.5">
		<a
			href={`/status/${sp.Subdomain}`}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center justify-center size-7 rounded bg-white/5 border border-border hover:bg-white/10 hover:text-primary transition-colors text-muted-foreground"
			aria-label="View status page"
		>
			<ExternalLink class="size-3.5" />
		</a>

		{#if onEdit || onDelete}
			<div class="relative">
				<button
					onclick={toggleMenu}
					class="flex items-center justify-center size-7 rounded bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors"
					aria-label="More options"
				>
					<MoreVertical class="size-3.5" />
				</button>

				{#if menuOpen}
					<div
						class="dropdown-menu absolute right-0 top-full mt-1.5 w-44 z-50 bg-card border border-border rounded-xl shadow-xl overflow-hidden"
					>
						{#if onGroups}
							<button
								onclick={handleGroups}
								class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
							>
								<FolderTree class="size-3.5 text-purple-400" />
								Groups
							</button>
						{/if}
						{#if onMaintenance}
							<button
								onclick={handleMaintenance}
								class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
							>
								<Wrench class="size-3.5 text-blue-400" />
								Maintenance
							</button>
						{/if}
						{#if onEdit}
							<button
								onclick={handleEdit}
								class="menu-item flex items-center gap-2.5 w-full px-3.5 py-2.5 text-[13px] text-foreground hover:bg-white/8 transition-colors text-left"
							>
								<Pencil class="size-3.5 text-muted-foreground" />
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

<style>
	/* Shared from page */
	.list-row {
		display: grid;
		grid-template-columns: 1fr 1fr 100px 100px 74px;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		transition: background 140ms ease;
	}

	.list-row:hover {
		background: var(--color-accent);
	}

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

	.menu-item:not(:last-child) {
		border-bottom: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
	}

	@media (prefers-reduced-motion: reduce) {
		.list-row {
			transition: none;
		}
		.dropdown-menu {
			animation: none;
		}
	}
</style>
