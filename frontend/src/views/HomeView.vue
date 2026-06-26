<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="home-shell min-h-screen text-zinc-900 dark:text-zinc-100 font-sans antialiased selection:bg-zinc-200 dark:selection:bg-zinc-800 transition-colors duration-300">
    
    <!-- Header iOS style -->
    <header class="glass-header sticky top-0 z-40 border-b border-zinc-200/40 bg-[#f5f5f7]/70 px-4 py-3 sm:px-6 backdrop-blur-md dark:border-zinc-800/40 dark:bg-[#000000]/70">
      <nav class="mx-auto flex max-w-6xl items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain p-1" />
          </div>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold leading-5 text-zinc-950 dark:text-white tracking-tight">{{ siteName }}</div>
            <div class="hidden text-[10px] text-zinc-450 dark:text-zinc-500 sm:block tracking-tight font-medium">AI API Gateway Platform</div>
          </div>
        </div>

        <div class="hidden items-center gap-8 text-xs font-semibold text-zinc-600 dark:text-zinc-400 md:flex tracking-tight">
          <a href="#features" class="hover:text-zinc-950 dark:hover:text-white transition-colors">核心优势</a>
          <a href="#models" class="hover:text-zinc-950 dark:hover:text-white transition-colors">模型生态</a>
        </div>

        <div class="flex items-center gap-1 sm:gap-2">
          <LocaleSwitcher class="shrink-0" />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-icon-btn hidden sm:inline-flex shrink-0"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="sm" />
          </a>
          <button
            class="nav-icon-btn hidden sm:inline-flex shrink-0"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex h-8 items-center gap-1 rounded-full bg-zinc-900 px-3 sm:px-4 text-xs font-semibold text-white transition-all hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-100 shadow-sm shrink-0 whitespace-nowrap active:scale-95"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
            <Icon name="arrowRight" size="xs" />
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative">
      
      <!-- Hero Section with background image -->
      <section class="hero-section flex min-h-[90svh] flex-col items-center justify-center px-4 text-center relative overflow-hidden">
        
        <!-- Mask overlay to ensure text readability -->
        <div class="absolute inset-0 bg-white/5 dark:bg-black/30 pointer-events-none z-0"></div>

        <div class="relative z-10 max-w-4xl mx-auto flex flex-col items-center pt-24 pb-16">
          <!-- Dynamic iOS style Badge -->
          <div class="animate-fade-up animate-float inline-flex items-center gap-1.5 px-3 py-1 mb-8 text-[11px] font-semibold rounded-full bg-white/80 text-zinc-800 dark:bg-zinc-900/60 dark:text-zinc-200 border border-zinc-200/50 dark:border-zinc-800/50 backdrop-blur-md shadow-sm hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors">
            <span class="flex h-1.5 w-1.5 relative">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500"></span>
            </span>
            <span>已全面适配 DeepSeek-R1 与 Claude 3.5 全系列大模型</span>
          </div>

          <!-- Large, spacious, Apple-style heading -->
          <h1 class="animate-fade-up animate-fade-up-2 leading-[1.1] tracking-[-0.03em] text-4xl sm:text-6xl md:text-7xl lg:text-8xl font-extrabold text-zinc-950 dark:text-white max-w-4xl mb-6">
            连接全球智慧<br class="hidden sm:inline"/>赋能 AI 应用创新
          </h1>
          
          <p class="animate-fade-up animate-fade-up-3 text-zinc-650 dark:text-zinc-300 text-base sm:text-lg md:text-xl max-w-2xl leading-relaxed mb-10 tracking-tight font-normal">
            {{ siteSubtitle || '专为开发者与企业打造的下一代 AI API 网关底座。统一聚合全球主流大模型，提供毫秒级智能调度、渠道灾备与精细计费。' }}
          </p>

          <div class="animate-fade-up animate-fade-up-4 flex flex-wrap justify-center items-center gap-4 sm:gap-6">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center gap-1.5 px-8 py-3.5 rounded-full text-xs font-semibold text-white bg-zinc-950 hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-100 transition-all duration-200 shadow-md hover:-translate-y-0.5 active:translate-y-0"
            >
              <span class="relative z-10 flex items-center gap-1.5">
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="xs" />
              </span>
            </router-link>
            
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 text-xs font-semibold text-zinc-600 hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white transition-colors py-3"
            >
              <span>查看开发文档</span>
              <Icon name="chevronRight" size="xs" />
            </a>
          </div>
        </div>
      </section>

      <!-- Features Section (Simplified and Refined) -->
      <section id="features" class="py-24 border-t border-zinc-200/20 dark:border-zinc-900/40 relative">
        <div class="mx-auto max-w-6xl px-4">
          <div class="reveal-element section-heading max-w-2xl mx-auto text-center space-y-4 mb-16">
            <p class="eyebrow text-[10px] font-bold tracking-widest uppercase text-zinc-450 dark:text-zinc-500">Core Features</p>
            <h2 class="text-3xl sm:text-4xl lg:text-5xl font-bold text-zinc-900 dark:text-white tracking-tight leading-tight">卓越网关性能，赋能业务增长</h2>
            <p class="text-zinc-555 dark:text-zinc-455 text-sm tracking-tight">为您提供大模型集成与交付的一站式基础设施服务。</p>
          </div>

          <div class="reveal-element delay-100 grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8">
            <!-- Card 1 -->
            <div class="feature-card p-8 rounded-3xl border border-zinc-200/50 dark:border-zinc-800 bg-white/40 dark:bg-zinc-900/20 backdrop-blur-sm shadow-sm transition-all duration-300">
              <div class="inline-flex p-3 rounded-2xl bg-zinc-100 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 border border-zinc-200/50 dark:border-zinc-700/50 shadow-sm mb-6">
                <Icon name="server" size="md" />
              </div>
              <h3 class="text-lg font-bold text-zinc-950 mb-3 dark:text-white">统一聚合 简易集成</h3>
              <p class="text-zinc-600 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight">
                统一多种主流大模型协议。支持流式传输与函数调用，通过单一 API 密钥，在几行代码内接入全球顶尖 AI 能力。
              </p>
            </div>

            <!-- Card 2 -->
            <div class="feature-card p-8 rounded-3xl border border-zinc-200/50 dark:border-zinc-800 bg-white/40 dark:bg-zinc-900/20 backdrop-blur-sm shadow-sm transition-all duration-300">
              <div class="inline-flex p-3 rounded-2xl bg-zinc-100 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 border border-zinc-200/50 dark:border-zinc-700/50 shadow-sm mb-6">
                <Icon name="trendingUp" size="md" />
              </div>
              <h3 class="text-lg font-bold text-zinc-950 mb-3 dark:text-white">智能调度 稳定灾备</h3>
              <p class="text-zinc-600 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight">
                基于并发数、可用性与耗时自动分发。首字延迟深度优化，支持故障自动重试与上游负载均衡，保障服务始终在线。
              </p>
            </div>

            <!-- Card 3 -->
            <div class="feature-card p-8 rounded-3xl border border-zinc-200/50 dark:border-zinc-800 bg-white/40 dark:bg-zinc-900/20 backdrop-blur-sm shadow-sm transition-all duration-300">
              <div class="inline-flex p-3 rounded-2xl bg-zinc-100 dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 border border-zinc-200/50 dark:border-zinc-700/50 shadow-sm mb-6">
                <Icon name="shield" size="md" />
              </div>
              <h3 class="text-lg font-bold text-zinc-950 mb-3 dark:text-white">精细管控 安全合规</h3>
              <p class="text-zinc-600 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight">
                支持密钥级额度限制、RPM/TPM 频率控制与 Token 细粒度审计。内置合规治理模块，有效规避滥用风险。
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- Model Ecosystem Section -->
      <section id="models" class="py-24 border-t border-zinc-200/20 dark:border-zinc-900/40 relative">
        <div class="mx-auto max-w-6xl px-4">
          <div class="reveal-element section-heading max-w-2xl mx-auto text-center space-y-4 mb-16">
            <p class="eyebrow text-[10px] font-bold tracking-widest uppercase text-zinc-450 dark:text-zinc-500">Model Ecosystem</p>
            <h2 class="text-3xl sm:text-4xl lg:text-5xl font-bold text-zinc-900 dark:text-white tracking-tight leading-tight">全球大模型，一键就绪</h2>
            <p class="text-zinc-555 dark:text-zinc-455 text-sm tracking-tight">精选全球顶尖 AI 模型，提供全生命周期接入调度。</p>
          </div>

          <div class="reveal-element delay-100 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
            <div
              v-for="model in modelStrip"
              :key="model"
              class="px-4 py-3 rounded-2xl text-center text-xs font-semibold bg-white/40 dark:bg-zinc-900/20 border border-zinc-200/50 dark:border-zinc-800 backdrop-blur-sm text-zinc-750 dark:text-zinc-350 shadow-sm hover:border-zinc-355 dark:hover:border-zinc-700 transition-colors"
            >
              {{ model }}
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="border-t border-zinc-200/20 bg-[#f5f5f7]/60 px-4 py-12 backdrop-blur-md dark:border-zinc-900/40 dark:bg-[#000000]/60">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 text-center sm:flex-row sm:text-left">
        <p class="text-xs font-medium text-zinc-500 dark:text-zinc-455">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-6 text-xs font-semibold">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-zinc-500 hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white transition-colors"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-zinc-500 hover:text-zinc-950 dark:text-zinc-400 dark:hover:text-white transition-colors"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'CCAPI')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/Wei-Shaw/ccapi'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const currentYear = computed(() => new Date().getFullYear())

const modelStrip = [
  'DeepSeek-R1',
  'Claude 3.5 Sonnet',
  'GPT-4o',
  'Gemini 1.5 Pro',
  'Qwen-Max',
  'GLM-4'
]

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  // Set up intersection observer for scroll reveal animations
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('reveal-active')
        observer.unobserve(entry.target)
      }
    })
  }, {
    root: null,
    rootMargin: '0px',
    threshold: 0.1
  })

  setTimeout(() => {
    document.querySelectorAll('.reveal-element').forEach((el) => {
      observer.observe(el)
    })
  }, 100)
})
</script>

<style scoped>
.nav-icon-btn {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: rgb(113 113 122);
  transition: all 180ms ease;
}

.nav-icon-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: rgb(9 9 11);
}

.nav-icon-btn:active {
  transform: scale(0.95);
}

.dark .nav-icon-btn {
  color: rgb(161 161 170);
}

.dark .nav-icon-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

/* Glass Header base style */
.glass-header {
  position: fixed !important;
  inset: 0 0 auto 0;
  z-index: 40;
}

/* Full screen fixed background for the entire home shell */
.home-shell {
  position: relative;
}

.home-shell::before {
  content: '';
  position: fixed;
  inset: 0;
  z-index: -1;
  background: #f5f5f7 url('/ccapi-home-bg.webp') center / cover no-repeat;
  pointer-events: none;
}

.dark .home-shell::before {
  background: #000000 url('/ccapi-home-bg.webp') center / cover no-repeat;
  background-blend-mode: multiply;
}

/* Hero Section */
.hero-section {
  position: relative;
  min-height: 90svh;
  width: 100%;
  overflow: hidden;
  isolation: isolate;
}

/* Feature card styles */
.feature-card {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.feature-card:hover {
  transform: translateY(-4px);
  border-color: rgba(161, 161, 170, 0.4);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.04);
}

.dark .feature-card:hover {
  border-color: rgba(161, 161, 170, 0.2);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

.dark .feature-card {
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.02);
}

/* Scroll Reveal Animations */
.reveal-element {
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}
.reveal-active {
  opacity: 1;
  transform: translateY(0);
}
.delay-100 { transition-delay: 100ms; }

/* Hero Load Animations */
.animate-fade-up {
  animation: fadeUp 1s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}
.animate-fade-up-1 { animation-delay: 0.1s; }
.animate-fade-up-2 { animation-delay: 0.2s; }
.animate-fade-up-3 { animation-delay: 0.3s; }
.animate-fade-up-4 { animation-delay: 0.4s; }

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Micro-animations */
.animate-float {
  animation: float 6s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-6px); }
}

</style>
