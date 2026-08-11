<script lang="ts">
	import { onMount } from 'svelte';
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { t } from '$lib/i18n/index.svelte';

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);
		gsap.from('.why-title', {
			scrollTrigger: {
				trigger: '.why-section',
				start: 'top 80%'
			},
			y: 30,
			opacity: 0,
			duration: 0.8,
			ease: 'power3.out'
		});

		gsap.from('.why-item', {
			scrollTrigger: {
				trigger: '.why-section',
				start: 'top 75%'
			},
			y: 20,
			opacity: 0,
			duration: 0.8,
			stagger: 0.15,
			ease: 'power3.out'
		});
	});
</script>

<section id="solutions" class="why-section py-2xl px-md bg-[#0e0e0f] text-on-surface border-t border-[#1a1a1c]">
	<div class="max-w-container-max mx-auto">
		<div class="grid grid-cols-1 lg:grid-cols-12 gap-xl lg:gap-md items-start">
			<!-- Left: Massive Statement -->
			<div class="lg:col-span-5 lg:sticky lg:top-32 why-title">
				<h2
					class="font-headline-lg text-[40px] md:text-[56px] leading-[1.1] tracking-tight mb-md"
					style="text-wrap: balance;"
				>
					{t('landing.why.headline')}
				</h2>
				<p class="font-body-md text-on-surface-variant text-[18px]">
					{t('landing.why.subtitle')}
				</p>
			</div>

			<!-- Right: Structural List -->
			<div class="lg:col-span-6 lg:col-start-7 flex flex-col">
				<!-- Item 1 -->
				<div class="why-item py-lg border-t border-[#2a2a2b] flex flex-col gap-sm">
					<div class="flex items-center gap-xs text-error">
						<span class="material-symbols-outlined text-[20px]">trending_down</span>
						<span class="font-mono-data text-[12px] uppercase tracking-wider font-bold"
							>{t('landing.why.items.financial.tag')}</span
						>
					</div>
					<h3 class="font-headline-md text-[24px] md:text-[32px] tracking-tight text-white">
						{t('landing.why.items.financial.title')}
					</h3>
					<p class="font-body-md text-on-surface-variant text-[16px] leading-relaxed max-w-[500px]">
						{t('landing.why.items.financial.desc')}
					</p>
				</div>

				<!-- Item 2 -->
				<div class="why-item py-lg border-t border-[#2a2a2b] flex flex-col gap-sm">
					<div class="flex items-center gap-xs text-primary">
						<span class="material-symbols-outlined text-[20px]">verified_user</span>
						<span class="font-mono-data text-[12px] uppercase tracking-wider font-bold"
							>{t('landing.why.items.reputation.tag')}</span
						>
					</div>
					<h3 class="font-headline-md text-[24px] md:text-[32px] tracking-tight text-white">
						{t('landing.why.items.reputation.title')}
					</h3>
					<p class="font-body-md text-on-surface-variant text-[16px] leading-relaxed max-w-[500px]">
						{t('landing.why.items.reputation.desc')}
					</p>
				</div>

				<!-- Item 3 -->
				<div class="why-item py-lg border-t border-[#2a2a2b] border-b flex flex-col gap-sm">
					<div class="flex items-center gap-xs text-tertiary">
						<span class="material-symbols-outlined text-[20px]">bolt</span>
						<span class="font-mono-data text-[12px] uppercase tracking-wider font-bold"
							>{t('landing.why.items.operational.tag')}</span
						>
					</div>
					<h3 class="font-headline-md text-[24px] md:text-[32px] tracking-tight text-white">
						{t('landing.why.items.operational.title')}
					</h3>
					<p class="font-body-md text-on-surface-variant text-[16px] leading-relaxed max-w-[500px]">
						{t('landing.why.items.operational.desc')}
					</p>
				</div>
			</div>
		</div>
	</div>
</section>
