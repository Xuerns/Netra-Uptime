<script lang="ts">
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { onMount } from 'svelte';

	const audiences = [
		{
			id: 1,
			role: 'Pemilik Toko Online & Bisnis Kecil',
			impact: 'Jangan biarkan downtime menghilangkan potensi pendapatan Anda.'
		},
		{
			id: 2,
			role: 'Developer & Backend Engineer',
			impact: 'Pantau microservices, API, dan server secara terperinci.'
		},
		{
			id: 3,
			role: 'Agency & Freelancer',
			impact:
				'Kelola dan pantau puluhan atau ratusan website klien Anda dari satu dashboard terpusat.'
		},
		{
			id: 4,
			role: 'Perusahaan SaaS',
			impact: 'Sediakan halaman status publik yang profesional untuk pengguna Anda.'
		}
	];

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		gsap.from('.audience-card', {
			scrollTrigger: {
				trigger: '.audience-section',
				start: 'top 80%'
			},
			x: -40,
			opacity: 0,
			duration: 0.7,
			stagger: 0.2,
			ease: 'power3.out'
		});
	});
</script>

<section class="audience-section py-32 px-8 lg:px-24 bg-zinc-900">
	<div class="max-w-7xl mx-auto flex flex-col lg:flex-row gap-16">
		<div class="lg:w-1/3 flex flex-col items-start gap-6">
			<h2 class="text-4xl font-bold text-white">Solusi Tepat Untuk</h2>
			<div class="w-16 h-1 bg-emerald-500 rounded-lg"></div>
		</div>

		<div class="lg:w-2/3 flex flex-col gap-6">
			{#each audiences as audience (audience.id)}
				<div
					class="audience-card bg-zinc-950 border border-zinc-800 p-8 rounded-lg flex flex-col md:flex-row md:items-center gap-4"
				>
					<h3 class="text-xl font-bold text-white md:w-1/3">{audience.role}</h3>
					<p class="text-zinc-400 md:w-2/3">{audience.impact}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
