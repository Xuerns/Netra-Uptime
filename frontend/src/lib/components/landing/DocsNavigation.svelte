<script lang="ts">
	import logo from '$lib/assets/Logo.svg';
	import { resolve } from '$app/paths';

	const homePath = resolve('/');
	const loginPath = resolve('/auth/login');
</script>

<header
	class="fixed top-0 left-0 right-0 z-50 bg-[#0a0a0b]/90 backdrop-blur-xl border-b border-white/8"
>
	<div class="max-w-[1440px] mx-auto px-6 h-16 flex items-center justify-between gap-6">
		<!-- Left: Logo + Title -->
		<a href={homePath} class="flex items-center gap-3 group shrink-0">
			<div class="relative">
				<img
					src={logo}
					alt="PulseWatch Logo"
					class="h-8 w-auto transition-transform duration-300 group-hover:scale-105"
				/>
				<!-- subtle green glow on hover -->
				<div
					class="absolute inset-0 rounded-full bg-primary/20 blur-md opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"
				></div>
			</div>
			<div class="flex items-center gap-2">
				<span class="font-semibold text-white text-[15px] tracking-tight">PulseWatch</span>
				<span class="text-white/20 text-[15px] font-light select-none">/</span>
				<span class="text-on-surface-variant text-[15px] font-medium">Documentation</span>
			</div>
		</a>

		<!-- Center: Nav links -->
		<nav class="hidden md:flex items-center gap-1 text-[13px] font-medium">
			<a
				href="#getting-started"
				class="text-on-surface-variant hover:text-white hover:bg-white/5 px-3 py-1.5 rounded-lg transition-all duration-200"
			>
				Getting Started
			</a>
			<a
				href="#guides"
				class="text-on-surface-variant hover:text-white hover:bg-white/5 px-3 py-1.5 rounded-lg transition-all duration-200"
			>
				Guides
			</a>
			<a
				href="#api-reference"
				class="text-on-surface-variant hover:text-white hover:bg-white/5 px-3 py-1.5 rounded-lg transition-all duration-200"
			>
				API Reference
			</a>
		</nav>

		<!-- Right: Search + Back to app -->
		<div class="flex items-center gap-3 shrink-0">
			<!-- Search shortcut hint -->
			<button
				class="hidden md:flex items-center gap-2 text-on-surface-variant hover:text-white bg-white/5 hover:bg-white/8 border border-white/8 hover:border-white/15 rounded-lg px-3 py-1.5 text-[13px] transition-all duration-200 group"
				aria-label="Search documentation"
			>
				<span class="material-symbols-outlined text-[16px]">search</span>
				<span class="text-on-surface-variant">Search</span>
				<kbd
					class="ml-1 flex items-center gap-0.5 text-[11px] text-on-surface-variant/60 group-hover:text-on-surface-variant/80 bg-white/5 rounded px-1 py-0.5 font-mono"
				>
					⌘K
				</kbd>
			</button>
		</div>
	</div>
</header>
