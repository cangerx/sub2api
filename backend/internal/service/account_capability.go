package service

import "strings"

type AccountCapability string

const (
	AccountCapabilityChat       AccountCapability = "chat"
	AccountCapabilityResponses  AccountCapability = "responses"
	AccountCapabilityMessages   AccountCapability = "messages"
	AccountCapabilityEmbeddings AccountCapability = "embeddings"
	AccountCapabilityImages     AccountCapability = "images"
	AccountCapabilityVideos     AccountCapability = "videos"
)

const (
	accountCapabilitiesCredentialKey         = "capabilities"
	accountPlatformCapabilitiesCredentialKey = "platform_capabilities"
)

func (a *Account) SupportsCapability(capability AccountCapability) bool {
	if a == nil {
		return false
	}
	capability = normalizeAccountCapability(capability)
	if capability == "" {
		return true
	}

	configured, found := a.accountCapabilitySet()
	if found {
		return configured[string(capability)]
	}

	switch capability {
	case AccountCapabilityVideos:
		return a.Platform == PlatformVideo
	case AccountCapabilityChat, AccountCapabilityResponses, AccountCapabilityMessages:
		return a.Platform == PlatformAnthropic || a.Platform == PlatformOpenAI || a.Platform == PlatformGemini || a.Platform == PlatformAntigravity
	case AccountCapabilityEmbeddings:
		return a.Platform == PlatformOpenAI && a.Type == AccountTypeAPIKey
	case AccountCapabilityImages:
		return a.Platform == PlatformOpenAI || a.Platform == PlatformGemini || a.Platform == PlatformAntigravity
	default:
		return false
	}
}

func (a *Account) accountCapabilitySet() (map[string]bool, bool) {
	if a == nil || a.Credentials == nil {
		return nil, false
	}
	result := make(map[string]bool)
	found := false
	if addCapabilities(result, a.Credentials[accountCapabilitiesCredentialKey]) {
		found = true
	}
	if addCapabilities(result, a.Credentials[accountPlatformCapabilitiesCredentialKey]) {
		found = true
	}
	return result, found
}

func addCapabilities(dst map[string]bool, raw any) bool {
	if raw == nil {
		return false
	}
	found := false
	add := func(value string) {
		value = string(normalizeAccountCapability(AccountCapability(value)))
		if value == "" {
			return
		}
		dst[value] = true
		found = true
	}
	switch capabilities := raw.(type) {
	case []any:
		for _, item := range capabilities {
			if value, ok := item.(string); ok {
				add(value)
			}
		}
	case []string:
		for _, value := range capabilities {
			add(value)
		}
	case map[string]any:
		for key, value := range capabilities {
			if enabled, ok := value.(bool); ok && enabled {
				add(key)
			}
		}
	case map[string]bool:
		for key, enabled := range capabilities {
			if enabled {
				add(key)
			}
		}
	case string:
		for _, value := range strings.Split(capabilities, ",") {
			add(value)
		}
	}
	return found
}

func normalizeAccountCapability(capability AccountCapability) AccountCapability {
	switch strings.ToLower(strings.TrimSpace(string(capability))) {
	case "chat", "chat_completions", "chat-completions":
		return AccountCapabilityChat
	case "responses":
		return AccountCapabilityResponses
	case "messages":
		return AccountCapabilityMessages
	case "embeddings":
		return AccountCapabilityEmbeddings
	case "image", "images", "image_generation", "image-generation":
		return AccountCapabilityImages
	case "video", "videos", "video_generation", "video-generation":
		return AccountCapabilityVideos
	default:
		return ""
	}
}
