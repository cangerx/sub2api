<template>
    <div class="space-y-6">
      <!-- S3 Storage Config -->
      <div class="card p-6">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.backup.s3.title') }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.backup.s3.descriptionPrefix') }}
              <button type="button" class="text-primary-600 underline hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300" @click="showR2Guide = true">Cloudflare R2</button>
              {{ t('admin.backup.s3.descriptionSuffix') }}
            </p>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.provider') }}</label>
            <select v-model="s3Form.provider" class="input w-full" @change="applyProviderDefaults">
              <option value="r2">Cloudflare R2</option>
              <option value="oss">阿里云 OSS</option>
              <option value="s3">S3 Compatible</option>
              <option value="local">{{ t('admin.backup.s3.local') }}</option>
            </select>
          </div>
          <div v-if="s3Form.provider === 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.localPath') }}</label>
            <input v-model="s3Form.local_path" class="input w-full" placeholder="/data/ccapi/storage" />
          </div>
          <div v-if="s3Form.provider !== 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.endpoint') }}</label>
            <input v-model="s3Form.endpoint" class="input w-full" :placeholder="endpointPlaceholder" />
          </div>
          <div v-if="s3Form.provider !== 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.region') }}</label>
            <input v-model="s3Form.region" class="input w-full" :placeholder="regionPlaceholder" />
          </div>
          <div v-if="s3Form.provider !== 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.bucket') }}</label>
            <input v-model="s3Form.bucket" class="input w-full" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.prefix') }}</label>
            <input v-model="s3Form.prefix" class="input w-full" placeholder="backups/" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.publicBaseUrl') }}</label>
            <input v-model="s3Form.public_base_url" class="input w-full" placeholder="https://cdn.example.com/backups" />
          </div>
          <div v-if="s3Form.provider !== 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.accessKeyId') }}</label>
            <input v-model="s3Form.access_key_id" class="input w-full" />
          </div>
          <div v-if="s3Form.provider !== 'local'">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.s3.secretAccessKey') }}</label>
            <input v-model="s3Form.secret_access_key" type="password" class="input w-full" :placeholder="s3SecretConfigured ? t('admin.backup.s3.secretConfigured') : ''" />
          </div>
          <label v-if="s3Form.provider !== 'local'" class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 md:col-span-2">
            <input v-model="s3Form.force_path_style" type="checkbox" />
            <span>{{ t('admin.backup.s3.forcePathStyle') }}</span>
          </label>
        </div>
        <div class="mt-4 flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="testingS3" @click="testS3">
            {{ testingS3 ? t('common.loading') : t('admin.backup.s3.testConnection') }}
          </button>
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingS3" @click="saveS3Config">
            {{ savingS3 ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Schedule Config -->
      <div class="card p-6">
        <div class="mb-4">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">
            {{ t('admin.backup.schedule.title') }}
          </h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.backup.schedule.description') }}
          </p>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 md:col-span-2">
            <input v-model="scheduleForm.enabled" type="checkbox" />
            <span>{{ t('admin.backup.schedule.enabled') }}</span>
          </label>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.schedule.cronExpr') }}</label>
            <input v-model="scheduleForm.cron_expr" class="input w-full" placeholder="0 2 * * *" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.backup.schedule.cronHint') }}</p>
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.schedule.retainDays') }}</label>
            <input v-model.number="scheduleForm.retain_days" type="number" min="0" class="input w-full" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.backup.schedule.retainDaysHint') }}</p>
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('admin.backup.schedule.retainCount') }}</label>
            <input v-model.number="scheduleForm.retain_count" type="number" min="0" class="input w-full" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.backup.schedule.retainCountHint') }}</p>
          </div>
        </div>
        <div class="mt-4">
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingSchedule" @click="saveSchedule">
            {{ savingSchedule ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Backup Operations -->
      <div class="card p-6">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.backup.operations.title') }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.backup.operations.description') }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <div class="flex items-center gap-1">
              <label class="text-xs text-gray-600 dark:text-gray-400">{{ t('admin.backup.operations.expireDays') }}</label>
              <input v-model.number="manualExpireDays" type="number" min="0" class="input w-20 text-xs" />
            </div>
            <button type="button" class="btn btn-primary btn-sm" :disabled="creatingBackup" @click="createBackup">
              {{ creatingBackup ? t('admin.backup.operations.backing') : t('admin.backup.operations.createBackup') }}
            </button>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingBackups" @click="loadBackups">
              {{ loadingBackups ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full min-w-[800px] text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
                <th class="py-2 pr-4">ID</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.status') }}</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.fileName') }}</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.size') }}</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.expiresAt') }}</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.triggeredBy') }}</th>
                <th class="py-2 pr-4">{{ t('admin.backup.columns.startedAt') }}</th>
                <th class="py-2">{{ t('admin.backup.columns.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in backups" :key="record.id" class="border-b border-gray-100 align-top dark:border-dark-800">
                <td class="py-3 pr-4 font-mono text-xs">{{ record.id }}</td>
                <td class="py-3 pr-4">
                  <span
                    class="rounded px-2 py-0.5 text-xs"
                    :class="statusClass(record.status)"
                  >
                    {{ record.status === 'running' && record.progress
                      ? t(`admin.backup.progress.${record.progress}`)
                      : t(`admin.backup.status.${record.status}`) }}
                  </span>
                </td>
                <td class="py-3 pr-4 text-xs">{{ record.file_name }}</td>
                <td class="py-3 pr-4 text-xs">{{ formatSize(record.size_bytes) }}</td>
                <td class="py-3 pr-4 text-xs">
                  {{ record.expires_at ? formatDate(record.expires_at) : t('admin.backup.neverExpire') }}
                </td>
                <td class="py-3 pr-4 text-xs">
                  {{ record.triggered_by === 'scheduled' ? t('admin.backup.trigger.scheduled') : t('admin.backup.trigger.manual') }}
                </td>
                <td class="py-3 pr-4 text-xs">{{ formatDate(record.started_at) }}</td>
                <td class="py-3 text-xs">
                  <div class="flex flex-wrap gap-1">
                    <button
                      v-if="record.status === 'completed'"
                      type="button"
                      class="btn btn-secondary btn-xs"
                      @click="downloadBackup(record.id)"
                    >
                      {{ t('admin.backup.actions.download') }}
                    </button>
                    <button
                      v-if="record.status === 'completed'"
                      type="button"
                      class="btn btn-secondary btn-xs"
                      :disabled="restoringId === record.id"
                      @click="restoreBackup(record.id)"
                    >
                      {{ restoringId === record.id ? t('common.loading') : t('admin.backup.actions.restore') }}
                    </button>
                    <button
                      type="button"
                      class="btn btn-danger btn-xs"
                      @click="removeBackup(record.id)"
                    >
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="backups.length === 0">
                <td colspan="8" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.backup.empty') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Cloudflare R2 Setup Guide Modal -->
    <teleport to="body">
      <transition name="modal">
        <div v-if="showR2Guide" class="fixed inset-0 z-50 flex items-center justify-center p-4" @mousedown.self="showR2Guide = false">
          <div class="fixed inset-0 bg-black/50" @click="showR2Guide = false"></div>
          <div class="relative max-h-[85vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white p-6 shadow-2xl dark:bg-dark-800">
            <button type="button" class="absolute right-4 top-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200" @click="showR2Guide = false">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>

            <h2 class="mb-4 text-lg font-bold text-gray-900 dark:text-white">{{ t('admin.backup.r2Guide.title') }}</h2>
            <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.backup.r2Guide.intro') }}</p>

            <!-- Step 1 -->
            <div class="mb-5">
              <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
                <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">1</span>
                {{ t('admin.backup.r2Guide.step1.title') }}
              </h3>
              <ol class="ml-8 list-decimal space-y-1 text-sm text-gray-600 dark:text-gray-300">
                <li>{{ t('admin.backup.r2Guide.step1.line1') }}</li>
                <li>{{ t('admin.backup.r2Guide.step1.line2') }}</li>
                <li>{{ t('admin.backup.r2Guide.step1.line3') }}</li>
              </ol>
            </div>

            <!-- Step 2 -->
            <div class="mb-5">
              <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
                <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">2</span>
                {{ t('admin.backup.r2Guide.step2.title') }}
              </h3>
              <ol class="ml-8 list-decimal space-y-1 text-sm text-gray-600 dark:text-gray-300">
                <li>{{ t('admin.backup.r2Guide.step2.line1') }}</li>
                <li>{{ t('admin.backup.r2Guide.step2.line2') }}</li>
                <li>{{ t('admin.backup.r2Guide.step2.line3') }}</li>
                <li>{{ t('admin.backup.r2Guide.step2.line4') }}</li>
              </ol>
              <div class="mt-2 rounded-lg bg-amber-50 p-3 text-xs text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">
                {{ t('admin.backup.r2Guide.step2.warning') }}
              </div>
            </div>

            <!-- Step 3 -->
            <div class="mb-5">
              <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
                <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">3</span>
                {{ t('admin.backup.r2Guide.step3.title') }}
              </h3>
              <p class="ml-8 text-sm text-gray-600 dark:text-gray-300">{{ t('admin.backup.r2Guide.step3.desc') }}</p>
              <code class="ml-8 mt-1 block rounded bg-gray-100 px-3 py-2 text-xs text-gray-800 dark:bg-dark-700 dark:text-gray-200">https://&lt;{{ t('admin.backup.r2Guide.step3.accountId') }}&gt;.r2.cloudflarestorage.com</code>
            </div>

            <!-- Step 4: Fill form -->
            <div class="mb-5">
              <h3 class="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
                <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">4</span>
                {{ t('admin.backup.r2Guide.step4.title') }}
              </h3>
              <div class="ml-8 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
                <table class="w-full text-sm">
                  <tbody>
                    <tr v-for="(row, i) in r2ConfigRows" :key="i" class="border-b border-gray-100 dark:border-dark-700 last:border-0">
                      <td class="whitespace-nowrap bg-gray-50 px-3 py-2 font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">{{ row.field }}</td>
                      <td class="px-3 py-2 text-gray-600 dark:text-gray-400"><code class="text-xs">{{ row.value }}</code></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Free tier note -->
            <div class="rounded-lg bg-green-50 p-3 text-xs text-green-700 dark:bg-green-900/20 dark:text-green-300">
              {{ t('admin.backup.r2Guide.freeTier') }}
            </div>

            <div class="mt-4 text-right">
              <button type="button" class="btn btn-primary btn-sm" @click="showR2Guide = false">{{ t('common.close') }}</button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { BackupS3Config, BackupScheduleConfig, BackupRecord } from '@/api/admin/backup'

const { t } = useI18n()
const appStore = useAppStore()

// S3 config
const s3Form = ref<BackupS3Config>({
  provider: 'r2',
  endpoint: '',
  region: 'auto',
  bucket: '',
  access_key_id: '',
  secret_access_key: '',
  prefix: 'backups/',
  force_path_style: false,
  local_path: '',
  public_base_url: '',
})
const s3SecretConfigured = ref(false)
const savingS3 = ref(false)
const testingS3 = ref(false)

// Schedule config
const scheduleForm = ref<BackupScheduleConfig>({
  enabled: false,
  cron_expr: '0 2 * * *',
  retain_days: 14,
  retain_count: 10,
})
const savingSchedule = ref(false)

// Backups
const backups = ref<BackupRecord[]>([])
const loadingBackups = ref(false)
const creatingBackup = ref(false)
const restoringId = ref('')
const manualExpireDays = ref(14)

// Polling
const pollingTimer = ref<ReturnType<typeof setInterval> | null>(null)
const restoringPollingTimer = ref<ReturnType<typeof setInterval> | null>(null)
const MAX_POLL_COUNT = 900

function updateRecordInList(updated: BackupRecord) {
  const idx = backups.value.findIndex(r => r.id === updated.id)
  if (idx >= 0) {
    backups.value[idx] = updated
  }
}

function startPolling(backupId: string) {
  stopPolling()
  let count = 0
  pollingTimer.value = setInterval(async () => {
    if (count++ >= MAX_POLL_COUNT) {
      stopPolling()
      creatingBackup.value = false
      appStore.showWarning(t('admin.backup.operations.backupRunning'))
      return
    }
    try {
      const record = await adminAPI.backup.getBackup(backupId)
      updateRecordInList(record)
      if (record.status === 'completed' || record.status === 'failed') {
        stopPolling()
        creatingBackup.value = false
        if (record.status === 'completed') {
          appStore.showSuccess(t('admin.backup.operations.backupCreated'))
        } else {
          appStore.showError(record.error_message || t('admin.backup.operations.backupFailed'))
        }
        await loadBackups()
      }
    } catch {
      // 轮询失败时不中断
    }
  }, 2000)
}

function stopPolling() {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
    pollingTimer.value = null
  }
}

function startRestorePolling(backupId: string) {
  stopRestorePolling()
  let count = 0
  restoringPollingTimer.value = setInterval(async () => {
    if (count++ >= MAX_POLL_COUNT) {
      stopRestorePolling()
      restoringId.value = ''
      appStore.showWarning(t('admin.backup.operations.restoreRunning'))
      return
    }
    try {
      const record = await adminAPI.backup.getBackup(backupId)
      updateRecordInList(record)
      if (record.restore_status === 'completed' || record.restore_status === 'failed') {
        stopRestorePolling()
        restoringId.value = ''
        if (record.restore_status === 'completed') {
          appStore.showSuccess(t('admin.backup.actions.restoreSuccess'))
        } else {
          appStore.showError(record.restore_error || t('admin.backup.operations.restoreFailed'))
        }
        await loadBackups()
      }
    } catch {
      // 轮询失败时不中断
    }
  }, 2000)
}

function stopRestorePolling() {
  if (restoringPollingTimer.value) {
    clearInterval(restoringPollingTimer.value)
    restoringPollingTimer.value = null
  }
}

function handleVisibilityChange() {
  if (document.hidden) {
    stopPolling()
    stopRestorePolling()
  } else {
    // 标签页恢复时刷新列表，检查是否仍有活跃操作
    loadBackups().then(() => {
      const running = backups.value.find(r => r.status === 'running')
      if (running) {
        creatingBackup.value = true
        startPolling(running.id)
      }
      const restoring = backups.value.find(r => r.restore_status === 'running')
      if (restoring) {
        restoringId.value = restoring.id
        startRestorePolling(restoring.id)
      }
    })
  }
}

// R2 guide
const showR2Guide = ref(false)
const r2ConfigRows = computed(() => [
  { field: t('admin.backup.s3.endpoint'), value: 'https://<account_id>.r2.cloudflarestorage.com' },
  { field: t('admin.backup.s3.region'), value: 'auto' },
  { field: t('admin.backup.s3.bucket'), value: t('admin.backup.r2Guide.step4.bucketValue') },
  { field: t('admin.backup.s3.prefix'), value: 'backups/' },
  { field: 'Access Key ID', value: t('admin.backup.r2Guide.step4.fromStep2') },
  { field: 'Secret Access Key', value: t('admin.backup.r2Guide.step4.fromStep2') },
  { field: t('admin.backup.s3.forcePathStyle'), value: t('admin.backup.r2Guide.step4.unchecked') },
])

const endpointPlaceholder = computed(() => {
  if (s3Form.value.provider === 'oss') return 'https://oss-cn-hangzhou.aliyuncs.com'
  if (s3Form.value.provider === 'r2') return 'https://<account_id>.r2.cloudflarestorage.com'
  return 'https://s3.example.com'
})

const regionPlaceholder = computed(() => {
  if (s3Form.value.provider === 'oss') return 'oss-cn-hangzhou'
  if (s3Form.value.provider === 'r2') return 'auto'
  return 'auto'
})

function backupMessage(errorOrMessage: unknown, fallbackKey: string): string {
  const err = errorOrMessage as {
    status?: number
    code?: string | number
    reason?: string
    message?: string
    error?: string
    response?: { status?: number; data?: { code?: string | number; message?: string; detail?: string } }
  } | null
  const code = String(err?.reason || err?.code || err?.response?.data?.code || '')
  const raw = typeof errorOrMessage === 'string'
    ? errorOrMessage
    : String(err?.message || err?.error || err?.response?.data?.message || err?.response?.data?.detail || '')

  const codeKeyMap: Record<string, string> = {
    BACKUP_S3_NOT_CONFIGURED: 'admin.backup.errors.storageNotConfigured',
    BACKUP_NOT_FOUND: 'admin.backup.errors.notFound',
    BACKUP_IN_PROGRESS: 'admin.backup.operations.alreadyInProgress',
    RESTORE_IN_PROGRESS: 'admin.backup.operations.restoreRunning',
    BACKUP_RECORDS_CORRUPT: 'admin.backup.errors.recordsCorrupt',
    BACKUP_S3_CONFIG_CORRUPT: 'admin.backup.errors.storageConfigCorrupt',
    BACKUP_NOT_COMPLETED: 'admin.backup.errors.notCompleted',
    INVALID_CRON: 'admin.backup.errors.invalidCron',
  }
  if (code && codeKeyMap[code]) return t(codeKeyMap[code])

  if (raw.includes('incomplete local storage config') || raw.includes('local_path is required')) {
    return t('admin.backup.s3.errors.localPathRequired')
  }
  if (raw.includes('incomplete object storage config') || raw.includes('incomplete S3 config')) {
    return t('admin.backup.s3.errors.objectStorageRequired')
  }
  if (raw.includes('connection successful')) {
    return t('admin.backup.s3.testSuccess')
  }
  if (raw.includes('cron expression is required')) {
    return t('admin.backup.errors.cronRequired')
  }
  if (raw.includes('invalid cron expression') || raw.includes('failed to schedule')) {
    return t('admin.backup.errors.invalidCron')
  }
  if (raw.includes('backup storage is not configured')) {
    return t('admin.backup.errors.storageNotConfigured')
  }
  if (raw.includes('backup record not found') || raw.includes('backup is not completed')) {
    return t('admin.backup.errors.notFound')
  }
  if (raw.includes('can only restore from a completed backup')) {
    return t('admin.backup.errors.notCompleted')
  }
  if (raw.includes('incorrect admin password')) {
    return t('admin.backup.errors.incorrectPassword')
  }
  if (raw.includes('password is required')) {
    return t('admin.backup.errors.passwordRequired')
  }
  if (raw.includes('a backup is already in progress')) {
    return t('admin.backup.operations.alreadyInProgress')
  }
  if (raw.includes('a restore is already in progress')) {
    return t('admin.backup.operations.restoreRunning')
  }
  if (raw.includes('Network error')) {
    return t('errors.networkError')
  }
  return raw || t(fallbackKey)
}

function applyProviderDefaults() {
  if (s3Form.value.provider === 'local') {
    s3Form.value.endpoint = ''
    s3Form.value.region = ''
    s3Form.value.bucket = ''
    s3Form.value.access_key_id = ''
    s3Form.value.secret_access_key = ''
    s3Form.value.force_path_style = false
    if (!s3Form.value.local_path) s3Form.value.local_path = './data/storage'
    return
  }
  if (s3Form.value.provider === 'r2') {
    if (!s3Form.value.region) s3Form.value.region = 'auto'
    s3Form.value.force_path_style = false
    return
  }
  if (s3Form.value.provider === 'oss') {
    if (!s3Form.value.region || s3Form.value.region === 'auto') s3Form.value.region = 'oss-cn-hangzhou'
    s3Form.value.force_path_style = true
    return
  }
  if (!s3Form.value.region) s3Form.value.region = 'auto'
}

function buildStoragePayload(): BackupS3Config {
  const provider = s3Form.value.provider || 'r2'
  const prefix = s3Form.value.prefix || 'backups/'
  if (provider === 'local') {
    return {
      provider: 'local',
      endpoint: '',
      region: '',
      bucket: '',
      access_key_id: '',
      secret_access_key: '',
      prefix,
      force_path_style: false,
      local_path: s3Form.value.local_path || './data/storage',
      public_base_url: s3Form.value.public_base_url || '',
    }
  }
  return {
    provider,
    endpoint: s3Form.value.endpoint || '',
    region: s3Form.value.region || 'auto',
    bucket: s3Form.value.bucket || '',
    access_key_id: s3Form.value.access_key_id || '',
    secret_access_key: s3Form.value.secret_access_key || '',
    prefix,
    force_path_style: Boolean(s3Form.value.force_path_style),
    local_path: '',
    public_base_url: s3Form.value.public_base_url || '',
  }
}

async function loadS3Config() {
  try {
    const cfg = await adminAPI.backup.getS3Config()
    const provider = cfg.provider || (cfg.local_path ? 'local' : 'r2')
    s3Form.value = {
      provider,
      endpoint: cfg.endpoint || '',
      region: cfg.region || 'auto',
      bucket: cfg.bucket || '',
      access_key_id: cfg.access_key_id || '',
      secret_access_key: '',
      prefix: cfg.prefix || 'backups/',
      force_path_style: cfg.force_path_style,
      local_path: cfg.local_path || '',
      public_base_url: cfg.public_base_url || '',
    }
    applyProviderDefaults()
    s3SecretConfigured.value = Boolean(cfg.access_key_id)
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  }
}

async function saveS3Config() {
  savingS3.value = true
  try {
    applyProviderDefaults()
    const payload = buildStoragePayload()
    const saved = await adminAPI.backup.updateS3Config(payload)
    s3Form.value = {
      provider: saved.provider || payload.provider,
      endpoint: saved.endpoint || payload.endpoint,
      region: saved.region || payload.region,
      bucket: saved.bucket || payload.bucket,
      access_key_id: saved.access_key_id || payload.access_key_id,
      secret_access_key: '',
      prefix: saved.prefix || payload.prefix,
      force_path_style: saved.force_path_style ?? payload.force_path_style,
      local_path: saved.local_path || payload.local_path || '',
      public_base_url: saved.public_base_url || payload.public_base_url || '',
    }
    applyProviderDefaults()
    s3SecretConfigured.value = s3Form.value.provider !== 'local' && Boolean(s3Form.value.access_key_id)
    appStore.showSuccess(t('admin.backup.s3.saved'))
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  } finally {
    savingS3.value = false
  }
}

async function testS3() {
  testingS3.value = true
  try {
    applyProviderDefaults()
    const result = await adminAPI.backup.testS3Connection(buildStoragePayload())
    if (result.ok) {
      appStore.showSuccess(backupMessage(result.message, 'admin.backup.s3.testSuccess'))
    } else {
      appStore.showError(backupMessage(result.message, 'admin.backup.s3.testFailed'))
    }
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  } finally {
    testingS3.value = false
  }
}

async function loadSchedule() {
  try {
    const cfg = await adminAPI.backup.getSchedule()
    scheduleForm.value = {
      enabled: cfg.enabled,
      cron_expr: cfg.cron_expr || '0 2 * * *',
      retain_days: cfg.retain_days || 14,
      retain_count: cfg.retain_count || 10,
    }
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  }
}

async function saveSchedule() {
  savingSchedule.value = true
  try {
    await adminAPI.backup.updateSchedule(scheduleForm.value)
    appStore.showSuccess(t('admin.backup.schedule.saved'))
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  } finally {
    savingSchedule.value = false
  }
}

async function loadBackups() {
  loadingBackups.value = true
  try {
    const result = await adminAPI.backup.listBackups()
    backups.value = result.items || []
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  } finally {
    loadingBackups.value = false
  }
}

async function createBackup() {
  creatingBackup.value = true
  try {
    const record = await adminAPI.backup.createBackup({ expire_days: manualExpireDays.value })
    // 插入到列表顶部
    backups.value.unshift(record)
    startPolling(record.id)
  } catch (error: any) {
    if (error?.status === 409 || error?.response?.status === 409) {
      appStore.showWarning(t('admin.backup.operations.alreadyInProgress'))
    } else {
      appStore.showError(backupMessage(error, 'errors.networkError'))
    }
    creatingBackup.value = false
  }
}

async function downloadBackup(id: string) {
  try {
    const result = await adminAPI.backup.getDownloadURL(id)
    window.open(result.url, '_blank')
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  }
}

async function restoreBackup(id: string) {
  if (!window.confirm(t('admin.backup.actions.restoreConfirm'))) return
  const password = window.prompt(t('admin.backup.actions.restorePasswordPrompt'))
  if (!password) return
  restoringId.value = id
  try {
    const record = await adminAPI.backup.restoreBackup(id, password)
    updateRecordInList(record)
    startRestorePolling(id)
  } catch (error: any) {
    if (error?.status === 409 || error?.response?.status === 409) {
      appStore.showWarning(t('admin.backup.operations.restoreRunning'))
    } else {
      appStore.showError(backupMessage(error, 'errors.networkError'))
    }
    restoringId.value = ''
  }
}

async function removeBackup(id: string) {
  if (!window.confirm(t('admin.backup.actions.deleteConfirm'))) return
  try {
    await adminAPI.backup.deleteBackup(id)
    appStore.showSuccess(t('admin.backup.actions.deleted'))
    await loadBackups()
  } catch (error) {
    appStore.showError(backupMessage(error, 'errors.networkError'))
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
    case 'running':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
    case 'failed':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'
  }
}

function formatSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

onMounted(async () => {
  document.addEventListener('visibilitychange', handleVisibilityChange)
  await Promise.all([loadS3Config(), loadSchedule(), loadBackups()])

  // 如果有正在 running 的备份，恢复轮询
  const runningBackup = backups.value.find(r => r.status === 'running')
  if (runningBackup) {
    creatingBackup.value = true
    startPolling(runningBackup.id)
  }
  const restoringBackup = backups.value.find(r => r.restore_status === 'running')
  if (restoringBackup) {
    restoringId.value = restoringBackup.id
    startRestorePolling(restoringBackup.id)
  }
})

onBeforeUnmount(() => {
  stopPolling()
  stopRestorePolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
