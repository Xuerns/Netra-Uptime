<script lang="ts">
	import { onMount } from 'svelte';
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { t } from '$lib/i18n/index.svelte';

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		gsap.from('.bento-header', {
			scrollTrigger: {
				trigger: '.bento-section',
				start: 'top 80%'
			},
			y: 30,
			opacity: 0,
			duration: 0.8,
			ease: 'power3.out'
		});

		gsap.from('.bento-element', {
			scrollTrigger: {
				trigger: '.bento-section',
				start: 'top 75%'
			},
			y: 40,
			opacity: 0,
			duration: 0.8,
			stagger: 0.1,
			ease: 'power3.out'
		});
	});
</script>

<section id="features" class="bento-section py-2xl px-md text-on-surface">
	<div class="max-w-container-max mx-auto">
		<div class="mb-xl bento-header flex flex-col lg:flex-row lg:items-end justify-between gap-lg">
			<div>
				<h2
					class="font-headline-lg text-[40px] md:text-[56px] leading-[1.1] tracking-tight text-white"
				>
					{t('landing.bento.headline_1')} <br class="hidden md:block" />{t('landing.bento.headline_2')}
				</h2>
			</div>
			<div class="max-w-[500px]">
				<p class="font-body-md text-on-surface-variant text-[18px] leading-relaxed">
					{t('landing.bento.subtitle')}
				</p>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-12 gap-md">
			<!-- Feature 1: Main -->
			<div
				class="bento-element md:col-span-8 bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg md:p-xl overflow-hidden"
			>
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-xl h-full items-center">
					<div class="flex flex-col justify-center h-full">
						<h3 class="font-headline-md text-[24px] md:text-[32px] text-white mb-md tracking-tight">
							{t('landing.bento.feature1.title')}
						</h3>
						<p class="font-body-md text-[16px] text-on-surface-variant mb-xl leading-relaxed">
							{t('landing.bento.feature1.desc')}
						</p>
						<ul
							class="space-y-4 font-mono-data text-[13px] text-on-surface-variant uppercase tracking-wide"
						>
							<li class="flex items-center gap-3">
								<span class="w-1.5 h-1.5 rounded-full bg-primary"></span> HTTP/S, TCP, UDP, ICMP
							</li>
							<li class="flex items-center gap-3">
								<span class="w-1.5 h-1.5 rounded-full bg-primary"></span> SSL & DNS Lifecycle
							</li>
							<li class="flex items-center gap-3">
								<span class="w-1.5 h-1.5 rounded-full bg-primary"></span> Response Body Matching
							</li>
						</ul>
					</div>
					<!-- Data Viz Component -->
					<div
						class="relative bg-[#0e0e0f] rounded-xl p-md border border-[#2a2a2b] flex flex-col justify-center aspect-square lg:aspect-auto lg:h-full"
					>
						<div
							class="absolute inset-0"
							style="background-image: radial-gradient(rgba(255, 255, 255, 0.04) 1.5px, transparent 1.5px); background-size: 16px 16px;"
						></div>
						<div class="space-y-4 relative z-10">
							<div
								class="flex justify-between items-center text-[11px] text-on-surface-variant uppercase tracking-widest font-bold mb-md"
							>
								<span>Network Nodes</span>
								<span class="flex items-center gap-1.5"
									><span class="w-2 h-2 rounded-full bg-primary"></span> Active</span
								>
							</div>
							<div class="space-y-3">
								<div
									class="flex items-center justify-between bg-[#131314] p-3 rounded-lg border border-[#2a2a2b]"
								>
									<div class="flex items-center gap-3">
										<span class="w-1.5 h-1.5 bg-primary rounded-full"></span>
										<span class="font-mono-data text-[13px] text-white">Singapore (SG-1)</span>
									</div>
									<span class="font-mono-data text-[13px] text-on-surface-variant">24ms</span>
								</div>
								<div
									class="flex items-center justify-between bg-[#131314] p-3 rounded-lg border border-[#2a2a2b]"
								>
									<div class="flex items-center gap-3">
										<span class="w-1.5 h-1.5 bg-primary rounded-full"></span>
										<span class="font-mono-data text-[13px] text-white">New York (US-E)</span>
									</div>
									<span class="font-mono-data text-[13px] text-on-surface-variant">142ms</span>
								</div>
								<div
									class="flex items-center justify-between bg-[#131314] p-3 rounded-lg border border-[#2a2a2b]"
								>
									<div class="flex items-center gap-3">
										<span class="w-1.5 h-1.5 bg-primary rounded-full"></span>
										<span class="font-mono-data text-[13px] text-white">Frankfurt (EU-1)</span>
									</div>
									<span class="font-mono-data text-[13px] text-on-surface-variant">189ms</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Feature 2 -->
			<div
				class="bento-element md:col-span-4 bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex flex-col"
			>
				<h3 class="font-headline-md text-[24px] text-white mb-sm tracking-tight">
					{t('landing.bento.feature2.title')}
				</h3>
				<p class="font-body-md text-[16px] text-on-surface-variant leading-relaxed">
					{t('landing.bento.feature2.desc')}
				</p>
				<div class="mt-auto pt-xl flex flex-col gap-3">
					<div
						class="px-4 py-3.5 bg-[#0e0e0f] rounded-lg border border-[#2a2a2b] flex items-center gap-3 text-[14px] font-medium text-white"
					>
						<span class="material-symbols-outlined text-[20px] text-on-surface-variant">mail</span> Slack
						Integration
					</div>
					<div
						class="px-4 py-3.5 bg-[#0e0e0f] rounded-lg border border-[#2a2a2b] flex items-center gap-3 text-[14px] font-medium text-white"
					>
						<span class="material-symbols-outlined text-[20px] text-on-surface-variant">chat</span> Telegram
						Bot
					</div>
				</div>
			</div>

			<!-- Feature 3 -->
			<div
				class="bento-element md:col-span-4 bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex flex-col"
			>
				<h3 class="font-headline-md text-[24px] text-white mb-sm tracking-tight">{t('landing.bento.feature3.title')}</h3>
				<p class="font-body-md text-[16px] text-on-surface-variant leading-relaxed">
					{t('landing.bento.feature3.desc')}
				</p>
				<div class="mt-auto pt-xl">
					<div
						class="h-32 bg-[#0e0e0f] rounded-xl border border-[#2a2a2b] flex flex-col overflow-hidden"
					>
						<div class="p-3 bg-[#131314] border-b border-[#2a2a2b] flex items-center gap-2">
							<div class="w-2.5 h-2.5 rounded-full bg-error"></div>
							<div class="w-2.5 h-2.5 rounded-full bg-primary"></div>
							<div class="w-2.5 h-2.5 rounded-full bg-primary"></div>
						</div>
						<div class="p-5 flex-1 flex flex-col gap-4 justify-center">
							<div class="w-1/2 h-2.5 bg-[#2a2a2b] rounded-full"></div>
							<div class="w-3/4 h-2.5 bg-[#2a2a2b] rounded-full"></div>
						</div>
					</div>
				</div>
			</div>

			<!-- Feature 4 -->
			<div
				class="bento-element md:col-span-4 bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex flex-col"
			>
				<h3 class="font-headline-md text-[24px] text-white mb-sm tracking-tight">
					{t('landing.bento.feature4.title')}
				</h3>
				<p class="font-body-md text-[16px] text-on-surface-variant leading-relaxed">
					{t('landing.bento.feature4.desc')}
				</p>
				<div class="mt-auto pt-xl">
					<div
						class="p-5 bg-[#0e0e0f] rounded-xl border border-[#2a2a2b] flex flex-col gap-4 font-mono-data text-[13px]"
					>
						<div class="flex gap-4">
							<span class="text-on-surface-variant w-12">10:42</span>
							<span class="text-error">[CRITICAL] Timeout</span>
						</div>
						<div class="flex gap-4">
							<span class="text-on-surface-variant w-12">10:48</span>
							<span class="text-tertiary">[INVESTIGATING]</span>
						</div>
						<div class="flex gap-4">
							<span class="text-on-surface-variant w-12">11:05</span>
							<span class="text-primary">[RESOLVED] Recovery</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Feature 5 -->
			<div
				class="bento-element md:col-span-4 bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex flex-col"
			>
				<h3 class="font-headline-md text-[24px] text-white mb-sm tracking-tight">{t('landing.bento.feature5.title')}</h3>
				<p class="font-body-md text-[16px] text-on-surface-variant leading-relaxed">
					{t('landing.bento.feature5.desc')}
				</p>
				<div class="mt-auto pt-xl flex items-center">
					<div class="flex -space-x-3">
						<div
							class="w-12 h-12 rounded-full border-2 border-[#131314] bg-[#2a2a2b] flex items-center justify-center text-[14px] font-bold text-white z-30"
						>
							JD
						</div>
						<div
							class="w-12 h-12 rounded-full border-2 border-[#131314] bg-primary/20 text-primary flex items-center justify-center text-[14px] font-bold z-20"
						>
							AS
						</div>
						<div
							class="w-12 h-12 rounded-full border-2 border-[#131314] bg-tertiary/20 text-tertiary flex items-center justify-center text-[14px] font-bold z-10"
						>
							MK
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>
