<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Building2,
		Plus,
		TriangleAlert,
		Search,
		X,
		UserPlus,
		Check,
		LayoutGrid,
		List,
		Link2,
		Users,
		Activity
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { OrgAPI } from '$lib/api/organization';
	import { debounce } from '$lib/helper/utils';
	import { validateOrgName, validateOrgSlug } from '$lib/utils/validation';
	import { DataSkeleton } from '$lib/components/common';
	import { toast } from 'svelte-sonner';
	import { activeOrgId, organizations } from '$lib/stores/organization';
	import { parseApiError } from '$lib/utils/parseApiError';

	// ── Types ────────────────────────────────────────────────────────────────────
	interface Organization {
		ID: string;
		Name: string;
		Slug: string;
		CreatedAt: string;
		Monitors?: Array<{ Status: string }>;
		OrganizationMembers?: Array<any>;
	}

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

	// ── State ────────────────────────────────────────────────────────────────────
	let orgs = $state<Organization[]>([]);
	let isLoading = $state(true);
	let fetchError = $state('');

	let viewMode = $state<'grid' | 'list'>('grid');

	// Dialog
	let dialogOpen = $state(false);
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

	// Invite
	let searchQuery = $state('');
	let searchResults = $state<UserResult[]>([]);
	let isSearching = $state(false);
	let showDropdown = $state(false);
	let invitedUsers = $state<InvitedUser[]>([]);
	let searchPage = $state(1);
	let hasMoreResults = $state(true);

	// ── Helpers ──────────────────────────────────────────────────────────────────
	function slugify(str: string): string {
		return (
			str
				.toLowerCase()
				.trim()
				.replace(/[^a-z0-9]+/g, '-')
				.replace(/^-+|-+$/g, '') || ''
		);
	}

	function handleNameInput() {
		orgSlug = slugify(orgName);
		handleNameValidation(orgName);
		handleSlugValidation(orgSlug);
	}

	function initials(name: string): string {
		const parts = name.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
	}

	// ── Search API Call ──────────────────────────────────────────────────────────
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
			if (newResults.length < 10) {
				hasMoreResults = false;
			} else {
				hasMoreResults = true;
			}
			if (pageNum === 1) {
				searchResults = newResults;
			} else {
				searchResults = [...searchResults, ...newResults];
			}
		} catch (error) {
			if (import.meta.env.DEV) {
				console.error('Failed to search users:', error);
			}
			if (pageNum === 1) searchResults = [];
		} finally {
			isSearching = false;
		}
	}

	const debouncedSearchUsers = debounce(searchUsers, 500);

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
		debouncedSearchUsers(searchQuery, searchPage);
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
		invitedUsers = invitedUsers.map((u) => (u.id === id ? { ...u, role } : u));
	}

	// ── Derived ──────────────────────────────────────────────────────────────────
	let canSubmit = $derived(
		orgName.trim().length >= 3 &&
			orgSlug.length >= 6 &&
			!isSubmitting &&
			!fieldErrors.orgName &&
			!fieldErrors.orgSlug
	);
	let filteredResults = $derived(
		searchResults.filter((r) => !invitedUsers.some((u) => u.id === r.id))
	);

	// ── Data fetching ─────────────────────────────────────────────────────────────
	async function loadOrgs() {
		isLoading = true;
		fetchError = '';
		try {
			const res = await OrgAPI.getAll();
			orgs = res.data ?? [];

			// Update global store
			organizations.set(orgs.map((o) => ({ id: o.ID, name: o.Name, slug: o.Slug })));
		} catch (err: any) {
			fetchError = parseApiError(err, 'Failed to load organizations.');
		} finally {
			isLoading = false;
		}
	}

	// ── Actions ──────────────────────────────────────────────────────────────────
	function openDialog() {
		orgName = '';
		orgSlug = '';
		createError = '';
		fieldErrors = { orgName: '', orgSlug: '' };
		shake = { orgName: false, orgSlug: false };
		searchQuery = '';
		searchResults = [];
		invitedUsers = [];
		showDropdown = false;
		dialogOpen = true;
	}

	async function handleCreate() {
		if (!canSubmit) return;
		isSubmitting = true;
		createError = '';
		try {
			const res = await OrgAPI.create({
				name: orgName.trim(),
				slug: orgSlug,
				...(invitedUsers.length > 0 && {
					invitations: invitedUsers.map((u) => ({ invitee_id: u.id, role: u.role }))
				})
			});
			toast.success('Organization created successfully');
			dialogOpen = false;
			await loadOrgs();

			// Navigate to the newly created org
			const newOrg = res.data;
			if (newOrg?.ID) {
				selectOrg(newOrg.ID);
			}
		} catch (err: any) {
			createError = parseApiError(err, 'Failed to create organization.');
			toast.error(createError);
		} finally {
			isSubmitting = false;
		}
	}

	function selectOrg(orgId: string) {
		activeOrgId.set(orgId);
		localStorage.setItem('pulsewatch_last_org', orgId);
		goto(`/dashboard/org/${orgId}/monitors`);
	}

	function setViewMode(mode: 'grid' | 'list') {
		viewMode = mode;
		localStorage.setItem('orgViewMode', mode);
	}

	onMount(() => {
		loadOrgs();
		const savedMode = localStorage.getItem('orgViewMode');
		if (savedMode === 'grid' || savedMode === 'list') {
			viewMode = savedMode;
		}
	});
</script>

<svelte:head>
	<title>Select Organization | PulseWatch</title>
</svelte:head>

<div class="text-foreground antialiased">
	<!-- Page Header -->
	<header class="flex flex-wrap items-start justify-between gap-4 mb-6">
		<div>
			<h1 class="text-[32px] font-semibold leading-tight tracking-tight text-foreground mb-1.5">
				Organizations
			</h1>
			<p class="text-base leading-relaxed text-muted-foreground max-w-[46ch]">
				Select an organization to manage its monitors, incidents, and status pages.
			</p>
		</div>
		<div class="flex items-center gap-3">
			<!-- View Toggle -->
			{#if orgs.length > 0}
				<div class="flex items-center bg-foreground/10 border border-border rounded-lg p-0.5">
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'grid'
							? 'bg-white/10 text-foreground'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('grid')}
						aria-label="Grid view"
					>
						<LayoutGrid class="size-4" />
					</button>
					<button
						class="p-1.5 rounded-md transition-colors {viewMode === 'list'
							? 'bg-white/10 text-foreground'
							: 'text-muted-foreground hover:text-foreground hover:bg-white/5'}"
						onclick={() => setViewMode('list')}
						aria-label="List view"
					>
						<List class="size-4" />
					</button>
				</div>
			{/if}

			<Button
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium"
				onclick={openDialog}
			>
				<Plus class="size-4" />
				New organization
			</Button>
		</div>
	</header>

	<!-- Loading state -->
	{#if isLoading}
		<DataSkeleton variant="cards" {viewMode} count={6} />

		<!-- Fetch error state -->
	{:else if fetchError}
		<div
			class="flex flex-col items-center justify-center gap-3 py-24 border border-border rounded-lg bg-white/3"
		>
			<TriangleAlert class="size-8 text-destructive" />
			<p class="text-sm text-muted-foreground">{fetchError}</p>
			<Button
				class="bg-secondary border border-border text-muted-foreground hover:text-foreground hover:bg-accent active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium"
				onclick={loadOrgs}
			>
				Try again
			</Button>
		</div>

		<!-- Empty state -->
	{:else if orgs.length === 0}
		<div
			class="flex flex-col items-center text-center py-24 px-4 border border-border rounded-lg bg-white/3"
		>
			<div
				class="flex items-center justify-center size-14 rounded-lg bg-primary/10 border border-primary/20 mb-4 text-primary"
			>
				<Building2 class="size-6" />
			</div>
			<h3 class="text-xl font-semibold text-foreground mb-1.5">No organizations yet</h3>
			<p class="text-body-md text-muted-foreground max-w-[40ch] mb-6">
				Create an organization to start adding monitors, inviting your team, and watching uptime
				together.
			</p>
			<Button
				class="bg-secondary border border-border text-primary hover:bg-primary hover:border-primary hover:text-primary-foreground active:scale-[0.98] transition-all duration-200 h-9 px-4 font-medium"
				onclick={openDialog}
			>
				<Plus class="size-4" />
				Create organization
			</Button>
		</div>

		<!-- Org list -->
	{:else}
		{#if viewMode === 'grid'}
			<div class="grid grid-cols-[repeat(auto-fill,minmax(430px,1fr))] gap-5">
				{#each orgs as org}
					<button
						class="relative flex flex-col gap-2 p-4 bg-white/3 border border-border rounded-lg transition-all hover:-translate-y-0.5 hover:border-primary/30 group text-left w-full cursor-pointer"
						onclick={() => selectOrg(org.ID)}
					>
						<div class="flex items-start gap-3">
							<div
								class="flex items-center justify-center size-11 rounded-md bg-primary/15 text-primary font-mono text-sm font-medium shrink-0"
							>
								{initials(org.Name)}
							</div>
							<div class="min-w-0 flex-1">
								<h3 class="text-[17px] font-semibold text-foreground leading-tight mb-0.5 truncate">
									{org.Name}
								</h3>
								<div
									class="flex items-center gap-1.5 text-[12.5px] text-muted-foreground font-mono truncate"
								>
									<Link2 class="size-3.5 opacity-70 shrink-0" />
									<span class="truncate">pulsewatch.app/o/{org.Slug}</span>
								</div>
							</div>
						</div>

						<div class="h-px bg-white/8 my-1"></div>

						<div class="flex gap-6">
							<div class="flex items-center gap-2">
								<Users class="size-[15px] text-muted-foreground shrink-0" />
								<span class="font-mono font-medium text-body-md text-foreground"
									>{org.OrganizationMembers?.length || 0}</span
								>
								<span class="text-[13px] text-muted-foreground"
									>member{org.OrganizationMembers?.length === 1 ? '' : 's'}</span
								>
							</div>
							<div class="flex items-center gap-2">
								<Activity class="size-[15px] text-muted-foreground shrink-0" />
								<span class="font-mono font-medium text-body-md text-foreground"
									>{org.Monitors?.length || 0}</span
								>
								<span class="text-[13px] text-muted-foreground"
									>monitor{org.Monitors?.length === 1 ? '' : 's'}</span
								>
							</div>
						</div>

						<!-- Status Monitor -->
						<div class="mt-4 flex flex-col gap-2">
							<div class="flex items-center justify-between text-[13px]">
								<span class="text-muted-foreground font-medium">Status monitor</span>
								{#if (org.Monitors?.length || 0) > 0}
									<span
										class="{(org.Monitors?.filter((m) => m.Status !== 'down').length || 0) ===
										(org.Monitors?.length || 0)
											? 'text-primary'
											: 'text-destructive'} font-semibold flex items-center gap-1.5 tracking-wide"
									>
										<Activity class="size-3.5" />
										{org.Monitors?.filter((m) => m.Status !== 'down').length || 0}/{org.Monitors
											?.length || 0} active
									</span>
								{:else}
									<span
										class="text-muted-foreground/50 font-semibold flex items-center gap-1.5 tracking-wide"
									>
										<Activity class="size-3.5" />
										0/0 active
									</span>
								{/if}
							</div>
							<div class="flex items-center gap-1 h-[14px]" aria-hidden="true">
								{#if org.Monitors && org.Monitors.length > 0}
									{#each org.Monitors as monitor}
										<span
											class="flex-1 h-full rounded-[2px] {monitor.Status === 'down'
												? 'bg-destructive'
												: 'bg-primary'} min-w-[4px]"
										></span>
									{/each}
								{:else}
									<span class="flex-1 h-full rounded-[2px] bg-white/10"></span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			</div>
		{:else}
			<div class="flex flex-col gap-2">
				{#each orgs as org}
					<button
						class="flex flex-col lg:flex-row lg:items-center gap-3 lg:gap-5 p-3 lg:py-2 lg:px-5 bg-white/3 border border-border rounded-lg transition-all hover:-translate-y-px hover:border-primary/30 group text-left w-full cursor-pointer"
						onclick={() => selectOrg(org.ID)}
					>
						<div class="flex items-start lg:items-center gap-3 flex-1 min-w-[240px]">
							<div
								class="flex items-center justify-center size-11 rounded-md bg-primary/15 text-primary font-mono text-sm font-medium shrink-0"
							>
								{initials(org.Name)}
							</div>
							<div class="min-w-0 flex-1">
								<h3 class="text-[17px] font-semibold text-foreground leading-tight mb-0.5 truncate">
									{org.Name}
								</h3>
								<div
									class="flex items-center gap-1.5 text-[12.5px] text-muted-foreground font-mono truncate"
								>
									<Link2 class="size-3.5 opacity-70 shrink-0" />
									<span class="truncate">pulsewatch.app/o/{org.Slug}</span>
								</div>
							</div>
						</div>

						<div class="h-px bg-white/8 w-full lg:hidden"></div>

						<div class="flex gap-5 lg:w-[180px] lg:pl-5 lg:border-l lg:border-border shrink-0">
							<div class="flex items-center gap-1.5">
								<Users class="size-[15px] text-muted-foreground shrink-0" />
								<span class="font-mono font-medium text-body-md text-foreground"
									>{org.OrganizationMembers?.length || 0}</span
								>
								<span class="text-[12px] text-muted-foreground"
									>member{org.OrganizationMembers?.length === 1 ? '' : 's'}</span
								>
							</div>
							<div class="flex items-center gap-1.5">
								<Activity class="size-[15px] text-muted-foreground shrink-0" />
								<span class="font-mono font-medium text-body-md text-foreground"
									>{org.Monitors?.length || 0}</span
								>
								<span class="text-[12px] text-muted-foreground"
									>monitor{org.Monitors?.length === 1 ? '' : 's'}</span
								>
							</div>
						</div>

						<div class="h-px bg-white/8 w-full lg:hidden my-0.5"></div>

						<div
							class="flex items-center justify-between lg:w-[200px] lg:pl-5 lg:border-l lg:border-border shrink-0"
						>
							<div class="flex items-end gap-0.5 h-[22px] w-[140px]" aria-hidden="true">
								{#if org.Monitors && org.Monitors.length > 0}
									{#each org.Monitors as monitor}
										<span
											class="flex-1 h-full rounded-[1px] {monitor.Status === 'down'
												? 'bg-destructive'
												: 'bg-primary'} opacity-85 min-w-[2px]"
										></span>
									{/each}
								{:else}
									<span class="flex-1 h-full rounded-[1px] bg-white/10"></span>
								{/if}
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<!-- ── Create Organization Dialog ──────────────────────────────────────────── -->
<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content class="bg-card! border-border! p-0! gap-0! text-foreground w-full max-w-[480px]!">
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
					<label class="text-[13px] font-medium text-muted-foreground" for="orgName">
						Organization name <span class="text-destructive font-bold">*</span>
					</label>
					{#if fieldErrors.orgName}
						<span class="text-xs text-destructive font-medium">{fieldErrors.orgName}</span>
					{/if}
				</div>
				<Input
					id="orgName"
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
						<label class="text-[13px] font-medium text-muted-foreground" for="orgSlug">
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
						pulsewatch.app/o/
					</span>
					<input
						id="orgSlug"
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
									{#each filteredResults as user}
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
						{#each invitedUsers as user}
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
					onclick={() => (dialogOpen = false)}
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
