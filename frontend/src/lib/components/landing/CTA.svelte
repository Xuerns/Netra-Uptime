<script lang="ts">
  import { onMount } from 'svelte';
  import gsap from 'gsap';
  import ScrollTrigger from 'gsap/ScrollTrigger';

  onMount(() => {
    gsap.registerPlugin(ScrollTrigger);
    
    gsap.from('.cta-element', {
      scrollTrigger: {
        trigger: '.cta-section',
        start: 'top 85%',
      },
      scale: 0.95,
      opacity: 0,
      duration: 0.6,
      ease: 'back.out(1.5)'
    });
  });
</script>

<section class="cta-section py-32 px-8 lg:px-24 bg-zinc-950 flex flex-col items-center text-center">
  <div class="cta-element bg-zinc-900 border border-zinc-800 p-12 md:p-24 rounded-lg flex flex-col items-center gap-8 w-full max-w-5xl">
    <h2 class="text-4xl md:text-5xl font-bold text-white">Siap Mengamankan Uptime Anda?</h2>
    <p class="text-xl text-zinc-300 max-w-2xl">
      Jangan tunggu sampai pengguna komplain. Serahkan tugas pemantauan kepada Pulsewatch dan fokuslah pada pengembangan bisnis Anda.
    </p>
    <button class="bg-emerald-500 hover:bg-emerald-400 text-zinc-950 font-bold px-10 py-5 rounded-lg text-lg transition-colors mt-4">
      Buat Akun Sekarang
    </button>
  </div>
</section>
