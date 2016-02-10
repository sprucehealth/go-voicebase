package deeplink

// https://docs.google.com/document/d/1kuJszqKi45z2WFly0xhWMOLyvw0S5Z7gFAu0K5AgCAk/edit#heading=h.5mqpvpoen3ud

import "fmt"

func deepLinkBase(webDomain string) string {
	return fmt.Sprintf("https://%s", webDomain)
}

// OrgURL returns a deeplink compatible URL to a particular organization
func OrgURL(webDomain, organizationID string) string {
	return fmt.Sprintf("%s/org/%s", deepLinkBase(webDomain), organizationID)
}

// OrgDetailsURL returns a deeplink compatible URL to a particular organization's details
func OrgDetailsURL(webDomain, organizationID string) string {
	return fmt.Sprintf("%s/details", OrgURL(webDomain, organizationID))
}

// SavedQueryURL returns a deeplink compatible URL to a saved query
func SavedQueryURL(webDomain, organizationID, savedQueryID string) string {
	return fmt.Sprintf("%s/list/%s", OrgURL(webDomain, organizationID), savedQueryID)
}

// SavedQueryDetailsURL returns a deeplink compatible URL to a saved query's details
func SavedQueryDetailsURL(webDomain, organizationID, savedQueryID string) string {
	return fmt.Sprintf("%s/details", SavedQueryURL(webDomain, organizationID, savedQueryID))
}

// ThreadURL returns a deeplink compatible URL to a particular thread
func ThreadURL(webDomain, organizationID, savedQueryID, threadID string) string {
	return fmt.Sprintf("%s/thread/%s", SavedQueryURL(webDomain, organizationID, savedQueryID), threadID)
}

// ThreadURLShareable returns a shareable deep link compatible URL to a particular thread
func ThreadURLShareable(webDomain, organizationID, threadID string) string {
	return fmt.Sprintf("%s/thread/%s", OrgURL(webDomain, organizationID), threadID)
}

// ThreadDetailsURL returns a deeplink compatible URL to a particular thread's details
func ThreadDetailsURL(webDomain, organizationID, savedQueryID, threadID string) string {
	return fmt.Sprintf("%s/details", ThreadURL(webDomain, organizationID, savedQueryID, threadID))
}

// ThreadDetailsURLShareable returns a shareable deep link compatible URL to a particular thread
func ThreadDetailsURLShareable(webDomain, organizationID, threadID string) string {
	return fmt.Sprintf("%s/details", ThreadURLShareable(webDomain, organizationID, threadID))
}

// ThreadMessageURL returns a deeplink compatible URL to a particular message in a thread
func ThreadMessageURL(webDomain, organizationID, savedQueryID, threadID, messageID string) string {
	return fmt.Sprintf("%s/message/%s", ThreadURL(webDomain, organizationID, savedQueryID, threadID), messageID)
}

// ThreadMessageURLShareable returns a shareable deeplink compatible URL to a particular message in a thread
func ThreadMessageURLShareable(webDomain, organizationID, threadID, messageID string) string {
	return fmt.Sprintf("%s/message/%s", ThreadURLShareable(webDomain, organizationID, threadID), messageID)
}

// ThreadMessageDetailsURL returns a deeplink compatible URL to a particular message in a thread's details
func ThreadMessageDetailsURL(webDomain, organizationID, savedQueryID, threadID, messageID string) string {
	return fmt.Sprintf("%s/details", ThreadMessageURL(webDomain, organizationID, savedQueryID, threadID, messageID))
}

// OrgSettingsEmailURL returns a deeplink compatible URL to the email settings for a particular organization
func OrgSettingsEmailURL(webDomain, organizationID string) string {
	return fmt.Sprintf("%s/settings/email", OrgURL(webDomain, organizationID))
}

// OrgSettingsPhoneURL returns a deeplink compatible URL to the phone settings for a particular organization
func OrgSettingsPhoneURL(webDomain, organizationID string) string {
	return fmt.Sprintf("%s/settings/phone", OrgURL(webDomain, organizationID))
}

// OrgSettingsNotificationsURL returns a deeplink compatible URL to the notification settings for a particular organization
func OrgSettingsNotificationsURL(webDomain, organizationID string) string {
	return fmt.Sprintf("%s/settings/notifications", OrgURL(webDomain, organizationID))
}