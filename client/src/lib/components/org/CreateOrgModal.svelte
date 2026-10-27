<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Search, UserPlus, X, TriangleAlert } from 'lucide-svelte';
	import { OrgAPI } from '$lib/api/organization';
	import { debounce } from '$lib/utils';
	import { validateOrgName, validateOrgSlug } from '$lib/utils/validation';
	import { slugify, getInitials as initials } from '$lib/utils/string';
	import { parseApiError } from '$lib/utils/parseApiError';
	import { toast } from 'svelte-sonner';

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
		onCreated?: (newOrg: any) => void;
	}

	let { open = $bindable(false), onCreated }: Props = $props();

	let orgName = $state('');
	let orgSlug = $state('');
	let isSubmitting = $state(false);
	let createError = $state('');

	let fieldErrors = $state({ orgName: '', orgSlug: '' });
	let shake = $state({ orgName: false, orgSlug: false });

	const handleNameValidation = debounce((val: string) => {
		const err = validateOrgName(val);
		if (err) {
			fieldErrors.orgName = err;
			shake.orgName = false;
			setTimeout(() => (shake.orgName = true), 10);
		} else {
			fieldErrors.orgName = '';
			shake.orgName = false;
		}
	}, 500);

	const handleSlugValidation = debounce((val: string) => {
		const err = validateOrgSlug(val);
		if (err) {
			fieldErrors.orgSlug = err;
			shake.orgSlug = false;
			setTimeout(() => (shake.orgSlug = true), 10);
		} else {
			fieldErrors.orgSlug = '';
			shake.orgSlug = false;
		}
	}, 500);

	function handleNameInput() {
		orgSlug = slugify(orgName);
		handleNameValidation(orgName);
		handleSlugValidation(orgSlug);
	}

	// Invite
	let searchQuery = $state('');
	let searchResults = $state<UserResult[]>([]);
	let isSearching = $state(false);
	let showDropdown = $state(false);
	let invitedUsers = $state<InvitedUser[]>([]);
	let searchPage = $state(1);
	let hasMoreResults = $state(true);

	async function searchUsers(query: string, pageNum: number = 1) {
		if (!query.trim()) {
			searchResults = [];
			isSearching = false;
			hasMoreResults = false;
			return;
		}
		try {
			const res = await OrgAPI.searchPersons(query, pageNum);
			const newResults = (res.data || []).map((u: any) => ({
				id: u.id ?? u.ID,
				name: u.name ?? u.Name,
				avatar_url: u.avatar_url ?? u.AvatarURL
			}));
			if (pageNum === 1) {
				searchResults = newResults;
			} else {
				searchResults = [...searchResults, ...newResults];
			}
			hasMoreResults = Boolean(res.pagination?.has_more);
		} catch {
			searchResults = [];
			hasMoreResults = false;
		} finally {
			isSearching = false;
		}
	}

	const debouncedSearch = debounce((q: string) => {
		searchPage = 1;
		hasMoreResults = true;
		searchUsers(q, 1);
	}, 300);

	function handleSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		if (searchQuery.trim()) {
			isSearching = true;
			showDropdown = true;
			debouncedSearch(searchQuery);
		} else {
			searchResults = [];
			showDropdown = false;
			isSearching = false;
		}
	}

	function handleScroll(e: Event) {
		const el = e.target as HTMLElement;
		if (el.scrollHeight - el.scrollTop - el.clientHeight < 50 && !isSearching && hasMoreResults) {
			searchPage += 1;
			searchUsers(searchQuery, searchPage);
		}
	}

	let filteredResults = $derived(
		searchResults.filter((u) => !invitedUsers.some((inv) => inv.id === u.id))
	);

	function addInvite(user: UserResult) {
		invitedUsers = [...invitedUsers, { ...user, role: 'member' }];
		searchQuery = '';
		searchResults = [];
		showDropdown = false;
	}

	function removeInvite(userId: string) {
		invitedUsers = invitedUsers.filter((u) => u.id !== userId);
	}

	function setRole(userId: string, role: 'admin' | 'member') {
		invitedUsers = invitedUsers.map((u) => (u.id === userId ? { ...u, role } : u));
	}

	let canSubmit = $derived(
		orgName.trim().length > 0 &&
			orgSlug.trim().length >= 6 &&
			!fieldErrors.orgName &&
			!fieldErrors.orgSlug &&
			!isSubmitting
	);

	function resetForm() {
		orgName = '';
		orgSlug = '';
		fieldErrors = { orgName: '', orgSlug: '' };
		shake = { orgName: false, orgSlug: false };
		searchQuery = '';
		searchResults = [];
		invitedUsers = [];
		createError = '';
	}

	$effect(() => {
		if (!open) {
			resetForm();
		}
	});

	async function handleCreate() {
		const nameErr = validateOrgName(orgName);
		const slugErr = validateOrgSlug(orgSlug);
		fieldErrors.orgName = nameErr;
		fieldErrors.orgSlug = slugErr;
		if (nameErr || slugErr) return;

		isSubmitting = true;
		createError = '';

		try {
			const res = await OrgAPI.create({
				name: orgName.trim(),
				slug: orgSlug.trim(),
				...(invitedUsers.length > 0 && {
					invitations: invitedUsers.map((u) => ({ invitee_id: u.id, role: u.role }))
				})
			});
			toast.success('Organization created successfully');
			open = false;
			if (onCreated) {
				onCreated(res.data);
			}
		} catch (err: any) {
			createError = parseApiError(err, 'Failed to create organization.');
			toast.error(createError);
		} finally {
			isSubmitting = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="bg-card! border-border! p-0! gap-0! text-foreground w-full max-w-120!">
		<!-- Header -->
		<div class="px-6 pt-6 pb-5">
			<Dialog.Title class="text-[17px] font-semibold text-foreground leading-snug">
				New organization
			</Dialog.Title>
			<p class="text-sm text-muted-foreground mt-1 leading-relaxed">
				Set up your workspace and optionally invite teammates right away.
			</p>
		</div>

		<div class="h-px bg-white/8"></div>

		<!-- Body -->
		<div class="px-6 py-5 flex flex-col gap-5">
			<!-- Organization name -->
			<div class="flex flex-col gap-2">
				<div class="flex items-center gap-3">
					<label class="text-[13px] font-medium text-muted-foreground" for="modalOrgName">
						Organization name <span class="text-destructive font-bold">*</span>
					</label>
					{#if fieldErrors.orgName}
						<span class="text-xs text-destructive font-medium">{fieldErrors.orgName}</span>
					{/if}
				</div>
				<Input
					id="modalOrgName"
					type="text"
					placeholder="e.g. Acme Corporation"
					autocomplete="off"
					bind:value={orgName}
					oninput={handleNameInput}
					class="h-10 bg-foreground/10 border-border text-foreground text-sm placeholder:text-muted-foreground/40 focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-primary/25 transition-colors {fieldErrors.orgName
						? 'border-destructive/70 focus-visible:border-destructive focus-visible:ring-error'
						: ''} {shake.orgName ? 'animate-shake' : ''}"
				/>
			</div>

			<!-- URL slug -->
			<div class="flex flex-col gap-2">
				<div class="flex items-baseline justify-between">
					<div class="flex items-center gap-3">
						<label class="text-[13px] font-medium text-muted-foreground" for="modalOrgSlug">
							URL slug <span class="text-destructive font-bold">*</span>
						</label>
						{#if fieldErrors.orgSlug}
							<span class="text-xs text-destructive font-medium">{fieldErrors.orgSlug}</span>
						{/if}
					</div>
					<span class="text-[11px] text-muted-foreground/50">min. 6 characters</span>
				</div>
				<div
					class="flex items-stretch h-10 rounded-lg border border-border bg-foreground/10 overflow-hidden transition-colors focus-within:border-primary focus-within:ring-1 focus-within:ring-primary/25 {fieldErrors.orgSlug
						? 'border-destructive/70 focus-within:border-destructive focus-within:ring-error'
						: ''} {shake.orgSlug ? 'animate-shake' : ''}"
				>
					<span
						class="flex items-center px-3 text-[12px] text-muted-foreground/50 font-mono bg-white/4 border-r border-border select-none shrink-0 whitespace-nowrap"
					>
						netra.app/o/
					</span>
					<input
						id="modalOrgSlug"
						type="text"
						placeholder="your-org"
						autocomplete="off"
						bind:value={orgSlug}
						oninput={() => handleSlugValidation(orgSlug)}
						class="flex-1 px-3 bg-transparent border-none outline-none text-primary text-[13px] font-mono placeholder:text-muted-foreground/30 min-w-0"
					/>
				</div>
			</div>

			<!-- Divider with label -->
			<div class="flex items-center gap-3">
				<div class="h-px flex-1 bg-white/8"></div>
				<span
					class="text-[11px] text-muted-foreground/50 uppercase tracking-widest font-medium shrink-0"
				>
					Invite teammates
				</span>
				<div class="h-px flex-1 bg-white/8"></div>
			</div>

			<!-- Invite search -->
			<div class="flex flex-col gap-2">
				<div class="relative">
					<!-- Search input -->
					<div
						class="flex items-center h-10 rounded-lg border border-border bg-foreground/10 overflow-hidden transition-colors focus-within:border-primary focus-within:ring-1 focus-within:ring-primary/25 gap-2 px-3"
					>
						{#if isSearching}
							<div
								class="size-4 rounded-full border-2 border-border border-t-primary animate-spin shrink-0"
							></div>
						{:else}
							<Search class="size-4 text-muted-foreground/50 shrink-0" />
						{/if}
						<input
							type="text"
							placeholder="Search by name or email..."
							autocomplete="off"
							bind:value={searchQuery}
							oninput={handleSearchInput}
							onfocus={() => {
								if (searchQuery.trim()) showDropdown = true;
							}}
							onblur={() =>
								setTimeout(() => {
									showDropdown = false;
								}, 150)}
							class="flex-1 h-full bg-transparent border-none outline-none text-foreground text-[13px] placeholder:text-muted-foreground/35 min-w-0"
						/>
					</div>

					<!-- Results dropdown -->
					{#if showDropdown && searchQuery.trim()}
						<div
							class="absolute top-[calc(100%+4px)] left-0 right-0 rounded-lg border border-border bg-card z-10 overflow-hidden"
						>
							{#if isSearching}
								<div class="px-4 py-3 text-[13px] text-muted-foreground/60">Searching...</div>
							{:else if filteredResults.length > 0}
								<ul class="max-h-60 overflow-y-auto" onscroll={handleScroll}>
									{#each filteredResults as user (user.id)}
										<li>
											<button
												type="button"
												class="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-white/5 transition-colors text-left"
												onmousedown={() => addInvite(user)}
											>
												{#if user.avatar_url}
													<img
														src={user.avatar_url}
														alt={user.name}
														class="size-7 rounded-lg object-cover shrink-0"
													/>
												{:else}
													<div
														class="size-7 rounded-lg bg-secondary/15 text-secondary text-[11px] font-semibold flex items-center justify-center shrink-0"
													>
														{initials(user.name)}
													</div>
												{/if}
												<div class="min-w-0">
													<p class="text-[13px] text-foreground font-medium truncate">
														{user.name}
													</p>
												</div>
											</button>
										</li>
									{/each}
								</ul>
							{:else}
								<div class="px-4 py-4 flex flex-col items-center gap-1.5 text-center">
									<UserPlus class="size-5 text-muted-foreground/40" />
									<p class="text-[13px] text-muted-foreground/60">
										No users found for "{searchQuery}"
									</p>
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<!-- Invited users list -->
				{#if invitedUsers.length > 0}
					<ul class="flex flex-col gap-1.5 mt-1">
						{#each invitedUsers as user (user.id)}
							<li
								class="flex items-center gap-3 px-3 py-2 rounded-lg bg-white/4 border border-border"
							>
								<!-- Avatar -->
								{#if user.avatar_url}
									<img
										src={user.avatar_url}
										alt={user.name}
										class="size-7 rounded-lg object-cover shrink-0"
									/>
								{:else}
									<div
										class="size-7 rounded-lg bg-secondary/15 text-secondary text-[11px] font-semibold flex items-center justify-center shrink-0"
									>
										{initials(user.name)}
									</div>
								{/if}

								<!-- Name -->
								<div class="min-w-0 flex-1">
									<p class="text-[13px] text-foreground font-medium truncate">{user.name}</p>
								</div>

								<!-- Role selector -->
								<div class="flex items-center gap-1 shrink-0">
									<button
										type="button"
										class="h-6 px-2 rounded text-[11px] font-medium transition-colors {user.role ===
										'member'
											? 'bg-white/10 text-foreground'
											: 'text-muted-foreground/50 hover:text-foreground hover:bg-white/5'}"
										onclick={() => setRole(user.id, 'member')}
									>
										Member
									</button>
									<button
										type="button"
										class="h-6 px-2 rounded text-[11px] font-medium transition-colors {user.role ===
										'admin'
											? 'bg-secondary/20 text-secondary'
											: 'text-muted-foreground/50 hover:text-foreground hover:bg-white/5'}"
										onclick={() => setRole(user.id, 'admin')}
									>
										Admin
									</button>
								</div>

								<!-- Remove -->
								<button
									type="button"
									class="size-6 flex items-center justify-center rounded text-muted-foreground/40 hover:text-foreground hover:bg-white/8 transition-colors shrink-0"
									onclick={() => removeInvite(user.id)}
									aria-label="Remove {user.name}"
								>
									<X class="size-3.5" />
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Error message -->
			{#if createError}
				<div
					class="flex items-start gap-2 rounded-lg bg-destructive/8 border border-destructive/20 px-3 py-2.5"
				>
					<TriangleAlert class="size-3.5 shrink-0 text-destructive mt-0.5" />
					<p class="text-[13px] text-destructive leading-snug">{createError}</p>
				</div>
			{/if}
		</div>

		<div class="h-px bg-white/8"></div>

		<!-- Footer -->
		<div class="px-6 py-4 flex items-center justify-between">
			<p class="text-[12px] text-muted-foreground/50">
				{#if invitedUsers.length > 0}
					{invitedUsers.length} teammate{invitedUsers.length > 1 ? 's' : ''} will be invited
				{:else}
					Invites are optional
				{/if}
			</p>
			<div class="flex items-center gap-2">
				<Button
					type="button"
					variant="ghost"
					class="h-9 px-4 text-sm text-foreground hover:text-foreground hover:bg-white/10 transition-colors"
					onclick={() => (open = false)}
				>
					Cancel
				</Button>
				<Button
					class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 text-sm font-medium disabled:opacity-50"
					disabled={!canSubmit}
					onclick={handleCreate}
				>
					{isSubmitting ? 'Creating...' : 'Create organization'}
				</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
