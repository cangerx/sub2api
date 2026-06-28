<template>
  <div class="auth-shell relative flex min-h-screen overflow-hidden">
    <!-- Shared home-style shader background -->
    <div class="auth-bg-overlay absolute inset-0 z-0 pointer-events-none">
      <div ref="authShaderEl" class="auth-grain-shader"></div>
    </div>
    <div class="auth-gradient-scrim absolute inset-0 z-0 pointer-events-none"></div>

    <!-- Top Navigation (Logo, Language) -->
    <div class="auth-topbar absolute top-0 w-full p-6 lg:px-12 flex justify-between items-center z-20">
      <router-link
        to="/"
        class="inline-flex items-center gap-2 rounded-full px-3 py-2 text-xs font-semibold text-white/82 transition hover:bg-white/12 hover:text-white"
      >
        <span class="text-base leading-none">&lsaquo;</span>
        <span>首页</span>
      </router-link>
      <!-- Right links -->
      <div class="text-white/60 text-xs tracking-wider cursor-pointer hover:text-white transition-colors">
        中 | En
      </div>
    </div>

    <!-- Main Content -->
    <div class="relative z-10 w-full min-h-screen px-6 py-24 flex items-center justify-center">
      <div class="auth-panel w-full max-w-[392px] relative">
        <div class="auth-brand mb-5 flex flex-col items-center gap-3">
          <img :src="siteLogo || '/logo.png'" class="h-9 w-auto drop-shadow-lg" alt="Logo" />
          <div class="text-sm font-semibold tracking-[0.18em] text-white drop-shadow-md">{{ siteName }}</div>
        </div>

        <div class="auth-card rounded-[1.25rem] p-7 sm:p-8 relative overflow-hidden">
          <slot />

          <div class="mt-7 text-center text-sm relative z-10 text-gray-500 dark:text-zinc-400">
            <slot name="footer" />
          </div>
        </div>

        <!-- Copyright -->
        <div class="mt-6 text-center text-[10px] tracking-wide text-white/35">
          &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, nextTick } from 'vue'
import {
  ShaderFitOptions,
  ShaderMount,
  defaultObjectSizing,
  grainGradientFragmentShader,
  GrainGradientShapes,
  getShaderColorFromString,
  getShaderNoiseTexture
} from '@paper-design/shaders'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const authShaderEl = ref<HTMLElement | null>(null)
let authShader: ShaderMount | null = null

const siteName = computed(() => appStore.siteName || 'CCAPI')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

const currentYear = computed(() => new Date().getFullYear())

function getAuthShaderTheme() {
  const isDark = document.documentElement.classList.contains('dark')

  return isDark
    ? {
        colorBack: '#030305',
        colors: ['#00D5FF', '#7C3AED', '#FF2E88', '#2563EB00'],
        intensity: 1,
        noise: 0.72
      }
    : {
        colorBack: '#F7FAFF',
        colors: ['#0EA5E9', '#8B5CF6', '#F43F5E', '#22D3EE00'],
        intensity: 0.96,
        noise: 0.5
      }
}

async function initAuthShader() {
  if (!authShaderEl.value || authShader) return

  try {
    await nextTick()

    const noiseTexture = getShaderNoiseTexture()
    if (noiseTexture) {
      await waitForShaderImage(noiseTexture)
    }

    const sizing = defaultObjectSizing
    const theme = getAuthShaderTheme()

    authShader = new ShaderMount(
      authShaderEl.value,
      grainGradientFragmentShader,
      {
        u_fit: ShaderFitOptions.cover,
        u_scale: 1,
        u_rotation: sizing.rotation,
        u_originX: sizing.originX,
        u_originY: sizing.originY,
        u_offsetX: sizing.offsetX,
        u_offsetY: sizing.offsetY,
        u_worldWidth: sizing.worldWidth,
        u_worldHeight: sizing.worldHeight,
        u_colorBack: getShaderColorFromString(theme.colorBack),
        u_colors: theme.colors.map(getShaderColorFromString),
        u_colorsCount: theme.colors.length,
        u_softness: 1,
        u_intensity: theme.intensity,
        u_noise: theme.noise,
        u_shape: GrainGradientShapes.corners,
        u_noiseTexture: noiseTexture
      },
      { alpha: true, antialias: true, premultipliedAlpha: false },
      1.85,
      0,
      1,
      1920 * 1080 * 2
    )
    authShaderEl.value?.classList.add('is-ready')
  } catch (error) {
    console.warn('[auth] Paper shader background disabled:', error)
    authShaderEl.value?.classList.remove('is-ready')
    authShader = null
  }
}

function waitForShaderImage(image: HTMLImageElement) {
  if (image.complete && image.naturalWidth > 0) {
    return Promise.resolve()
  }

  return new Promise<void>((resolve, reject) => {
    image.onload = () => resolve()
    image.onerror = () => reject(new Error('Paper shader noise texture failed to load'))
  })
}

onMounted(() => {
  appStore.fetchPublicSettings()
  void initAuthShader()
})

onBeforeUnmount(() => {
  authShader?.dispose()
  authShader = null
})
</script>

<style scoped>
.auth-shell {
  background-color: #f7f8fb;
}

.dark .auth-shell {
  background-color: #050505;
}

.auth-bg-overlay {
  overflow: hidden;
  background:
    radial-gradient(circle at 16% 18%, rgba(14, 165, 233, 0.88), transparent 34%),
    radial-gradient(circle at 78% 20%, rgba(139, 92, 246, 0.8), transparent 36%),
    radial-gradient(circle at 58% 76%, rgba(244, 63, 94, 0.72), transparent 38%),
    #f8fafc;
}

.dark .auth-bg-overlay {
  background:
    radial-gradient(circle at 18% 18%, rgba(0, 213, 255, 0.68), transparent 34%),
    radial-gradient(circle at 78% 24%, rgba(124, 58, 237, 0.62), transparent 36%),
    radial-gradient(circle at 58% 76%, rgba(255, 46, 136, 0.54), transparent 38%),
    #050505;
}

.auth-grain-shader {
  position: absolute;
  inset: 0;
  opacity: 1;
}

.auth-grain-shader canvas {
  display: block;
  width: 100%;
  height: 100%;
}

.auth-gradient-scrim {
  background:
    radial-gradient(circle at 50% 42%, rgba(255, 255, 255, 0.02), rgba(255, 255, 255, 0.24) 72%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.1), rgba(248, 250, 252, 0.34));
}

.dark .auth-gradient-scrim {
  background:
    radial-gradient(circle at 50% 42%, rgba(0, 0, 0, 0.06), rgba(0, 0, 0, 0.34) 68%),
    linear-gradient(180deg, rgba(0, 0, 0, 0.08), rgba(0, 0, 0, 0.38));
}

.auth-topbar {
  animation: authTopbarEnter 700ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

.auth-brand {
  animation: authBrandEnter 760ms cubic-bezier(0.16, 1, 0.3, 1) 80ms both;
}

.auth-card {
  transform-origin: center;
  animation: authCardEnter 760ms cubic-bezier(0.16, 1, 0.3, 1) 120ms both;
  border: 1px solid rgba(255, 255, 255, 0.84);
  background: rgba(255, 255, 255, 0.94);
  box-shadow:
    0 28px 80px rgba(37, 64, 120, 0.22),
    0 8px 24px rgba(15, 23, 42, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.96);
  -webkit-backdrop-filter: blur(18px) saturate(1.18);
  backdrop-filter: blur(18px) saturate(1.18);
  transition:
    transform 260ms cubic-bezier(0.2, 0.8, 0.2, 1),
    box-shadow 260ms ease,
    border-color 260ms ease;
}

.auth-card:hover {
  transform: translateY(-2px);
  box-shadow:
    0 34px 92px rgba(37, 64, 120, 0.26),
    0 10px 28px rgba(15, 23, 42, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.98);
}

.dark .auth-card {
  border-color: rgba(255, 255, 255, 0.16);
  background: rgba(15, 15, 18, 0.84);
  box-shadow:
    0 28px 82px rgba(0, 0, 0, 0.46),
    0 8px 24px rgba(0, 0, 0, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.14);
}

.dark .auth-card:hover {
  box-shadow:
    0 34px 92px rgba(0, 0, 0, 0.54),
    0 10px 28px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.auth-panel {
  animation: authPanelEnter 760ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes authCardEnter {
  from {
    opacity: 0;
    transform: translate3d(0, 18px, 0) scale(0.975);
    filter: blur(8px);
  }
  to {
    opacity: 1;
    transform: translate3d(0, 0, 0) scale(1);
    filter: blur(0);
  }
}

@keyframes authPanelEnter {
  from {
    opacity: 0;
    transform: translate3d(0, 18px, 0);
  }
  to {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

@keyframes authBrandEnter {
  from {
    opacity: 0;
    transform: translate3d(0, 10px, 0);
    filter: blur(6px);
  }
  to {
    opacity: 1;
    transform: translate3d(0, 0, 0);
    filter: blur(0);
  }
}

@keyframes authTopbarEnter {
  from {
    opacity: 0;
    transform: translate3d(0, -10px, 0);
  }
  to {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-card,
  .auth-brand,
  .auth-topbar,
  .auth-panel {
    animation: none;
  }
}

@media (max-width: 640px) {
  .auth-shell {
    min-height: 100svh;
    overflow-x: hidden;
    overflow-y: auto;
  }

  .auth-topbar {
    padding: max(0.9rem, env(safe-area-inset-top)) 1rem 0;
  }

  .auth-topbar > a,
  .auth-topbar > div {
    background: rgba(255, 255, 255, 0.13);
    border: 1px solid rgba(255, 255, 255, 0.16);
    -webkit-backdrop-filter: blur(16px) saturate(1.25);
    backdrop-filter: blur(16px) saturate(1.25);
  }

  .auth-shell > .relative.z-10 {
    min-height: 100svh;
    align-items: flex-start;
    padding: 5.25rem 1rem 2rem;
  }

  .auth-panel {
    max-width: 24rem;
  }

  .auth-brand {
    margin-bottom: 1rem;
    gap: 0.55rem;
  }

  .auth-brand img {
    height: 2rem;
  }

  .auth-card {
    border-radius: 1.15rem;
    padding: 1.35rem;
    box-shadow:
      0 22px 62px rgba(37, 64, 120, 0.2),
      0 6px 20px rgba(15, 23, 42, 0.08),
      inset 0 1px 0 rgba(255, 255, 255, 0.96);
  }

  .auth-card:hover {
    transform: none;
  }
}

@media (max-width: 380px) {
  .auth-shell > .relative.z-10 {
    padding-left: 0.75rem;
    padding-right: 0.75rem;
  }

  .auth-card {
    padding: 1.1rem;
  }
}
</style>
