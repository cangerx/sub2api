<template>
  <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('videoStudio.tasks') }}</h2>
      <div class="flex items-center gap-2">
        <select class="input h-8 w-20 text-xs" :value="refreshInterval" :disabled="!autoRefresh" @change="$emit('set-refresh-interval', Number(($event.target as HTMLSelectElement).value))">
          <option v-for="option in refreshIntervals" :key="option" :value="option">{{ option }}s</option>
        </select>
        <button class="btn btn-secondary btn-sm" type="button" :disabled="!hasApiKey || loadingTasks" @click="$emit('refresh')">
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div class="mb-3 flex items-center justify-between rounded-md bg-gray-50 px-3 py-2 text-xs dark:bg-dark-700">
      <span class="text-gray-500 dark:text-dark-300">
        {{ activeCount }} {{ t('videoStudio.activeTasks') }}
        <span v-if="autoRefresh && countdown > 0"> · {{ countdown }}s</span>
      </span>
      <label class="inline-flex items-center gap-2 text-gray-600 dark:text-dark-200">
        <input type="checkbox" class="rounded border-gray-300" :checked="autoRefresh" @change="$emit('toggle-auto-refresh', ($event.target as HTMLInputElement).checked)" />
        <span>Auto</span>
      </label>
    </div>

    <div v-if="error" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300">
      {{ error }}
    </div>

    <div v-if="tasks.length === 0" class="rounded-lg border border-dashed border-gray-200 p-8 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-dark-300">
      {{ t('videoStudio.emptyTasks') }}
    </div>

    <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
      <div v-for="task in tasks" :key="task.id" class="grid gap-3 py-4 lg:grid-cols-[minmax(0,1fr)_170px]">
        <button class="min-w-0 text-left" type="button" @click="$emit('select', task)">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ task.id }}</span>
            <span class="rounded px-2 py-0.5 text-xs font-medium" :class="statusClass(task.status)">
              {{ task.status }}
            </span>
            <span class="text-xs text-gray-500 dark:text-dark-300">{{ task.model }}</span>
          </div>
          <div class="mt-2 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
            <div class="h-full bg-blue-500 transition-all" :style="{ width: `${Math.max(0, Math.min(100, task.progress || 0))}%` }" />
          </div>
          <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-dark-300">
            <span>{{ formatDate(task.created_at) }}</span>
            <span v-if="task.seconds">{{ task.seconds }}s</span>
            <span v-if="task.size">{{ task.size }}</span>
            <span v-if="task.error?.message" class="text-red-600 dark:text-red-300">{{ task.error.message }}</span>
          </div>
        </button>

        <div class="flex items-center justify-end gap-2">
          <button class="btn btn-secondary btn-sm" type="button" @click="$emit('refresh-task', task.id)">
            {{ t('common.refresh') }}
          </button>
          <button v-if="task.status === 'queued' || task.status === 'in_progress'" class="btn btn-danger btn-sm" type="button" @click="$emit('cancel', task.id)">
            {{ t('common.cancel') }}
          </button>
          <a v-if="task.status === 'completed'" class="btn btn-primary btn-sm" :href="downloadHref(task.id)" target="_blank" rel="noopener">
            {{ t('videoStudio.download') }}
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { videoContentURL, type VideoObject } from '@/api/video'

defineProps<{
  tasks: VideoObject[]
  hasApiKey: boolean
  loadingTasks: boolean
  error: string
  activeCount: number
  autoRefresh: boolean
  refreshInterval: number
  refreshIntervals: readonly number[]
  countdown: number
}>()

defineEmits<{
  (event: 'refresh'): void
  (event: 'refresh-task', id: string): void
  (event: 'cancel', id: string): void
  (event: 'select', task: VideoObject): void
  (event: 'toggle-auto-refresh', value: boolean): void
  (event: 'set-refresh-interval', value: number): void
}>()

const { t } = useI18n()

function downloadHref(id: string): string {
  return videoContentURL(id)
}

function formatDate(timestamp?: number | null): string {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString()
}

function statusClass(status: string): string {
  if (status === 'completed') return 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300'
  if (status === 'failed' || status === 'expired') return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
  if (status === 'cancelled') return 'bg-gray-100 text-gray-700 dark:bg-dark-600 dark:text-dark-200'
  return 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
}
</script>
