export interface VideoCreateRequest {
  model: string
  prompt: string
  seconds?: string
  size?: string
  input_reference?: string
  extra_body?: Record<string, unknown>
}

export interface VideoObjectError {
  code?: string
  message?: string
}

export interface VideoObject {
  id: string
  object: 'video'
  model: string
  status: 'queued' | 'in_progress' | 'completed' | 'failed' | 'cancelled' | 'expired'
  progress: number
  created_at: number
  completed_at?: number | null
  expires_at?: number | null
  seconds?: string
  size?: string
  error?: VideoObjectError | null
}

export interface VideoListResponse {
  object: 'list'
  data: VideoObject[]
}

export interface VideoBilling {
  mode: string
  unit_price: number
  unit_seconds?: number
  currency: string
}

export interface VideoModelObject {
  id: string
  object: 'model'
  display_name?: string
  status: string
  supports?: string[]
  seconds?: number[]
  sizes?: string[]
  limits?: Record<string, unknown>
  extra_body_allow?: string[]
  billing?: VideoBilling
}

export interface VideoModelListResponse {
  object: 'list'
  data: VideoModelObject[]
}

function gatewayHeaders(apiKey: string, idempotencyKey?: string): HeadersInit {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json'
  }
  if (idempotencyKey) {
    headers['Idempotency-Key'] = idempotencyKey
  }
  return headers
}

async function parseGatewayResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get('content-type') || ''
  const payload = contentType.includes('application/json') ? await response.json().catch(() => null) : null
  if (!response.ok) {
    const message = payload?.error?.message || payload?.message || response.statusText || 'Request failed'
    throw new Error(message)
  }
  return payload as T
}

export async function listVideoModels(apiKey: string): Promise<VideoModelListResponse> {
  const response = await fetch('/v1/videos/models', {
    method: 'GET',
    headers: gatewayHeaders(apiKey)
  })
  return parseGatewayResponse<VideoModelListResponse>(response)
}

export async function createVideo(apiKey: string, payload: VideoCreateRequest, idempotencyKey?: string): Promise<VideoObject> {
  const response = await fetch('/v1/videos', {
    method: 'POST',
    headers: gatewayHeaders(apiKey, idempotencyKey),
    body: JSON.stringify(payload)
  })
  return parseGatewayResponse<VideoObject>(response)
}

export async function listVideos(apiKey: string, limit = 20, after = ''): Promise<VideoListResponse> {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  if (after) params.set('after', after)
  const response = await fetch(`/v1/videos?${params.toString()}`, {
    method: 'GET',
    headers: gatewayHeaders(apiKey)
  })
  return parseGatewayResponse<VideoListResponse>(response)
}

export async function getVideo(apiKey: string, id: string): Promise<VideoObject> {
  const response = await fetch(`/v1/videos/${encodeURIComponent(id)}`, {
    method: 'GET',
    headers: gatewayHeaders(apiKey)
  })
  return parseGatewayResponse<VideoObject>(response)
}

export async function cancelVideo(apiKey: string, id: string): Promise<VideoObject> {
  const response = await fetch(`/v1/videos/${encodeURIComponent(id)}/cancel`, {
    method: 'POST',
    headers: gatewayHeaders(apiKey)
  })
  return parseGatewayResponse<VideoObject>(response)
}

export function videoContentURL(id: string): string {
  return `/v1/videos/${encodeURIComponent(id)}/content.mp4`
}
