<script lang="ts">
	import logo from '$lib/assets/Logo.svg';
	import { t } from '$lib/i18n/index.svelte';
</script>

<footer class="bg-surface-container-lowest border-t border-white/5 py-lg px-md">
	<div class="max-w-container-max mx-auto grid grid-cols-1 md:grid-cols-4 gap-lg">
		<div class="col-span-1">
			<div class="flex items-center gap-2 mb-xs">
				<img src={logo} alt="Pulsewatch Logo" class="h-10 w-auto" />
				<div class="font-headline-md text-[20px] font-bold text-white tracking-tight">
					Pulsewatch
				</div>
			</div>
			<p class="font-body-md text-[14px] text-on-surface-variant leading-relaxed">
				{t('landing.footer.desc')}
			</p>
		</div>

		<div>
			<h4
				class="font-label-md text-[13px] font-bold text-primary-container mb-sm uppercase tracking-wider"
			>
				{t('landing.footer.product')}
			</h4>
			<div class="flex flex-col gap-2 font-body-md text-[14px]">
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/">{t('landing.footer.product_links.features')}</a>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/">{t('landing.footer.product_links.solutions')}</a>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/">{t('landing.footer.product_links.pricing')}</a>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/">{t('landing.footer.product_links.status')}</a>
			</div>
		</div>

		<div>
			<h4
				class="font-label-md text-[13px] font-bold text-primary-container mb-sm uppercase tracking-wider"
			>
				{t('landing.footer.company')}
			</h4>
			<div class="flex flex-col gap-2 font-body-md text-[14px]">
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/">{t('landing.footer.company_links.about')}</a>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/"
					>{t('landing.footer.company_links.privacy')}</a
				>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/"
					>{t('landing.footer.company_links.terms')}</a
				>
			</div>
		</div>

		<div>
			<h4
				class="font-label-md text-[13px] font-bold text-primary-container mb-sm uppercase tracking-wider"
			>
				{t('landing.footer.support')}
			</h4>
			<div class="flex flex-col gap-2 font-body-md text-[14px]">
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/"
					>{t('landing.footer.support_links.docs')}</a
				>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/"
					>{t('landing.footer.support_links.api')}</a
				>
				<a class="text-on-surface-variant hover:text-white transition-colors" href="/"
					>{t('landing.footer.support_links.contact')}</a
				>
			</div>
		</div>
	</div>

	<div
		class="max-w-container-max mx-auto mt-lg pt-sm border-t border-white/5 flex flex-col md:flex-row items-center justify-between gap-sm"
	>
		<p class="font-body-md text-[13px] text-on-surface-variant/60">
			{t('landing.footer.copyright')}
		</p>
	</div>
</footer>
