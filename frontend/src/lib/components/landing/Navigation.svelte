<script lang="ts">
	import { resolve } from '$app/paths';
	import logo from '$lib/assets/Logo.svg';
	import { t, getLocale, setLocale } from '$lib/i18n/index.svelte';

	const LoginPath = resolve('/auth/login');
	const RegisterPath = resolve('/auth/register');
</script>

<div class="fixed top-6 right-4 lg:right-8 z-50 flex items-center gap-2">
	<!-- GitHub button -->
	<a
		href="https://github.com/Xuerns/PulseWatch"
		target="_blank"
		rel="noopener noreferrer"
		aria-label="View on GitHub"
		class="flex items-center justify-center w-9 h-9 bg-[#0a0a0b]/80 backdrop-blur-xl border border-white/10 rounded-full shadow-lg text-on-surface-variant hover:text-white hover:border-white/30 hover:bg-white/10 transition-all duration-200 active:scale-95"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			class="w-4.5 h-4.5 fill-current"
			aria-hidden="true"
		>
			<path
				d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"
			/>
		</svg>
	</a>

	<!-- Language toggle pill -->
	<div
		class="relative flex items-center bg-[#0a0a0b]/80 backdrop-blur-xl border border-white/10 rounded-full p-1 shadow-lg"
	>
		<button
			onclick={() => setLocale('id')}
			class="relative w-9 h-6 rounded-full text-[10.5px] font-bold tracking-wider transition-colors z-10 {getLocale() ===
			'id'
				? 'text-white'
				: 'text-on-surface-variant hover:text-white/80'}"
		>
			ID
		</button>
		<button
			onclick={() => setLocale('en')}
			class="relative w-9 h-6 rounded-full text-[10.5px] font-bold tracking-wider transition-colors z-10 {getLocale() ===
			'en'
				? 'text-white'
				: 'text-on-surface-variant hover:text-white/80'}"
		>
			EN
		</button>

		<!-- Sliding pill background -->
		<div
			class="absolute top-1 left-1 w-9 h-6 bg-white/15 rounded-full transition-transform duration-300 ease-out z-0"
			style="transform: translateX({getLocale() === 'id' ? '0' : '100%'})"
		></div>
	</div>
</div>

<nav
	class="fixed top-4 left-1/2 -translate-x-1/2 w-[95%] max-w-5xl z-50 bg-[#0a0a0b]/80 backdrop-blur-xl border border-white/10 shadow-lg rounded-full"
>
	<div class="flex items-center w-full px-md py-3 justify-between">
		<div class="flex items-center gap-2">
			<img src={logo} alt="Pulsewatch Logo" class="h-10 w-auto" />
			<div class="font-headline-md text-headline-md font-bold text-on-surface tracking-tight">
				Pulsewatch
			</div>
		</div>
		<div class="hidden md:flex gap-sm items-center font-label-md text-label-md">
			<a class="text-on-surface-variant hover:text-on-surface transition-colors" href="/#features"
				>{t('landing.nav.features')}</a
			>
			<a class="text-on-surface-variant hover:text-on-surface transition-colors" href="/#solutions"
				>Solutions</a
			>
			<a class="text-on-surface-variant hover:text-on-surface transition-colors" href="/pricing"
				>{t('landing.nav.pricing')}</a
			>
			<a
				class="text-on-surface-variant hover:text-on-surface transition-colors"
				href="/documentation">Documentation</a
			>
		</div>
		<div class="flex gap-sm items-center">
			<a
				href={LoginPath}
				class="font-label-md text-label-md text-on-surface-variant hover:text-on-surface px-sm py-xs transition-colors"
				>{t('landing.nav.login')}</a
			>
			<a
				href={RegisterPath}
				class="bg-primary-container/10 border border-primary-container/30 text-primary-container hover:bg-primary-container hover:text-white font-label-md text-label-md px-md py-xs rounded-lg active:scale-95 transition-all"
				>Register</a
			>
		</div>
	</div>
</nav>
