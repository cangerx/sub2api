<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="row"
        class="fixed inset-0 z-[10000] flex items-end justify-center bg-gray-950/65 p-0 backdrop-blur-md sm:items-center sm:p-6"
        @click.self="close"
      >
        <div class="flex h-[96dvh] w-full max-w-7xl flex-col overflow-hidden rounded-t-2xl border border-white/70 bg-white shadow-2xl ring-1 ring-black/5 dark:border-white/10 dark:bg-dark-900 dark:ring-white/10 sm:h-auto sm:max-h-[90vh] sm:rounded-2xl">
          <header class="flex flex-col gap-4 border-b border-gray-200/80 bg-white/95 px-5 py-4 backdrop-blur-xl dark:border-dark-700/80 dark:bg-dark-900/95 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex min-w-0 items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200">
                <Icon name="sparkles" size="md" :stroke-width="2" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('usage.mediaDetails') }}</h3>
                  <span v-if="imageURLs.length > 0" class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                    {{ imageURLs.length }}{{ t('usage.imageUnit') }}
                  </span>
                </div>
                <div class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                  <span class="max-w-[220px] truncate font-medium text-gray-700 dark:text-gray-200" :title="row.model || ''">{{ row.model || '-' }}</span>
                  <span v-if="row.created_at" class="text-gray-300 dark:text-dark-600">/</span>
                  <span v-if="row.created_at">{{ formatDateTime(row.created_at) }}</span>
                  <span v-if="row.request_id" class="text-gray-300 dark:text-dark-600">/</span>
                  <span v-if="row.request_id" class="max-w-[180px] truncate font-mono" :title="row.request_id">{{ formatRequestID(row.request_id) }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-center justify-end gap-2">
              <a
                v-if="selectedURL"
                :href="selectedURL"
                target="_blank"
                rel="noreferrer"
                class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-xs font-medium text-gray-700 transition hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-600 dark:hover:bg-dark-700"
              >
                <Icon name="externalLink" size="xs" />
                {{ t('usage.openImage') }}
              </a>
              <button
                v-if="selectedURL"
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-xs font-medium text-gray-700 transition hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-600 dark:hover:bg-dark-700"
                @click="copySelectedURL"
              >
                <Icon :name="copiedURL === selectedURL ? 'check' : 'copy'" size="xs" :stroke-width="2" />
                {{ copiedURL === selectedURL ? t('usage.copied') : t('usage.copyImageUrl') }}
              </button>
              <button
                type="button"
                class="flex h-9 w-9 items-center justify-center rounded-full text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-white"
                @click="close"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
            </div>
          </header>

          <div class="grid min-h-0 flex-1 overflow-hidden xl:grid-cols-[minmax(0,1fr)_380px]">
            <section class="flex min-h-0 flex-col bg-gray-50 dark:bg-dark-950">
              <div v-if="selectedURL" class="relative flex min-h-0 flex-1 items-center justify-center overflow-hidden p-4 sm:p-6">
                <div class="absolute inset-0 bg-[size:24px_24px] opacity-[0.45] [background-image:linear-gradient(to_right,rgba(148,163,184,0.16)_1px,transparent_1px),linear-gradient(to_bottom,rgba(148,163,184,0.16)_1px,transparent_1px)] dark:opacity-[0.18]" />
                <div class="relative flex h-full max-h-full w-full items-center justify-center rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
                  <img
                    v-if="!failedURLs.includes(selectedURL)"
                    :src="selectedURL"
                    :alt="`${t('usage.generatedImage')} ${activeIndex + 1}`"
                    class="max-h-[62dvh] w-full object-contain p-3 sm:max-h-[68vh] sm:p-5"
                    @error="markImageFailed(selectedURL)"
                  />
                  <div v-else class="flex min-h-[360px] flex-col items-center justify-center gap-3 px-4 text-center text-sm text-gray-500 dark:text-gray-400">
                    <Icon name="exclamationCircle" size="lg" class="text-gray-400" />
                    <span>{{ t('usage.imagePreviewFailed') }}</span>
                  </div>

                  <div class="absolute left-3 top-3 rounded-full border border-white/80 bg-white/90 px-3 py-1 text-xs font-medium text-gray-700 shadow-sm backdrop-blur dark:border-white/10 dark:bg-dark-800/90 dark:text-gray-200">
                    {{ activeIndex + 1 }} / {{ imageURLs.length }}
                  </div>
                  <div class="absolute bottom-3 left-3 right-3 flex min-w-0 items-center gap-2 rounded-lg border border-white/80 bg-white/90 px-3 py-2 text-xs text-gray-600 shadow-sm backdrop-blur dark:border-white/10 dark:bg-dark-800/90 dark:text-gray-300">
                    <Icon name="link" size="xs" class="shrink-0 text-gray-400" />
                    <span class="min-w-0 flex-1 truncate font-mono" :title="selectedURL">{{ formatMediaURL(selectedURL) }}</span>
                    <span class="shrink-0 rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                      {{ isDataURL(selectedURL) ? t('usage.dataUrl') : t('usage.publicUrl') }}
                    </span>
                  </div>
                </div>
              </div>

              <div v-else class="flex min-h-[460px] flex-1 flex-col items-center justify-center px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                <div class="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-white text-gray-400 shadow-sm ring-1 ring-gray-200 dark:bg-dark-900 dark:ring-dark-700">
                  <Icon name="link" size="lg" />
                </div>
                {{ t('usage.noImageUrlRecorded') }}
              </div>

              <div v-if="imageURLs.length > 1" class="border-t border-gray-200 bg-white/90 p-3 backdrop-blur dark:border-dark-700 dark:bg-dark-900/90">
                <div class="flex gap-3 overflow-x-auto pb-1">
                  <button
                    v-for="(url, index) in imageURLs"
                    :key="`${index}-${url.slice(0, 48)}`"
                    type="button"
                    class="relative h-20 w-20 shrink-0 overflow-hidden rounded-lg border bg-gray-100 transition dark:bg-dark-800"
                    :class="activeIndex === index ? 'border-primary-500 ring-2 ring-primary-500/20' : 'border-gray-200 hover:border-gray-300 dark:border-dark-700 dark:hover:border-dark-600'"
                    @click="activeIndex = index"
                  >
                    <img
                      v-if="!failedURLs.includes(url)"
                      :src="url"
                      :alt="`${t('usage.generatedImage')} ${index + 1}`"
                      class="h-full w-full object-cover"
                      loading="lazy"
                      @error="markImageFailed(url)"
                    />
                    <div v-else class="flex h-full items-center justify-center text-gray-400">
                      <Icon name="exclamationCircle" size="sm" />
                    </div>
                    <div class="absolute left-1.5 top-1.5 rounded bg-gray-950/65 px-1.5 py-0.5 text-[10px] font-medium text-white">
                      {{ index + 1 }}
                    </div>
                  </button>
                </div>
              </div>
            </section>

            <aside class="min-h-0 overflow-y-auto border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900 xl:border-l xl:border-t-0 sm:p-5">
              <div class="grid grid-cols-2 gap-2">
                <div
                  v-for="item in summaryItems"
                  :key="item.label"
                  class="rounded-lg border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800/70"
                >
                  <div class="text-[11px] font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ item.label }}</div>
                  <div class="mt-1 truncate text-sm font-semibold text-gray-900 dark:text-gray-100" :title="item.value">{{ item.value }}</div>
                </div>
              </div>

              <section v-if="selectedURL" class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800/70">
                <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="link" size="xs" />
                  {{ t('usage.imageUrlCount') }}
                </div>
                <div class="flex items-center gap-2 rounded-lg bg-white px-3 py-2 ring-1 ring-gray-200 dark:bg-dark-900 dark:ring-dark-700">
                  <span class="min-w-0 flex-1 truncate font-mono text-xs text-gray-600 dark:text-gray-300" :title="selectedURL">{{ formatMediaURL(selectedURL) }}</span>
                  <a
                    :href="selectedURL"
                    download
                    class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-white"
                    :title="t('usage.downloadImage')"
                  >
                    <Icon name="download" size="xs" />
                  </a>
                </div>
              </section>

              <section class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
                <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="chatBubble" size="xs" />
                  {{ t('usage.prompt') }}
                </div>
                <p class="max-h-56 overflow-auto whitespace-pre-wrap break-words text-sm leading-6 text-gray-950 dark:text-gray-100">
                  {{ row.image_prompt || t('usage.noPromptRecorded') }}
                </p>
              </section>

              <section class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
                <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="sparkles" size="xs" />
                  {{ t('usage.revisedPrompts') }}
                </div>
                <div v-if="revisedPrompts.length > 0" class="space-y-3">
                  <p
                    v-for="(prompt, index) in revisedPrompts"
                    :key="`${index}-${prompt.slice(0, 32)}`"
                    class="max-h-40 overflow-auto whitespace-pre-wrap break-words rounded-lg bg-white p-3 text-sm leading-6 text-gray-900 ring-1 ring-gray-200 dark:bg-dark-900 dark:text-gray-100 dark:ring-dark-700"
                  >
                    {{ prompt }}
                  </p>
                </div>
                <p v-else class="text-sm text-gray-500 dark:text-gray-400">{{ t('usage.noRevisedPromptRecorded') }}</p>
              </section>
            </aside>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeSource,
} from '@/utils/imageUsage'
import type { UsageLog } from '@/types'

type ImageDetailRow = Pick<
  UsageLog,
  | 'request_id'
  | 'model'
  | 'created_at'
  | 'image_count'
  | 'image_size'
  | 'image_input_size'
  | 'image_output_size'
  | 'image_size_source'
  | 'image_size_breakdown'
  | 'image_prompt'
  | 'image_urls'
  | 'image_revised_prompts'
>

const props = defineProps<{
  row: ImageDetailRow | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const activeIndex = ref(0)
const copiedURL = ref<string | null>(null)
const failedURLs = ref<string[]>([])

const sanitizeStringArray = (value: string[] | null | undefined): string[] => {
  if (!Array.isArray(value)) return []
  return value.map((item) => item?.trim()).filter((item): item is string => Boolean(item))
}

const imageURLs = computed(() => sanitizeStringArray(props.row?.image_urls))
const revisedPrompts = computed(() => sanitizeStringArray(props.row?.image_revised_prompts))
const selectedURL = computed(() => imageURLs.value[activeIndex.value] || imageURLs.value[0] || '')

const summaryItems = computed(() => {
  if (!props.row) return []
  return [
    { label: t('usage.imageCount'), value: `${props.row.image_count ?? imageURLs.value.length}${t('usage.imageUnit')}` },
    { label: t('usage.imageBillingSize'), value: formatImageBillingSize(props.row, t) },
    { label: t('usage.imageInputSize'), value: formatImageInputSize(props.row, t) },
    { label: t('usage.imageOutputSize'), value: formatImageOutputSize(props.row, t) },
    { label: t('usage.imageSizeSource'), value: formatImageSizeSource(props.row, t) },
    { label: t('usage.imageUrlCount'), value: String(imageURLs.value.length) },
  ]
})

watch(
  () => props.row?.request_id,
  () => {
    activeIndex.value = 0
    copiedURL.value = null
    failedURLs.value = []
  }
)

watch(imageURLs, (urls) => {
  if (activeIndex.value >= urls.length) activeIndex.value = 0
})

let copyTimer: number | null = null

const close = () => {
  emit('close')
}

const isDataURL = (url: string): boolean => url.startsWith('data:')

const formatRequestID = (requestID: string): string => {
  if (requestID.length <= 18) return requestID
  return `${requestID.slice(0, 8)}...${requestID.slice(-6)}`
}

const formatMediaURL = (url: string): string => {
  if (isDataURL(url)) return t('usage.dataUrl')
  try {
    const parsed = new URL(url)
    return `${parsed.host}${parsed.pathname}`
  } catch {
    return url
  }
}

const copyURL = async (url: string) => {
  try {
    await navigator.clipboard?.writeText(url)
    copiedURL.value = url
    if (copyTimer) window.clearTimeout(copyTimer)
    copyTimer = window.setTimeout(() => {
      copiedURL.value = null
    }, 1600)
  } catch {
    copiedURL.value = null
  }
}

const copySelectedURL = async () => {
  if (!selectedURL.value) return
  await copyURL(selectedURL.value)
}

const markImageFailed = (url: string) => {
  if (failedURLs.value.includes(url)) return
  failedURLs.value = [...failedURLs.value, url]
}
</script>
