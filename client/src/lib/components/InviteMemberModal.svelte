<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { OrgAPI, type InviteMemberRequest } from '$lib/api/organization';
	import { debounce } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { Search, X, TriangleAlert, Loader2 } from 'lucide-svelte';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { getInitials as initials } from '$lib/utils/string';
	import { logger } from '$lib/utils/logger';

	interface UserResult {
		id: string;
		name: string;
		avatar_url?: string;
	}

	interface InvitedUser {
		id: string;
		name: string;
		avatar_url?: string;
		role: 'admin' | 'member';
	}

	interface Props {
		open?: boolean;
		orgId?: string;
		isOwner?: boolean;
		onInvited?: () => void;
	}

	let { open = $bindable(false), orgId = '', isOwner = false, onInvited }: Props = $props();

	let searchQuery = $state('');
	let searchResults = $state<UserResult[]>([]);
	let isSearching = $state(false);
	let showDropdown = $state(false);
	let invitedUsers = $state<InvitedUser[]>([]);
	let isSubmitting = $state(false);
	let errorMsg = $state('');
	let searchPage = $state(1);
	let hasMoreResults = $state(true);
	let searchContainer = $state<HTMLElement | null>(null);

	function handleFocusOut(e: FocusEvent) {
		if (e.relatedTarget && !searchContainer?.contains(e.relatedTarget as Node)) {
			showDropdown = false;
		}
	}

	onMount(() => {
		function handleClickOutside(event: PointerEvent) {
			if (showDropdown && searchContainer && !searchContainer.contains(event.target as Node)) {
				showDropdown = false;
			}
		}

		document.addEventListener('pointerdown', handleClickOutside);
		return () => {
			document.removeEventListener('pointerdown', handleClickOutside);
		};
	});

	async function searchUsers(query: string, pageNum: number = 1) {
		if (!query.trim()) {
			searchResults = [];
			isSearching = false;
			hasMoreResults = false;
			return;
		}
		try {
			const res = await OrgAPI.searchPersons(query, pageNum, orgId);
			const newResults = (res.data || []).map((u: any) => ({
				id: u.id ?? u.ID,
				name: u.name ?? u.Name,
				avatar_url: u.avatar_url ?? u.AvatarURL
			}));

			hasMoreResults = newResults.length >= 10;

			if (pageNum === 1) {
				searchResults = newResults;
			} else {
				searchResults = [...searchResults, ...newResults];
			}
		} catch (error) {
			logger.error('Failed to search users:', error);
			if (pageNum === 1) searchResults = [];
		} finally {
			isSearching = false;
		}
	}

	const debouncedSearch = debounce(searchUsers, 400);

	function handleSearchInput() {
		showDropdown = true;
		searchPage = 1;
		hasMoreResults = true;
		if (!searchQuery.trim()) {
			searchResults = [];
			isSearching = false;
			return;
		}
		isSearching = true;
		debouncedSearch(searchQuery, searchPage);
	}

	function handleScroll(e: Event) {
		const target = e.currentTarget as HTMLUListElement;
		if (target.scrollHeight - target.scrollTop - target.clientHeight < 20) {
			if (hasMoreResults && !isSearching) {
				isSearching = true;
				searchPage++;
				searchUsers(searchQuery, searchPage);
			}
		}
	}

	function addInvite(user: UserResult) {
		if (invitedUsers.some((u) => u.id === user.id)) return;
		invitedUsers = [...invitedUsers, { ...user, role: 'member' }];
		searchQuery = '';
		searchResults = [];
		showDropdown = false;
	}

	function removeInvite(id: string) {
		invitedUsers = invitedUsers.filter((u) => u.id !== id);
	}

	function setRole(id: string, role: 'admin' | 'member') {
		if (role === 'admin' && !isOwner) {
			toast.error('Only organization owners can invite admins');
			return;
		}
		invitedUsers = invitedUsers.map((u) => (u.id === id ? { ...u, role } : u));
	}

	let filteredResults = $derived(
		searchResults.filter((r) => !invitedUsers.some((u) => u.id === r.id))
	);

	async function handleSendInvitations() {
		if (invitedUsers.length === 0) return;
		isSubmitting = true;
		errorMsg = '';
		try {
			const payload: InviteMemberRequest[] = invitedUsers.map((u) => ({
				invitee_id: u.id,
				role: u.role
			}));
			await OrgAPI.inviteMember(orgId, payload as any);
			toast.success(
				`Sent invitation${invitedUsers.length > 1 ? 's' : ''} to ${invitedUsers.length} member${invitedUsers.length > 1 ? 's' : ''}`
			);
			invitedUsers = [];
			searchQuery = '';
			open = false;
			onInvited?.();
		} catch (err: any) {
			errorMsg = parseApiError(err, 'Failed to send invitations');
			toast.error(errorMsg);
		} finally {
			isSubmitting = false;
		}
	}

	function handleClose() {
		invitedUsers = [];
		searchQuery = '';
		searchResults = [];
		errorMsg = '';
		open = false;
	}
</script>

<Dialog.Root bind:open onOpenChange={(v) => (!v ? handleClose() : null)}>
	<Dialog.Content class="sm:max-w-120 bg-card border border-border text-foreground p-6 rounded-lg">
		<Dialog.Header>
			<div class="flex items-center gap-2.5">
				<Dialog.Title class="text-xl font-semibold tracking-tight">Invite Members</Dialog.Title>
			</div>
			<Dialog.Description class="text-muted-foreground text-sm mt-1">
				Search users by name and invite them to join this organization.
			</Dialog.Description>
		</Dialog.Header>

		<!-- Divider -->
		<div class="h-px w-full bg-border my-4"></div>

		<div class="flex flex-col gap-4">
			<!-- Error message -->
			{#if errorMsg}
				<div
					class="p-3 text-[13px] text-destructive bg-destructive/10 border border-destructive/20 rounded-lg font-medium flex items-center gap-2"
				>
					<TriangleAlert class="size-4 shrink-0" />
					{errorMsg}
				</div>
			{/if}

			<!-- Search bar -->
			<div class="space-y-1.5">
				<label for="search-user-input" class="text-sm font-medium text-foreground"
					>Search User</label
				>
				<div class="relative" bind:this={searchContainer} onfocusout={handleFocusOut}>
					<div
						class="flex items-center h-10 rounded-lg border border-border bg-background overflow-hidden transition-colors focus-within:ring-1 focus-within:ring-primary focus-within:border-primary gap-2 px-3"
					>
						{#if isSearching}
							<div
								class="size-4 rounded-full border-2 border-border border-t-primary animate-spin shrink-0"
							></div>
						{:else}
							<Search class="size-4 text-muted-foreground/50 shrink-0" />
						{/if}
						<input
							id="search-user-input"
							type="text"
							placeholder="Search user by name..."
							autocomplete="off"
							bind:value={searchQuery}
							oninput={handleSearchInput}
							onfocus={() => {
								if (searchQuery.trim()) showDropdown = true;
							}}
							onclick={() => {
								if (searchQuery.trim()) showDropdown = true;
							}}
							onkeydown={(e) => {
								if (e.key === 'Escape') {
									showDropdown = false;
								}
							}}
							class="flex-1 h-full bg-transparent border-none outline-none text-foreground text-[13px] placeholder:text-muted-foreground/40 min-w-0"
						/>
					</div>

					<!-- Search dropdown results -->
					{#if showDropdown && searchQuery.trim()}
						<div
							data-lenis-prevent
							class="absolute top-[calc(100%+4px)] left-0 right-0 rounded-lg border border-border bg-popover z-50 overflow-hidden"
						>
							{#if isSearching && searchResults.length === 0}
								<div class="px-4 py-3 text-[13px] text-muted-foreground">Searching...</div>
							{:else if filteredResults.length > 0}
								<ul
									data-lenis-prevent
									class="max-h-52 overflow-y-auto py-1"
									onscroll={handleScroll}
								>
									{#each filteredResults as user (user.id)}
										<li>
											<button
												type="button"
												class="w-full flex items-center gap-3 px-3 py-2.5 hover:bg-accent/50 transition-colors text-left cursor-pointer"
												onmousedown={(e) => e.preventDefault()}
												onclick={() => addInvite(user)}
											>
												{#if user.avatar_url}
													<img
														src={user.avatar_url}
														alt={user.name}
														class="size-7 rounded-full object-cover shrink-0"
													/>
												{:else}
													<div
														class="size-7 rounded-full bg-primary/15 text-primary text-[11px] font-semibold flex items-center justify-center shrink-0"
													>
														{initials(user.name)}
													</div>
												{/if}
												<div class="min-w-0 flex-1">
													<p class="text-[13px] text-foreground font-medium truncate">
														{user.name}
													</p>
												</div>
											</button>
										</li>
									{/each}
								</ul>
							{:else}
								<div class="px-4 py-4 flex flex-col items-center gap-1 text-center">
									<p class="text-[13px] text-muted-foreground">
										No users found for "{searchQuery}"
									</p>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- List of selected users to invite -->
			{#if invitedUsers.length > 0}
				<div class="space-y-1.5">
					<span class="text-sm font-medium text-foreground">
						Selected
						<span class="text-muted-foreground font-normal">({invitedUsers.length})</span>
					</span>
					<ul data-lenis-prevent class="flex flex-col gap-2 max-h-48 overflow-y-auto pr-0.5">
						{#each invitedUsers as user (user.id)}
							<li
								class="flex items-center gap-3 px-3 py-2 rounded-lg bg-background border border-border"
							>
								{#if user.avatar_url}
									<img
										src={user.avatar_url}
										alt={user.name}
										class="size-7 rounded-full object-cover shrink-0"
									/>
								{:else}
									<div
										class="size-7 rounded-full bg-primary/15 text-primary text-[11px] font-semibold flex items-center justify-center shrink-0"
									>
										{initials(user.name)}
									</div>
								{/if}

								<div class="min-w-0 flex-1">
									<p class="text-[13px] text-foreground font-medium truncate">{user.name}</p>
								</div>

								<!-- Role selector -->
								<div
									class="flex items-center gap-0.5 shrink-0 bg-card p-0.5 rounded-md border border-border"
								>
									<button
										type="button"
										class="h-6 px-2.5 rounded text-[11px] font-medium transition-colors {user.role ===
										'member'
											? 'bg-primary/20 text-primary font-semibold'
											: 'text-muted-foreground hover:text-foreground hover:bg-accent/40'}"
										onclick={() => setRole(user.id, 'member')}
									>
										Member
									</button>
									{#if isOwner}
										<button
											type="button"
											class="h-6 px-2.5 rounded text-[11px] font-medium transition-colors {user.role ===
											'admin'
												? 'bg-blue-500/20 text-blue-400 font-semibold'
												: 'text-muted-foreground hover:text-foreground hover:bg-accent/40'}"
											onclick={() => setRole(user.id, 'admin')}
										>
											Admin
										</button>
									{:else}
										<button
											type="button"
											disabled
											title="Only organization owners can invite admins"
											class="h-6 px-2.5 rounded text-[11px] font-medium text-muted-foreground/40 cursor-not-allowed"
										>
											Admin
										</button>
									{/if}
								</div>

								<!-- Remove button -->
								<button
									type="button"
									class="size-6 flex items-center justify-center rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors shrink-0"
									onclick={() => removeInvite(user.id)}
									aria-label="Remove {user.name}"
								>
									<X class="size-3.5" />
								</button>
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		</div>

		<!-- Divider -->
		<div class="h-px w-full bg-border my-4"></div>

		<!-- Footer -->
		<div class="flex items-center justify-end gap-3">
			<Button
				variant="outline"
				onclick={handleClose}
				disabled={isSubmitting}
				class="h-9 px-4 border-border text-muted-foreground hover:text-foreground"
			>
				Cancel
			</Button>
			<Button
				onclick={handleSendInvitations}
				disabled={invitedUsers.length === 0 || isSubmitting}
				class="h-9 px-4 bg-primary text-primary-foreground hover:bg-primary/90 font-medium"
			>
				{#if isSubmitting}
					<Loader2 class="size-4 mr-2 animate-spin" />
					Sending...
				{:else}
					Send Invitation{invitedUsers.length > 1 ? 's' : ''}
					{#if invitedUsers.length > 0}
						<span
							class="ml-1.5 flex items-center justify-center size-4 rounded-full bg-primary-foreground/20 text-[10px] font-bold"
						>
							{invitedUsers.length}
						</span>
					{/if}
				{/if}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
