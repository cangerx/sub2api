import upstreamLocale from './en/index'

type LocaleRecord = Record<string, unknown>

const customLocale = {
  "setup": {
    "title": "CCAPI Setup",
    "description": "Configure your CCAPI instance"
  },
  "nav": {
    "videoStudio": "Video Studio",
    "accountList": "Account List",
    "videoGateway": "Video Gateway",
    "videoTemplates": "Video Templates",
    "videoModels": "Video Models",
    "videoTasks": "Video Tasks"
  },
  "videoStudio": {
    "title": "Video Studio",
    "description": "Create, poll, cancel, and download video tasks through the unified OpenAI-style video API",
    "apiKeyPlaceholder": "Paste a gateway API key with video access enabled",
    "models": "Models",
    "activeTasks": "Active",
    "createTask": "Create Task",
    "model": "Model",
    "selectModel": "Select model",
    "seconds": "Seconds",
    "size": "Size",
    "prompt": "Prompt",
    "promptPlaceholder": "Describe the shot, subject, movement, and style",
    "reference": "Reference Asset URL",
    "estimate": "Estimated Cost",
    "submit": "Create Video Task",
    "tasks": "Tasks",
    "emptyTasks": "No video tasks yet",
    "download": "Download",
    "loadFailed": "Failed to load video data",
    "submitFailed": "Failed to create video task",
    "submitted": "Video task submitted"
  },
  "keys": {
    "multiGroup": {
      "title": "Multi-group routing",
      "hint": "Bind this key to multiple groups and route by priority/weight (non-subscription groups only)",
      "addGroup": "Add group",
      "group": "Group",
      "priority": "Priority",
      "weight": "Weight",
      "enabled": "Enabled",
      "noBindings": "No groups added yet. Use the button below to add one.",
      "priorityHint": "Lower priority value is preferred; traffic splits by weight within the same priority; a failed group cools down then auto-recovers"
    },
    "forceImageUrlResponse": "Force image URL response",
    "forceImageUrlResponseHint": "When enabled, image generation/edit requests using this key return accessible image URLs even if the client requests base64."
  },
  "usage": {
    "viewMedia": "View",
    "mediaDetails": "Image details",
    "generatedImage": "Generated image",
    "openImage": "Open image",
    "copyImageUrl": "Copy link",
    "downloadImage": "Download",
    "copied": "Copied",
    "publicUrl": "Public URL",
    "imageUrlCount": "URL count",
    "imagePreviewFailed": "Image preview failed",
    "noImageUrlRecorded": "No image URL recorded",
    "prompt": "Prompt",
    "revisedPrompts": "Revised prompts",
    "noPromptRecorded": "No prompt recorded",
    "noRevisedPromptRecorded": "No revised prompt recorded",
    "dataUrl": "Data URL image",
    "videoGeneration": "Video generation",
    "videoDetails": "Video details",
    "videoTaskId": "Video task ID",
    "copyVideoTaskId": "Copy task ID",
    "videoSeconds": "Duration",
    "videoSize": "Video size",
    "videoBillingUnitCount": "Billing units",
    "videoPlaybackUrl": "Playback URL",
    "noVideoUrlRecorded": "No video URL recorded",
    "noVideoUrlRecordedHint": "This usage record currently stores the video task ID, duration, size, and billing data, but not the video file URL.",
    "videoBillingUnits": "{count} billing units",
    "billingModeSecond": "Video per second",
    "billingModeSegment": "Video per segment"
  },
  "admin": {
    "video": {
      "title": "Video Gateway",
      "upstreamModelIdOptional": "Leave empty (recommended)",
      "upstreamModelIdHint": "Recommended empty: configure the public → upstream model name mapping under Account Management → the video account → Model Mapping, same as chat/image. This field is only a fallback when no account mapping exists.",
      "selectTemplatePreset": "Select a built-in preset",
      "applyPreset": "Apply preset",
      "blankTemplate": "Blank template",
      "selectAccount": "Select an account",
      "aiRecognize": "AI Recognize",
      "aiRecognizeTitle": "AI: recognize upstream docs",
      "aiRecognizeHint": "Paste the upstream video API docs and pick a video account as the model source. The AI parses the docs into a draft call template for you to review.",
      "aiModel": "Recognition model",
      "aiDocument": "Upstream API docs",
      "aiDocumentPlaceholder": "Paste the upstream video generation API docs here: create/query/content/cancel paths, request params, task status fields, result fields, etc.",
      "aiRecognizing": "Recognizing…",
      "aiRecognizeSuccess": "Recognized. Review the template before saving.",
      "aiRecognizeFailed": "AI recognition failed. Check the docs or fill the template manually.",
      "subtitle": "Manage video upstream accounts, call templates, public models, and async tasks inside Account Management.",
      "templates": "Call Templates",
      "models": "Video Models",
      "tasks": "Video Tasks",
      "templateTest": "Template Test",
      "templateTestHint": "Select a Video account and test the saved upstream create/query template directly.",
      "upstreamAccount": "Upstream Account",
      "selectVideoAccount": "Select a video account",
      "upstreamAccountHint": "Accounts come from Account List > Add Account > Video. Raw account ID input is no longer needed.",
      "template": "Template",
      "createBodyJson": "Create Body JSON",
      "createBodyJsonHint": "Raw request body sent to the upstream create endpoint for checking template parameters and auth.",
      "testCreate": "Test Create",
      "upstreamTaskId": "Upstream Task ID",
      "testQuery": "Test Query",
      "filterModel": "Model",
      "filterUserId": "User ID",
      "filterKeyId": "Key ID",
      "reservedCost": "Reserved",
      "actualCost": "Actual",
      "details": "Details",
      "requeue": "Requeue",
      "failAndRefund": "Fail and refund",
      "editTemplate": "Edit Template",
      "createTemplate": "Create Template",
      "name": "Name",
      "statusLabel": "Status",
      "createMethod": "Create Method",
      "createPath": "Create Path",
      "queryEndpoint": "Query Endpoint",
      "queryMethod": "Query Method",
      "queryPath": "Query Path",
      "contentMethod": "Content Method",
      "contentPath": "Content Path",
      "cancelMethod": "Cancel Method",
      "cancelPath": "Cancel Path",
      "statusMappingJson": "Status Mapping JSON",
      "statusMappingJsonHint": "Map upstream statuses to queued, in_progress, completed, failed, cancelled, or expired.",
      "resultMappingJson": "Result Field Mapping JSON",
      "resultMappingJsonHint": "Configure where to read video URL, duration, progress, and other result fields from upstream responses.",
      "errorMappingJson": "Error Field Mapping JSON",
      "errorMappingJsonHint": "Configure where to read error code and message from upstream responses.",
      "pollConfigJson": "Polling Config JSON",
      "pollConfigJsonHint": "Configure polling interval, max backoff, max attempts, and async query strategy.",
      "timeoutConfigJson": "Timeout Config JSON",
      "timeoutConfigJsonHint": "Configure upstream HTTP timeouts for create, query, content download, and related calls.",
      "editModel": "Edit Video Model",
      "createModel": "Create Video Model",
      "publicModel": "Public Model Name",
      "model": "Model",
      "displayName": "Display Name",
      "requestShape": "Request Shape",
      "requestShapeShort": "Shape",
      "upstreamModelId": "Upstream Model ID",
      "sortOrder": "Sort Order",
      "extraBodyAllow": "Allowed Passthrough Fields",
      "extraBodyAllowPlaceholder": "e.g. seed, camera, watermark",
      "capabilitiesJson": "Capabilities JSON",
      "capabilitiesJsonHint": "Describe model capabilities such as reference image, first/last frame, audio, or download support.",
      "defaultsJson": "Default Params JSON",
      "defaultsJsonHint": "Defaults added when the user request omits them, such as seconds or size.",
      "limitsJson": "Limit Config JSON",
      "limitsJsonHint": "Configure accepted duration, size, concurrency, file size, and other constraints.",
      "supportedOptionsJson": "Supported Options JSON",
      "supportedOptionsJsonHint": "Options that callers or UI can expose, such as seconds and sizes.",
      "taskDetail": "Video Task Detail",
      "task": "Task",
      "userKey": "User / Key",
      "account": "Account",
      "requestedModel": "Requested Model",
      "upstreamModel": "Upstream Model",
      "billing": "Billing",
      "cost": "Cost",
      "polls": "Polls",
      "error": "Error",
      "requestPayload": "Request Payload",
      "upstreamRequest": "Upstream Request",
      "upstreamResponse": "Upstream Response",
      "resultPayload": "Result Payload",
      "noTaskSelected": "No task selected",
      "loadError": "Failed to load video gateway config",
      "tasksLoadError": "Failed to load video tasks",
      "templateTestFailed": "Template test failed",
      "taskActionFailed": "Video task action failed",
      "taskLoadError": "Failed to load video task",
      "invalidJson": "{label} is not valid JSON",
      "deleteTemplateConfirm": "Delete template \"{name}\"?",
      "deleteModelConfirm": "Delete video model \"{name}\"?",
      "requeueConfirm": "Requeue task {id}?",
      "failConfirm": "Force fail task {id} and refund reserved cost?",
      "adminForcedFailed": "Forced failed by admin",
      "selectAccountAndTemplate": "Select an upstream account and template first",
      "selectAccountTemplateAndTask": "Select an upstream account, template, and upstream task ID first",
      "status": {
        "active": "Active",
        "disabled": "Disabled",
        "deprecated": "Deprecated",
        "queued": "Queued",
        "inProgress": "In progress",
        "completed": "Completed",
        "failed": "Failed",
        "cancelled": "Cancelled",
        "expired": "Expired"
      }
    },
    "backup": {
      "description": "Full database backup to object storage with local, R2, OSS, S3, scheduled backup and restore",
      "s3": {
        "title": "Object Storage Configuration",
        "description": "Configure local, Cloudflare R2, Aliyun OSS, or S3-compatible storage",
        "descriptionPrefix": "Configure object storage (supports",
        "descriptionSuffix": ", OSS, local)",
        "enabled": "Enable Object Storage",
        "provider": "Storage Type",
        "local": "Local Storage",
        "localPath": "Local Path",
        "publicBaseUrl": "Public Base URL",
        "testSuccess": "Storage connection test successful",
        "testFailed": "Storage connection test failed",
        "saved": "Storage configuration saved",
        "errors": {
          "localPathRequired": "Please enter a local storage path",
          "objectStorageRequired": "Please enter bucket, Access Key ID, and Secret Access Key"
        }
      },
      "errors": {
        "storageNotConfigured": "Please configure object storage first",
        "notFound": "Backup record not found",
        "notCompleted": "Only completed backups can be restored",
        "recordsCorrupt": "Backup records data is corrupted",
        "storageConfigCorrupt": "Object storage configuration data is corrupted",
        "cronRequired": "Please enter a cron expression when scheduled backup is enabled",
        "invalidCron": "Invalid cron expression",
        "incorrectPassword": "Incorrect admin password",
        "passwordRequired": "Admin password is required to restore a backup"
      },
      "r2Guide": {
        "step1": {
          "line2": "Click \"Create bucket\", enter a name (e.g. ccapi-backups), choose a region"
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
        "second": "Video (Per Second)",
        "segment": "Video (Per Segment)"
      },
      "form": {
        "videoSecondPrice": "Price per second",
        "videoSegmentPrice": "Price per segment",
        "unitSeconds": "Seconds per segment",
        "unitSecondsPlaceholder": "e.g. 5",
        "videoSecondHint": "Cost = price per second × video seconds",
        "videoSegmentHint": "Cost = ⌈seconds / unit seconds⌉ × price per segment"
      }
    },
    "accounts": {
      "videoModelMappingHint": "Select a video interface template here, then configure the model whitelist or public → upstream mapping. The template defines the upstream paths used to create, query, and download video tasks.",
      "videoFetchUpstreamModels": "Fetch upstream models",
      "videoFetchUpstreamModelsNeedKey": "Enter the API Key first to fetch upstream models",
      "videoTemplateSelect": "Video interface template",
      "videoTemplateSelectPlaceholder": "Select video interface template",
      "videoTemplateGlobalHint": "A video account usually uses one template. The template configures upstream create, query, and download paths.",
      "videoTemplateRequired": "Create at least one video interface template first",
      "videoTemplateRequiredForModel": "Select a video interface template for video model {model}",
      "videoModelConfigLoadFailed": "Failed to load video model configuration: {message}",
      "platforms": {
        "video": "Video"
      },
      "usageWindowsHint": "\"5h / 7d\" are the upstream account's official rolling usage windows (e.g. OpenAI ChatGPT, Claude). They are imposed by the upstream provider on the account itself — not configured by ccapi, and unrelated to the models you map. Usage resets automatically once each window rolls over, and the limit cannot be lifted from within ccapi.",
      "syncUpstreamModelsNoChanges": "All {count} upstream model(s) are already in the current configuration",
      "poolModeInfo": "When enabled, upstream 429/403/401 errors will auto-retry without marking the account as rate-limited or errored. Suitable for upstream pointing to another ccapi instance."
    },
    "settings": {
      "linuxdo": {
        "description": "Configure LinuxDo Connect OAuth for CCAPI end-user login"
      },
      "dingtalk": {
        "description": "Configure DingTalk OAuth for CCAPI end-user login"
      },
      "site": {
        "siteNamePlaceholder": "CCAPI",
        "apiBaseUrlHint": "Used for \"Use Key\" and \"Import to CC Switch\" features. Leave empty to use current site URL."
      },
      "payment": {
        "providerTianque": "SuixingPay",
        "field_orgId": "Organization ID",
        "field_mno": "Merchant Number",
        "field_version": "API Version",
        "field_tianqueApiBaseHint": "Default sandbox endpoint is https://openapi-test.tianquetech.com. Use the production openapi endpoint provided by SuixingPay for live payments.",
        "field_tianqueOrgIdHint": "Organization ID assigned by SuixingPay Open Platform, usually an 8 or 10 digit number.",
        "field_tianqueMnoHint": "SuixingPay merchant number used to identify the receiving merchant.",
        "field_tianquePrivateKeyHint": "Paste the merchant RSA private key. Full PEM and base64 body formats are both accepted; it will not be returned in plaintext after saving.",
        "field_tianqueVersionHint": "SuixingPay API version. Keep 1.2 unless your account manager requires a different version.",
        "tianqueNotifyHint": "Configure the generated full callback URL as the async notify URL in SuixingPay. It must be reachable from the public internet.",
        "tianqueGuideSummary": "SuixingPay acts as an aggregate channel for Alipay and WeChat Pay while the checkout page still shows the standard Alipay/WeChat entries.",
        "tianqueGuideNote": "After saving, switch the Alipay or WeChat visible-method source to SuixingPay so frontend payments use this channel.",
        "tianqueGuideMerchantTitle": "Merchant Credentials",
        "tianqueGuideMerchantOpen": "Prepare the organization ID, merchant number, and RSA private key from SuixingPay Open Platform.",
        "tianqueGuideMerchantCall": "Orders automatically send ALIPAY or WECHAT based on the selected payment method and are signed with the merchant private key.",
        "tianqueGuideMerchantFallback": "Full PEM private keys are supported. If your console only provides the base64 body, paste it directly.",
        "tianqueGuideCallbackTitle": "Async Notification",
        "tianqueGuideCallbackOpen": "Copy the generated callback URL into SuixingPay as the async notify URL.",
        "tianqueGuideCallbackCall": "Successful payment notifications resolve the original provider instance by order number and complete the recharge or subscription.",
        "tianqueGuideCallbackFallback": "Use a public tunnel for local tests. Use an HTTPS domain in production."
      },
      "smtp": {
        "fromNamePlaceholder": "CCAPI"
      }
    }
  },
  "onboarding": {
    "admin": {
      "welcome": {
        "title": "👋 Welcome to CCAPI",
        "description": "<div style=\"line-height: 1.8;\"><p style=\"margin-bottom: 16px;\">CCAPI is a powerful AI service gateway platform that helps you easily manage and distribute AI services.</p><p style=\"margin-bottom: 12px;\"><b>🎯 Core Features:</b></p><ul style=\"margin-left: 20px; margin-bottom: 16px;\"><li>📦 <b>Group Management</b> - Create service tiers (VIP, Free Trial, etc.)</li><li>🔗 <b>Account Pool</b> - Connect multiple upstream AI service accounts</li><li>🔑 <b>Key Distribution</b> - Generate independent API Keys for users</li><li>💰 <b>Billing Control</b> - Flexible rate and quota management</li></ul><p style=\"color: #10b981; font-weight: 600;\">Let's complete the initial setup in 3 minutes →</p></div>"
      },
      "groupManage": {
        "description": "<div style=\"line-height: 1.7;\"><p style=\"margin-bottom: 12px;\"><b>What is a Group?</b></p><p style=\"margin-bottom: 12px;\">Groups are the core concept of CCAPI, like a \"service package\":</p><ul style=\"margin-left: 20px; margin-bottom: 12px; font-size: 13px;\"><li>🎯 Each group can contain multiple upstream accounts</li><li>💰 Each group has independent billing multiplier</li><li>👥 Can be set as public or exclusive</li></ul><p style=\"margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;\"><b>💡 Example:</b> You can create \"VIP Premium\" (high rate) and \"Free Trial\" (low rate) groups</p><p style=\"margin-top: 16px; color: #10b981; font-weight: 600;\">👉 Click \"Group Management\" on the left sidebar</p></div>"
      }
    },
    "user": {
      "welcome": {
        "title": "👋 Welcome to CCAPI",
        "description": "<div style=\"line-height: 1.8;\"><p style=\"margin-bottom: 16px;\">Hello! Welcome to the CCAPI AI service platform.</p><p style=\"margin-bottom: 12px;\"><b>🎯 Quick Start:</b></p><ul style=\"margin-left: 20px; margin-bottom: 16px;\"><li>🔑 Create API Key</li><li>📋 Copy key to your application</li><li>🚀 Start using AI services</li></ul><p style=\"color: #10b981; font-weight: 600;\">Just 1 minute, let's get started →</p></div>"
      }
    }
  },
  "payment": {
    "methods": {
      "tianque": "SuixingPay"
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
