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
        <div class="flex h-[96dvh] w-full max-w-6xl flex-col overflow-hidden rounded-t-2xl border border-white/70 bg-white shadow-2xl ring-1 ring-black/5 dark:border-white/10 dark:bg-dark-900 dark:ring-white/10 sm:h-auto sm:max-h-[90vh] sm:rounded-2xl">
          <header class="flex flex-col gap-4 border-b border-gray-200/80 bg-white/95 px-5 py-4 backdrop-blur-xl dark:border-dark-700/80 dark:bg-dark-900/95 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex min-w-0 items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200">
                <Icon name="video" size="md" :stroke-width="2" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('usage.videoDetails') }}</h3>
                  <span class="rounded-full bg-rose-50 px-2 py-0.5 text-xs font-medium text-rose-600 dark:bg-rose-500/10 dark:text-rose-300">
                    {{ billingModeLabel }}
                  </span>
                </div>
                <div class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                  <span class="max-w-[220px] truncate font-medium text-gray-700 dark:text-gray-200" :title="row.model || ''">{{ row.model || '-' }}</span>
                  <span v-if="row.created_at" class="text-gray-300 dark:text-dark-600">/</span>
                  <span v-if="row.created_at">{{ formatDateTime(row.created_at) }}</span>
                  <span v-if="row.request_id" class="text-gray-300 dark:text-dark-600">/</span>
                  <span v-if="row.request_id" class="max-w-[180px] truncate font-mono" :title="row.request_id">{{ formatShortID(row.request_id) }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-center justify-end gap-2">
              <button
                v-if="row.video_task_id"
                type="button"
                class="inline-flex h-9 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-xs font-medium text-gray-700 transition hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-600 dark:hover:bg-dark-700"
                @click="copyTaskID"
              >
                <Icon :name="copiedTask ? 'check' : 'copy'" size="xs" :stroke-width="2" />
                {{ copiedTask ? t('usage.copied') : t('usage.copyVideoTaskId') }}
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

          <div class="grid min-h-0 flex-1 overflow-hidden lg:grid-cols-[minmax(0,1fr)_360px]">
            <section class="flex min-h-0 flex-col bg-gray-50 dark:bg-dark-950">
              <div class="relative flex min-h-[420px] flex-1 items-center justify-center overflow-hidden p-4 sm:p-6">
                <div class="absolute inset-0 bg-[size:24px_24px] opacity-[0.45] [background-image:linear-gradient(to_right,rgba(148,163,184,0.16)_1px,transparent_1px),linear-gradient(to_bottom,rgba(148,163,184,0.16)_1px,transparent_1px)] dark:opacity-[0.18]" />
                <div class="relative flex h-full max-h-full w-full items-center justify-center overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
                  <video
                    v-if="videoURL"
                    class="max-h-[62dvh] w-full bg-black object-contain sm:max-h-[68vh]"
                    :src="videoURL"
                    controls
                    playsinline
                    preload="metadata"
                  />
                  <div v-else class="flex max-w-md flex-col items-center justify-center px-6 py-16 text-center">
                    <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-gray-100 text-gray-500 ring-1 ring-gray-200 dark:bg-dark-800 dark:text-gray-300 dark:ring-dark-700">
                      <Icon name="video" size="lg" />
                    </div>
                    <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('usage.noVideoUrlRecorded') }}</div>
                    <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
                      {{ t('usage.noVideoUrlRecordedHint') }}
                    </p>
                  </div>

                  <div class="absolute left-3 top-3 rounded-full border border-white/80 bg-white/90 px-3 py-1 text-xs font-medium text-gray-700 shadow-sm backdrop-blur dark:border-white/10 dark:bg-dark-800/90 dark:text-gray-200">
                    {{ t('usage.videoGeneration') }}
                  </div>
                  <div v-if="row.video_task_id" class="absolute bottom-3 left-3 right-3 flex min-w-0 items-center gap-2 rounded-lg border border-white/80 bg-white/90 px-3 py-2 text-xs text-gray-600 shadow-sm backdrop-blur dark:border-white/10 dark:bg-dark-800/90 dark:text-gray-300">
                    <Icon name="terminal" size="xs" class="shrink-0 text-gray-400" />
                    <span class="min-w-0 flex-1 truncate font-mono" :title="row.video_task_id">{{ row.video_task_id }}</span>
                  </div>
                </div>
              </div>
            </section>

            <aside class="min-h-0 overflow-y-auto border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900 lg:border-l lg:border-t-0 sm:p-5">
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

              <section class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
                <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="terminal" size="xs" />
                  {{ t('usage.videoTaskId') }}
                </div>
                <div class="flex items-center gap-2 rounded-lg bg-white px-3 py-2 ring-1 ring-gray-200 dark:bg-dark-900 dark:ring-dark-700">
                  <span class="min-w-0 flex-1 truncate font-mono text-xs text-gray-600 dark:text-gray-300" :title="row.video_task_id || ''">
                    {{ row.video_task_id || '-' }}
                  </span>
                  <button
                    v-if="row.video_task_id"
                    type="button"
                    class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-white"
                    :title="t('usage.copyVideoTaskId')"
                    @click="copyTaskID"
                  >
                    <Icon :name="copiedTask ? 'check' : 'copy'" size="xs" />
                  </button>
                </div>
              </section>

              <section class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
                <div class="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="link" size="xs" />
                  {{ t('usage.videoPlaybackUrl') }}
                </div>
                <p class="break-words text-sm leading-6 text-gray-500 dark:text-gray-400">
                  {{ videoURL || t('usage.noVideoUrlRecorded') }}
                </p>
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
import { getBillingModeLabel } from '@/utils/billingMode'
import type { UsageLog } from '@/types'

type VideoDetailRow = Pick<
  UsageLog,
  | 'request_id'
  | 'model'
  | 'created_at'
  | 'billing_mode'
  | 'video_task_id'
  | 'video_seconds'
  | 'video_size'
  | 'video_billing_units'
>

const props = defineProps<{
  row: VideoDetailRow | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const copiedTask = ref(false)
let copyTimer: number | null = null

const videoURL = computed(() => '')

const billingModeLabel = computed(() => getBillingModeLabel(props.row?.billing_mode, t))

const summaryItems = computed(() => {
  if (!props.row) return []
  return [
    { label: t('usage.videoSeconds'), value: props.row.video_seconds != null ? `${props.row.video_seconds}s` : '-' },
    { label: t('usage.videoSize'), value: props.row.video_size || '-' },
    { label: t('usage.videoBillingUnitCount'), value: props.row.video_billing_units != null ? String(props.row.video_billing_units) : '-' },
    { label: t('admin.usage.billingMode'), value: billingModeLabel.value },
    { label: t('usage.model'), value: props.row.model || '-' },
    { label: t('usage.time'), value: props.row.created_at ? formatDateTime(props.row.created_at) : '-' },
  ]
})

watch(
  () => props.row?.request_id,
  () => {
    copiedTask.value = false
  }
)

const close = () => {
  emit('close')
}

const formatShortID = (value: string): string => {
  if (value.length <= 18) return value
  return `${value.slice(0, 8)}...${value.slice(-6)}`
}

const copyTaskID = async () => {
  const taskID = props.row?.video_task_id
  if (!taskID) return
  try {
    await navigator.clipboard?.writeText(taskID)
    copiedTask.value = true
    if (copyTimer) window.clearTimeout(copyTimer)
    copyTimer = window.setTimeout(() => {
      copiedTask.value = false
    }, 1600)
  } catch {
    copiedTask.value = false
  }
}
</script>
