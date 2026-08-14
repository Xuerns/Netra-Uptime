<script lang="ts">
	import { onMount } from 'svelte';
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { t } from '$lib/i18n/index.svelte';

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		gsap.from('.pricing-header', {
			scrollTrigger: {
				trigger: '.pricing-section',
				start: 'top 80%'
			},
			y: 30,
			opacity: 0,
			duration: 0.8,
			ease: 'power3.out'
		});

		gsap.from('.pricing-card', {
			scrollTrigger: {
				trigger: '.pricing-section',
				start: 'top 75%'
			},
			y: 40,
			opacity: 0,
			duration: 0.8,
			stagger: 0.15,
			ease: 'power3.out'
		});
	});
</script>

<section
	id="pricing"
	class="pricing-section pt-32 py-xl px-md text-on-surface border-t border-[#1a1a1c]"
>
	<div class="max-w-container-max mx-auto">
		<div class="text-center mb-xl pricing-header">
			<h2
				class="font-headline-lg text-[40px] md:text-[56px] leading-[1.1] tracking-tight mb-sm text-white"
				style="text-wrap: balance;"
			>
				{t('landing.pricing.headline')}
			</h2>
			<p class="font-body-md text-[18px] text-on-surface-variant max-w-[672px] mx-auto">
				{t('landing.pricing.subtitle')}
			</p>
		</div>

		<div
			class="grid grid-cols-1 lg:grid-cols-3 gap-6 xl:gap-8 max-w-6xl mx-auto items-stretch mt-8 lg:mt-12"
		>
			<!-- Free Plan -->
			<div
				class="pricing-card group bg-[#131314] rounded-3xl border border-surface-container-high hover:border-[#3a3a3c] transition-colors duration-500 p-md flex flex-col h-full relative overflow-hidden"
			>
				<!-- Subtle hover glow -->
				<div
					class="absolute inset-0 bg-linear-to-b from-white/2 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"
				></div>

				<div class="mb-lg relative z-10">
					<h3 class="font-headline-md text-[28px] text-white mb-2 font-medium tracking-tight">
						{t('landing.pricing.free.title')}
					</h3>
					<p class="font-body-md text-[15px] text-on-surface-variant">
						{t('landing.pricing.free.desc')}
					</p>
				</div>
				<div class="mb-xl pb-lg border-b border-surface-container-high relative z-10">
					<div class="flex items-baseline gap-1">
						<span class="font-display-lg text-[48px] font-bold text-white tracking-tighter"
							>{t('landing.pricing.free.price')}</span
						>
						<span class="font-body-md text-[15px] text-on-surface-variant"
							>{t('landing.pricing.free.period')}</span
						>
					</div>
				</div>
				<ul
					class="space-y-4 mb-xl flex-1 font-body-md text-[15px] text-on-surface-variant relative z-10"
				>
					<li class="flex items-start gap-3">
						<div
							class="w-5 h-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center shrink-0 mt-0.5"
						>
							<span class="material-symbols-outlined text-[12px] text-white/50 font-bold"
								>check</span
							>
						</div>
						<span>{t('landing.pricing.free.feature1')}</span>
					</li>
					<li class="flex items-start gap-3">
						<div
							class="w-5 h-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center shrink-0 mt-0.5"
						>
							<span class="material-symbols-outlined text-[12px] text-white/50 font-bold"
								>check</span
							>
						</div>
						<span>{t('landing.pricing.free.feature2')}</span>
					</li>
					<li class="flex items-start gap-3">
						<div
							class="w-5 h-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center shrink-0 mt-0.5"
						>
							<span class="material-symbols-outlined text-[12px] text-white/50 font-bold"
								>check</span
							>
						</div>
						<span>{t('landing.pricing.free.feature3')}</span>
					</li>
					<li class="flex items-start gap-3">
						<div
							class="w-5 h-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center shrink-0 mt-0.5"
						>
							<span class="material-symbols-outlined text-[12px] text-white/50 font-bold"
								>check</span
							>
						</div>
						<span>{t('landing.pricing.free.feature4')}</span>
					</li>
				</ul>
			</div>

			<!-- Pro Plan -->
			<div
				class="pricing-card group glass-card relative rounded-3xl p-md flex flex-col h-full overflow-hidden border-primary/20 hover:border-primary/50 transition-colors duration-500 z-10"
			>
				<!-- Grid pattern overlay -->
				<div
					class="absolute inset-0"
					style="background-image: radial-gradient(rgba(255, 255, 255, 0.03) 1.5px, transparent 1.5px); background-size: 24px 24px;"
				></div>

				<div class="mb-lg relative z-20">
					<div class="flex items-center justify-between mb-2">
						<h3 class="font-headline-md text-[28px] text-white font-medium tracking-tight">
							{t('landing.pricing.pro.title')}
						</h3>
					</div>
					<p class="font-body-md text-[15px] text-on-surface-variant">
						{t('landing.pricing.pro.desc')}
					</p>
				</div>
				<div class="mb-xl pb-lg border-b border-surface-container-high relative z-20">
					<div class="flex items-baseline gap-1">
						<span class="font-display-lg text-[48px] font-bold tracking-tighter text-white"
							>{t('landing.pricing.pro.price')}</span
						>
						<span class="font-body-md text-[15px] text-on-surface-variant"
							>{t('landing.pricing.pro.period')}</span
						>
					</div>
				</div>
				<ul
					class="space-y-4 mb-xl flex-1 font-body-md text-[15px] text-on-surface-variant relative z-20"
				>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/10 border border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{@html t('landing.pricing.pro.feature1')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/10 border border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{@html t('landing.pricing.pro.feature2')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/10 border border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.pro.feature3')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/10 border border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.pro.feature4')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/10 border border-primary/20 group-hover/item:bg-primary/20 group-hover/item:border-primary/40 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.pro.feature5')}</span>
					</li>
				</ul>
			</div>

			<!-- Max Plan -->
			<div
				class="pricing-card group glass-card relative rounded-3xl p-md flex flex-col h-full overflow-hidden border-primary/20 hover:border-primary transition-colors duration-500 z-10 hover:shadow-[0_0_40px_rgba(78,222,163,0.15)]"
			>
				<!-- Subtle premium background tint for Max Plan -->
				<div
					class="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none"
				></div>

				<!-- Grid pattern overlay -->
				<div
					class="absolute inset-0 pointer-events-none"
					style="background-image: radial-gradient(rgba(255, 255, 255, 0.03) 1.5px, transparent 1.5px); background-size: 24px 24px;"
				></div>

				<!-- Slow shimmer sweep on hover -->
				<div
					class="absolute inset-0 overflow-hidden rounded-3xl pointer-events-none"
					aria-hidden="true"
				>
					<div
						class="w-[200%] h-full bg-linear-to-r from-transparent via-white/10 to-transparent absolute top-0 left-[-200%] group-hover:left-full transition-all duration-1000 ease-out skew-x-12"
					></div>
				</div>

				<!-- "Best Value" badge — lives in the heading row, not as an eyebrow -->
				<div class="mb-lg relative z-20">
					<div class="flex items-center justify-between mb-2">
						<h3 class="font-headline-md text-[28px] text-white font-medium tracking-tight">
							{t('landing.pricing.max.title')}
						</h3>
					</div>
					<p class="font-body-md text-[15px] text-on-surface-variant">
						{t('landing.pricing.max.desc')}
					</p>
				</div>

				<!-- Price -->
				<div class="mb-xl pb-lg border-b border-surface-container-high relative z-20">
					<div class="flex items-baseline gap-1">
						<span class="font-display-lg text-[48px] font-bold tracking-tighter text-white">
							{t('landing.pricing.max.price')}
						</span>
						<span class="font-body-md text-[15px] text-on-surface-variant"
							>{t('landing.pricing.max.period')}</span
						>
					</div>
				</div>

				<!-- Features -->
				<ul
					class="space-y-4 mb-xl flex-1 font-body-md text-[15px] text-on-surface-variant relative z-20"
				>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{@html t('landing.pricing.max.feature1')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{@html t('landing.pricing.max.feature2')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.max.feature3')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.max.feature4')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.max.feature5')}</span>
					</li>
					<li class="flex items-start gap-3 group/item">
						<div
							class="w-5 h-5 rounded-full bg-primary/20 border border-primary/30 group-hover/item:bg-primary/40 group-hover/item:border-primary/60 flex items-center justify-center shrink-0 mt-0.5 transition-colors"
						>
							<span class="material-symbols-outlined text-[12px] text-primary font-bold">check</span
							>
						</div>
						<span class="text-white">{t('landing.pricing.max.feature6')}</span>
					</li>
				</ul>
			</div>
		</div>
	</div>
</section>
