<script lang="ts">
	import gsap from 'gsap';
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n/index.svelte';

	onMount(() => {
		gsap.from('.hero-content > *', {
			y: 50,
			opacity: 0,
			duration: 1.2,
			stagger: 0.2,
			ease: 'power4.out',
			delay: 0.1
		});
	});
</script>

<header class="relative overflow-hidden">
	<div class="w-full h-screen relative overflow-hidden">
		<video src="/Hero.mp4" class="w-full h-full object-cover" autoplay loop muted playsinline
		></video>
		<div
			class="hero-content absolute bottom-12 md:bottom-20 left-4 md:left-8 lg:left-12 z-10 max-w-180"
		>
			<h1
				class="text-primary-container font-headline-md text-4xl md:text-5xl lg:text-6xl font-bold mb-4 tracking-tight drop-shadow-lg"
			>
				{t('landing.hero.headline')}
			</h1>
			<p class="text-on-surface/95 font-body-md text-lg md:text-xl leading-relaxed drop-shadow-md">
				{t('landing.hero.subtitle')}
			</p>
		</div>
	</div>
</header>
