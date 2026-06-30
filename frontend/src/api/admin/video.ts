import { apiClient } from '../client'

export interface VideoTemplate {
  id: number
  name: string
  create_method: string
  create_path: string
  query_method: string
  query_path: string
  content_method?: string | null
  content_path?: string | null
  cancel_method?: string | null
  cancel_path?: string | null
  status_mapping: Record<string, string>
  result_mapping: Record<string, string>
  error_mapping: Record<string, string>
  poll_config: Record<string, unknown>
  timeout_config: Record<string, unknown>
  status: string
  created_at: string
  updated_at: string
}

export interface VideoModel {
  id: number
  public_model: string
  display_name?: string | null
  template_id: number
  upstream_model_id: string
  request_shape: string
  status: string
  capabilities: Record<string, unknown>
  defaults: Record<string, unknown>
  limits: Record<string, unknown>
  supported_options: Record<string, unknown>
  extra_body_allow: string[]
  sort_order: number
  template?: VideoTemplate | null
  created_at: string
  updated_at: string
}

export interface VideoTask {
  id: number
  public_id: string
  user_id: number
  api_key_id: number
  account_id: number
  group_id?: number | null
  channel_id?: number | null
  video_model_id: number
  requested_model: string
  upstream_model: string
  upstream_task_id?: string | null
  status: string
  progress: number
  billing_state: string
  request_payload?: Record<string, unknown>
  upstream_request_payload?: Record<string, unknown>
  upstream_response_payload?: Record<string, unknown>
  result_payload?: Record<string, unknown>
  content_url?: string | null
  upstream_content_url?: string | null
  local_content_url?: string | null
  billing_mode?: string
  unit_price?: number
  unit_seconds?: number | null
  requested_seconds?: number | null
  billable_seconds?: number | null
  reserved_cost: number
  estimated_cost: number
  actual_cost: number
  submitted_at?: string | null
  started_at?: string | null
  next_poll_at?: string | null
  locked_until?: string | null
  poll_count: number
  error_code?: string | null
  error_message?: string | null
  created_at: string
  completed_at?: string | null
  expires_at?: string | null
  video_model?: VideoModel | null
}

export interface VideoTemplateCreateTestPayload {
  account_id: number
  template_id: number
  body: Record<string, unknown>
}

export interface VideoTemplateQueryTestPayload {
  account_id: number
  template_id: number
  upstream_task_id: string
}

export interface VideoTemplateCreateTestResult {
  task_id: string
  response: Record<string, unknown>
}

export interface VideoTemplateQueryTestResult {
  status: string
  progress: number
  content_url?: string | null
  seconds?: number | null
  response: Record<string, unknown>
  error_code?: string
  error_msg?: string
}

export interface VideoTemplateRecognizePayload {
  account_id: number
  model: string
  document: string
}

export type VideoTemplatePayload = Omit<VideoTemplate, 'id' | 'created_at' | 'updated_at'>
export type VideoModelPayload = Omit<VideoModel, 'id' | 'template' | 'created_at' | 'updated_at'>

export async function listTemplates(): Promise<{ items: VideoTemplate[] }> {
  const { data } = await apiClient.get<{ items: VideoTemplate[] }>('/admin/video/templates')
  return data
}

export async function createTemplate(payload: VideoTemplatePayload): Promise<VideoTemplate> {
  const { data } = await apiClient.post<VideoTemplate>('/admin/video/templates', payload)
  return data
}

export async function updateTemplate(id: number, payload: VideoTemplatePayload): Promise<VideoTemplate> {
  const { data } = await apiClient.put<VideoTemplate>(`/admin/video/templates/${id}`, payload)
  return data
}

export async function deleteTemplate(id: number): Promise<void> {
  await apiClient.delete(`/admin/video/templates/${id}`)
}

export async function testTemplateCreate(payload: VideoTemplateCreateTestPayload): Promise<VideoTemplateCreateTestResult> {
  const { data } = await apiClient.post<VideoTemplateCreateTestResult>('/admin/video/templates/test-create', payload)
  return data
}

export async function testTemplateQuery(payload: VideoTemplateQueryTestPayload): Promise<VideoTemplateQueryTestResult> {
  const { data } = await apiClient.post<VideoTemplateQueryTestResult>('/admin/video/templates/test-query', payload)
  return data
}

// recognizeTemplate asks a video account's OpenAI-compatible chat endpoint to
// parse pasted upstream docs into a draft template (id is 0; not persisted
// until the admin saves it).
export async function recognizeTemplate(payload: VideoTemplateRecognizePayload): Promise<VideoTemplate> {
  const { data } = await apiClient.post<VideoTemplate>('/admin/video/templates/recognize', payload)
  return data
}

export async function listModels(): Promise<{ items: VideoModel[] }> {
  const { data } = await apiClient.get<{ items: VideoModel[] }>('/admin/video/models')
  return data
}

export async function listRequestShapes(): Promise<{ items: string[] }> {
  const { data } = await apiClient.get<{ items: string[] }>('/admin/video/request-shapes')
  return data
}

export async function createModel(payload: VideoModelPayload): Promise<VideoModel> {
  const { data } = await apiClient.post<VideoModel>('/admin/video/models', payload)
  return data
}

export async function updateModel(id: number, payload: VideoModelPayload): Promise<VideoModel> {
  const { data } = await apiClient.put<VideoModel>(`/admin/video/models/${id}`, payload)
  return data
}

export async function deleteModel(id: number): Promise<void> {
  await apiClient.delete(`/admin/video/models/${id}`)
}

export async function listTasks(params?: {
  page?: number
  page_size?: number
  status?: string
  model?: string
  user_id?: number | string
  api_key_id?: number | string
  start_at?: string
  end_at?: string
}): Promise<{ items: VideoTask[]; total: number; page: number; page_size: number }> {
  const { data } = await apiClient.get<{ items: VideoTask[]; total: number; page: number; page_size: number }>('/admin/video/tasks', { params })
  return data
}

export async function getTask(id: string): Promise<VideoTask> {
  const { data } = await apiClient.get<VideoTask>(`/admin/video/tasks/${encodeURIComponent(id)}`)
  return data
}

export async function requeueTask(id: string): Promise<VideoTask> {
  const { data } = await apiClient.post<VideoTask>(`/admin/video/tasks/${encodeURIComponent(id)}/requeue`)
  return data
}

export async function failTask(id: string, payload?: { code?: string; message?: string }): Promise<VideoTask> {
  const { data } = await apiClient.post<VideoTask>(`/admin/video/tasks/${encodeURIComponent(id)}/fail`, payload || {})
  return data
}

const videoAPI = {
  listTemplates,
  createTemplate,
  updateTemplate,
  deleteTemplate,
  testTemplateCreate,
  testTemplateQuery,
  recognizeTemplate,
  listModels,
  listRequestShapes,
  createModel,
  updateModel,
  deleteModel,
  listTasks,
  getTask,
  requeueTask,
  failTask,
}

export default videoAPI
