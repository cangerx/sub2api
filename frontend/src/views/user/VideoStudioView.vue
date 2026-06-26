<template>
  <div class="space-y-6">
    <section class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div class="space-y-4">
          <div>
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('videoStudio.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('videoStudio.description') }}</p>
          </div>

          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
            <input
              v-model="apiKeyInput"
              type="password"
              class="input"
              :placeholder="t('videoStudio.apiKeyPlaceholder')"
              autocomplete="off"
              @keyup.enter="saveApiKey"
            />
            <button class="btn btn-primary whitespace-nowrap" type="button" @click="saveApiKey">
              {{ t('common.save') }}
            </button>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
            <div class="text-xs text-gray-500 dark:text-dark-300">{{ t('videoStudio.models') }}</div>
            <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ videoStore.models.length }}</div>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700">
            <div class="text-xs text-gray-500 dark:text-dark-300">{{ t('videoStudio.activeTasks') }}</div>
            <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ videoStore.activeTasks.length }}</div>
          </div>
        </div>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[420px_minmax(0,1fr)]">
      <VideoCreateForm
        :models="videoStore.models"
        :has-api-key="videoStore.hasApiKey"
        :creating="videoStore.creating"
        :refresh-disabled="!videoStore.hasApiKey || videoStore.loadingModels"
        @refresh="loadAll"
        @submit="submit"
      />

      <VideoTaskList
        :tasks="videoStore.tasks"
        :has-api-key="videoStore.hasApiKey"
        :loading-tasks="videoStore.loadingTasks"
        :error="videoStore.error"
        :active-count="videoStore.activeTasks.length"
        :auto-refresh="autoRefresh.enabled.value"
        :refresh-interval="autoRefresh.intervalSeconds.value"
        :refresh-intervals="autoRefresh.intervals"
        :countdown="autoRefresh.countdown.value"
        @refresh="videoStore.loadTasks"
        @refresh-task="videoStore.refreshTask"
        @cancel="videoStore.cancelTask"
        @select="selectedTask = $event"
        @toggle-auto-refresh="autoRefresh.setEnabled"
        @set-refresh-interval="autoRefresh.setInterval"
      />
    </section>

    <VideoTaskDetailDialog
      :task="selectedTask"
      @close="selectedTask = null"
      @refresh="videoStore.refreshTask"
      @cancel="videoStore.cancelTask"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useVideoStore } from '@/stores'
import type { VideoCreateRequest, VideoObject } from '@/api/video'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import VideoCreateForm from '@/components/video/VideoCreateForm.vue'
import VideoTaskList from '@/components/video/VideoTaskList.vue'
import VideoTaskDetailDialog from '@/components/video/VideoTaskDetailDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
const videoStore = useVideoStore()

const apiKeyInput = ref(videoStore.apiKey)
const selectedTask = ref<VideoObject | null>(null)

const autoRefresh = useAutoRefresh({
  storageKey: 'ccapi_video_auto_refresh',
  intervals: [5, 10, 15, 30],
  defaultInterval: 5,
  onRefresh: () => videoStore.pollActiveTasks(),
  shouldPause: () => document.hidden || !videoStore.hasApiKey || videoStore.activeTasks.length === 0,
})

watch(() => videoStore.activeTasks.length, (count) => {
  if (count > 0 && autoRefresh.enabled.value) {
    autoRefresh.start()
  }
})

function saveApiKey(): void {
  videoStore.setApiKey(apiKeyInput.value)
  void loadAll()
}

async function loadAll(): Promise<void> {
  if (!videoStore.hasApiKey) return
  try {
    await Promise.all([videoStore.loadModels(), videoStore.loadTasks()])
  } catch (err) {
    appStore.showError(err instanceof Error ? err.message : t('videoStudio.loadFailed'))
  }
}

async function submit(payload: VideoCreateRequest): Promise<void> {
  try {
    await videoStore.submit(payload)
    appStore.showSuccess(t('videoStudio.submitted'))
    if (autoRefresh.enabled.value) autoRefresh.resetCountdown()
  } catch (err) {
    appStore.showError(err instanceof Error ? err.message : t('videoStudio.submitFailed'))
  }
}

onMounted(() => {
  void loadAll()
  if (autoRefresh.enabled.value) autoRefresh.start()
})
</script>
