<script lang="ts">
	import { Globe, ExternalLink, Building2, Pencil, Trash2, MoreVertical } from 'lucide-svelte';
	import type { StatusPage } from '$lib/api/statuspage';

	let {
		sp,
		onEdit,
		onDelete
	}: {
		sp: StatusPage;
		onEdit?: (sp: StatusPage) => void;
		onDelete?: (sp: StatusPage) => void;
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
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="card-wrapper flex flex-col bg-card border border-border rounded-xl p-5 hover:border-border/80 hover:bg-card/80 transition-colors shadow-sm h-full w-full"
	onclick={closeMenu}
>
	<!-- Top row: icon + title + badge + 3-dot menu -->
	<div class="flex items-start justify-between gap-3 mb-3">
		<!-- Left: icon + title/subdomain -->
		<div class="flex items-center gap-3 min-w-0 flex-1">
			<div
				class="flex items-center justify-center size-10 bg-white/5 rounded-lg shrink-0 border border-border"
			>
				<Globe class="size-5 text-muted-foreground" />
			</div>
			<div class="min-w-0 flex flex-col">
				<h3 class="text-base font-semibold text-foreground truncate m-0">{sp.Title}</h3>
				<p class="text-sm text-muted-foreground truncate m-0">{sp.Subdomain}</p>
			</div>
		</div>

		<!-- Right: badge + 3-dot menu -->
		<div class="flex items-center gap-2 shrink-0">
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

			{#if onEdit || onDelete}
				<div class="relative">
					<button
						onclick={toggleMenu}
						class="flex items-center justify-center size-7 rounded-md bg-white/5 border border-border hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors"
						aria-label="More options"
					>
						<MoreVertical class="size-4" />
					</button>

					{#if menuOpen}
						<div
							class="dropdown-menu absolute right-0 top-full mt-1.5 w-40 z-20 bg-card border border-border rounded-xl shadow-xl overflow-hidden"
						>
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
