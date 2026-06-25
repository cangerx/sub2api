# 视频异步网关规划

本文档规划 Sub2API 的视频模型异步调用能力。目标是对用户提供统一的 OpenAI 风格视频接口，对内通过「调用模板 + 视频模型目录」适配不同上游的视频任务 API，并支持按次、按秒、按区间计费。

> 设计基线：**最大复用现有 Account / Group / Channel / 并发 / 计费 / 后台 / 安全基础设施**，只为视频特有的「异步任务 + 上游形态适配」新增最小的三张表。不另起一套与现有账号/渠道平行的体系。

---

## 目标

视频生成模型通常是异步任务模式：创建任务、查询状态、下载成片。不同上游的接口路径、任务状态、请求参数、结果字段、计费方式都不一致。

本功能要解决：

- 用户只面对一套统一的视频调用协议（OpenAI 风格）
- 管理员可以配置不同上游和不同模型，且复用已有的账号/渠道/分组管理
- 上游模型 ID、价格、分辨率、时长、上下架变化时不需要改代码
- 支持按次、按秒、按区间计费，且复用现有余额/订阅/限额体系
- 支持后台轮询任务状态和任务完成后结算
- 后续可以平滑接入国内外更多视频模型

---

## 与现有架构的关系（复用清单）

视频网关不重新发明轮子。下表是落地时必须复用的现有能力，以及与本规划的对应关系：

| 现有能力 | 位置 | 视频网关如何复用 |
|----------|------|------------------|
| **上游连接（凭证/base_url/代理/并发/状态/限流冷却）** | `Account`（`ent/schema/account.go`，`platform`/`type`/`credentials`/`base_url`/`proxy_id`/`concurrency`/`rate_multiplier`/`status`） | 视频上游账号即 `Account`，约定 `platform = "video"`。**不新建 `video_providers` 表。** |
| **路由（分组 → 账号、模型映射）** | `Group`（`model_routing`/`supported_model_scopes`）+ 渠道模型映射 | 复用分组路由与映射，`supported_model_scopes` 增加 `"video"`。 |
| **定价（含 token/per_request/image 模式、通配符、生效）** | `Channel` + `channel_model_pricing`（迁移 081/085，已有 `per_request_price`、`billing_mode`） | 视频定价复用渠道定价，**扩展两种计费模式 `second`/`segment`**，新增 `unit_seconds` 列。不新建 `video_price_rules` 表。 |
| **异步后台 worker（ticker + 跨实例单飞）** | `account_expiry_service.go` + `tryAcquireSingletonLeaderLock`（Redis 锁，Postgres advisory 兜底） | 视频 poll worker 直接套用该范式（见「异步任务状态机」）。 |
| **并发槽（Redis）** | `ConcurrencyService.AcquireUserSlot / AcquireAccountSlot` | 视频任务并发复用，新增 slot key 维度（见「并发控制」）。 |
| **原子计费 + 计费幂等** | `UsageBillingRepository.Apply(cmd)` + `usage_billing_dedup`（单事务内扣余额/订阅/Key 配额，按 `request_id+api_key_id` 去重） | 预扣与对账都走此路径，用不同 `RequestID` 区分阶段；**不新建冻结余额子系统**。退款用 `UserRepository.UpdateBalance` 正数加回。 |
| **请求幂等** | `idempotency` 表 + `idempotency.go` | `POST /v1/videos` 强制幂等，重试不重复创建付费任务。 |
| **出站安全 / SSRF 防护** | `security.url_allowlist` + `channel_monitor_ssrf.go` | 用户传入的素材 URL、代理下载上游 URL 全部经过校验。 |
| **网关中间件** | `apiKeyAuth` / `bodyLimit` / `opsErrorLogger` / `endpointNorm` / 分组平台校验 | 视频路由复用同一套中间件链。 |
| **用量日志** | `usage_log`（已含 `billing_mode`/`image_count`/`channel_id`/`group_id`） | 视频完成后写 `usage_log`，扩展视频字段（见「数据模型」）。 |
| **记录清理** | `UsageCleanupService`（timing wheel + 批量删除 + 管理任务） | 视频终态任务按保留期清理。 |
| **ID / 路径库** | `google/uuid`、`tidwall/gjson`（均已是依赖） | `public_id` 用 uuid；模板结果映射用 gjson 路径，不引新库。 |

**结论：本规划只新增 3 张表** —— `video_call_templates`（上游形态适配）、`video_models`（对外模型目录）、`video_generation_tasks`（异步任务）。货币基准统一为 **USD**，与现有 `rate_multiplier` 体系一致。

---

## 核心原则

### 对外统一

用户、前端、第三方客户端统一使用 Sub2API 的 OpenAI 风格视频协议，不直接暴露各家上游的原始调用方式。

```text
用户请求 -> Sub2API 统一视频协议 -> video_model -> 调用模板 -> 上游真实接口（账号/代理走 Account）
```

### 对内适配

每家上游做一个**调用模板（video_call_template）**，负责描述其视频任务 API 形态：

- 创建任务路径
- 查询任务路径
- 下载成片路径
- 取消任务路径
- 状态字段映射
- 结果字段映射
- 错误字段映射

鉴权方式、base_url、代理、并发、限流冷却**不进模板**，而是落在它关联的 `Account` 上，复用现有调度。

### 模型配置化

代码只认识：

```text
template       # 上游 API 形态
request_shape  # 创建任务时的请求字段形态
billing_mode   # request / second / segment
```

对外模型名、上游模型 ID、价格、时长、分辨率、是否启用、是否隐藏、能力限制全部走配置。

---

## 对外协议

严格对齐 OpenAI 的 video 对象，使现有 OpenAI SDK 无需改造即可调用。**不自创 `url`/`video_url`/`metadata.content_url` 字段。**

### 创建视频任务

```http
POST /v1/videos
Authorization: Bearer sk-xxx
Content-Type: application/json
Idempotency-Key: <可选，客户端重试用同一 key>
```

请求示例：

```json
{
  "model": "video-fast",
  "prompt": "一只橘猫在阳光充足的窗台上缓慢转身，镜头稳定",
  "seconds": "10",
  "size": "720x1280",
  "input_reference": "https://example.com/cat.jpg"
}
```

说明：
- `seconds`、`size` 是 OpenAI 视频接口的标准字段（字符串）。内部会把 `size` 解析为分辨率/比例供能力校验与计费使用。
- 文档其余更丰富的素材字段（首尾帧、参考图组等）通过 `request_shape` 在内部映射，见「参数适配」。

响应示例（OpenAI video 对象）：

```json
{
  "id": "video_01jz0000000000000000000000",
  "object": "video",
  "model": "video-fast",
  "status": "queued",
  "progress": 0,
  "created_at": 1780909131,
  "completed_at": null,
  "expires_at": null,
  "seconds": "10",
  "size": "720x1280",
  "error": null
}
```

- `id` 为本地任务 ID（前缀 `video_` + ULID/雪花），**不透出上游 task_id**。
- 成片不放在创建响应里，统一通过 `/content` 端点获取，与 OpenAI 一致。

### 查询视频任务

```http
GET /v1/videos/{video_id}
```

统一状态：

| 状态 | 说明 |
|------|------|
| `queued` | 已创建，等待上游处理 |
| `in_progress` | 上游正在生成 |
| `completed` | 已完成 |
| `failed` | 生成失败（`error` 字段给出原因） |
| `cancelled` | 已取消 |
| `expired` | 结果已过期或本地记录不可用 |

### 列出视频任务

```http
GET /v1/videos?limit=20&after=video_xxx
```

返回当前 API Key/用户维度的任务分页列表，复用现有 `usage`/列表分页约定。

### 下载成片

```http
GET /v1/videos/{video_id}/content
```

下载策略：

1. 优先返回本地缓存或对象存储文件（第四阶段）
2. 没有缓存时按调用模板的 `content` 路径代理上游下载接口
3. 如果上游只在查询结果返回视频 URL，则代理该 URL
4. **代理任意 URL 前必须经过 `security.url_allowlist` / SSRF 守卫**，禁止访问私网/环回/链路本地地址

### 取消任务

```http
POST /v1/videos/{video_id}/cancel
```

如果上游不支持取消，则只允许取消本地 `queued` 或尚未提交的任务，并按规则释放预扣额度。

### 动态模型能力

```http
GET /v1/videos/models
```

> 注意路径用 `/v1/videos/models`（与视频资源同命名空间），避免与现有 `/v1/models`（聊天）混淆。响应见「前端用户页面」。

---

## 内部调用模板（video_call_templates）

调用模板只描述某家上游的**视频任务 API 形态**。鉴权、base_url、代理、并发由关联的 `Account` 提供。模型只引用模板，不重复填写路径。

### 模板字段（落库形态见「数据模型」）

```yaml
name: openai_videos_mp4
# 路径相对 Account.base_url；{task_id} 为上游任务 ID 占位符
create:
  method: POST
  path: /v1/videos
query:
  method: GET
  path: /v1/videos/{task_id}
content:
  method: GET
  path: /v1/videos/{task_id}/content
cancel:
  method: POST
  path: /v1/videos/{task_id}/cancel
poll:
  interval_seconds: 5      # 基础轮询间隔
  backoff_max_seconds: 30  # 指数退避上限
  max_attempts: 240        # 超过即判 failed/expired
timeout:
  create_seconds: 60
  query_seconds: 30
  content_seconds: 300
status_mapping:            # 上游状态 -> 统一状态（不区分大小写匹配）
  succeeded: completed
  processing: in_progress
  queued: queued
  failed: failed
result_mapping:            # 用 gjson 路径从查询响应取值（项目已依赖 tidwall/gjson）
  content_url: data.video_url
  seconds: data.duration
  progress: data.progress
error_mapping:
  code: error.code
  message: error.message
```

映射执行约定（避免落地歧义）：

- **路径语法是 gjson**（`data.video_url`、`data.items.0.url`），不是 JSONPath。项目已依赖 `tidwall/gjson v1.18.0`，不引新库。
- **未知上游状态**：`status_mapping` 未命中时，任务保持 `in_progress` 并照常退避轮询，同时记一条 ops 告警；连续 N 次未知状态或超 `max_attempts` 才判 `failed`，不武断终态。
- **响应非 JSON / 路径取空**：`content_url` 取不到时不转 `completed`（completed 必须拿到可下载来源）；其余字段取空用默认（progress=0）。
- **HTTP 错误码**：上游 4xx/5xx 先按 `error_mapping` 取错误信息；429/5xx 视为可重试（退避后重试，计入 `poll_count`），4xx 参数类错误直接判 `failed`。

### 建议内置模板

| 模板 | 说明 |
|------|------|
| `openai_videos` | `POST /v1/videos`、`GET /v1/videos/{task_id}`、`GET /v1/videos/{task_id}/content` |
| `openai_videos_mp4` | 下载路径使用 `/content.mp4` |
| `result_url_polling` | 查询结果里直接返回视频 URL，没有独立下载接口（代理该 URL 下载） |
| `custom` | 管理员手动配置创建、查询、下载、取消路径与字段映射 |

> 模板与上游账号解耦：同一个 `openai_videos_mp4` 模板可被多个 `platform="video"` 的 Account 复用，账号差异（域名、密钥、代理）由 Account 承载。

---

## 视频模型目录（video_models）

`video_models` 描述对外模型名、上游模型 ID、参数形态、能力限制，并引用调用模板。**定价不进此表**，走 Channel 渠道定价（见「计费设计」）。

### 配置示例

```yaml
video_models:
  - public_model: video-fast          # 对外模型名
    display_name: Video Fast
    template: openai_videos_mp4         # 引用调用模板
    upstream_model_id: videos-fast      # 传给上游的 model
    request_shape: videos               # 创建任务请求形态
    status: active
    capabilities:
      text_to_video: true
      image_to_video: true
      first_last_frame: true
      reference_images: true
      reference_videos: true
    limits:
      min_seconds: 4
      max_seconds: 15
      max_reference_images: 4
      max_reference_videos: 3
    defaults:
      seconds: 10
      size: "1280x720"
    supported:
      seconds: [4, 5, 10, 15]
      sizes: ["854x480", "1280x720", "720x1280"]
    extra_body_allow:                   # extra_body 透传白名单
      - watermark
      - prompt_extend
```

### 状态

| 状态 | 说明 |
|------|------|
| `active` | 用户可见且可调用 |
| `hidden` | 可调用但不在模型列表展示 |
| `deprecated` | 可调用但标记即将下线 |
| `disabled` | 不可调用 |

### 模型映射关系

视频模型复用现有「分组/渠道模型映射」做名称路由，能力与计费不进映射：

```text
用户请求 model -> 分组/渠道模型映射 -> video_models.public_model -> upstream_model_id
```

示例：

```text
video-fast        -> videos-fast
seedance-full     -> seedance2.0
grok-image-video  -> grok-imagine-video-1.5-preview
```

---

## 参数适配（request_shape）

不同上游创建任务时字段不同，因此需要 `request_shape` 把统一字段映射成上游字段。

### 统一标准字段（对外）

| 字段 | 说明 |
|------|------|
| `model` | 对外模型名 |
| `prompt` | 提示词 |
| `seconds` | 视频秒数（OpenAI 标准，字符串） |
| `size` | 分辨率，如 `1280x720`（OpenAI 标准） |
| `input_reference` | 图生视频/参考输入（OpenAI 标准） |
| `extra_body` | 上游特殊参数（受白名单约束） |

> 首尾帧、参考图组/视频组等高级素材，在标准字段不足时通过 `extra_body` + `request_shape` 内部映射进上游字段，对外保持 OpenAI 兼容。

### request_shape 是代码内置的，不是配置

明确边界，避免误解「全配置化」：

- **`request_shape` 是 Go 代码里的一组 builder**（`map[string]ShapeBuilder`，按字符串注册），不是模板里的字段映射表。每个 builder 接收统一的 `VideoCreateRequest`（标准字段已解析），输出该上游创建任务的 JSON body。
- 因此：**新增「同 `request_shape`」的模型只改配置**（建模型目录即可）；**新增「新 `request_shape`」的上游必须写一个 builder + 注册**，这是有代码改动的。验收标准已按此口径表述。
- 模板（`video_call_templates`）只管 HTTP 形态（路径/方法/结果映射）；`request_shape` 管创建 body 的字段形态。二者正交：`(template, request_shape)` 组合覆盖绝大多数上游。

```go
// 形态注册（内置，编译期确定）
type ShapeBuilder func(req *VideoCreateRequest, m *VideoModel) (map[string]any, error)

var shapeBuilders = map[string]ShapeBuilder{
    "videos":       buildVideosShape,
    "seedance":     buildSeedanceShape,
    "grok_imagine": buildGrokImagineShape,
}
```

### size / seconds 的解析与换算规则

统一字段是 OpenAI 风格的 `size`(像素 `宽x高`) 与 `seconds`(字符串)。各上游需要的是「分辨率档位 + 比例」，**换算必须有唯一确定规则**，否则计费与能力校验会错：

```text
size "1280x720" -> width=1280 height=720
  ratio       = 由 width:height 归约（1280:720 -> 16:9；720:1280 -> 9:16；1:1 -> 1:1）
  resolution  = 由 max(width,height) 映射档位：
                  <=854   -> 480P
                  <=1280  -> 720P
                  <=1920  -> 1080P
                  其余    -> 拒绝（不在 supported.sizes 内一律拒绝）
seconds "10" -> 解析为整数 10；非数字或不在 supported.seconds 内 -> 拒绝创建
```

- 校验顺序：先按 `video_models.supported`（`sizes`/`seconds` 白名单）校验，**不在白名单直接 4xx**，不做就近取整猜测。
- builder 内部再把 `(resolution, ratio)` 写成该上游的字段名（如 seedance 的 `metadata.resolution`/`metadata.ratio`）。

### request_shape 示例

#### videos

```json
{
  "model": "videos-fast",
  "prompt": "...",
  "duration": 10,
  "ratio": "9:16",
  "resolution": "720p",
  "first_image": "https://example.com/first.jpg",
  "last_image": "https://example.com/last.jpg",
  "referenceImages": ["https://example.com/ref.jpg"],
  "referenceVideos": ["https://example.com/ref.mp4"]
}
```

#### seedance

```json
{
  "model": "seedance2.0",
  "prompt": "...",
  "duration": 10,
  "input_reference": "https://example.com/ref.jpg",
  "metadata": {
    "resolution": "1080P",
    "ratio": "9:16",
    "prompt_extend": false,
    "watermark": false,
    "media": [{ "type": "first_frame", "url": "https://example.com/first.jpg" }]
  }
}
```

#### grok_imagine

```json
{
  "model": "grok-imagine-video-1.5-preview",
  "prompt": "...",
  "image_url": "https://example.com/cat.jpg",
  "seconds": "4",
  "size": "720x1280"
}
```

### extra_body 合并规则

1. 标准字段优先
2. `extra_body` 只能合并到模型配置 `extra_body_allow` 白名单内的字段
3. 禁止覆盖 `model`、鉴权、内部任务 ID 等敏感字段
4. 所有素材 URL（标准字段与 extra_body 内）在提交上游前都要过 SSRF 校验

---

## 异步任务状态机

视频任务不阻塞 HTTP 请求直到完成。创建接口只提交任务并返回本地任务 ID。

### 状态流转

```text
queued -> in_progress -> completed
queued -> in_progress -> failed
queued -> cancelled
in_progress -> cancelled
completed -> expired
```

### Worker 流程（复用现有后台范式）

poll worker 套用 `account_expiry_service.go` 的 ticker 循环，并用 `tryAcquireSingletonLeaderLock`（Redis 锁 + Postgres advisory 兜底）做跨实例单飞调度。任务抢占用 **`SELECT ... FOR UPDATE SKIP LOCKED`**，支持多 worker 横向扩展、无单点。

```text
每个 tick：
  SELECT * FROM video_generation_tasks
    WHERE status IN ('queued','in_progress')
      AND next_poll_at <= now()
    ORDER BY next_poll_at
    LIMIT batch_size
    FOR UPDATE SKIP LOCKED            -- 多 worker 并行抢占，互不阻塞
  对每条任务：
    -> 取 video_models -> template -> Account（含 base_url/凭证/代理）
    -> queued 且未提交：调用 create，写 upstream_task_id，转 in_progress
    -> in_progress：调用 query，映射状态/进度
    -> completed：记录 content_url + billable_seconds，按实际费用对账结算
    -> failed/cancelled：按规则释放预扣额度（或按配置仍扣费）
    -> 未终态：next_poll_at = now + 指数退避间隔；poll_count++
    -> poll_count 超过 max_attempts：判 failed，释放预扣
```

> 提交上游创建与本地任务落库要保证：先落库（status=queued）再异步提交，避免「上游已创建但本地无记录」导致的孤儿任务与漏计费。

### 创建流程：同步提交一次，失败立即返回

「先落库再异步提交」若直接返回 `queued`，用户对参数非法（模型不存在、size 不支持）也要等几秒轮询才看到 `failed`，体验差。因此创建采用**同步首次提交**：

```text
POST /v1/videos：
  1. 解析校验（model 存在/active、size·seconds 在 supported 白名单、素材 URL 过 SSRF）
     -> 任一不过：立即 4xx，不落库、不计费
  2. 幂等占位（Idempotency-Key 命中已存在任务则直接返回该任务）
  3. 预扣 estimated_cost（balance >= estimated_cost 校验）-> 不足 402
  4. 落库 status=queued, billing_state=reserved
  5. 同步调用上游 create（带 create_seconds 超时）：
       成功 -> 写 upstream_task_id，转 in_progress，返回 200 + 任务对象
       明确失败(4xx) -> 转 failed，全额退预扣，返回对应 4xx
       超时/5xx/网络抖动 -> 保持 queued（不退预扣），返回 200 + queued
         （由 worker 后续重试 create；用 idempotency_key 防上游重复创建）
  6. 之后状态推进全部交给 poll worker
```

这样「确定性失败」当场返回，「不确定」才落入队列异步重试，兼顾体验与一致性。

### 账号选择与长任务退化

账号在**创建时**由 `Group.model_routing` + 现有调度在 `platform="video"` 账号中选定，写入 `task.account_id` 固化。但视频是长任务，轮询时原账号可能已停用/限流：

- **查询/下载跟随创建账号**：`upstream_task_id` 属于该账号，不可换账号查询。
- **原账号不可用时**：任务**不立即失败**，按退避继续重试该账号（账号可能只是临时限流）；超过 `max_attempts` 或账号被删除才判 `failed` 并退预扣。
- **创建阶段无可用账号**：直接 4xx（不落库），与聊天无可用账号一致。
- 不做创建后的跨账号故障切换（列入第四阶段）。

### 任务 ID、过期与清理

- **public_id**：`"video_" + uuid.NewString()`（去掉 `-`）。项目统一用 `google/uuid`，不引入 ULID/雪花。`upstream_task_id` 仅内部存储，不对外暴露。
- **expires_at**：`completed` 时设为 `completed_at + content_retention`（配置，默认如 24h）。`/content` 在 `now > expires_at` 时返回 410，状态对外显示 `expired`。
- **记录清理**：复用 `UsageCleanupService`（timing wheel + 批量删除）按保留期清理 `video_generation_tasks` 终态记录；`usage_log` 视频账单随其既有保留策略，不单独清理。

### 列表与取消的边界

- **`GET /v1/videos` 鉴权范围**：默认按**当前 API Key** 维度返回（与按 Key 计量一致）；用户在前端页面查看时走 user 维度的管理接口。跨 Key 不可见。
- **取消与轮询竞争**：`cancel` 用 `UPDATE ... WHERE status IN ('queued','in_progress') AND ... RETURNING` 抢占式置为 `cancelled`，与 worker 的 `FOR UPDATE SKIP LOCKED` 互斥；抢到才退预扣，抢不到（worker 正在推进/已终态）则返回当前状态，不重复退款。

### 并发控制（复用 ConcurrencyService）

复用 `ConcurrencyService` 的 Redis 槽机制，新增视频维度的 slot key，不复用聊天短连接并发：

| 维度 | slot key | 说明 |
|------|----------|------|
| 用户视频并发 | `video:user:{user_id}` | 同一用户同时运行的视频任务数 |
| API Key 视频并发 | `video:key:{key_id}` | 同一 Key 同时运行的视频任务数 |
| 账号视频并发 | 复用 `Account.concurrency` | 同一上游账号同时运行任务数 |
| Worker 并发 | 配置项 | 后台轮询 worker goroutine 数 |

---

## 计费设计

视频计费独立于 token 计费，但**复用现有渠道定价、余额扣费、限额、倍率体系**。货币基准统一为 **USD**，最终金额经现有 `rate_multiplier`（分组/账号/用户）调整。

### 计费模式（扩展 channel_model_pricing）

现有渠道定价已支持 `token` / `per_request` / `image`。视频在此基础上扩展，新增 `unit_seconds` 列：

| 模式 | 公式 | 复用现状 |
|------|------|----------|
| `request` | `cost = unit_price` | 等价于现有 `per_request_price` |
| `second` | `cost = unit_price * seconds` | 新增，沿用渠道定价存储 |
| `segment` | `cost = ceil(seconds / unit_seconds) * unit_price` | 新增 `unit_seconds` 列 |

> 落地：`channel_model_pricing` 增加 `billing_mode IN ('second','segment')` 取值与 `unit_seconds NUMERIC` 列，复用其通配符匹配、生效逻辑与缓存。不新建 `video_price_rules`。

### 秒数取值

`billable_seconds` 优先级：

1. 上游完成结果返回的 `seconds`
2. 用户请求的 `seconds`
3. 模型配置 `defaults.seconds`
4. 无法确定时**拒绝创建任务**（避免无法计费的任务进入队列）

### 预扣与对账（复用 UsageBillingRepository.Apply 的事务+幂等路径）

异步任务跑几分钟，必须先预授权。落地**不新建冻结余额子系统**，而是复用现有的原子计费路径 `UsageBillingRepository.Apply(cmd)`：它在单个 DB 事务内完成「`usage_billing_dedup` 幂等占位 + `users.balance` 扣减 + 订阅/Key 配额累加」，天然解决并发与重复扣费。

**关键约定：一个视频任务产生两条计费命令，用不同 `RequestID` 区分阶段，各自幂等。**

```text
阶段一 预扣（创建任务时，同步执行）：
  RequestID = "videores:" + task.public_id        # 预扣专用幂等键
  BalanceCost = estimated_cost（按最大可能秒数，见下）
    request: unit_price
    second:  max_seconds * unit_price
    segment: ceil(max_seconds / unit_seconds) * unit_price
  -> 入口已有 balance<=0 拦截（middleware）；这里再对 estimated_cost 做一次显式
     "balance >= estimated_cost" 预检查（视频是高单价长任务，沿用聊天的"允许透支"不合适）。
  -> Apply 成功后 task.reserved_cost = estimated_cost，task.billing_state = 'reserved'

阶段二 对账（任务进入终态时，由 worker 执行）：
  RequestID = "videofin:" + task.public_id        # 结算专用幂等键，与预扣键不冲突
  delta = actual_cost - reserved_cost
    delta < 0（多退）：UpdateBalance(userID, -delta)   // AddBalance 正数 = 退款
    delta > 0（少补）：再发一条 BalanceCost=delta 的 Apply（极少见，按完成秒数>预估时）
    delta == 0：仅写 usage_log
  -> task.billing_state = 'settled'（状态机单向，worker 重入时见 settled 直接跳过）
```

终态处理：

| 终态 | actual_cost | 处理 |
|------|-------------|------|
| `completed` | 按 `billable_seconds` 实算 | 多退少补，写 `usage_log` |
| `failed` | 0（默认） | 全额退回 `reserved_cost` |
| `cancelled` | 0（默认） | 全额退回 `reserved_cost` |
| 上游失败仍收费 | 按模型/渠道配置 | 退回部分或不退 |

**防重入三道闸**：① `usage_billing_dedup` 对 `videofin:` 键去重；② `task.billing_state` 单向 `reserved → settled`，worker `UPDATE ... WHERE billing_state='reserved'` 抢占式推进；③ poll worker 的 `FOR UPDATE SKIP LOCKED` 保证同一任务同时只有一个 worker 处理。三者任一即可防重复退款，叠加是为崩溃恢复留冗余。

> 计费失败时按 `billing.circuit_breaker` 配置 fail-closed：预扣失败则创建接口返回错误、不入队；结算失败则保留 `billing_state='reserved'`，下个 tick 重试（幂等键保证不会重复扣）。

---

## 数据模型

**复用**：上游连接 = `Account`（`platform="video"`），定价 = `Channel`+`channel_model_pricing`（扩展 `second`/`segment`+`unit_seconds`），路由 = `Group`，用量 = `usage_log`（扩展字段）。

**新增 3 张 Ent 表 + 1 个迁移**（编号接续 153，如 `154_video_gateway.sql`）。

### video_call_templates

```text
id
name                 # 唯一
create_method / create_path
query_method  / query_path
content_method / content_path
cancel_method / cancel_path
status_mapping       # jsonb：上游状态 -> 统一状态
result_mapping       # jsonb：content_url/seconds/progress 的取值路径
error_mapping        # jsonb
poll_config          # jsonb：interval/backoff/max_attempts
timeout_config       # jsonb
status               # active/disabled
created_at / updated_at
```

### video_models

```text
id
public_model         # 对外模型名，唯一
display_name
template_id          # -> video_call_templates
upstream_model_id    # 传给上游的 model
request_shape        # videos/seedance/grok_imagine/...
status               # active/hidden/deprecated/disabled
capabilities         # jsonb
defaults             # jsonb（seconds/size）
limits               # jsonb（min/max_seconds、参考素材上限）
supported_options    # jsonb（seconds/sizes 枚举）
extra_body_allow     # jsonb（透传白名单）
sort_order
created_at / updated_at
```

> 上游账号选择：由 `Group.model_routing` + 现有调度在 `platform="video"` 的 Account 中选取，与聊天一致。`video_models` 不直接绑定单个账号。

### video_generation_tasks

```text
id
public_id            # 对外任务 ID："video_" + uuid（去 -），唯一
user_id / api_key_id / group_id / account_id / channel_id
video_model_id       # -> video_models
requested_model      # 用户请求 model
upstream_model       # 实际上游 model
upstream_task_id     # 上游任务 ID（不对外暴露）
status / progress
billing_state        # reserved / settled（结算状态机，防重复退款）
request_payload          # jsonb：对外请求（脱敏）
upstream_request_payload # jsonb：实际发上游的请求
upstream_response_payload# jsonb：最近一次上游查询响应
result_payload           # jsonb
error_code / error_message
content_url              # 统一对外下载来源（优先本地，回退上游）
upstream_content_url
local_content_url        # 第四阶段转存后填充
billing_mode / unit_price / unit_seconds
requested_seconds / billable_seconds
reserved_cost / estimated_cost / actual_cost   # USD
idempotency_key          # 关联幂等记录，防重复创建
submitted_at / started_at / completed_at / expires_at
next_poll_at / poll_count / locked_until
created_at / updated_at
```

索引建议：`(status, next_poll_at)`（worker 扫描）、`(user_id, created_at)`、`(api_key_id, created_at)`、`unique(public_id)`、`unique(idempotency_key)`、`(billing_state)`（清理/对账重试扫描）。

### usage_log 扩展（复用现有表）

视频完成后写入现有 `usage_log`（已含 `billing_mode`/`channel_id`/`group_id`），新增列：

```text
video_task_id
video_seconds
video_size
video_billing_units     # segment 模式的区间数
```

> `billing_mode` 复用现有列，新增取值 `video_request`/`video_second`/`video_segment` 或直接复用 `second`/`segment`（落地时统一约定）。

---

## 路由与中间件

视频路由挂在网关下，**复用现有中间件链**，不新造鉴权/限流/日志：

```text
v1.POST  /videos                 -> bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, apiKeyAuth, requireGroupVideo, idempotency
v1.GET   /videos                 -> apiKeyAuth（列表）
v1.GET   /videos/:id             -> apiKeyAuth
v1.GET   /videos/:id/content     -> apiKeyAuth（SSRF 守卫）
v1.POST  /videos/:id/cancel      -> apiKeyAuth
v1.GET   /videos/models          -> apiKeyAuth
```

- `requireGroupVideo`：参照现有 `requireGroupAnthropic`/平台校验，要求分组 `supported_model_scopes` 含 `"video"`。
- `idempotency`：复用现有幂等中间件/服务；同一 `Idempotency-Key` 重试返回同一任务，不重复扣费。
- `/content` 端点在代理上游/任意 URL 前，强制走 `security.url_allowlist` + SSRF 校验。

---

## 后台管理页面

前端技术栈与约定（落地必须对齐）：**Vue 3 + Pinia**；视图放 `src/views/admin/`；路由在 `src/router/index.ts` 用 lazy import + `meta{requiresAuth:true, requiresAdmin:true, titleKey}` 注册；API 按域拆到 `src/api/admin/`（导出 TS interface）；文案进 `src/i18n/locales/{en,zh}.ts`；列表复用 `useTableLoader`/`usePersistedPageSize`，表单复用 `useForm`。

复用现有「账号管理 / 渠道管理 / 分组管理」，**只新增两块视频专属配置**。

### 新增视图与路由

| 路由 path | name | 视图文件 | 说明 |
|-----------|------|----------|------|
| `/admin/video-templates` | `AdminVideoTemplates` | `views/admin/VideoTemplatesView.vue` | 调用模板管理 |
| `/admin/video-models` | `AdminVideoModels` | `views/admin/VideoModelsView.vue` | 视频模型目录管理 |
| `/admin/video-tasks` | `AdminVideoTasks` | `views/admin/VideoTasksView.vue` | 全局任务监控（只读，排障用） |

API 模块：`src/api/admin/videoTemplates.ts`、`videoModels.ts`、`videoTasks.ts`（仿 `src/api/channels.ts` 的 interface 风格）。后台导航入口加在「系统/网关」分组下。

### 调用模板管理（VideoTemplatesView）

- 列表：名称、状态、被引用模型数、更新时间；行操作 编辑/禁用/删除（被引用时禁止删除，给提示）。
- 表单（抽屉或弹窗，对齐现有 admin 风格）：
  - 创建/查询/下载/取消的 method + path（path 输入框旁标注 `{task_id}` 占位说明）
  - 状态映射（key-value 列表编辑：上游状态 → 统一状态下拉）
  - 结果映射 `content_url`/`seconds`/`progress`（gjson 路径输入，带「示例响应里取值」校验提示）
  - 错误映射 `code`/`message`
  - 轮询：interval / backoff_max / max_attempts；超时：create/query/content
- **测试联调区**（管理员价值最高）：选一个绑定账号 + 填一个最小请求 → 「测试创建」显示上游原始响应与映射后结果；「测试查询」输入 task_id 看状态映射是否正确。仅管理员、不计费、走后端专用 admin 测试接口。

### 视频模型管理（VideoModelsView）

- 列表：对外模型名、展示名、上游模型 ID、`request_shape`、模板、状态徽标、是否展示给用户、sort_order；支持拖拽或编辑排序。
- 表单：
  - 基本：public_model / display_name / upstream_model_id / 模板下拉 / `request_shape` 下拉（取后端已注册 builder 列表，避免填错）
  - 状态：active/hidden/deprecated/disabled
  - 能力开关：text_to_video / image_to_video / first_last_frame / reference_images / reference_videos
  - 枚举：supported.seconds（多选/标签输入）、supported.sizes（标签输入，校验 `宽x高`）
  - 限制：min/max_seconds、max_reference_images/videos
  - 默认值：defaults.seconds / defaults.size
  - extra_body 透传白名单（标签输入）
- 校验：`request_shape` 必须是后端已注册项；`sizes` 必须能被换算规则解析（前端做一次本地校验，与后端一致）。

### 全局任务监控（VideoTasksView，只读）

- 筛选：状态、模型、用户、API Key、时间范围（复用现有 usage 筛选组件）。
- 列：public_id、用户、模型、状态/进度、billing_state、reserved/actual_cost、poll_count、创建/完成时间、错误信息。
- 详情抽屉：脱敏请求、上游最近响应、错误详情；提供「手动重新入队」和「强制判失败并退款」两个管理员动作（高危，二次确认）。

### 复用现有页面（不新增视图）

- **上游账号**：账号管理新增 `platform="video"` 账号；表单复用现有 base_url/凭证/代理/并发/状态字段，仅平台下拉增加 `video` 选项。
- **定价**：渠道定价表单为视频模型增加 `second`/`segment` 计费模式与 `unit_seconds` 输入（在现有 `billing_mode` 选择器扩展选项）。
- **路由/分组**：分组的 `supported_model_scopes` 增加 `video` 选项（复用 `groupsSupportedModelScopes.ts`），模型路由配置沿用现有 UI。

---

## 前端用户页面

用户页面不暴露上游协议，只暴露统一视频生成能力。技术约定同上，视图放 `src/views/user/`，路由 `meta.requiresAdmin:false`。

### 新增视图与路由

| 路由 path | name | 视图文件 | 说明 |
|-----------|------|----------|------|
| `/video` | `UserVideoStudio` | `views/user/VideoStudioView.vue` | 视频生成工作台（左表单 + 右任务列表） |

- API 模块：`src/api/video.ts`（`createVideo` / `getVideo` / `listVideos` / `cancelVideo` / `listVideoModels`，导出对应 interface）。
- Pinia store：`src/stores/video.ts`（当前模型列表缓存、进行中任务集合、轮询节流）。
- 导航入口：用户侧菜单新增「视频」(图标 + `titleKey: video.title`)，**受开关控制**——分组无 `video` scope 或后端未启用视频时不显示入口。

### 页面结构（VideoStudioView）

单页两栏，组件拆到 `src/components/video/`：

- **VideoCreateForm.vue**（左栏）：
  - 模型选择（来自 `GET /v1/videos/models`）
  - 生成方式 tab：文生视频 / 图生视频 / 首尾帧 / 参考素材 —— **按所选模型 `supports` 动态显隐**
  - 素材：**阶段一只支持填 URL**（无上传组件，与现状一致；上传留到第四阶段）
  - 秒数 / 分辨率(size) / 比例 —— 选项来自模型 `supported`，非法组合禁选
  - 提示词输入
  - **价格预估**：随秒数/size 实时按模型 `billing` 计算（`request`/`second`/`segment` 三公式），展示「预扣上限」与「预计实际」
  - 提交按钮：余额不足时禁用并提示充值
- **VideoTaskList.vue**（右栏）：
  - 进行中任务卡片：缩略/占位、状态徽标、进度条、耗时
  - 历史任务分页（复用 `useTableLoader`/`usePersistedPageSize`）
  - 行操作：取消（仅 queued/in_progress）、下载（completed）、查看失败原因
- **VideoTaskDetailDialog.vue**：预览播放器、下载按钮、参数回显、失败原因。

### 轮询与状态刷新

- 复用 **`useAutoRefresh`** 组合式（项目已有），只轮询「存在进行中任务时」：
  - 有 `queued/in_progress` 任务 → 每 5s 拉一次 `listVideos`（或逐个 `getVideo`）；全部终态 → 自动停。
  - 页面隐藏（`visibilitychange`）暂停轮询，省请求。
- `completed` 后用 `GET /v1/videos/{id}/content` 作为 `<video>` src 播放/下载；`expired`（410）显示「已过期」并禁用下载。

### 动态模型能力接口

```http
GET /v1/videos/models
```

响应示例：

```json
{
  "object": "list",
  "data": [
    {
      "id": "video-fast",
      "object": "model",
      "display_name": "Video Fast",
      "status": "active",
      "supports": ["text_to_video", "image_to_video"],
      "seconds": [4, 5, 10, 15],
      "sizes": ["854x480", "1280x720", "720x1280"],
      "limits": { "max_reference_images": 4, "max_reference_videos": 3 },
      "billing": { "mode": "segment", "unit_price": 0.21, "unit_seconds": 15, "currency": "USD" }
    }
  ]
}
```

前端据此动态渲染表单与价格预估，模型增减/调价无需发版。`billing` 直接取自渠道定价，货币为 USD。

### i18n 与可访问性

- 文案 key 统一前缀 `video.*`（`video.title`/`video.form.*`/`video.status.*`/`video.billing.*`），en/zh 同步补齐，不硬编码中文。
- 状态徽标除颜色外带文字（色盲友好）；播放器、下载按钮补 `aria-label`。

---


## 实施阶段

### 阶段一：基础异步视频网关（最小闭环）

> 收敛范围：**1 个视频账号 + 1 个模板 + 1 种 request_shape + 1 个模型**，先打通端到端再扩。

- 新增 `video_call_templates` / `video_models` / `video_generation_tasks` 三表 + 迁移
- 新增 `/v1/videos` 创建（幂等）、`/v1/videos/{id}` 查询、`/v1/videos/{id}/content` 下载（SSRF 守卫）
- 复用 `account_expiry_service` 范式实现 poll worker（`FOR UPDATE SKIP LOCKED` + 单飞锁）
- 复用 `Account`(`platform="video"`) 承载上游连接；复用 `ConcurrencyService` 控制视频并发
- 计费：渠道定价扩展 `second`/`segment`+`unit_seconds`；创建预扣 + 终态对账，复用 `DeductUserBalance`
- 仅接入 1 个模型（建议 `grok-image-video` 按次，计费最简单）跑通

### 阶段二：管理后台配置化

- 新增 `VideoTemplatesView` / `VideoModelsView` 两个 admin 视图 + 对应 `src/api/admin/` 模块与路由
- 调用模板的「测试创建/查询」联调区
- 模型 active/hidden/deprecated/disabled 切换、排序
- 视频账号/定价/分组复用现有页面完成接入（平台下拉加 `video`、计费模式加 `second`/`segment`、scope 加 `video`）
- 可选：`VideoTasksView` 全局任务监控（只读排障）

### 阶段三：用户视频页面

- 新增 `VideoStudioView` + `components/video/*`（CreateForm / TaskList / DetailDialog）
- `src/api/video.ts` + `src/stores/video.ts`，复用 `useAutoRefresh` 做进行中任务轮询
- 动态表单（按模型 `supports`/`supported` 渲染）、实时价格预估、预览下载、失败展示
- `video.*` i18n（en/zh）、导航入口按开关显隐

### 阶段四：更多上游与结果转存

- 接入更多国内外上游与 request_shape
- webhook、对象存储转存、结果过期清理、取消任务、重试策略

---

## 第一阶段建议范围

第一阶段先接入一家上游，用 `platform="video"` 的 Account 承载，跑通最小闭环。

### 支持模型（逐步打开）

| 对外模型 | 上游模型 | request_shape | 计费 | 备注 |
|----------|----------|---------------|------|------|
| `grok-image-video` | `grok-imagine-video-1.5-preview` | `grok_imagine` | 按次 | **首个打通**（计费最简单） |
| `video-standard` | `videos-standard` | `videos` | 按区间 | 验证 segment 计费 |
| `video-fast` | `videos-fast` | `videos` | 按区间 | 复用同模板/形态 |
| `seedance-full` | `seedance2.0` | `seedance` | 按区间或按秒 | 验证第二种 request_shape |

### 不做或延后

- 暂不做 webhook
- 暂不做对象存储转存（`content` 先代理上游）
- 暂不做复杂素材上传（先支持素材 URL）
- 暂不做多上游自动故障切换
- 暂不做用户自定义官方协议透传

### 验收标准

- 用户用统一 `/v1/videos` 创建不同模型的视频任务，OpenAI SDK 直接可用
- 用户可查询任务状态、下载完成后的视频
- 创建即按最大预估**预扣**，终态**对账多退少补**，账目可在 `usage_log` 追溯
- 按次模型正确扣费；按秒/按区间模型按完成秒数结算
- 上游模型 ID 与价格变更只需改配置（账号/渠道/模型目录），不改代码
- 新增同 `request_shape` 模型不需要改代码；新增**新 `request_shape`** 需新增一个 builder（已知有代码改动）
- 重复提交（同 `Idempotency-Key`）不产生重复任务与重复扣费
- 确定性失败（模型不存在/size·seconds 非法/SSRF 拒绝）在创建时同步返回 4xx，不入队
- 终态对账幂等：worker 崩溃重启或重复轮询不会重复退款（`billing_state` + `videofin:` 幂等键双保险）
- 素材 URL 与上游 URL 代理均通过 SSRF 校验
- poll worker 多实例部署不重复处理同一任务
- 完成视频超 `expires_at` 后 `/content` 返回 410、状态显示 `expired`，终态记录由 `UsageCleanupService` 按保留期清理
