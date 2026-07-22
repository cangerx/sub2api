import upstreamLocale from './zh/index'

type LocaleRecord = Record<string, unknown>

const customLocale = {
  "setup": {
    "title": "CCAPI 安装向导",
    "description": "配置您的 CCAPI 实例"
  },
  "nav": {
    "videoStudio": "视频生成",
    "accountList": "账号列表",
    "videoGateway": "视频网关",
    "videoTemplates": "视频模板",
    "videoModels": "视频模型",
    "videoTasks": "视频任务"
  },
  "videoStudio": {
    "title": "视频生成",
    "description": "使用统一 OpenAI 风格视频接口创建、轮询、取消和下载视频任务",
    "apiKeyPlaceholder": "粘贴一个已启用 video 权限的网关 API Key",
    "models": "模型",
    "activeTasks": "进行中",
    "createTask": "创建任务",
    "model": "模型",
    "selectModel": "选择模型",
    "seconds": "时长",
    "size": "尺寸",
    "prompt": "提示词",
    "promptPlaceholder": "描述镜头、主体、动作和风格",
    "reference": "参考素材 URL",
    "estimate": "预估费用",
    "submit": "创建视频任务",
    "tasks": "任务",
    "emptyTasks": "暂无视频任务",
    "download": "下载",
    "loadFailed": "加载视频数据失败",
    "submitFailed": "创建视频任务失败",
    "submitted": "视频任务已提交"
  },
  "keys": {
    "multiGroup": {
      "title": "多分组路由",
      "hint": "开启后可为该 Key 绑定多个分组，按优先级和权重自动切换（仅支持非订阅分组）",
      "addGroup": "添加分组",
      "group": "分组",
      "priority": "优先级",
      "weight": "权重",
      "enabled": "启用",
      "noBindings": "尚未添加分组，点击下方按钮添加",
      "priorityHint": "优先级数字越小越优先；同优先级内按权重分配流量；某分组失败会冷却后自动恢复"
    },
    "forceImageUrlResponse": "图片强制返回 URL",
    "forceImageUrlResponseHint": "开启后此密钥调用图片生成/编辑接口时，即使客户端请求 base64，也会返回可访问的图片地址。"
  },
  "usage": {
    "viewMedia": "查看",
    "mediaDetails": "图片详情",
    "generatedImage": "生成图片",
    "openImage": "打开原图",
    "copyImageUrl": "复制链接",
    "downloadImage": "下载",
    "copied": "已复制",
    "publicUrl": "公网 URL",
    "imageUrlCount": "URL 数量",
    "imagePreviewFailed": "图片预览加载失败",
    "noImageUrlRecorded": "未记录图片地址",
    "prompt": "提示词",
    "revisedPrompts": "优化提示词",
    "noPromptRecorded": "未记录提示词",
    "noRevisedPromptRecorded": "未记录优化提示词",
    "dataUrl": "Data URL 图片",
    "videoGeneration": "视频生成",
    "videoDetails": "视频详情",
    "videoTaskId": "视频任务 ID",
    "copyVideoTaskId": "复制任务 ID",
    "videoSeconds": "视频时长",
    "videoSize": "视频尺寸",
    "videoBillingUnitCount": "计费单位",
    "videoPlaybackUrl": "视频播放地址",
    "noVideoUrlRecorded": "未记录视频地址",
    "noVideoUrlRecordedHint": "当前用量记录只保存了视频任务 ID、时长、尺寸和计费信息，暂未保存视频文件地址。",
    "videoBillingUnits": "{count} 个计费单位",
    "billingModeSecond": "视频按秒",
    "billingModeSegment": "视频按区间"
  },
  "admin": {
    "video": {
      "title": "视频网关",
      "upstreamModelIdOptional": "可留空（推荐）",
      "upstreamModelIdHint": "推荐留空：对外模型名→上游模型名的映射请在「账号管理 → 对应视频账号 → 模型映射」里配置，与聊天/图片一致。此处仅作无账号映射时的兜底。",
      "selectTemplatePreset": "选择内置模板预设",
      "applyPreset": "应用预设",
      "blankTemplate": "空白模板",
      "selectAccount": "请选择账号",
      "aiRecognize": "AI 识别",
      "aiRecognizeTitle": "AI 识别上游文档",
      "aiRecognizeHint": "粘贴上游视频接口文档，选择一个视频账号作为识别用的模型来源，AI 会把文档解析成调用模板草稿供你确认。",
      "aiModel": "识别模型",
      "aiDocument": "上游接口文档",
      "aiDocumentPlaceholder": "在此粘贴上游视频生成接口的文档：创建/查询/下载/取消接口的路径、请求参数、任务状态字段、结果字段等。",
      "aiRecognizing": "识别中…",
      "aiRecognizeSuccess": "识别完成，请核对模板内容后保存",
      "aiRecognizeFailed": "AI 识别失败，请检查文档或改用手动填写",
      "subtitle": "在账号管理里维护视频上游账号、调用模板、对外模型和异步任务。",
      "templates": "调用模板",
      "models": "视频模型",
      "tasks": "视频任务",
      "templateTest": "模板测试",
      "templateTestHint": "选择一个“视频”平台账号，用已保存的模板直接测试上游创建和查询接口。",
      "upstreamAccount": "上游账号",
      "selectVideoAccount": "请选择视频账号",
      "upstreamAccountHint": "账号来自“账号列表 > 添加账号 > Video”，这里不再手填账号 ID。",
      "template": "调用模板",
      "createBodyJson": "创建请求体 JSON",
      "createBodyJsonHint": "这里填写发给上游创建接口的原始请求体，用于验证模板参数和鉴权是否正确。",
      "testCreate": "测试创建",
      "upstreamTaskId": "上游任务 ID",
      "testQuery": "测试查询",
      "filterModel": "模型名",
      "filterUserId": "用户 ID",
      "filterKeyId": "密钥 ID",
      "reservedCost": "预扣费用",
      "actualCost": "实际费用",
      "details": "详情",
      "requeue": "重新入队",
      "failAndRefund": "置失败并退预扣",
      "editTemplate": "编辑调用模板",
      "createTemplate": "创建调用模板",
      "name": "名称",
      "statusLabel": "状态",
      "createMethod": "创建接口方法",
      "createPath": "创建接口路径",
      "queryEndpoint": "查询接口",
      "queryMethod": "查询接口方法",
      "queryPath": "查询接口路径",
      "contentMethod": "成片下载方法",
      "contentPath": "成片下载路径",
      "cancelMethod": "取消接口方法",
      "cancelPath": "取消接口路径",
      "statusMappingJson": "状态映射 JSON",
      "statusMappingJsonHint": "把上游状态映射成 queued、in_progress、completed、failed、cancelled、expired。",
      "resultMappingJson": "结果字段映射 JSON",
      "resultMappingJsonHint": "配置成片地址、时长、进度等字段从上游响应哪个路径读取。",
      "errorMappingJson": "错误字段映射 JSON",
      "errorMappingJsonHint": "配置错误码和错误信息从上游响应哪个路径读取。",
      "pollConfigJson": "轮询配置 JSON",
      "pollConfigJsonHint": "配置轮询间隔、最大退避时间、最大轮询次数等异步查询策略。",
      "timeoutConfigJson": "超时配置 JSON",
      "timeoutConfigJsonHint": "配置创建、查询、下载等上游 HTTP 请求超时时间。",
      "editModel": "编辑视频模型",
      "createModel": "创建视频模型",
      "publicModel": "对外模型名",
      "model": "模型",
      "displayName": "显示名称",
      "requestShape": "请求形态",
      "requestShapeShort": "形态",
      "upstreamModelId": "上游模型 ID",
      "sortOrder": "排序",
      "extraBodyAllow": "允许透传字段",
      "extraBodyAllowPlaceholder": "例如 seed, camera, watermark",
      "capabilitiesJson": "能力配置 JSON",
      "capabilitiesJsonHint": "描述模型能力，例如是否支持参考图、首尾帧、音频、下载等。",
      "defaultsJson": "默认参数 JSON",
      "defaultsJsonHint": "用户请求未填写时自动补的默认值，例如 seconds、size。",
      "limitsJson": "限制配置 JSON",
      "limitsJsonHint": "配置模型可接受的时长、尺寸、并发、文件大小等限制。",
      "supportedOptionsJson": "可选项 JSON",
      "supportedOptionsJsonHint": "配置前端或调用方可展示的可选参数，例如 seconds、sizes。",
      "taskDetail": "视频任务详情",
      "task": "任务",
      "userKey": "用户 / 密钥",
      "account": "账号",
      "requestedModel": "请求模型",
      "upstreamModel": "上游模型",
      "billing": "计费",
      "cost": "费用",
      "polls": "轮询次数",
      "error": "错误",
      "requestPayload": "用户请求",
      "upstreamRequest": "上游请求",
      "upstreamResponse": "上游响应",
      "resultPayload": "结果数据",
      "noTaskSelected": "未选择任务",
      "loadError": "加载视频网关配置失败",
      "tasksLoadError": "加载视频任务失败",
      "templateTestFailed": "模板测试失败",
      "taskActionFailed": "视频任务操作失败",
      "taskLoadError": "加载视频任务详情失败",
      "invalidJson": "{label} 不是合法 JSON",
      "deleteTemplateConfirm": "确定要删除调用模板“{name}”吗？",
      "deleteModelConfirm": "确定要删除视频模型“{name}”吗？",
      "requeueConfirm": "确定要把任务 {id} 重新入队吗？",
      "failConfirm": "确定要强制失败任务 {id} 并退还预扣费用吗？",
      "adminForcedFailed": "管理员强制置为失败",
      "selectAccountAndTemplate": "请先选择上游账号和调用模板",
      "selectAccountTemplateAndTask": "请先选择上游账号、调用模板并填写上游任务 ID",
      "status": {
        "active": "启用",
        "disabled": "禁用",
        "deprecated": "已废弃",
        "queued": "排队中",
        "inProgress": "处理中",
        "completed": "已完成",
        "failed": "失败",
        "cancelled": "已取消",
        "expired": "已过期"
      }
    },
    "backup": {
      "description": "全量数据库备份到对象存储，支持本地、R2、OSS、S3 与定时恢复",
      "s3": {
        "title": "对象存储配置",
        "description": "配置本地、Cloudflare R2、阿里云 OSS 或 S3 兼容存储",
        "descriptionPrefix": "配置对象存储（支持",
        "descriptionSuffix": "、OSS、本地）",
        "enabled": "启用对象存储",
        "provider": "存储类型",
        "local": "本地存储",
        "localPath": "本地目录",
        "publicBaseUrl": "公开访问地址",
        "testSuccess": "存储连接测试成功",
        "testFailed": "存储连接测试失败",
        "saved": "存储配置已保存",
        "errors": {
          "localPathRequired": "请填写本地存储目录",
          "objectStorageRequired": "请填写存储桶、Access Key ID 和 Secret Access Key"
        }
      },
      "errors": {
        "storageNotConfigured": "请先配置对象存储",
        "notFound": "备份记录不存在",
        "notCompleted": "只能恢复已完成的备份",
        "recordsCorrupt": "备份记录数据异常",
        "storageConfigCorrupt": "对象存储配置数据异常",
        "cronRequired": "启用定时备份时请填写 Cron 表达式",
        "invalidCron": "Cron 表达式格式不正确",
        "incorrectPassword": "管理员密码不正确",
        "passwordRequired": "恢复备份需要输入管理员密码"
      },
      "r2Guide": {
        "step1": {
          "line2": "点击「创建存储桶」，输入名称（如 ccapi-backups），选择区域"
        }
      }
    },
    "groups": {
      "platforms": {
        "video": "Video"
      }
    },
    "channels": {
      "billingMode": {
        "second": "视频（按秒）",
        "segment": "视频（按区间）"
      },
      "form": {
        "videoSecondPrice": "每秒价格",
        "videoSegmentPrice": "每区间价格",
        "unitSeconds": "区间单位秒数",
        "unitSecondsPlaceholder": "如 5",
        "videoSecondHint": "费用 = 每秒价格 × 视频秒数",
        "videoSegmentHint": "费用 = ⌈秒数 / 单位秒数⌉ × 每区间价格"
      }
    },
    "accounts": {
      "videoModelMappingHint": "视频账号在此选择视频接口模板，并配置模型白名单或「对外模型名 → 上游模型名」映射。接口模板决定创建、查询、下载视频任务时调用哪些上游路径。",
      "videoFetchUpstreamModels": "一键拉取上游模型",
      "videoFetchUpstreamModelsNeedKey": "请先填写 API Key 再拉取上游模型",
      "videoTemplateSelect": "视频接口模板",
      "videoTemplateSelectPlaceholder": "请选择视频接口模板",
      "videoTemplateGlobalHint": "一个视频账号通常选择一个模板；模板里配置创建、查询、下载任务的上游路径。",
      "videoTemplateRequired": "请先在视频配置中创建至少一个视频接口模板",
      "videoTemplateRequiredForModel": "请为视频模型 {model} 选择视频接口模板",
      "videoModelConfigLoadFailed": "加载视频模型配置失败：{message}",
      "usageWindowsHint": "“5h / 7d”是上游账号（如 OpenAI ChatGPT、Claude）官方的滚动用量窗口限制，由上游对账号设定，并非 ccapi 配置，也与你映射的模型无关。窗口滚动到期后用量会自动重置，无法在 ccapi 端解除该限制。",
      "platforms": {
        "video": "Video"
      },
      "syncUpstreamModelsNoChanges": "上游 {count} 个模型均已在当前配置中",
      "poolModeInfo": "启用后，上游 429/403/401 错误将自动重试而不标记账号限流或错误，适用于上游指向另一个 ccapi 实例的场景。"
    },
    "settings": {
      "linuxdo": {
        "description": "配置 LinuxDo Connect OAuth，用于 CCAPI 用户登录"
      },
      "dingtalk": {
        "description": "配置钉钉 OAuth，用于 CCAPI 用户登录"
      },
      "site": {
        "siteNamePlaceholder": "CCAPI",
        "apiBaseUrlHint": "用于\"使用密钥\"和\"导入到 CC Switch\"功能，留空则使用当前站点地址"
      },
      "payment": {
        "providerTianque": "随行付",
        "field_orgId": "机构号",
        "field_mno": "商户号",
        "field_version": "接口版本",
        "field_tianqueApiBaseHint": "测试环境默认 https://openapi-test.tianquetech.com，生产环境请按随行付后台提供的 openapi 地址填写。",
        "field_tianqueOrgIdHint": "随行付开放平台分配的机构号，通常为 8 或 10 位数字。",
        "field_tianqueMnoHint": "随行付商户编号，用于实际收款商户识别。",
        "field_tianquePrivateKeyHint": "填写商户 RSA 私钥，支持完整 PEM 或去掉头尾后的 Base64 内容；保存后不会明文回显。",
        "field_tianqueVersionHint": "随行付接口版本，未特殊要求保持 1.2。",
        "tianqueNotifyHint": "将生成的完整回调地址配置到随行付后台异步通知地址，公网必须可访问。",
        "tianqueGuideSummary": "随行付作为聚合通道承载支付宝和微信收款，前台仍展示标准支付宝/微信入口。",
        "tianqueGuideNote": "保存后到“可见支付方式”里把支付宝或微信来源切换为随行付，前台才会走该通道。",
        "tianqueGuideMerchantTitle": "商户资料",
        "tianqueGuideMerchantOpen": "在随行付开放平台准备机构号、商户号和 RSA 私钥。",
        "tianqueGuideMerchantCall": "下单时系统按支付方式自动传 ALIPAY 或 WECHAT，并使用商户私钥签名。",
        "tianqueGuideMerchantFallback": "私钥支持完整 PEM；如果后台只给 Base64 内容，也可以直接粘贴。",
        "tianqueGuideCallbackTitle": "异步通知",
        "tianqueGuideCallbackOpen": "把表单生成的回调地址填入随行付后台通知地址。",
        "tianqueGuideCallbackCall": "支付成功通知会按订单号回查原服务商实例并完成充值或订阅发放。",
        "tianqueGuideCallbackFallback": "本地调试需使用公网穿透地址，生产环境请使用 HTTPS 域名。"
      },
      "smtp": {
        "fromNamePlaceholder": "CCAPI"
      }
    }
  },
  "onboarding": {
    "admin": {
      "welcome": {
        "title": "👋 欢迎使用 CCAPI",
        "description": "<div style=\"line-height: 1.8;\"><p style=\"margin-bottom: 16px;\">CCAPI 是一个强大的 AI 服务中转平台，让您轻松管理和分发 AI 服务。</p><p style=\"margin-bottom: 12px;\"><b>🎯 核心功能：</b></p><ul style=\"margin-left: 20px; margin-bottom: 16px;\"><li>📦 <b>分组管理</b> - 创建不同的服务套餐（VIP、免费试用等）</li><li>🔗 <b>账号池</b> - 连接多个上游 AI 服务商账号</li><li>🔑 <b>密钥分发</b> - 为用户生成独立的 API Key</li><li>💰 <b>计费管理</b> - 灵活的费率和配额控制</li></ul><p style=\"color: #10b981; font-weight: 600;\">接下来，我们将用 3 分钟带您完成首次配置 →</p></div>"
      },
      "groupManage": {
        "description": "<div style=\"line-height: 1.7;\"><p style=\"margin-bottom: 12px;\"><b>什么是分组？</b></p><p style=\"margin-bottom: 12px;\">分组是 CCAPI 的核心概念，它就像一个\"服务套餐\"：</p><ul style=\"margin-left: 20px; margin-bottom: 12px; font-size: 13px;\"><li>🎯 每个分组可以包含多个上游账号</li><li>💰 每个分组有独立的计费倍率</li><li>👥 可以设置为公开或专属分组</li></ul><p style=\"margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;\"><b>💡 示例：</b>您可以创建\"VIP专线\"（高倍率）和\"免费试用\"（低倍率）两个分组</p><p style=\"margin-top: 16px; color: #10b981; font-weight: 600;\">👉 点击左侧的\"分组管理\"开始</p></div>"
      }
    },
    "user": {
      "welcome": {
        "title": "👋 欢迎使用 CCAPI",
        "description": "<div style=\"line-height: 1.8;\"><p style=\"margin-bottom: 16px;\">您好！欢迎来到 CCAPI AI 服务平台。</p><p style=\"margin-bottom: 12px;\"><b>🎯 快速开始：</b></p><ul style=\"margin-left: 20px; margin-bottom: 16px;\"><li>🔑 创建 API 密钥</li><li>📋 复制密钥到您的应用</li><li>🚀 开始使用 AI 服务</li></ul><p style=\"color: #10b981; font-weight: 600;\">只需 1 分钟，让我们开始吧 →</p></div>"
      }
    }
  },
  "payment": {
    "methods": {
      "tianque": "随行付"
    }
  }
} as LocaleRecord

function mergeLocale(base: LocaleRecord, overrides: LocaleRecord): LocaleRecord {
  const merged: LocaleRecord = { ...base }

  for (const [key, value] of Object.entries(overrides)) {
    const baseValue = merged[key]
    if (
      value !== null &&
      typeof value === 'object' &&
      !Array.isArray(value) &&
      baseValue !== null &&
      typeof baseValue === 'object' &&
      !Array.isArray(baseValue)
    ) {
      merged[key] = mergeLocale(baseValue as LocaleRecord, value as LocaleRecord)
    } else {
      merged[key] = value
    }
  }

  return merged
}

export default mergeLocale(upstreamLocale, customLocale)
