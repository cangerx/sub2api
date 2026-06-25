<template>
  <form class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800" @submit.prevent="emitSubmit">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('videoStudio.createTask') }}</h2>
      <button class="btn btn-secondary btn-sm" type="button" :disabled="refreshDisabled" @click="$emit('refresh')">
        {{ t('common.refresh') }}
      </button>
    </div>

    <div class="space-y-4">
      <label class="block">
        <span class="input-label">{{ t('videoStudio.model') }}</span>
        <select v-model="form.model" class="input">
          <option value="">{{ t('videoStudio.selectModel') }}</option>
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.display_name || model.id }}
          </option>
        </select>
      </label>

      <div v-if="modeOptions.length > 1">
        <span class="input-label">{{ t('videoStudio.generationMode', 'Mode') }}</span>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <button
            v-for="option in modeOptions"
            :key="option.value"
            type="button"
            class="rounded-md border px-3 py-2 text-sm font-medium transition"
            :class="form.mode === option.value
              ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-500/10 dark:text-primary-300'
              : 'border-gray-200 text-gray-600 hover:bg-gray-50 dark:border-dark-600 dark:text-dark-300 dark:hover:bg-dark-700'"
            @click="form.mode = option.value"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label class="block">
          <span class="input-label">{{ t('videoStudio.seconds') }}</span>
          <select v-if="secondOptions.length" v-model="form.seconds" class="input">
            <option v-for="option in secondOptions" :key="option" :value="String(option)">{{ option }}s</option>
          </select>
          <input
            v-else
            v-model="form.seconds"
            class="input"
            type="number"
            inputmode="numeric"
            min="1"
            :max="maxSeconds || undefined"
          />
        </label>
        <label class="block">
          <span class="input-label">{{ t('videoStudio.size') }}</span>
          <select v-if="sizeOptions.length" v-model="form.size" class="input">
            <option v-for="option in sizeOptions" :key="option" :value="option">{{ option }}</option>
          </select>
          <input v-else v-model="form.size" class="input" placeholder="1280x720" />
        </label>
      </div>

      <label class="block">
        <span class="input-label">{{ t('videoStudio.prompt') }}</span>
        <textarea
          v-model="form.prompt"
          rows="7"
          class="input resize-y"
          :maxlength="maxPromptLength || undefined"
          :placeholder="t('videoStudio.promptPlaceholder')"
        />
        <span v-if="maxPromptLength" class="mt-1 block text-xs text-gray-500 dark:text-dark-300">
          {{ form.prompt.length }}/{{ maxPromptLength }}
        </span>
      </label>

      <label v-if="requiresSingleReference" class="block">
        <span class="input-label">{{ singleReferenceLabel }}</span>
        <input v-model="form.input_reference" class="input" placeholder="https://example.com/frame.jpg" />
      </label>

      <div v-if="requiresFirstLastFrame" class="grid gap-3 sm:grid-cols-2">
        <label class="block">
          <span class="input-label">{{ t('videoStudio.firstFrame', 'First frame URL') }}</span>
          <input v-model="form.first_frame_url" class="input" placeholder="https://example.com/first.jpg" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('videoStudio.lastFrame', 'Last frame URL') }}</span>
          <input v-model="form.last_frame_url" class="input" placeholder="https://example.com/last.jpg" />
        </label>
      </div>

      <div v-if="extraFields.length" class="rounded-md border border-gray-200 p-3 dark:border-dark-700">
        <div class="mb-3 text-sm font-medium text-gray-900 dark:text-white">{{ t('videoStudio.advanced', 'Advanced') }}</div>
        <div class="grid gap-3 sm:grid-cols-2">
          <label v-for="field in extraFields" :key="field" class="block">
            <span class="input-label">{{ field }}</span>
            <input v-model="form.extraBody[field]" class="input" :placeholder="t('videoStudio.extraValuePlaceholder', 'number, boolean, text, or JSON')" />
          </label>
        </div>
      </div>

      <div class="rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-700">
        <div class="flex items-center justify-between">
          <span class="text-gray-500 dark:text-dark-300">{{ t('videoStudio.estimate') }}</span>
          <span class="font-semibold text-gray-900 dark:text-white">{{ estimatedCost }}</span>
        </div>
        <div v-if="selectedModel?.billing" class="mt-1 text-xs text-gray-500 dark:text-dark-300">
          {{ selectedModel.billing.mode }} · {{ selectedModel.billing.unit_price }} {{ selectedModel.billing.currency }}
        </div>
      </div>

      <button class="btn btn-primary w-full" type="submit" :disabled="!canSubmit || creating">
        {{ creating ? t('common.loading') : t('videoStudio.submit') }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { VideoCreateRequest, VideoModelObject } from '@/api/video'

type VideoMode = 'text_to_video' | 'image_to_video' | 'first_last_frame' | 'reference_image' | 'reference_video'

const props = defineProps<{
  models: VideoModelObject[]
  hasApiKey: boolean
  creating: boolean
  refreshDisabled: boolean
}>()

const emit = defineEmits<{
  (event: 'submit', payload: VideoCreateRequest): void
  (event: 'refresh'): void
}>()

const { t } = useI18n()
const form = reactive({
  model: '',
  mode: 'text_to_video' as VideoMode,
  prompt: '',
  seconds: '5',
  size: '1280x720',
  input_reference: '',
  first_frame_url: '',
  last_frame_url: '',
  extraBody: {} as Record<string, string>,
})

const modeLabels: Record<VideoMode, string> = {
  text_to_video: t('videoStudio.modeText', 'Text'),
  image_to_video: t('videoStudio.modeImage', 'Image'),
  first_last_frame: t('videoStudio.modeFirstLast', 'First/Last'),
  reference_image: t('videoStudio.modeReferenceImage', 'Reference image'),
  reference_video: t('videoStudio.modeReferenceVideo', 'Reference video'),
}

const selectedModel = computed(() => props.models.find((model) => model.id === form.model))
const modelSupports = computed(() => {
  const supports = selectedModel.value?.supports?.filter(Boolean) || []
  return supports.length ? supports : ['text_to_video']
})
const modeOptions = computed(() => {
  const values = modelSupports.value.filter((value): value is VideoMode => isVideoMode(value))
  const normalized = values.length ? values : ['text_to_video' as VideoMode]
  return normalized.map((value) => ({ value, label: modeLabels[value] }))
})
const secondOptions = computed(() => selectedModel.value?.seconds?.length ? selectedModel.value.seconds : [])
const sizeOptions = computed(() => selectedModel.value?.sizes?.length ? selectedModel.value.sizes : [])
const maxSeconds = computed(() => numberFromLimit('max_seconds'))
const maxPromptLength = computed(() => numberFromLimit('max_prompt_length') || numberFromLimit('prompt_max_length'))
const requiresSingleReference = computed(() => ['image_to_video', 'reference_image', 'reference_video'].includes(form.mode))
const requiresFirstLastFrame = computed(() => form.mode === 'first_last_frame')
const singleReferenceLabel = computed(() => {
  if (form.mode === 'reference_video') return t('videoStudio.referenceVideo', 'Reference video URL')
  if (form.mode === 'reference_image') return t('videoStudio.referenceImage', 'Reference image URL')
  return t('videoStudio.reference')
})
const extraFields = computed(() => {
  const blocked = new Set([
    'input_reference',
    'image_url',
    'reference_url',
    'reference_image_url',
    'reference_video_url',
    'first_frame_url',
    'last_frame_url',
    'prompt',
    'model',
    'seconds',
    'size',
    'duration',
  ])
  return (selectedModel.value?.extra_body_allow || [])
    .map((field) => field.trim())
    .filter((field, index, all) => field && !blocked.has(field) && all.indexOf(field) === index)
})
const canSubmit = computed(() => {
  if (!props.hasApiKey || !form.model || !form.prompt.trim()) return false
  if (maxPromptLength.value && form.prompt.length > maxPromptLength.value) return false
  if (requiresSingleReference.value && !form.input_reference.trim()) return false
  if (requiresFirstLastFrame.value && (!form.first_frame_url.trim() || !form.last_frame_url.trim())) return false
  return true
})

const estimatedCost = computed(() => {
  const billing = selectedModel.value?.billing
  if (!billing) return '-'
  const seconds = Number.parseFloat(form.seconds || '0') || 0
  let total = billing.unit_price
  if (billing.mode === 'second') total = billing.unit_price * seconds
  if (billing.mode === 'segment') {
    const unitSeconds = billing.unit_seconds || 1
    total = billing.unit_price * Math.ceil(seconds / unitSeconds)
  }
  return `${billing.currency || 'USD'} ${total.toFixed(4)}`
})

watch(() => props.models, (models) => {
  if (!form.model && models.length > 0) {
    form.model = models[0].id
  }
}, { deep: true, immediate: true })

watch(selectedModel, (model) => {
  if (!model) return
  if (!modeOptions.value.some((option) => option.value === form.mode)) {
    form.mode = modeOptions.value[0]?.value || 'text_to_video'
  }
  if (model.seconds?.length && !model.seconds.includes(Number.parseInt(form.seconds, 10))) {
    form.seconds = String(model.seconds[0])
  }
  if (!model.seconds?.length && (!form.seconds || Number.parseInt(form.seconds, 10) <= 0)) {
    form.seconds = '5'
  }
  if (model.sizes?.length && !model.sizes.includes(form.size)) {
    form.size = model.sizes[0]
  }
  if (!model.sizes?.length && !form.size) {
    form.size = '1280x720'
  }
  form.extraBody = {}
})

watch(modeOptions, (options) => {
  if (!options.some((option) => option.value === form.mode)) {
    form.mode = options[0]?.value || 'text_to_video'
  }
})

function emitSubmit(): void {
  if (!canSubmit.value) return
  const extra_body = buildExtraBody()
  const inputReference = resolveInputReference()
  const payload: VideoCreateRequest = {
    model: form.model,
    prompt: form.prompt.trim(),
    seconds: String(form.seconds || '').trim() || undefined,
    size: form.size.trim() || undefined,
    input_reference: inputReference || undefined,
    extra_body: Object.keys(extra_body).length ? extra_body : undefined,
  }
  emit('submit', payload)
  form.prompt = ''
  form.input_reference = ''
  form.first_frame_url = ''
  form.last_frame_url = ''
  form.extraBody = {}
}

function buildExtraBody(): Record<string, unknown> {
  const extra: Record<string, unknown> = {}
  for (const field of extraFields.value) {
    const raw = form.extraBody[field]?.trim()
    if (!raw) continue
    extra[field] = parseExtraValue(raw)
  }
  addAllowedExtra(extra, 'generation_mode', form.mode)
  if (requiresFirstLastFrame.value) {
    addAllowedExtra(extra, 'first_frame_url', form.first_frame_url.trim())
    addAllowedExtra(extra, 'last_frame_url', form.last_frame_url.trim())
  }
  return extra
}

function resolveInputReference(): string {
  if (requiresFirstLastFrame.value) return form.first_frame_url.trim()
  if (requiresSingleReference.value) return form.input_reference.trim()
  return ''
}

function addAllowedExtra(extra: Record<string, unknown>, key: string, value: unknown): void {
  if (!selectedModel.value?.extra_body_allow?.includes(key)) return
  if (typeof value === 'string' && value.trim() === '') return
  extra[key] = value
}

function parseExtraValue(raw: string): unknown {
  if (raw === 'true') return true
  if (raw === 'false') return false
  if (raw === 'null') return null
  if (/^-?\d+(\.\d+)?$/.test(raw)) return Number(raw)
  if ((raw.startsWith('{') && raw.endsWith('}')) || (raw.startsWith('[') && raw.endsWith(']'))) {
    try {
      return JSON.parse(raw)
    } catch {
      return raw
    }
  }
  return raw
}

function numberFromLimit(key: string): number {
  const raw = selectedModel.value?.limits?.[key]
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}

function isVideoMode(value: string): value is VideoMode {
  return ['text_to_video', 'image_to_video', 'first_last_frame', 'reference_image', 'reference_video'].includes(value)
}
</script>
