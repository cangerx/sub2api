<template>
  <component :is="embedded ? 'div' : AppLayout">
    <div class="space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.video.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.video.subtitle') }}
          </p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" @click="loadAll">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div class="grid gap-6 xl:grid-cols-2">
        <section id="video-templates" class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
          <div class="flex flex-col gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 class="font-medium text-gray-900 dark:text-white">{{ t('admin.video.templates') }}</h2>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ templates.length }} {{ t('common.items', 'items') }}</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <select v-model="selectedTemplatePresetKey" class="input h-10 min-w-[260px]">
                <option value="">{{ t('admin.video.selectTemplatePreset') }}</option>
                <option v-for="preset in videoTemplatePresets" :key="preset.key" :value="preset.key">
                  {{ preset.provider }} - {{ preset.name }}
                </option>
              </select>
              <button class="btn btn-secondary" :disabled="!selectedTemplatePresetKey" @click="applyTemplatePresetAction">
                <Icon name="copy" size="sm" class="mr-2" />
                {{ t('admin.video.applyPreset') }}
              </button>
              <button class="btn btn-secondary" @click="openRecognizeDialog">
                <Icon name="sparkles" size="sm" class="mr-2" />
                {{ t('admin.video.aiRecognize') }}
              </button>
              <button class="btn btn-primary" @click="openTemplateDialog()">
                <Icon name="plus" size="sm" class="mr-2" />
                {{ t('admin.video.blankTemplate') }}
              </button>
            </div>
          </div>
          <DataTable :columns="templateColumns" :data="templates" :loading="loading">
            <template #cell-name="{ row }">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div class="text-xs text-gray-500">{{ row.create_method }} {{ row.create_path }}</div>
            </template>
            <template #cell-query_path="{ row }">
              <span class="font-mono text-xs">{{ row.query_method }} {{ row.query_path }}</span>
            </template>
            <template #cell-status="{ row }">
              <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex gap-1">
                <button class="icon-btn" @click="openTemplateDialog(row)" :title="t('common.edit', 'Edit')">
                  <Icon name="edit" size="sm" />
                </button>
                <button class="icon-btn danger" @click="removeTemplate(row)" :title="t('common.delete', 'Delete')">
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </section>

        <section id="video-models" class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
          <div class="flex items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-dark-700">
            <div>
              <h2 class="font-medium text-gray-900 dark:text-white">{{ t('admin.video.models') }}</h2>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ models.length }} {{ t('common.items', 'items') }}</p>
            </div>
            <button class="btn btn-primary" :disabled="templates.length === 0" @click="openModelDialog()">
              <Icon name="plus" size="sm" class="mr-2" />
              {{ t('common.create', 'Create') }}
            </button>
          </div>
          <DataTable :columns="modelColumns" :data="models" :loading="loading">
            <template #cell-public_model="{ row }">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.public_model }}</div>
              <div class="text-xs text-gray-500">{{ row.upstream_model_id }}</div>
            </template>
            <template #cell-template_id="{ row }">
              <span class="text-sm">{{ templateName(row.template_id) }}</span>
            </template>
            <template #cell-request_shape="{ row }">
              <span class="rounded bg-gray-100 px-2 py-0.5 text-xs dark:bg-dark-700">{{ row.request_shape }}</span>
            </template>
            <template #cell-status="{ row }">
              <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex gap-1">
                <button class="icon-btn" @click="openModelDialog(row)" :title="t('common.edit', 'Edit')">
                  <Icon name="edit" size="sm" />
                </button>
                <button class="icon-btn danger" @click="removeModel(row)" :title="t('common.delete', 'Delete')">
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </section>
      </div>

      <section id="video-template-test" class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <h2 class="font-medium text-gray-900 dark:text-white">{{ t('admin.video.templateTest') }}</h2>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.video.templateTestHint') }}</p>
        </div>
        <div class="grid gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <div class="space-y-3">
            <div class="grid gap-3 md:grid-cols-2">
              <label class="space-y-1">
                <span class="input-label">{{ t('admin.video.upstreamAccount') }}</span>
                <select v-model.number="templateTestForm.account_id" class="input">
                  <option :value="0">{{ t('admin.video.selectVideoAccount') }}</option>
                  <option v-for="account in videoAccounts" :key="account.id" :value="account.id">
                    {{ videoAccountLabel(account) }}
                  </option>
                </select>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.video.upstreamAccountHint') }}</span>
              </label>
              <label class="space-y-1">
                <span class="input-label">{{ t('admin.video.template') }}</span>
                <select v-model.number="templateTestForm.template_id" class="input">
                  <option :value="0">{{ t('common.select', '请选择') }}</option>
                  <option v-for="tpl in templates" :key="tpl.id" :value="tpl.id">{{ tpl.name }}</option>
                </select>
              </label>
            </div>
            <JsonField v-model="templateTestForm.body_json" :label="t('admin.video.createBodyJson')" :hint="t('admin.video.createBodyJsonHint')" />
            <div class="flex flex-wrap gap-2">
              <button class="btn btn-primary" type="button" :disabled="templateTestLoading || !templateTestForm.account_id || !templateTestForm.template_id" @click="testTemplateCreateAction">{{ t('admin.video.testCreate') }}</button>
              <input v-model="templateTestForm.upstream_task_id" class="input h-10 min-w-[240px] flex-1 font-mono text-xs" :placeholder="t('admin.video.upstreamTaskId')" />
              <button class="btn btn-secondary" type="button" :disabled="templateTestLoading || !templateTestForm.account_id || !templateTestForm.template_id || !templateTestForm.upstream_task_id" @click="testTemplateQueryAction">{{ t('admin.video.testQuery') }}</button>
            </div>
          </div>
          <div class="min-h-[220px] rounded-lg bg-gray-950 p-3 text-xs text-gray-100">
            <pre class="whitespace-pre-wrap break-words">{{ templateTestResult || '{}' }}</pre>
          </div>
        </div>
      </section>

      <section id="video-tasks" class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 class="font-medium text-gray-900 dark:text-white">{{ t('admin.video.tasks') }}</h2>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ taskTotal }} {{ t('common.items', 'items') }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <select v-model="taskFilters.status" class="input h-9 w-36">
              <option value="">{{ t('common.all', 'All') }}</option>
              <option value="queued">{{ t('admin.video.status.queued') }}</option>
              <option value="in_progress">{{ t('admin.video.status.inProgress') }}</option>
              <option value="completed">{{ t('admin.video.status.completed') }}</option>
              <option value="failed">{{ t('admin.video.status.failed') }}</option>
              <option value="cancelled">{{ t('admin.video.status.cancelled') }}</option>
              <option value="expired">{{ t('admin.video.status.expired') }}</option>
            </select>
            <input v-model="taskFilters.model" class="input h-9 w-48" :placeholder="t('admin.video.filterModel')" @keyup.enter="loadTasks" />
            <input v-model="taskFilters.user_id" class="input h-9 w-28" :placeholder="t('admin.video.filterUserId')" @keyup.enter="loadTasks" />
            <input v-model="taskFilters.api_key_id" class="input h-9 w-28" :placeholder="t('admin.video.filterKeyId')" @keyup.enter="loadTasks" />
            <input v-model="taskFilters.start_at" class="input h-9 w-44" type="datetime-local" />
            <input v-model="taskFilters.end_at" class="input h-9 w-44" type="datetime-local" />
            <button class="btn btn-secondary" :disabled="taskLoading" @click="loadTasks">{{ t('common.refresh', 'Refresh') }}</button>
          </div>
        </div>
        <DataTable :columns="taskColumns" :data="tasks" :loading="taskLoading">
          <template #cell-public_id="{ row }">
            <div class="font-mono text-xs font-medium text-gray-900 dark:text-white">{{ row.public_id }}</div>
            <div class="text-xs text-gray-500">{{ row.requested_model }}</div>
          </template>
          <template #cell-status="{ row }">
            <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
            <div class="mt-1 h-1.5 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
              <div class="h-full bg-blue-500" :style="{ width: `${Math.max(0, Math.min(100, row.progress || 0))}%` }" />
            </div>
          </template>
          <template #cell-cost="{ row }">
            <div class="text-xs">{{ t('admin.video.reservedCost') }} {{ money(row.reserved_cost) }}</div>
            <div class="text-xs">{{ t('admin.video.actualCost') }} {{ money(row.actual_cost) }}</div>
          </template>
          <template #cell-error="{ row }">
            <div class="max-w-[260px] truncate text-xs text-red-600 dark:text-red-300" :title="row.error_message || row.error_code || ''">
              {{ row.error_message || row.error_code || '-' }}
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex gap-1">
              <button class="icon-btn" @click="openTaskDetail(row)" :title="t('admin.video.details')">
                <Icon name="eye" size="sm" />
              </button>
              <button class="icon-btn" @click="requeueVideoTask(row)" :title="t('admin.video.requeue')">
                <Icon name="refresh" size="sm" />
              </button>
              <button class="icon-btn danger" @click="failVideoTask(row)" :title="t('admin.video.failAndRefund')">
                <Icon name="x" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </section>
    </div>

    <BaseDialog :show="templateDialogOpen" :title="templateEditing ? t('admin.video.editTemplate') : t('admin.video.createTemplate')" width="extra-wide" @close="templateDialogOpen = false">
      <form class="space-y-4" @submit.prevent="saveTemplate">
        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.name') }}</span>
            <input v-model="templateForm.name" class="input" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.statusLabel') }}</span>
            <select v-model="templateForm.status" class="input">
              <option value="active">{{ t('admin.video.status.active') }}</option>
              <option value="disabled">{{ t('admin.video.status.disabled') }}</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.createMethod') }}</span>
            <input v-model="templateForm.create_method" class="input" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.createPath') }}</span>
            <input v-model="templateForm.create_path" class="input font-mono" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.queryMethod') }}</span>
            <input v-model="templateForm.query_method" class="input" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.queryPath') }}</span>
            <input v-model="templateForm.query_path" class="input font-mono" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.contentMethod') }}</span>
            <input v-model="templateForm.content_method" class="input" />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.contentPath') }}</span>
            <input v-model="templateForm.content_path" class="input font-mono" />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.cancelMethod') }}</span>
            <input v-model="templateForm.cancel_method" class="input" />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.cancelPath') }}</span>
            <input v-model="templateForm.cancel_path" class="input font-mono" />
          </label>
        </div>
        <JsonGrid v-model:statusMapping="templateForm.status_mapping_json" v-model:resultMapping="templateForm.result_mapping_json" v-model:errorMapping="templateForm.error_mapping_json" v-model:pollConfig="templateForm.poll_config_json" v-model:timeoutConfig="templateForm.timeout_config_json" />
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="templateDialogOpen = false">{{ t('common.cancel', 'Cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ t('common.save', 'Save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="recognizeDialogOpen" :title="t('admin.video.aiRecognizeTitle')" width="wide" @close="recognizeDialogOpen = false">
      <form class="space-y-4" @submit.prevent="recognizeTemplateAction">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.video.aiRecognizeHint') }}</p>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.account') }}</span>
            <select v-model.number="recognizeForm.account_id" class="input" required>
              <option :value="0" disabled>{{ t('admin.video.selectAccount') }}</option>
              <option v-for="account in videoAccounts" :key="account.id" :value="account.id">{{ account.name }}</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.aiModel') }}</span>
            <input v-model="recognizeForm.model" class="input font-mono" placeholder="video-chat-model" required />
          </label>
        </div>
        <label class="space-y-1 block">
          <span class="input-label">{{ t('admin.video.aiDocument') }}</span>
          <textarea v-model="recognizeForm.document" class="input font-mono h-64" :placeholder="t('admin.video.aiDocumentPlaceholder')" required></textarea>
        </label>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="recognizeDialogOpen = false">{{ t('common.cancel', 'Cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="recognizing || !recognizeForm.account_id || !recognizeForm.model || !recognizeForm.document">
            <Icon name="sparkles" size="sm" class="mr-2" :class="recognizing ? 'animate-pulse' : ''" />
            {{ recognizing ? t('admin.video.aiRecognizing') : t('admin.video.aiRecognize') }}
          </button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="modelDialogOpen" :title="modelEditing ? t('admin.video.editModel') : t('admin.video.createModel')" width="extra-wide" @close="modelDialogOpen = false">
      <form class="space-y-4" @submit.prevent="saveModel">
        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.publicModel') }}</span>
            <input v-model="modelForm.public_model" class="input" required />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.displayName') }}</span>
            <input v-model="modelForm.display_name" class="input" />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.template') }}</span>
            <select v-model.number="modelForm.template_id" class="input" required>
              <option v-for="tpl in templates" :key="tpl.id" :value="tpl.id">{{ tpl.name }}</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.requestShape') }}</span>
            <select v-model="modelForm.request_shape" class="input">
              <option v-for="shape in requestShapes" :key="shape" :value="shape">{{ shape }}</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.upstreamModelId') }}</span>
            <input v-model="modelForm.upstream_model_id" class="input" :placeholder="t('admin.video.upstreamModelIdOptional')" />
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.video.upstreamModelIdHint') }}</span>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.statusLabel') }}</span>
            <select v-model="modelForm.status" class="input">
              <option value="active">{{ t('admin.video.status.active') }}</option>
              <option value="deprecated">{{ t('admin.video.status.deprecated') }}</option>
              <option value="disabled">{{ t('admin.video.status.disabled') }}</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.sortOrder') }}</span>
            <input v-model.number="modelForm.sort_order" type="number" class="input" />
          </label>
          <label class="space-y-1">
            <span class="input-label">{{ t('admin.video.extraBodyAllow') }}</span>
            <input v-model="modelForm.extra_body_allow_text" class="input" :placeholder="t('admin.video.extraBodyAllowPlaceholder')" />
          </label>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <JsonField v-model="modelForm.capabilities_json" :label="t('admin.video.capabilitiesJson')" :hint="t('admin.video.capabilitiesJsonHint')" />
          <JsonField v-model="modelForm.defaults_json" :label="t('admin.video.defaultsJson')" :hint="t('admin.video.defaultsJsonHint')" />
          <JsonField v-model="modelForm.limits_json" :label="t('admin.video.limitsJson')" :hint="t('admin.video.limitsJsonHint')" />
          <JsonField v-model="modelForm.supported_options_json" :label="t('admin.video.supportedOptionsJson')" :hint="t('admin.video.supportedOptionsJsonHint')" />
        </div>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="modelDialogOpen = false">{{ t('common.cancel', 'Cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ t('common.save', 'Save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="taskDetailOpen" :title="t('admin.video.taskDetail')" width="full" @close="taskDetailOpen = false">
      <div v-if="selectedTask" class="space-y-4">
        <div class="grid gap-3 text-sm md:grid-cols-2 lg:grid-cols-4">
          <div><span class="text-gray-500">{{ t('admin.video.task') }}</span><div class="font-mono text-xs">{{ selectedTask.public_id }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.userKey') }}</span><div>{{ selectedTask.user_id }} / {{ selectedTask.api_key_id }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.account') }}</span><div>{{ selectedTask.account_id }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.statusLabel') }}</span><div><span :class="statusClass(selectedTask.status)">{{ statusText(selectedTask.status) }}</span></div></div>
          <div><span class="text-gray-500">{{ t('admin.video.requestedModel') }}</span><div>{{ selectedTask.requested_model }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.upstreamModel') }}</span><div>{{ selectedTask.upstream_model }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.billing') }}</span><div>{{ selectedTask.billing_mode || '-' }} {{ money(selectedTask.actual_cost) }}</div></div>
          <div><span class="text-gray-500">{{ t('admin.video.polls') }}</span><div>{{ selectedTask.poll_count }}</div></div>
        </div>
        <div class="grid gap-4 lg:grid-cols-2">
          <JsonPreview :title="t('admin.video.requestPayload')" :value="selectedTask.request_payload" />
          <JsonPreview :title="t('admin.video.upstreamRequest')" :value="selectedTask.upstream_request_payload" />
          <JsonPreview :title="t('admin.video.upstreamResponse')" :value="selectedTask.upstream_response_payload" />
          <JsonPreview :title="t('admin.video.resultPayload')" :value="selectedTask.result_payload" />
        </div>
      </div>
      <div v-else class="py-10 text-center text-sm text-gray-500">
        {{ taskDetailLoading ? t('common.loading', '加载中...') : t('admin.video.noTaskSelected') }}
      </div>
    </BaseDialog>
  </component>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, nextTick, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import videoAPI from '@/api/admin/video'
import accountsAPI from '@/api/admin/accounts'
import type { VideoModel, VideoModelPayload, VideoTask, VideoTemplate, VideoTemplatePayload } from '@/api/admin/video'
import type { Account } from '@/types'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

defineProps<{
  embedded?: boolean
}>()

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const templates = ref<VideoTemplate[]>([])
const models = ref<VideoModel[]>([])
const videoAccounts = ref<Account[]>([])
const requestShapes = ref<string[]>(['videos'])
const tasks = ref<VideoTask[]>([])
const taskTotal = ref(0)
const taskLoading = ref(false)
const taskDetailLoading = ref(false)
const templateTestLoading = ref(false)
const templateTestResult = ref('')
const templateDialogOpen = ref(false)
const modelDialogOpen = ref(false)
const taskDetailOpen = ref(false)
const templateEditing = ref<VideoTemplate | null>(null)
const modelEditing = ref<VideoModel | null>(null)
const selectedTask = ref<VideoTask | null>(null)
const selectedTemplatePresetKey = ref('')
const recognizeDialogOpen = ref(false)
const recognizing = ref(false)
const recognizeForm = reactive({
  account_id: 0,
  model: '',
  document: '',
})

interface VideoTemplatePreset {
  key: string
  provider: string
  name: string
  payload: VideoTemplatePayload
  testBody?: Record<string, unknown>
}

const baseStatusMapping = {
  queued: 'queued',
  pending: 'queued',
  created: 'queued',
  processing: 'in_progress',
  running: 'in_progress',
  in_progress: 'in_progress',
  submitted: 'in_progress',
  succeeded: 'completed',
  success: 'completed',
  succeed: 'completed',
  completed: 'completed',
  done: 'completed',
  failed: 'failed',
  error: 'failed',
  cancelled: 'cancelled',
  canceled: 'cancelled',
  expired: 'expired',
}

const basePollConfig = {
  interval_seconds: 5,
  backoff_max_seconds: 30,
  max_attempts: 240,
}

const baseTimeoutConfig = {
  create_seconds: 60,
  query_seconds: 30,
  content_seconds: 300,
}

const megaByAIVideoTestBody = {
  model: 'videos-mini',
  prompt: 'A cinematic close-up shot, realistic lighting',
  duration: 5,
  ratio: '16:9',
  resolution: '720p',
}

const videoTemplatePresets: VideoTemplatePreset[] = [
  {
    key: 'openai-video',
    provider: '内置',
    name: 'OpenAI 风格视频接口',
    payload: {
      name: '内置 - OpenAI 风格视频接口',
      create_method: 'POST',
      create_path: '/v1/videos',
      query_method: 'GET',
      query_path: '/v1/videos/{task_id}',
      content_method: 'GET',
      content_path: '/v1/videos/{task_id}/content.mp4',
      cancel_method: 'POST',
      cancel_path: '/v1/videos/{task_id}/cancel',
      status_mapping: { ...baseStatusMapping },
      result_mapping: {
        content_url: 'content_url',
        seconds: 'seconds',
        progress: 'progress',
      },
      error_mapping: {
        code: 'error.code',
        message: 'error.message',
      },
      poll_config: { ...basePollConfig },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
  {
    key: 'megabyai-video',
    provider: 'MegaByAI',
    name: '异步视频接口',
    payload: {
      name: 'MegaByAI - 异步视频接口',
      create_method: 'POST',
      create_path: '/v1/videos',
      query_method: 'GET',
      query_path: '/v1/videos/{task_id}',
      content_method: null,
      content_path: null,
      cancel_method: null,
      cancel_path: null,
      status_mapping: {
        queued: 'queued',
        in_progress: 'in_progress',
        completed: 'completed',
        failed: 'failed',
      },
      result_mapping: {
        content_url: 'video_url|url|metadata.content_url',
        seconds: 'seconds',
        progress: 'progress',
      },
      error_mapping: {
        code: 'error.code',
        message: 'error.message',
      },
      poll_config: { ...basePollConfig, backoff_max_seconds: 10, max_attempts: 180 },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
    testBody: megaByAIVideoTestBody,
  },
  {
    key: 'volcengine-ark-video',
    provider: '内置',
    name: '火山 Ark / 豆包视频',
    payload: {
      name: '内置 - 火山 Ark / 豆包官方视频',
      create_method: 'POST',
      create_path: '/api/v3/contents/generations/tasks',
      query_method: 'GET',
      query_path: '/api/v3/contents/generations/tasks/{task_id}',
      content_method: null,
      content_path: null,
      cancel_method: null,
      cancel_path: null,
      status_mapping: { ...baseStatusMapping },
      result_mapping: {
        content_url: 'data.video_url',
        seconds: 'data.duration',
        progress: 'data.progress',
      },
      error_mapping: {
        code: 'error.code',
        message: 'error.message',
      },
      poll_config: { ...basePollConfig },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
  {
    key: 'google-veo-video',
    provider: '内置',
    name: 'Google Gemini / Veo 视频',
    payload: {
      name: '内置 - Google Gemini / Veo 官方视频',
      create_method: 'POST',
      create_path: '/v1beta/models/{model}:predictLongRunning',
      query_method: 'GET',
      query_path: '/v1beta/{task_id}',
      content_method: null,
      content_path: null,
      cancel_method: 'POST',
      cancel_path: '/v1beta/{task_id}:cancel',
      status_mapping: {
        ...baseStatusMapping,
        done: 'completed',
      },
      result_mapping: {
        content_url: 'response.generateVideoResponse.generatedSamples.0.video.uri',
        seconds: 'response.generateVideoResponse.generatedSamples.0.video.duration',
        progress: 'metadata.progressPercentage',
      },
      error_mapping: {
        code: 'error.code',
        message: 'error.message',
      },
      poll_config: { ...basePollConfig, interval_seconds: 10, backoff_max_seconds: 60, max_attempts: 180 },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
  {
    key: 'duomi-seedance-video',
    provider: '多米API',
    name: 'SEEDANCE 视频（官方格式）',
    payload: {
      name: '多米API - SEEDANCE 视频',
      create_method: 'POST',
      create_path: '/api/v3/contents/generations/tasks',
      query_method: 'GET',
      query_path: '/api/v3/contents/generations/tasks/{task_id}',
      content_method: 'GET',
      content_path: '',
      cancel_method: 'POST',
      cancel_path: '',
      status_mapping: { ...baseStatusMapping },
      result_mapping: {
        content_url: 'data.videos.0.url',
        seconds: 'data.duration',
        progress: 'progress',
      },
      error_mapping: {
        code: 'state',
        message: 'message',
      },
      poll_config: { ...basePollConfig },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
  {
    key: 'duomi-kling-video',
    provider: '多米API',
    name: 'KLING 文生视频（官方格式）',
    payload: {
      name: '多米API - KLING 文生视频',
      create_method: 'POST',
      create_path: '/api/video/kling/v1/videos/text2video',
      query_method: 'GET',
      query_path: '/api/video/kling/v1/videos/text2video/{task_id}',
      content_method: 'GET',
      content_path: '',
      cancel_method: 'POST',
      cancel_path: '',
      status_mapping: { ...baseStatusMapping },
      result_mapping: {
        content_url: 'data.task_result.videos.0.url',
        seconds: 'data.task_result.videos.0.duration',
        progress: 'data.progress',
      },
      error_mapping: {
        code: 'data.task_status',
        message: 'data.task_status_msg',
      },
      poll_config: { ...basePollConfig },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
  {
    key: 'duomi-gpt-image-2',
    provider: '多米API',
    name: '图片生成 gpt-image-2 / NANO-BANANA（异步）',
    payload: {
      name: '多米API - 图片生成（gpt-image-2）',
      create_method: 'POST',
      create_path: '/v1/images/generations?async=true',
      query_method: 'GET',
      query_path: '/v1/tasks/{task_id}',
      content_method: 'GET',
      content_path: '',
      cancel_method: 'POST',
      cancel_path: '',
      status_mapping: { ...baseStatusMapping },
      result_mapping: {
        content_url: 'data.images.0.url',
        seconds: '',
        progress: 'progress',
      },
      error_mapping: {
        code: 'state',
        message: 'message',
      },
      poll_config: { ...basePollConfig },
      timeout_config: { ...baseTimeoutConfig },
      status: 'active',
    },
  },
]

const templateColumns = computed(() => [
  { key: 'name', label: t('admin.video.name') },
  { key: 'query_path', label: t('admin.video.queryEndpoint') },
  { key: 'status', label: t('admin.video.statusLabel') },
  { key: 'actions', label: t('common.actions') },
])
const modelColumns = computed(() => [
  { key: 'public_model', label: t('admin.video.model') },
  { key: 'template_id', label: t('admin.video.template') },
  { key: 'request_shape', label: t('admin.video.requestShapeShort') },
  { key: 'status', label: t('admin.video.statusLabel') },
  { key: 'actions', label: t('common.actions') },
])
const taskColumns = computed(() => [
  { key: 'public_id', label: t('admin.video.task') },
  { key: 'status', label: t('admin.video.statusLabel') },
  { key: 'billing_state', label: t('admin.video.billing') },
  { key: 'cost', label: t('admin.video.cost') },
  { key: 'poll_count', label: t('admin.video.polls') },
  { key: 'error', label: t('admin.video.error') },
  { key: 'actions', label: t('common.actions') },
])

const taskFilters = reactive({
  page: 1,
  page_size: 20,
  status: '',
  model: '',
  user_id: '',
  api_key_id: '',
  start_at: '',
  end_at: '',
})

const templateForm = reactive({
  name: '',
  create_method: 'POST',
  create_path: '/v1/videos',
  query_method: 'GET',
  query_path: '/v1/videos/{task_id}',
  content_method: 'GET',
  content_path: '',
  cancel_method: 'POST',
  cancel_path: '',
  status: 'active',
  status_mapping_json: '{\n  "queued": "queued",\n  "processing": "in_progress",\n  "succeeded": "completed",\n  "failed": "failed"\n}',
  result_mapping_json: '{\n  "content_url": "data.video_url",\n  "seconds": "data.duration",\n  "progress": "data.progress"\n}',
  error_mapping_json: '{\n  "code": "error.code",\n  "message": "error.message"\n}',
  poll_config_json: '{\n  "interval_seconds": 5,\n  "backoff_max_seconds": 30,\n  "max_attempts": 240\n}',
  timeout_config_json: '{\n  "create_seconds": 60,\n  "query_seconds": 30,\n  "content_seconds": 300\n}',
})

const modelForm = reactive({
  public_model: '',
  display_name: '',
  template_id: 0,
  upstream_model_id: '',
  request_shape: 'videos',
  status: 'active',
  sort_order: 0,
  extra_body_allow_text: '',
  capabilities_json: '{}',
  defaults_json: '{\n  "seconds": 5,\n  "size": "1280x720"\n}',
  limits_json: '{}',
  supported_options_json: '{\n  "seconds": [5, 10],\n  "sizes": ["1280x720", "720x1280"]\n}',
})

const templateTestForm = reactive({
  account_id: 0,
  template_id: 0,
  upstream_task_id: '',
  body_json: '{\n  "model": "upstream-model-id",\n  "prompt": "A short cinematic video",\n  "duration": 5\n}',
})

const JsonField = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    hint: { type: String, required: false, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'input-label' }, props.label),
      h('textarea', {
        class: 'input min-h-[140px] font-mono text-xs',
        value: props.modelValue,
        onInput: (event: Event) => emit('update:modelValue', (event.target as HTMLTextAreaElement).value),
      }),
      props.hint ? h('span', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.hint) : null,
    ])
  },
})

const JsonGrid = defineComponent({
  props: {
    statusMapping: { type: String, required: true },
    resultMapping: { type: String, required: true },
    errorMapping: { type: String, required: true },
    pollConfig: { type: String, required: true },
    timeoutConfig: { type: String, required: true },
  },
  emits: ['update:statusMapping', 'update:resultMapping', 'update:errorMapping', 'update:pollConfig', 'update:timeoutConfig'],
  setup(props, { emit }) {
    const field = (label: string, value: string, eventName: string, hint = '') => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'input-label' }, label),
      h('textarea', {
        class: 'input min-h-[130px] font-mono text-xs',
        value,
        onInput: (event: Event) => emit(eventName as any, (event.target as HTMLTextAreaElement).value),
      }),
      hint ? h('span', { class: 'text-xs text-gray-500 dark:text-gray-400' }, hint) : null,
    ])
    return () => h('div', { class: 'grid gap-4 md:grid-cols-2' }, [
      field(t('admin.video.statusMappingJson'), props.statusMapping, 'update:statusMapping', t('admin.video.statusMappingJsonHint')),
      field(t('admin.video.resultMappingJson'), props.resultMapping, 'update:resultMapping', t('admin.video.resultMappingJsonHint')),
      field(t('admin.video.errorMappingJson'), props.errorMapping, 'update:errorMapping', t('admin.video.errorMappingJsonHint')),
      field(t('admin.video.pollConfigJson'), props.pollConfig, 'update:pollConfig', t('admin.video.pollConfigJsonHint')),
      field(t('admin.video.timeoutConfigJson'), props.timeoutConfig, 'update:timeoutConfig', t('admin.video.timeoutConfigJsonHint')),
    ])
  },
})

const JsonPreview = defineComponent({
  props: { title: { type: String, required: true }, value: { type: null, required: false } },
  setup(props) {
    return () => h('div', { class: 'space-y-1' }, [
      h('div', { class: 'text-xs font-medium text-gray-500 dark:text-gray-400' }, props.title),
      h('pre', { class: 'max-h-80 overflow-auto rounded bg-gray-950 p-3 text-xs text-gray-100 whitespace-pre-wrap break-words' }, prettyJSON(props.value)),
    ])
  },
})

function parseJSONField<T>(raw: string, label: string): T {
  try {
    return JSON.parse(raw || '{}') as T
  } catch {
    throw new Error(t('admin.video.invalidJson', { label }))
  }
}

function stringify(value: unknown, fallback: string) {
  return JSON.stringify(value ?? JSON.parse(fallback), null, 2)
}

async function loadAll() {
  loading.value = true
  try {
    const [tplResp, modelResp, shapeResp, accountResp] = await Promise.all([
      videoAPI.listTemplates(),
      videoAPI.listModels(),
      videoAPI.listRequestShapes(),
      accountsAPI.list(1, 200, { platform: 'video', type: 'apikey', status: 'active' }),
    ])
    templates.value = tplResp.items || []
    models.value = modelResp.items || []
    videoAccounts.value = accountResp.items || []
    requestShapes.value = shapeResp.items?.length ? shapeResp.items : ['videos']
    if (!templateTestForm.template_id && templates.value.length > 0) {
      templateTestForm.template_id = templates.value[0].id
    }
    if (!templateTestForm.account_id && videoAccounts.value.length > 0) {
      templateTestForm.account_id = videoAccounts.value[0].id
    }
    await loadTasks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.loadError')))
  } finally {
    loading.value = false
  }
}

async function loadTasks() {
  taskLoading.value = true
  try {
    const resp = await videoAPI.listTasks({
      page: taskFilters.page,
      page_size: taskFilters.page_size,
      status: taskFilters.status || undefined,
      model: taskFilters.model || undefined,
      user_id: taskFilters.user_id || undefined,
      api_key_id: taskFilters.api_key_id || undefined,
      start_at: localDateTimeToRFC3339(taskFilters.start_at),
      end_at: localDateTimeToRFC3339(taskFilters.end_at),
    })
    tasks.value = resp.items || []
    taskTotal.value = resp.total || 0
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.tasksLoadError')))
  } finally {
    taskLoading.value = false
  }
}

function openTemplateDialog(template?: VideoTemplate) {
  templateEditing.value = template || null
  Object.assign(templateForm, {
    name: template?.name || '',
    create_method: template?.create_method || 'POST',
    create_path: template?.create_path || '/v1/videos',
    query_method: template?.query_method || 'GET',
    query_path: template?.query_path || '/v1/videos/{task_id}',
    content_method: template ? (template.content_method || '') : 'GET',
    content_path: template?.content_path || '',
    cancel_method: template ? (template.cancel_method || '') : 'POST',
    cancel_path: template?.cancel_path || '',
    status: template?.status || 'active',
    status_mapping_json: stringify(template?.status_mapping, '{}'),
    result_mapping_json: stringify(template?.result_mapping, '{}'),
    error_mapping_json: stringify(template?.error_mapping, '{}'),
    poll_config_json: stringify(template?.poll_config, '{}'),
    timeout_config_json: stringify(template?.timeout_config, '{}'),
  })
  templateDialogOpen.value = true
}

function applyTemplatePresetAction() {
  const preset = videoTemplatePresets.find((item) => item.key === selectedTemplatePresetKey.value)
  if (!preset) return
  const payload = preset.payload
  templateEditing.value = null
  Object.assign(templateForm, {
    name: payload.name || '',
    create_method: payload.create_method || 'POST',
    create_path: payload.create_path || '/v1/videos',
    query_method: payload.query_method || 'GET',
    query_path: payload.query_path || '/v1/videos/{task_id}',
    content_method: payload.content_method || '',
    content_path: payload.content_path || '',
    cancel_method: payload.cancel_method || '',
    cancel_path: payload.cancel_path || '',
    status: payload.status || 'active',
    status_mapping_json: stringify(payload.status_mapping, '{}'),
    result_mapping_json: stringify(payload.result_mapping, '{}'),
    error_mapping_json: stringify(payload.error_mapping, '{}'),
    poll_config_json: stringify(payload.poll_config, '{}'),
    timeout_config_json: stringify(payload.timeout_config, '{}'),
  })
  if (preset.testBody) {
    templateTestForm.body_json = stringify(preset.testBody, '{}')
  }
  templateDialogOpen.value = true
}

function openRecognizeDialog() {
  if (!recognizeForm.account_id && videoAccounts.value.length > 0) {
    recognizeForm.account_id = videoAccounts.value[0].id
  }
  recognizeDialogOpen.value = true
}

async function recognizeTemplateAction() {
  if (!recognizeForm.account_id || !recognizeForm.model.trim() || !recognizeForm.document.trim()) return
  recognizing.value = true
  try {
    const template = await videoAPI.recognizeTemplate({
      account_id: recognizeForm.account_id,
      model: recognizeForm.model.trim(),
      document: recognizeForm.document.trim(),
    })
    // Prefill the create-template form with the recognized draft (id is 0 so it
    // saves as a new template after the admin reviews it).
    openTemplateDialog(template)
    templateEditing.value = null
    recognizeDialogOpen.value = false
    appStore.showSuccess(t('admin.video.aiRecognizeSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.aiRecognizeFailed')))
  } finally {
    recognizing.value = false
  }
}

function templatePayload(): VideoTemplatePayload {
  return {
    name: templateForm.name,
    create_method: templateForm.create_method,
    create_path: templateForm.create_path,
    query_method: templateForm.query_method,
    query_path: templateForm.query_path,
    content_method: templateForm.content_method || null,
    content_path: templateForm.content_path || null,
    cancel_method: templateForm.cancel_method || null,
    cancel_path: templateForm.cancel_path || null,
    status_mapping: parseJSONField(templateForm.status_mapping_json, t('admin.video.statusMappingJson')),
    result_mapping: parseJSONField(templateForm.result_mapping_json, t('admin.video.resultMappingJson')),
    error_mapping: parseJSONField(templateForm.error_mapping_json, t('admin.video.errorMappingJson')),
    poll_config: parseJSONField(templateForm.poll_config_json, t('admin.video.pollConfigJson')),
    timeout_config: parseJSONField(templateForm.timeout_config_json, t('admin.video.timeoutConfigJson')),
    status: templateForm.status,
  }
}

async function saveTemplate() {
  saving.value = true
  try {
    const payload = templatePayload()
    if (templateEditing.value) {
      await videoAPI.updateTemplate(templateEditing.value.id, payload)
    } else {
      await videoAPI.createTemplate(payload)
    }
    templateDialogOpen.value = false
    appStore.showSuccess(t('common.saved', 'Saved'))
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, error instanceof Error ? error.message : t('common.saveFailed', 'Save failed')))
  } finally {
    saving.value = false
  }
}

async function removeTemplate(template: VideoTemplate) {
  if (!window.confirm(t('admin.video.deleteTemplateConfirm', { name: template.name }))) return
  try {
    await videoAPI.deleteTemplate(template.id)
    appStore.showSuccess(t('common.deleted', 'Deleted'))
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.deleteFailed', 'Delete failed')))
  }
}

async function testTemplateCreateAction() {
  if (!templateTestForm.account_id || !templateTestForm.template_id) {
    appStore.showWarning(t('admin.video.selectAccountAndTemplate'))
    return
  }
  templateTestLoading.value = true
  try {
    const result = await videoAPI.testTemplateCreate({
      account_id: templateTestForm.account_id,
      template_id: templateTestForm.template_id,
      body: parseJSONField<Record<string, unknown>>(templateTestForm.body_json, t('admin.video.createBodyJson')),
    })
    templateTestResult.value = JSON.stringify(result, null, 2)
    if (result.task_id) {
      templateTestForm.upstream_task_id = result.task_id
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, error instanceof Error ? error.message : t('admin.video.templateTestFailed')))
  } finally {
    templateTestLoading.value = false
  }
}

async function testTemplateQueryAction() {
  if (!templateTestForm.account_id || !templateTestForm.template_id || !templateTestForm.upstream_task_id) {
    appStore.showWarning(t('admin.video.selectAccountTemplateAndTask'))
    return
  }
  templateTestLoading.value = true
  try {
    const result = await videoAPI.testTemplateQuery({
      account_id: templateTestForm.account_id,
      template_id: templateTestForm.template_id,
      upstream_task_id: templateTestForm.upstream_task_id,
    })
    templateTestResult.value = JSON.stringify(result, null, 2)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, error instanceof Error ? error.message : t('admin.video.templateTestFailed')))
  } finally {
    templateTestLoading.value = false
  }
}

function openModelDialog(model?: VideoModel) {
  modelEditing.value = model || null
  Object.assign(modelForm, {
    public_model: model?.public_model || '',
    display_name: model?.display_name || '',
    template_id: model?.template_id || templates.value[0]?.id || 0,
    upstream_model_id: model?.upstream_model_id || '',
    request_shape: model?.request_shape || 'videos',
    status: model?.status || 'active',
    sort_order: model?.sort_order || 0,
    extra_body_allow_text: (model?.extra_body_allow || []).join(', '),
    capabilities_json: stringify(model?.capabilities, '{}'),
    defaults_json: stringify(model?.defaults, '{}'),
    limits_json: stringify(model?.limits, '{}'),
    supported_options_json: stringify(model?.supported_options, '{}'),
  })
  modelDialogOpen.value = true
}

function modelPayload(): VideoModelPayload {
  return {
    public_model: modelForm.public_model,
    display_name: modelForm.display_name || null,
    template_id: modelForm.template_id,
    upstream_model_id: modelForm.upstream_model_id,
    request_shape: modelForm.request_shape,
    status: modelForm.status,
    sort_order: modelForm.sort_order,
    extra_body_allow: modelForm.extra_body_allow_text.split(',').map((v) => v.trim()).filter(Boolean),
    capabilities: parseJSONField(modelForm.capabilities_json, t('admin.video.capabilitiesJson')),
    defaults: parseJSONField(modelForm.defaults_json, t('admin.video.defaultsJson')),
    limits: parseJSONField(modelForm.limits_json, t('admin.video.limitsJson')),
    supported_options: parseJSONField(modelForm.supported_options_json, t('admin.video.supportedOptionsJson')),
  }
}

async function saveModel() {
  saving.value = true
  try {
    const payload = modelPayload()
    if (modelEditing.value) {
      await videoAPI.updateModel(modelEditing.value.id, payload)
    } else {
      await videoAPI.createModel(payload)
    }
    modelDialogOpen.value = false
    appStore.showSuccess(t('common.saved', 'Saved'))
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, error instanceof Error ? error.message : t('common.saveFailed', 'Save failed')))
  } finally {
    saving.value = false
  }
}

async function removeModel(model: VideoModel) {
  if (!window.confirm(t('admin.video.deleteModelConfirm', { name: model.public_model }))) return
  try {
    await videoAPI.deleteModel(model.id)
    appStore.showSuccess(t('common.deleted', 'Deleted'))
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.deleteFailed', 'Delete failed')))
  }
}

function templateName(id: number) {
  return templates.value.find((tpl) => tpl.id === id)?.name || `#${id}`
}

function videoAccountLabel(account: Account) {
  return `#${account.id} ${account.name}`
}

function statusText(status: string) {
  const keyMap: Record<string, string> = {
    active: 'active',
    disabled: 'disabled',
    deprecated: 'deprecated',
    queued: 'queued',
    in_progress: 'inProgress',
    completed: 'completed',
    failed: 'failed',
    cancelled: 'cancelled',
    expired: 'expired',
  }
  const key = keyMap[status]
  return key ? t(`admin.video.status.${key}`) : status
}

function statusClass(status: string) {
  return [
    'inline-flex rounded px-2 py-0.5 text-xs font-medium',
    status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
]
}

function money(value: number | string | null | undefined) {
  const n = Number(value || 0)
  return `$${n.toFixed(4)}`
}

async function requeueVideoTask(task: VideoTask) {
  if (!window.confirm(t('admin.video.requeueConfirm', { id: task.public_id }))) return
  try {
    await videoAPI.requeueTask(task.public_id)
    appStore.showSuccess(t('common.saved', 'Saved'))
    await loadTasks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.taskActionFailed')))
  }
}

async function failVideoTask(task: VideoTask) {
  if (!window.confirm(t('admin.video.failConfirm', { id: task.public_id }))) return
  try {
    await videoAPI.failTask(task.public_id, { code: 'admin_forced_failed', message: t('admin.video.adminForcedFailed') })
    appStore.showSuccess(t('common.saved', 'Saved'))
    await loadTasks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.taskActionFailed')))
  }
}

async function openTaskDetail(task: VideoTask) {
  taskDetailOpen.value = true
  taskDetailLoading.value = true
  selectedTask.value = task
  try {
    selectedTask.value = await videoAPI.getTask(task.public_id)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.video.taskLoadError')))
  } finally {
    taskDetailLoading.value = false
  }
}

function prettyJSON(value: unknown) {
  if (value == null) return '{}'
  return JSON.stringify(value, null, 2)
}

function localDateTimeToRFC3339(value: string) {
  if (!value) return undefined
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return undefined
  return date.toISOString()
}

function scrollToCurrentSection() {
  const map: Record<string, string> = {
    AdminVideoTemplates: 'video-templates',
    AdminVideoModels: 'video-models',
    AdminVideoTasks: 'video-tasks',
  }
  const id = map[String(route.name || '')]
  if (!id) return
  nextTick(() => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

onMounted(async () => {
  await loadAll()
  scrollToCurrentSection()
})
</script>

<style scoped>
.icon-btn {
  @apply rounded p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400;
}

.icon-btn.danger {
  @apply hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400;
}
</style>
