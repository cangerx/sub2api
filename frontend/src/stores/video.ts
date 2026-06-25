import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  cancelVideo,
  createVideo,
  getVideo,
  listVideoModels,
  listVideos,
  type VideoCreateRequest,
  type VideoModelObject,
  type VideoObject
} from '@/api/video'

const STORAGE_KEY = 'sub2api_video_gateway_key'
const ACTIVE_STATUSES = new Set(['queued', 'in_progress'])

export const useVideoStore = defineStore('video', () => {
  const apiKey = ref<string>(localStorage.getItem(STORAGE_KEY) || '')
  const models = ref<VideoModelObject[]>([])
  const tasks = ref<VideoObject[]>([])
  const loadingModels = ref(false)
  const loadingTasks = ref(false)
  const creating = ref(false)
  const error = ref<string>('')

  const activeTasks = computed(() => tasks.value.filter((task) => ACTIVE_STATUSES.has(task.status)))
  const completedTasks = computed(() => tasks.value.filter((task) => task.status === 'completed'))
  const hasApiKey = computed(() => apiKey.value.trim().length > 0)

  function setApiKey(value: string): void {
    apiKey.value = value.trim()
    if (apiKey.value) {
      localStorage.setItem(STORAGE_KEY, apiKey.value)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  function upsertTask(task: VideoObject): void {
    const index = tasks.value.findIndex((item) => item.id === task.id)
    if (index >= 0) {
      tasks.value[index] = task
      return
    }
    tasks.value.unshift(task)
  }

  async function loadModels(): Promise<void> {
    if (!hasApiKey.value) return
    loadingModels.value = true
    error.value = ''
    try {
      const response = await listVideoModels(apiKey.value)
      models.value = response.data || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load video models'
      throw err
    } finally {
      loadingModels.value = false
    }
  }

  async function loadTasks(): Promise<void> {
    if (!hasApiKey.value) return
    loadingTasks.value = true
    error.value = ''
    try {
      const response = await listVideos(apiKey.value, 50)
      tasks.value = response.data || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load video tasks'
      throw err
    } finally {
      loadingTasks.value = false
    }
  }

  async function submit(payload: VideoCreateRequest): Promise<VideoObject> {
    creating.value = true
    error.value = ''
    try {
      const idempotencyKey = `video-${Date.now()}-${Math.random().toString(36).slice(2)}`
      const task = await createVideo(apiKey.value, payload, idempotencyKey)
      upsertTask(task)
      return task
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create video task'
      throw err
    } finally {
      creating.value = false
    }
  }

  async function refreshTask(id: string): Promise<VideoObject> {
    const task = await getVideo(apiKey.value, id)
    upsertTask(task)
    return task
  }

  async function cancelTask(id: string): Promise<VideoObject> {
    const task = await cancelVideo(apiKey.value, id)
    upsertTask(task)
    return task
  }

  async function pollActiveTasks(): Promise<void> {
    if (!hasApiKey.value || activeTasks.value.length === 0) return
    await Promise.allSettled(activeTasks.value.map((task) => refreshTask(task.id)))
  }

  return {
    apiKey,
    models,
    tasks,
    loadingModels,
    loadingTasks,
    creating,
    error,
    activeTasks,
    completedTasks,
    hasApiKey,
    setApiKey,
    loadModels,
    loadTasks,
    submit,
    refreshTask,
    cancelTask,
    pollActiveTasks
  }
})
