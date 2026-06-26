<template>
  <div class="relative flex min-h-screen bg-[#060a1f] overflow-hidden">
    <!-- Full Screen Background (Left side focus) -->
    <div class="absolute inset-0 z-0 pointer-events-none">
      <!-- Deep space gradients -->
      <div class="absolute -left-[20%] -bottom-[20%] w-[120%] h-[120%] bg-[radial-gradient(ellipse_at_bottom_left,rgba(0,188,212,0.12)_0%,transparent_50%)]"></div>
      <div class="absolute top-0 right-0 w-[100%] h-[100%] bg-[radial-gradient(ellipse_at_top_right,rgba(11,29,74,0.6)_0%,transparent_70%)]"></div>
      <!-- Subtle starlight grid -->
      <div class="absolute inset-0 bg-[radial-gradient(rgba(255,255,255,0.15)_1px,transparent_1px)] bg-[length:32px_32px] opacity-[0.15]"></div>
    </div>

    <!-- Top Navigation (Logo, Language) -->
    <div class="absolute top-0 w-full p-6 lg:px-12 flex justify-between items-center z-20">
      <!-- Logo -->
      <div class="flex items-center gap-3">
        <template v-if="settingsLoaded">
          <img :src="siteLogo || '/logo.png'" class="h-8 w-auto filter drop-shadow-md" alt="Logo" />
          <span class="text-white font-bold text-xl tracking-wide drop-shadow-md">{{ siteName }}</span>
        </template>
      </div>
      <!-- Right links -->
      <div class="text-white/60 text-xs tracking-wider cursor-pointer hover:text-white transition-colors">
        中 | En
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="relative z-10 w-full flex flex-col lg:flex-row items-center justify-between px-6 lg:px-20 py-24 min-h-screen">
      
      <!-- Left Typography (Hidden on small screens) -->
      <div class="hidden lg:flex flex-col justify-center flex-1 pl-10 xl:pl-24 h-full">
        <h1 class="text-white text-5xl xl:text-[5.5rem] font-extralight tracking-[0.15em] leading-[1.3] mb-6 drop-shadow-2xl" style="text-shadow: 0 10px 30px rgba(0,0,0,0.5)">
          科学发现<br />
          <span class="font-normal">探索未来</span>
        </h1>
        <p class="text-cyan-100/40 text-lg tracking-widest mt-6 max-w-md">
          {{ siteSubtitle }}
        </p>
      </div>

      <!-- Right Auth Card Container -->
      <div class="w-full max-w-[420px] lg:w-[460px] lg:max-w-none lg:mr-12 xl:mr-32 relative">
        <div class="bg-[#f8f9fc] dark:bg-zinc-900 rounded-[1.5rem] p-8 sm:p-10 shadow-[0_20px_50px_rgba(0,0,0,0.3)] relative overflow-hidden">
          <slot />
        </div>

        <!-- Footer slot -->
        <div class="mt-6 text-center text-sm relative z-10 text-white/80">
          <slot name="footer" />
        </div>

        <!-- Copyright -->
        <div class="mt-8 text-center text-[10px] text-white/20">
          &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'CCAPI')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>
