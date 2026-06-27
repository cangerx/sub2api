<template>
  <AuthLayout>
    <div class="auth-form-flow space-y-5">
      <!-- Title / Single Tab -->
      <div class="auth-step mb-6 text-center">
        <h1 class="text-[1.45rem] font-semibold tracking-tight text-gray-950 dark:text-white">欢迎回来</h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-zinc-400">登录账户，继续管理你的 API 网关</p>
      </div>
      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-5">
        <!-- Email Input -->
        <div class="auth-step">
          <label for="email" class="auth-label">邮箱</label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="authActionDisabled"
              class="input auth-input pl-11"
              :class="{ 'input-error': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
        </div>

        <!-- Password Input -->
        <div class="auth-step">
          <label for="password" class="auth-label">密码</label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="lock" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="authActionDisabled"
              class="input auth-input pl-11 pr-11"
              :class="{ 'input-error': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              :disabled="authActionDisabled"
              class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-dark-300"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
          <div class="mt-4 flex items-center justify-between">
            <!-- Left side empty to push Forgot Password to the right, or we can add a visual-only Remember me -->
            <label class="flex items-center gap-2 cursor-pointer">
               <input type="checkbox" checked class="rounded border-gray-300 text-blue-600 focus:ring-blue-600" />
               <span class="text-xs text-gray-500 dark:text-gray-400">14天内免登录</span>
            </label>
            <router-link
              v-if="passwordResetEnabled && !backendModeEnabled"
              to="/forgot-password"
              class="text-xs font-medium text-gray-500 transition-colors hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>
        </div>

        <!-- Turnstile Widget -->
        <div v-if="turnstileEnabled && turnstileSiteKey" class="auth-step">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="authActionDisabled || (turnstileEnabled && !turnstileToken)"
          class="auth-submit auth-step w-full rounded-xl bg-[#4f63f6] text-white py-3 font-medium hover:bg-[#4355df] transition-colors focus:ring-4 focus:ring-blue-500/20 flex justify-center items-center mt-6"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}
        </button>

        <LoginAgreementPrompt
          v-if="loginAgreementEnabled"
          :accepted="agreementAccepted"
          :documents="loginAgreementDocuments"
          :mode="loginAgreementMode"
          :updated-at="loginAgreementUpdatedAt"
          :visible="showAgreementModal"
          @accept="acceptLoginAgreement"
          @reject="rejectLoginAgreement"
          @open="showAgreementModal = true"
        />

        <div v-if="showOAuthLogin" class="auth-step space-y-4 pt-4">
          <div class="flex items-center gap-3">
            <div class="h-px flex-1 bg-gray-200 dark:bg-zinc-800"></div>
            <span class="text-xs text-gray-400 dark:text-gray-500">
              {{ t('auth.oauthOrContinue') }}
            </span>
            <div class="h-px flex-1 bg-gray-200 dark:bg-zinc-800"></div>
          </div>

          <EmailOAuthButtons
            :disabled="authActionDisabled"
            :github-enabled="githubOAuthEnabled"
            :google-enabled="googleOAuthEnabled"
            :show-divider="false"
          />

          <LinuxDoOAuthSection
            v-if="linuxdoOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <DingTalkOAuthSection
            v-if="dingtalkOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <WechatOAuthSection
            v-if="wechatOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <OidcOAuthSection
            v-if="oidcOAuthEnabled"
            :disabled="authActionDisabled"
            :provider-name="oidcOAuthProviderName"
            :show-divider="false"
          />
        </div>
      </form>
    </div>

    <!-- Footer -->
    <template v-if="!backendModeEnabled" #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>

  <!-- 2FA Modal -->
  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import DingTalkOAuthSection from '@/components/auth/DingTalkOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, isTotp2FARequired, isWeChatWebOAuthEnabled } from '@/api/auth'
import type { LoginAgreementDocument, TotpLoginResponse } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'

const { t } = useI18n()
const LOGIN_AGREEMENT_STORAGE_KEY = 'ccapi_login_agreement_consent'

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const publicSettingsLoaded = ref<boolean>(false)

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const dingtalkOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const githubOAuthEnabled = ref<boolean>(false)
const googleOAuthEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)
const loginAgreementEnabled = ref<boolean>(false)
const loginAgreementMode = ref<'modal' | 'checkbox' | string>('modal')
const loginAgreementUpdatedAt = ref<string>('')
const loginAgreementRevision = ref<string>('')
const loginAgreementDocuments = ref<LoginAgreementDocument[]>([])
const agreementAccepted = ref<boolean>(false)
const showAgreementModal = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

// 2FA state
const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

const validationToastMessage = computed(
  () => errors.email || errors.password || errors.turnstile || ''
)

const agreementGateActive = computed(
  () => loginAgreementEnabled.value && !agreementAccepted.value
)

const authActionDisabled = computed(
  () => isLoading.value || !publicSettingsLoaded.value || agreementGateActive.value
)

const showOAuthLogin = computed(
  () =>
    !backendModeEnabled.value &&
    (linuxdoOAuthEnabled.value ||
      dingtalkOAuthEnabled.value ||
      wechatOAuthEnabled.value ||
      oidcOAuthEnabled.value ||
      githubOAuthEnabled.value ||
      googleOAuthEnabled.value)
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    backendModeEnabled.value = settings.backend_mode_enabled
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    githubOAuthEnabled.value = settings.github_oauth_enabled
    googleOAuthEnabled.value = settings.google_oauth_enabled
    backendModeEnabled.value = settings.backend_mode_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
    applyLoginAgreementSettings(settings)
  } catch (error) {
    console.error('Failed to load public settings:', error)
    loginAgreementEnabled.value = false
    agreementAccepted.value = true
  } finally {
    publicSettingsLoaded.value = true
  }
})

// ==================== Login Agreement ====================

function applyLoginAgreementSettings(settings: {
  login_agreement_enabled?: boolean
  login_agreement_mode?: string
  login_agreement_updated_at?: string
  login_agreement_revision?: string
  login_agreement_documents?: LoginAgreementDocument[]
}): void {
  const documents = Array.isArray(settings.login_agreement_documents)
    ? settings.login_agreement_documents.filter((doc) => doc.title?.trim())
    : []
  loginAgreementDocuments.value = documents
  loginAgreementEnabled.value = settings.login_agreement_enabled === true && documents.length > 0
  loginAgreementMode.value = settings.login_agreement_mode === 'checkbox' ? 'checkbox' : 'modal'
  loginAgreementUpdatedAt.value = settings.login_agreement_updated_at || ''
  loginAgreementRevision.value =
    settings.login_agreement_revision ||
    `${loginAgreementUpdatedAt.value}:${documents.map((doc) => `${doc.id}:${doc.title}`).join('|')}`

  agreementAccepted.value = !loginAgreementEnabled.value || hasAcceptedLoginAgreement(loginAgreementRevision.value)
  showAgreementModal.value =
    loginAgreementEnabled.value && !agreementAccepted.value && loginAgreementMode.value !== 'checkbox'
}

function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return false
    }
    const parsed = JSON.parse(raw) as { revision?: string }
    return parsed.revision === revision
  } catch {
    return false
  }
}

function acceptLoginAgreement(): void {
  if (loginAgreementRevision.value) {
    localStorage.setItem(
      LOGIN_AGREEMENT_STORAGE_KEY,
      JSON.stringify({
        revision: loginAgreementRevision.value,
        accepted_at: new Date().toISOString()
      })
    )
  }
  agreementAccepted.value = true
  showAgreementModal.value = false
}

function rejectLoginAgreement(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
  agreementAccepted.value = false
  showAgreementModal.value = false
  appStore.showWarning('未同意最新条款前，无法输入账号密码或使用快捷登录。')
}

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

// ==================== Validation ====================

function validateForm(): boolean {
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (agreementGateActive.value) {
    appStore.showWarning('请先阅读并同意最新条款后再登录。')
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return false
  }

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleLogin(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    // Call auth store login
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    // Check if 2FA is required
    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    // Show success toast
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    // Reset Turnstile on error
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', t('auth.loginFailed'))

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

// ==================== 2FA Handlers ====================

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    // Close modal and show success
    show2FAModal.value = false
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
}
</script>

<style scoped>
.auth-step {
  opacity: 0;
  transform: translate3d(0, 14px, 0);
  animation: authStepEnter 620ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.auth-step:nth-child(1) { animation-delay: 80ms; }
.auth-step:nth-child(2) { animation-delay: 150ms; }
.auth-step:nth-child(3) { animation-delay: 220ms; }
.auth-step:nth-child(4) { animation-delay: 290ms; }
.auth-step:nth-child(5) { animation-delay: 360ms; }

.auth-label {
  display: block;
  margin-bottom: 0.45rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: rgb(31 41 55);
}

.dark .auth-label {
  color: rgb(228 228 231);
}

.auth-input {
  height: 2.9rem;
  border-radius: 0.8rem;
  border-color: rgb(229 231 235);
  background: rgba(249, 250, 251, 0.72);
  color: rgb(17 24 39);
}

.auth-input:focus {
  border-color: rgba(79, 99, 246, 0.58);
  background: rgba(255, 255, 255, 0.98);
}

.dark .auth-input {
  border-color: rgba(63, 63, 70, 0.9);
  background: rgba(24, 24, 27, 0.68);
  color: rgb(244 244 245);
}

.dark .auth-input:focus {
  border-color: rgba(129, 140, 248, 0.72);
  background: rgba(24, 24, 27, 0.9);
}

.auth-form-flow :deep(.input) {
  transition:
    background-color 180ms ease,
    border-color 180ms ease,
    box-shadow 180ms ease,
    transform 180ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

.auth-form-flow :deep(.input:focus) {
  transform: translateY(-1px);
  box-shadow:
    0 0 0 4px rgba(59, 130, 246, 0.12),
    0 10px 24px rgba(15, 23, 42, 0.08);
}

.auth-submit {
  position: relative;
  overflow: hidden;
  box-shadow: 0 14px 30px rgba(79, 99, 246, 0.28);
  transition:
    transform 180ms cubic-bezier(0.2, 0.8, 0.2, 1),
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.auth-submit::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(110deg, transparent 0%, rgba(255, 255, 255, 0.24) 42%, transparent 66%);
  transform: translateX(-120%);
  transition: transform 520ms cubic-bezier(0.16, 1, 0.3, 1);
}

.auth-submit:hover {
  transform: translateY(-1px);
  box-shadow: 0 18px 38px rgba(79, 99, 246, 0.34);
}

.auth-submit:hover::after {
  transform: translateX(120%);
}

.auth-submit:active {
  transform: scale(0.985);
}

@keyframes authStepEnter {
  to {
    opacity: 1;
    transform: translate3d(0, 0, 0);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@media (prefers-reduced-motion: reduce) {
  .auth-step {
    opacity: 1;
    transform: none;
    animation: none;
  }
}

@media (max-width: 640px) {
  .auth-form-flow {
    gap: 1rem;
  }

  .auth-form-flow h1 {
    font-size: 1.28rem;
  }

  .auth-form-flow h1 + p {
    font-size: 0.8rem;
    line-height: 1.55;
  }

  .auth-input {
    height: 2.75rem;
    border-radius: 0.75rem;
    font-size: 0.875rem;
  }

  .auth-submit {
    min-height: 2.75rem;
    border-radius: 0.85rem;
  }
}

@media (max-width: 380px) {
  .auth-step .mt-4.flex {
    align-items: flex-start;
    gap: 0.75rem;
  }

  .auth-step .mt-4.flex label {
    min-width: 0;
  }
}
</style>
