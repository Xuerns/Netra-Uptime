<script lang="ts">
	import { onMount } from 'svelte';
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { t } from '$lib/i18n/index.svelte';

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		gsap.from('.deep-dive-title', {
			scrollTrigger: {
				trigger: '.deep-dive-section',
				start: 'top 80%'
			},
			y: 30,
			opacity: 0,
			duration: 0.8,
			ease: 'power3.out'
		});

		gsap.utils.toArray('.deep-dive-row').forEach((row: any) => {
			gsap.from(row, {
				scrollTrigger: {
					trigger: row,
					start: 'top 75%'
				},
				y: 40,
				opacity: 0,
				duration: 0.8,
				ease: 'power3.out'
			});
		});
	});
</script>

<section
	class="deep-dive-section py-2xl px-md bg-[#0e0e0f] text-on-surface border-t border-[#1a1a1c]"
>
	<div class="max-w-container-max mx-auto">
		<div class="text-center mb-2xl deep-dive-title">
			<h2
				class="font-headline-lg text-[40px] md:text-[56px] leading-[1.1] tracking-tight mb-md text-white"
				style="text-wrap: balance;"
			>
				{t('landing.deepdive.headline')}
			</h2>
			<p class="font-body-md text-on-surface-variant text-[18px] max-w-[672px] mx-auto">
				{t('landing.deepdive.subtitle')}
			</p>
		</div>

		<div class="flex flex-col gap-xl md:gap-2xl">
			<!-- Row 1 -->
			<div class="deep-dive-row grid grid-cols-1 md:grid-cols-2 gap-xl items-center">
				<div>
					<h3 class="font-headline-md text-[32px] text-white mb-sm tracking-tight">
						{t('landing.deepdive.row1.title')}
					</h3>
					<p class="font-body-md text-[18px] text-on-surface-variant leading-relaxed max-w-[448px]">
						{t('landing.deepdive.row1.desc')}
					</p>
				</div>
				<div class="bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex flex-col gap-3">
					<div
						class="flex items-center justify-between bg-[#0e0e0f] border border-[#2a2a2b] p-4 rounded-xl"
					>
						<span class="font-mono-data text-white font-bold">HTTP/S</span>
						<span class="text-primary font-mono-data">200 OK</span>
					</div>
					<div
						class="flex items-center justify-between bg-[#0e0e0f] border border-[#2a2a2b] p-4 rounded-xl"
					>
						<span class="font-mono-data text-white font-bold">TCP port 5432</span>
						<span class="text-primary font-mono-data">CONNECTED</span>
					</div>
					<div
						class="flex items-center justify-between bg-[#0e0e0f] border border-[#2a2a2b] p-4 rounded-xl"
					>
						<span class="font-mono-data text-white font-bold">ICMP ping</span>
						<span class="text-primary font-mono-data">12ms</span>
					</div>
				</div>
			</div>

			<!-- Row 2 -->
			<div class="deep-dive-row grid grid-cols-1 md:grid-cols-2 gap-xl items-center">
				<div class="md:order-2">
					<h3 class="font-headline-md text-[32px] text-white mb-sm tracking-tight">
						{t('landing.deepdive.row2.title')}
					</h3>
					<p class="font-body-md text-[18px] text-on-surface-variant leading-relaxed max-w-[448px]">
						{t('landing.deepdive.row2.desc')}
					</p>
				</div>
				<div
					class="bg-[#131314] rounded-2xl border border-[#2a2a2b] p-lg flex items-center justify-center md:order-1 h-[240px] relative overflow-hidden"
				>
					<div
						class="absolute inset-0 bg-[#0e0e0f]"
						style="background-image: linear-gradient(to right, rgba(255,255,255,0.02) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.02) 1px, transparent 1px); background-size: 24px 24px;"
					></div>
					<div
						class="relative z-10 bg-[#131314] border border-[#2a2a2b] p-6 rounded-2xl shadow-xl w-full max-w-[384px] flex justify-between items-center"
					>
						<div>
							<div
								class="text-[12px] text-on-surface-variant uppercase font-bold tracking-widest mb-1"
							>
								SSL Certificate
							</div>
							<div class="font-mono-data text-white text-[16px]">*.pulsewatch.com</div>
						</div>
						<div class="text-right">
							<div class="text-[32px] font-bold text-primary leading-none">84</div>
							<div class="text-[12px] text-on-surface-variant uppercase mt-1">Days Left</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Row 3 -->
			<div class="deep-dive-row grid grid-cols-1 md:grid-cols-2 gap-xl items-center">
				<div>
					<h3 class="font-headline-md text-[32px] text-white mb-sm tracking-tight">
						{t('landing.deepdive.row3.title')}
					</h3>
					<p class="font-body-md text-[18px] text-on-surface-variant leading-relaxed max-w-[448px]">
						{t('landing.deepdive.row3.desc')}
					</p>
				</div>
				<div class="bg-[#131314] rounded-2xl border border-[#2a2a2b] p-xl flex flex-col gap-6">
					<div class="border-l-2 border-error pl-5 py-1 relative">
						<div class="absolute -left-[5px] top-3 w-2 h-2 rounded-full bg-error"></div>
						<div class="font-mono-data text-[12px] text-on-surface-variant mb-1">
							Oct 14, 09:32 UTC
						</div>
						<div class="text-white font-medium text-[16px]">Database connection timeout</div>
					</div>
					<div class="border-l-2 border-tertiary pl-5 py-1 relative">
						<div class="absolute -left-[5px] top-3 w-2 h-2 rounded-full bg-tertiary"></div>
						<div class="font-mono-data text-[12px] text-on-surface-variant mb-1">
							Oct 14, 09:38 UTC
						</div>
						<div class="text-white font-medium text-[16px]">Scaling read replicas...</div>
					</div>
					<div class="border-l-2 border-primary pl-5 py-1 relative">
						<div class="absolute -left-[5px] top-3 w-2 h-2 rounded-full bg-primary"></div>
						<div class="font-mono-data text-[12px] text-on-surface-variant mb-1">
							Oct 14, 09:45 UTC
						</div>
						<div class="text-white font-medium text-[16px]">Systems fully recovered</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>
