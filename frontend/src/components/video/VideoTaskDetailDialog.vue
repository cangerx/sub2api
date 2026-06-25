<template>
  <BaseDialog :show="!!task" :title="task?.id || ''" width="extra-wide" @close="$emit('close')">
    <div v-if="task" class="space-y-4">
      <video
        v-if="task.status === 'completed'"
        class="aspect-video w-full rounded-lg bg-black"
        controls
        :src="downloadHref(task.id)"
      />
      <div class="grid gap-3 text-sm md:grid-cols-2">
        <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
          <div class="text-xs text-gray-500 dark:text-dark-300">Status</div>
          <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ task.status }} · {{ task.progress }}%</div>
        </div>
        <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
          <div class="text-xs text-gray-500 dark:text-dark-300">Model</div>
          <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ task.model }}</div>
        </div>
        <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
          <div class="text-xs text-gray-500 dark:text-dark-300">Created</div>
          <div class="mt-1 text-gray-900 dark:text-white">{{ formatDate(task.created_at) }}</div>
        </div>
        <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
          <div class="text-xs text-gray-500 dark:text-dark-300">Expires</div>
          <div class="mt-1 text-gray-900 dark:text-white">{{ formatDate(task.expires_at) }}</div>
        </div>
      </div>
      <div v-if="task.error" class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300">
        {{ task.error.message || task.error.code }}
      </div>
      <div class="flex justify-end gap-2">
        <button class="btn btn-secondary" type="button" @click="$emit('refresh', task.id)">{{ t('common.refresh') }}</button>
        <button v-if="task.status === 'queued' || task.status === 'in_progress'" class="btn btn-danger" type="button" @click="$emit('cancel', task.id)">
          {{ t('common.cancel') }}
        </button>
        <a v-if="task.status === 'completed'" class="btn btn-primary" :href="downloadHref(task.id)" target="_blank" rel="noopener">
          {{ t('videoStudio.download') }}
        </a>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { videoContentURL, type VideoObject } from '@/api/video'

defineProps<{ task: VideoObject | null }>()
defineEmits<{
  (event: 'close'): void
  (event: 'refresh', id: string): void
  (event: 'cancel', id: string): void
}>()

const { t } = useI18n()

function downloadHref(id: string): string {
  return videoContentURL(id)
}

function formatDate(timestamp?: number | null): string {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString()
}
</script>
