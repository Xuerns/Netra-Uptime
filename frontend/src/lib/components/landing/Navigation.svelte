<script lang="ts">
	import { resolve } from '$app/paths';
	import logo from '$lib/assets/Logo.svg';
	import { t, getLocale, setLocale } from '$lib/i18n/index.svelte';

	const LoginPath = resolve('/auth/login');
	const RegisterPath = resolve('/auth/register');
</script>

<div class="fixed top-6 right-4 lg:right-8 z-50 flex items-center bg-[#0a0a0b]/80 backdrop-blur-xl border border-white/10 rounded-full p-1 shadow-lg">
	<button 
		onclick={() => setLocale('id')}
		class="relative w-9 h-6 rounded-full text-[10.5px] font-bold tracking-wider transition-colors z-10 {getLocale() === 'id' ? 'text-white' : 'text-on-surface-variant hover:text-white/80'}"
	>
		ID
	</button>
	<button 
		onclick={() => setLocale('en')}
		class="relative w-9 h-6 rounded-full text-[10.5px] font-bold tracking-wider transition-colors z-10 {getLocale() === 'en' ? 'text-white' : 'text-on-surface-variant hover:text-white/80'}"
	>
		EN
	</button>
	
	<!-- Sliding pill background -->
	<div 
		class="absolute top-1 left-1 w-9 h-6 bg-white/15 rounded-full transition-transform duration-300 ease-out z-0"
		style="transform: translateX({getLocale() === 'id' ? '0' : '100%'})"
	></div>
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
			<a class="text-on-surface-variant hover:text-on-surface transition-colors" href="/#pricing"
				>{t('landing.nav.pricing')}</a
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
