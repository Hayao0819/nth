package northapi

import "github.com/Hayao0819/go-north"

type tokenKindReporter interface {
	TokenKind() north.TokenKind
}

func (c *Hybrid) officialSupportsScopedAPI() bool {
	if c.official == nil {
		return false
	}
	reporter, ok := c.official.(tokenKindReporter)

	return !ok || reporter.TokenKind() != north.TokenLegacy
}

func (c *Hybrid) webConfigured() bool {
	return c.currentWeb() != nil || c.refresh != nil
}

func (c *Hybrid) SupportsNotifications() bool {
	return c.official != nil || c.webConfigured()
}

func (c *Hybrid) SupportsNotificationStream() bool {
	_, supported := c.official.(officialNotificationStreamAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsBookmarks() bool {
	_, supported := c.official.(officialBookmarkAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsBookmarkFolders() bool {
	_, supported := c.official.(officialBookmarkFolderAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsMessages() bool {
	_, supported := c.official.(officialMessageAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsDMWriting() bool {
	_, supported := c.official.(officialMessageWriterAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsDMRequests() bool {
	_, supported := c.official.(officialMessageRequestAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsDMGroups() bool {
	_, supported := c.official.(officialMessageGroupAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsDMRecipients() bool {
	_, supported := c.official.(officialMessageRecipientAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsDMUnread() bool {
	_, supported := c.official.(officialMessageUnreadAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsTrends() bool {
	_, supported := c.official.(officialTrendAPI)

	return (supported && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsTrendDismiss() bool {
	_, supported := c.official.(officialTrendDismissAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsLists() bool {
	_, supported := c.official.(officialListAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsListEditing() bool {
	_, supported := c.official.(officialListEditorAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsListMembers() bool {
	_, supported := c.official.(officialListMemberAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsSavedPosts() bool {
	_, supported := c.official.(officialSavedPostAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsMedia() bool {
	_, supported := c.official.(officialMediaAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsAccountSafety() bool {
	_, supported := c.official.(officialAccountSafetyAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsMutedKeywords() bool {
	_, supported := c.official.(officialMutedKeywordAPI)

	return supported && c.officialSupportsScopedAPI()
}
