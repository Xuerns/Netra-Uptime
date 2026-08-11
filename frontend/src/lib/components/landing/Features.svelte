<script lang="ts">
	import gsap from 'gsap';
	import ScrollTrigger from 'gsap/ScrollTrigger';
	import { onMount } from 'svelte';

	const features = [
		{
			id: 1,
			title: 'Deteksi Cepat & Akurat',
			description:
				'Memantau ribuan layanan secara bersamaan. Bebas alarm palsu dengan sistem konfirmasi otomatis. Dukungan HTTP/HTTPS, TCP, dan Ping.'
		},
		{
			id: 2,
			title: 'Notifikasi Real-Time',
			description:
				'Dapatkan notifikasi kurang dari 30 detik setelah perubahan status terdeteksi melalui Email, Slack, dan Webhook.'
		},
		{
			id: 3,
			title: 'Halaman Status Publik',
			description:
				'Bangun kepercayaan pengguna dengan halaman status terdedikasi. Tampilkan histori uptime 90 hari terakhir secara transparan.'
		},
		{
			id: 4,
			title: 'Manajemen Insiden',
			description:
				'Pencatatan insiden otomatis. Tambahkan catatan manual untuk menginformasikan progres perbaikan kepada tim atau publik.'
		},
		{
			id: 5,
			title: 'Kolaborasi Tim',
			description:
				'Kelola semua monitor dalam satu organisasi. Undang kolega atau klien untuk memantau dashboard bersama.'
		}
	];

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		gsap.from('.feature-card', {
			scrollTrigger: {
				trigger: '.features-section',
				start: 'top 75%'
			},
			y: 40,
			opacity: 0,
			duration: 0.6,
			stagger: 0.15,
			ease: 'power2.out'
		});
	});
</script>

<section class="features-section py-32 px-8 lg:px-24 bg-zinc-950">
	<div class="max-w-7xl mx-auto">
		<div class="flex flex-col items-start gap-6 mb-16">
			<h2 class="text-4xl font-bold text-white">Fitur Utama</h2>
			<p class="text-xl text-zinc-400">Esensial, cepat, tanpa kerumitan.</p>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
			{#each features as feature (feature.id)}
				<div
					class="feature-card bg-zinc-900 border border-zinc-800 p-8 rounded-lg flex flex-col gap-4"
				>
					<h3 class="text-xl font-semibold text-emerald-400">{feature.title}</h3>
					<p class="text-zinc-400 leading-relaxed">{feature.description}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
