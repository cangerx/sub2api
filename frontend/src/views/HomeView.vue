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
  <div v-else class="home-shell min-h-screen text-zinc-900 dark:text-zinc-150 font-sans antialiased selection:bg-zinc-200 dark:selection:bg-zinc-800 transition-colors duration-300">
    
    <!-- Header iOS style -->
    <header class="glass-header sticky top-0 z-30 border-b border-zinc-200/40 bg-[#f5f5f7]/70 px-5 py-3.5 backdrop-blur-md dark:border-zinc-800/40 dark:bg-[#000000]/70">
      <nav class="mx-auto flex max-w-6xl items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-sm dark:border-zinc-850 dark:bg-zinc-900">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain p-1" />
          </div>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold leading-5 text-zinc-950 dark:text-white tracking-tight">{{ siteName }}</div>
            <div class="hidden text-[10px] text-zinc-450 dark:text-zinc-500 sm:block tracking-tight font-medium">AI API Gateway Platform</div>
          </div>
        </div>

        <div class="hidden items-center gap-8 text-xs font-semibold text-zinc-650 dark:text-zinc-400 md:flex tracking-tight">
          <a href="#platform" class="hover:text-zinc-950 dark:hover:text-white transition-colors">平台能力</a>
          <a href="#model-matrix" class="hover:text-zinc-950 dark:hover:text-white transition-colors">模型聚合</a>
          <a href="#video-models" class="hover:text-zinc-950 dark:hover:text-white transition-colors">视频模型</a>
          <a href="#china-models" class="hover:text-zinc-950 dark:hover:text-white transition-colors">国产模型</a>
        </div>

        <div class="flex items-center gap-1.5">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-icon-btn"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="sm" />
          </a>
          <button
            class="nav-icon-btn"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex h-8 items-center gap-1 rounded-full bg-zinc-900 px-4 text-xs font-semibold text-white transition-all hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-100 shadow-sm ml-2 active:scale-95"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
            <Icon name="arrowRight" size="xs" />
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative">
      
      <!-- Hero Section with background image -->
      <section class="hero-section flex min-h-screen flex-col items-center justify-center px-5 text-center relative overflow-hidden">
        
        <!-- Mask overlay to ensure text readability -->
        <div class="absolute inset-0 bg-white/10 dark:bg-black/40 pointer-events-none z-0"></div>

        <div class="relative z-10 max-w-5xl mx-auto flex flex-col items-center pt-24 pb-20">
          <!-- Dynamic iOS style Badge -->
          <div class="inline-flex items-center gap-1.5 px-3.5 py-1.5 mb-10 text-[11px] font-semibold rounded-full bg-white/90 text-zinc-850 dark:bg-zinc-900/80 dark:text-zinc-200 border border-zinc-200/60 dark:border-zinc-800/80 backdrop-blur-md shadow-sm hover:border-zinc-350 dark:hover:border-zinc-700 transition-colors">
            <span class="flex h-1.5 w-1.5 relative">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-blue-500"></span>
            </span>
            <span>已全面适配 DeepSeek-R1 与 Claude 3.5 全系列大模型</span>
          </div>

          <p class="hero-eyebrow text-xs sm:text-sm font-bold tracking-widest text-zinc-500 dark:text-zinc-400 uppercase mb-4">{{ siteName }} &middot; 下一代 AI API 路由与分发底座</p>
          
          <!-- Large, spacious, Apple-style heading -->
          <h1 class="leading-[1.03] tracking-[-0.035em] text-5xl sm:text-7xl md:text-8xl lg:text-[6.25rem] font-bold text-zinc-950 dark:text-white max-w-5xl mb-8">
            一个 API 入口，<br/>接入全球大模型智能。
          </h1>
          
          <p class="hero-subtitle text-zinc-600 dark:text-zinc-350 text-base sm:text-lg md:text-xl max-w-3xl leading-relaxed mb-12 tracking-tight font-normal">
            {{ siteSubtitle }}。专为企业与开发者打造的高性能 API 网关底座，统一聚合 ChatGPT、Claude、Gemini、DeepSeek 及国产与视频生成模型，提供超低延迟的智能分发、精细计费与企业级安全治理能力。
          </p>

          <div class="hero-actions flex flex-wrap justify-center items-center gap-8">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="inline-flex items-center gap-1.5 px-9 py-4 rounded-full text-xs font-semibold text-white bg-zinc-950 hover:bg-zinc-800 dark:bg-white dark:text-zinc-950 dark:hover:bg-zinc-100 transition-all duration-205 shadow-sm hover:-translate-y-0.5 active:translate-y-0"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="xs" />
            </router-link>
            
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 dark:text-blue-400 hover:underline"
            >
              <span>查看开发文档</span>
              <Icon name="chevronRight" size="xs" class="mt-0.5" />
            </a>
          </div>

          <div class="hero-models flex flex-wrap justify-center gap-2 max-w-2xl mx-auto mt-20" aria-label="Supported model families">
            <span
              v-for="model in modelStrip"
              :key="model"
              class="px-4.5 py-2 rounded-full text-[10px] font-bold bg-white/70 dark:bg-zinc-900/50 border border-zinc-200/50 dark:border-zinc-800/60 backdrop-blur-sm text-zinc-550 dark:text-zinc-400 shadow-sm"
            >
              {{ model }}
            </span>
          </div>
        </div>
      </section>

      <!-- Bento Grid Features Section (Apple grid style) -->
      <section id="platform" class="py-36 border-t border-zinc-200/30 dark:border-zinc-900/60 relative">
        <div class="mx-auto max-w-6xl px-5">
          <div class="section-heading max-w-2xl mx-auto text-center space-y-4 mb-20">
            <p class="eyebrow text-[11px] font-bold tracking-wider uppercase text-zinc-450 dark:text-zinc-500">Platform Features</p>
            <h2 class="text-3xl sm:text-5xl lg:text-6xl font-bold text-zinc-900 dark:text-white tracking-tight leading-tight">专为大模型交付与商业化而生</h2>
            <p class="text-zinc-500 dark:text-zinc-400 text-sm sm:text-base tracking-tight font-medium">提供全链路的大模型聚合、高可靠调度、精细化计费及安全合规策略，助您构建稳健的 AI 基础设施。</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            
            <!-- Bento Card 1: API Aggregation (Spans 2 columns on desktop) -->
            <div class="col-span-1 md:col-span-2 bento-card p-8 sm:p-10 lg:p-12 rounded-[2.5rem] border border-zinc-200/60 dark:border-zinc-900 bg-white dark:bg-[#111112] shadow-sm hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-800 transition-all duration-300 flex flex-col md:flex-row gap-8 items-center overflow-hidden">
              <div class="flex-1 space-y-4">
                <div class="inline-flex p-3.5 rounded-2xl bg-zinc-50 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200/50 dark:border-zinc-800 shadow-sm">
                  <Icon name="server" size="lg" />
                </div>
                <h3 class="text-xl sm:text-2xl lg:text-3xl font-bold text-zinc-900 dark:text-white tracking-tight">API 统一聚合</h3>
                <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight font-medium">
                  统一多种主流大模型协议，支持流式传输与函数调用。通过单一入口和标准化鉴权，使您的应用在几行代码内获得全球顶尖 AI 能力。
                </p>
              </div>
              
              <!-- Interactive Terminal mockup -->
              <div class="w-full md:w-80 shrink-0 font-mono text-[10px] rounded-2xl border border-zinc-200/80 dark:border-zinc-800/80 bg-zinc-950 text-zinc-350 shadow-xl overflow-hidden self-stretch flex flex-col justify-between min-h-[11rem]">
                <div class="flex items-center justify-between px-4 py-2.5 border-b border-zinc-900 bg-zinc-900/50">
                  <div class="flex items-center gap-1.5">
                    <span class="w-2.5 h-2.5 rounded-full bg-zinc-800"></span>
                    <span class="w-2.5 h-2.5 rounded-full bg-zinc-800"></span>
                    <span class="w-2.5 h-2.5 rounded-full bg-zinc-800"></span>
                  </div>
                  <span class="text-zinc-500 text-[9px] font-bold tracking-tight">Unified Request</span>
                </div>
                <div class="p-4.5 space-y-1.5 overflow-x-auto flex-1 select-none">
                  <div class="text-zinc-500">// POST /v1/chat/completions</div>
                  <div><span class="text-pink-450 font-bold">const</span> res = <span class="text-pink-450">await</span> ccapi.chat({</div>
                  <div class="pl-4">model: <span class="text-emerald-400">"deepseek-r1"</span>,</div>
                  <div class="pl-4">messages: [{</div>
                  <div class="pl-8">role: <span class="text-emerald-400">"user"</span>,</div>
                  <div class="pl-8">content: <span class="text-emerald-400">"你好"</span></div>
                  <div class="pl-4">}]</div>
                  <div>});</div>
                </div>
              </div>
            </div>

            <!-- Bento Card 2: Routing -->
            <div class="bento-card p-8 sm:p-10 lg:p-12 rounded-[2.5rem] border border-zinc-200/60 dark:border-zinc-900 bg-white dark:bg-[#111112] shadow-sm hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-800 transition-all duration-300 flex flex-col justify-between gap-6">
              <div class="space-y-4">
                <div class="inline-flex p-3.5 rounded-2xl bg-zinc-50 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200/50 dark:border-zinc-800 shadow-sm">
                  <Icon name="users" size="lg" />
                </div>
                <h3 class="text-xl sm:text-2xl lg:text-3xl font-bold text-zinc-900 dark:text-white tracking-tight">智能路由与分发</h3>
                <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight font-medium">
                  基于权重、优先级和可用性进行秒级智能调度，自动规避限频限流与供应商故障，保障服务持续在线。
                </p>
              </div>
              
              <!-- Visual Routing Graph -->
              <div class="h-28 flex items-center justify-between px-6 bg-zinc-50 dark:bg-zinc-900/30 rounded-2xl border border-zinc-100 dark:border-zinc-900/60 relative overflow-hidden">
                <div class="flex items-center z-10">
                  <div class="w-8 h-8 rounded-lg bg-zinc-900 dark:bg-zinc-800 text-white flex items-center justify-center font-bold text-[9px] shadow-sm">API</div>
                </div>
                <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
                  <svg class="w-full h-full" viewBox="0 0 200 60">
                    <path d="M 40,30 L 160,15" stroke="currentColor" class="text-zinc-200 dark:text-zinc-800" stroke-width="1.2" stroke-dasharray="3 3" />
                    <path d="M 40,30 L 160,30" stroke="currentColor" class="text-zinc-200 dark:text-zinc-800" stroke-width="1.2" stroke-dasharray="3 3" />
                    <path d="M 40,30 L 160,45" stroke="currentColor" class="text-zinc-200 dark:text-zinc-800" stroke-width="1.2" stroke-dasharray="3 3" />
                    <circle r="1.8" fill="#a1a1aa">
                      <animateMotion path="M 40,30 L 160,15" dur="2s" repeatCount="indefinite" />
                    </circle>
                    <circle r="1.8" fill="#a1a1aa">
                      <animateMotion path="M 40,30 L 160,30" dur="2.8s" repeatCount="indefinite" />
                    </circle>
                    <circle r="1.8" fill="#a1a1aa">
                      <animateMotion path="M 40,30 L 160,45" dur="2.4s" repeatCount="indefinite" />
                    </circle>
                  </svg>
                </div>
                <div class="flex flex-col gap-2.5 z-10 text-[7.5px] font-bold">
                  <div class="px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-850 text-zinc-550 dark:text-zinc-400 border border-zinc-200/50 dark:border-zinc-800/80">OpenAI</div>
                  <div class="px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-850 text-zinc-550 dark:text-zinc-400 border border-zinc-200/50 dark:border-zinc-800/80">Claude</div>
                  <div class="px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-850 text-zinc-550 dark:text-zinc-400 border border-zinc-200/50 dark:border-zinc-800/80">DeepSeek</div>
                </div>
              </div>
            </div>

            <!-- Bento Card 3: Billing -->
            <div class="bento-card p-8 sm:p-10 lg:p-12 rounded-[2.5rem] border border-zinc-200/60 dark:border-zinc-900 bg-white dark:bg-[#111112] shadow-sm hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-800 transition-all duration-300 flex flex-col justify-between gap-6">
              <div class="space-y-4">
                <div class="inline-flex p-3.5 rounded-2xl bg-zinc-50 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200/50 dark:border-zinc-800 shadow-sm">
                  <Icon name="chart" size="lg" />
                </div>
                <h3 class="text-xl sm:text-2xl lg:text-3xl font-bold text-zinc-900 dark:text-white tracking-tight">计费与用量核算</h3>
                <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight font-medium">
                  提供高并发、低延迟的计费计量引擎，支持实时 Token 消耗审计与精细余额核算，为企业内部划转与代理分销奠定数据基础。
                </p>
              </div>
              
              <!-- Billing indicator mockup -->
              <div class="p-4 bg-zinc-50 dark:bg-zinc-900/30 rounded-2xl border border-zinc-100 dark:border-zinc-900/60 space-y-2">
                <div class="flex items-center justify-between text-[9px] font-bold text-zinc-500 dark:text-zinc-400">
                  <span>账户剩余配额</span>
                  <span class="text-zinc-800 dark:text-zinc-200">84.2%</span>
                </div>
                <div class="h-1.5 w-full bg-zinc-250 dark:bg-zinc-800 rounded-full overflow-hidden">
                  <div class="h-full bg-zinc-900 dark:bg-zinc-100 rounded-full" style="width: 84.2%"></div>
                </div>
                <div class="flex justify-between items-center text-[8px] text-zinc-400 font-bold">
                  <span>已消费: $15.80</span>
                  <span>总计: $100.00</span>
                </div>
              </div>
            </div>

            <!-- Bento Card 4: Security (Spans 2 columns on desktop) -->
            <div class="col-span-1 md:col-span-2 bento-card p-8 sm:p-10 lg:p-12 rounded-[2.5rem] border border-zinc-200/60 dark:border-zinc-900 bg-white dark:bg-[#111112] shadow-sm hover:shadow-md hover:border-zinc-300 dark:hover:border-zinc-800 transition-all duration-300 flex flex-col md:flex-row gap-8 items-center overflow-hidden">
              <div class="flex-1 space-y-4">
                <div class="inline-flex p-3.5 rounded-2xl bg-zinc-50 dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 border border-zinc-200/50 dark:border-zinc-800 shadow-sm">
                  <Icon name="shield" size="lg" />
                </div>
                <h3 class="text-xl sm:text-2xl lg:text-3xl font-bold text-zinc-900 dark:text-white tracking-tight">安全与治理体系</h3>
                <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed tracking-tight font-medium">
                  提供密钥细粒度额度控制、高频调用速率限制（RPM/TPM）及全面行为审计拦截，有效防御恶意滥用并确保数据合规。
                </p>
              </div>
              
              <!-- Security UI control board -->
              <div class="w-full md:w-80 shrink-0 p-4.5 bg-zinc-50 dark:bg-zinc-900/30 rounded-2xl border border-zinc-100 dark:border-zinc-900/60 space-y-3.5 text-xs select-none">
                <div class="flex items-center justify-between border-b border-zinc-200/40 dark:border-zinc-800/40 pb-2">
                  <span class="font-bold text-zinc-800 dark:text-zinc-200 text-[10px]">Access Policy Rule</span>
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[8.5px] bg-zinc-150 dark:bg-zinc-800 text-zinc-650 dark:text-zinc-300 font-bold border border-zinc-200/20 dark:border-zinc-700/30">Active</span>
                </div>
                <div class="space-y-2.5 font-medium">
                  <div class="flex justify-between items-center text-[9.5px]">
                    <span class="text-zinc-505">模型分组:</span>
                    <div class="flex gap-1 font-bold text-[8.5px]">
                      <span class="px-1.5 py-0.5 rounded bg-zinc-200 dark:bg-zinc-800 text-zinc-650 dark:text-zinc-400">Chat</span>
                      <span class="px-1.5 py-0.5 rounded bg-zinc-200 dark:bg-zinc-800 text-zinc-650 dark:text-zinc-400">Video</span>
                    </div>
                  </div>
                  <div class="flex justify-between items-center text-[9.5px]">
                    <span class="text-zinc-505">频率限制 (Rate Limit):</span>
                    <span class="font-bold text-zinc-700 dark:text-zinc-300">60 RPM / 3000 TPM</span>
                  </div>
                  <div class="flex justify-between items-center text-[9.5px]">
                    <span class="text-zinc-505">安全拦截 (Auditing):</span>
                    <span class="inline-flex items-center gap-1 text-zinc-855 dark:text-zinc-200 font-bold">
                      <Icon name="checkCircle" size="xs" /> 已启用
                    </span>
                  </div>
                </div>
              </div>
            </div>
            
          </div>
        </div>
      </section>

      <!-- Interactive Model Matrix Section -->
      <section id="model-matrix" class="py-36 border-t border-b border-zinc-200/30 dark:border-zinc-900/60 relative">
        <div class="mx-auto max-w-6xl px-5">
          <div class="section-heading max-w-2xl mx-auto text-center space-y-4 mb-20">
            <p class="eyebrow text-[11px] font-bold tracking-wider uppercase text-zinc-450 dark:text-zinc-500">Model Aggregation</p>
            <h2 class="text-3xl sm:text-5xl lg:text-6xl font-bold text-zinc-900 dark:text-white tracking-tight leading-tight">全球顶尖大模型，一键就绪</h2>
            <p class="text-zinc-500 dark:text-zinc-400 text-sm sm:text-base tracking-tight font-medium">统一聚合全球主流语言、推理与多模态模型，通过标准的 OpenAI 协议提供极速分发与切换。</p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-6 mt-12">
            <div
              v-for="provider in providers"
              :key="provider.name"
              class="group p-6.5 rounded-[2rem] border border-zinc-200/60 dark:border-zinc-900 bg-white/50 dark:bg-zinc-900/50 hover:bg-[#ffffff] dark:hover:bg-[#1c1c1e] transition-all duration-350 hover:-translate-y-1 shadow-sm hover:shadow-md flex flex-col justify-between min-h-[14.5rem]"
            >
              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <div class="font-extrabold text-[10px] px-2.5 py-0.5 rounded-md bg-zinc-905 dark:bg-zinc-800 text-white dark:text-zinc-300 tracking-wider">
                    {{ provider.short }}
                  </div>
                  <span class="text-[9px] font-bold text-zinc-500 dark:text-zinc-450">
                    {{ provider.tag }}
                  </span>
                </div>
                <h3 class="text-base font-bold text-zinc-900 dark:text-white mt-1">{{ provider.name }}</h3>
                <p class="text-zinc-500 dark:text-zinc-400 text-xs leading-relaxed tracking-tight">{{ provider.description }}</p>
              </div>

              <!-- Mini metrics footer inside card -->
              <div class="pt-3 border-t border-zinc-200/50 dark:border-zinc-800/85 text-[9px] font-bold text-zinc-400 flex items-center justify-between">
                <span>可用率承诺:</span>
                <span class="text-zinc-700 dark:text-zinc-300">
                  {{ provider.short === 'GPT' ? '99.98%' : provider.short === 'C' ? '99.95%' : provider.short === 'G' ? '99.9%' : '99.99%' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Video Generation Section -->
      <section id="video-models" class="py-36 border-b border-zinc-200/30 dark:border-zinc-900/60 relative">
        <div class="mx-auto grid max-w-6xl gap-12 px-5 lg:grid-cols-[0.9fr_1.1fr] items-center">
          <div class="section-heading section-heading-left space-y-4">
            <p class="eyebrow text-[11px] font-bold tracking-wider uppercase text-zinc-450 dark:text-zinc-500">Video Generation</p>
            <h2 class="text-3xl sm:text-5xl lg:text-6xl font-bold text-zinc-900 dark:text-white tracking-tight leading-none">
              视频生成大模型，<br/>在同一套 API 内重塑想象
            </h2>
            <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed max-w-md tracking-tight font-medium">
              全面适配文生视频、图生视频及长视频生成协议。底座封装异步任务处理、查询、取消与状态轮询管理，让复杂的多媒体创作如文本对话般顺畅。
            </p>
          </div>

          <div class="grid grid-cols-1 gap-5">
            <div
              v-for="model in videoModels"
              :key="model.name"
              class="group p-6 rounded-[1.75rem] border border-zinc-200/50 dark:border-zinc-900 bg-white/50 dark:bg-zinc-900/50 shadow-sm hover:shadow-md transition-all duration-300 flex gap-5 items-center"
            >
              <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-zinc-50 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 border border-zinc-200/50 dark:border-zinc-800 shadow-sm transition-colors group-hover:bg-zinc-100 dark:group-hover:bg-zinc-750">
                <Icon name="video" size="md" />
              </div>
              <div class="space-y-0.5">
                <h3 class="text-sm font-bold text-zinc-900 dark:text-white tracking-tight">{{ model.name }}</h3>
                <p class="text-zinc-500 dark:text-zinc-400 text-xs leading-relaxed tracking-tight">{{ model.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Chinese Models Section -->
      <section id="china-models" class="py-36 border-b border-zinc-200/30 dark:border-zinc-900/60 relative">
        <div class="mx-auto grid max-w-6xl gap-12 px-5 lg:grid-cols-[0.9fr_1.1fr] items-center">
          <div class="section-heading section-heading-left space-y-4">
            <p class="eyebrow text-[11px] font-bold tracking-wider uppercase text-zinc-450 dark:text-zinc-500">China Models</p>
            <h2 class="text-3xl sm:text-5xl lg:text-6xl font-bold text-zinc-900 dark:text-white tracking-tight leading-none">
              国产大模型精选，<br/>兼顾合规与极致效能
            </h2>
            <p class="text-zinc-550 dark:text-zinc-400 text-xs sm:text-sm leading-relaxed max-w-md tracking-tight font-medium">
              为中文语境深度定制，提供极高性价比与网络合规保障。内置通义千问、DeepSeek、智谱 GLM、文心一言等，助力垂直行业与本土化业务快速落地。
            </p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div
              v-for="model in chinaModels"
              :key="model.name"
              class="group relative p-6 rounded-2xl border border-zinc-200/60 dark:border-zinc-900 bg-[#f5f5f7] dark:bg-[#111112] hover:bg-white dark:hover:bg-[#1c1c1e] hover:border-zinc-300 dark:hover:border-zinc-800 transition-all duration-300 min-h-[9.5rem] flex flex-col justify-between overflow-hidden shadow-sm hover:shadow-md"
            >
              <!-- Elegant subtle watermark monogram as text bg -->
              <span class="absolute -right-2 -bottom-5 font-mono font-bold text-7xl text-zinc-200/50 dark:text-zinc-900/10 pointer-events-none select-none transition-transform duration-300 group-hover:scale-105">
                {{ model.short }}
              </span>
              <div class="relative z-10 space-y-2">
                <div class="flex items-center justify-between">
                  <h3 class="text-sm font-bold text-zinc-900 dark:text-white tracking-tight">{{ model.name }}</h3>
                  <span class="text-[9px] font-bold text-zinc-550 dark:text-zinc-400 bg-white/90 dark:bg-zinc-800 px-2 py-0.5 rounded-full border border-zinc-200/50 dark:border-zinc-700/50">
                    {{ model.focus }}
                  </span>
                </div>
                <p class="text-zinc-550 dark:text-zinc-400 text-xs leading-relaxed max-w-[85%] tracking-tight">
                  {{ model.description }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="border-t border-zinc-200/40 bg-[#f5f5f7]/70 px-5 py-14 backdrop-blur-md dark:border-zinc-800/40 dark:bg-[#000000]/70">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 text-center sm:flex-row sm:text-left">
        <p class="text-xs font-semibold text-zinc-500 dark:text-zinc-550">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-6 text-xs font-bold">
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
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
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

const providers = [
  { short: 'GPT', name: 'ChatGPT / OpenAI', tag: '通用智能', description: '覆盖对话、工具调用、多模态理解和广泛生态集成。' },
  { short: 'C', name: 'Claude', tag: '推理 / 编码', description: '适合复杂推理、长文本、编码代理和高质量内容生成。' },
  { short: 'G', name: 'Gemini', tag: '多模态', description: '适合长上下文、视觉理解和 Google 生态工作流。' },
  { short: 'DS', name: 'DeepSeek', tag: '国产 / 推理', description: '面向中文、推理、代码与高并发极低成本调用场景。' }
]

const modelStrip = [
  'GPT-4o',
  'Claude 3.5 Sonnet',
  'Gemini 1.5 Pro',
  'DeepSeek-R1',
  'Qwen-Plus',
  'GLM-4',
  'Sora Video',
  'CogVideo'
]

const videoModels = [
  {
    name: '文生视频',
    description: '把文案、脚本和提示词转换成异步视频生成任务。'
  },
  {
    name: '图生视频',
    description: '以静态图片资产为起点生成动态影视级内容，适合产品展示和创意广告。'
  },
  {
    name: '视频工作流',
    description: '统一处理异步任务创建、查询、取消、下载以及上游厂商映射。'
  }
]

const chinaModels = [
  { short: 'QW', name: '通义千问', focus: '中文 / 企业', description: '适合中文场景问答、办公自动化、本地知识库检索与集成。' },
  { short: 'DS', name: 'DeepSeek', focus: '推理 / 高性价比', description: '在推理、代码和商业高并发调用中拥有无与伦比的性价比优势。' },
  { short: 'GLM', name: '智谱 GLM', focus: '知识 / 智能体', description: '适合 Agent、知识检索库、多轮任务编排和垂直行业场景。' },
  { short: 'ERN', name: '文心一言', focus: '生态 / 内容', description: '适合文案创意内容生成、百度搜索增强和中文原生应用开发。' }
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
  min-height: 100svh;
  width: 100%;
  overflow: hidden;
  isolation: isolate;
}

/* Bento grid custom designs */
.bento-card {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-card:hover {
  transform: translateY(-2px) scale(1.005);
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.03);
}

.dark .bento-card:hover {
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.2);
}

.dark .bento-card {
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

/* Apple style syntax highlighting overrides */
.text-pink-450 {
  color: #ff3b30;
}
.dark .text-pink-450 {
  color: #ff453a;
}
.text-emerald-400 {
  color: #34c759;
}
.dark .text-emerald-400 {
  color: #30d158;
}

.text-zinc-505 {
  color: rgb(113 113 122);
}
.dark .text-zinc-505 {
  color: rgb(161 161 170);
}

.text-zinc-905 {
  background-color: rgb(24 24 27);
}
.dark .text-zinc-905 {
  background-color: rgb(39 39 42);
}

.text-zinc-855 {
  color: rgb(39 39 42);
}
.dark .text-zinc-855 {
  color: rgb(228 228 231);
}

@media (max-width: 768px) {
  .hero-section {
    min-height: 100svh;
  }
}
</style>
