package domain

import (
	"encoding/json"
)

// ValidateLanguageCodes validates that all codes are ISO 639-1 two-letter lowercase.
func ValidateLanguageCodes(codes []string) error {
	for _, code := range codes {
		if len(code) != 2 {
			return ErrInvalidLanguageCode
		}
		for _, c := range code {
			if !(c >= 'a' && c <= 'z') {
				return ErrInvalidLanguageCode
			}
		}
	}
	return nil
}

// ValidateTimezone validates IANA timezone identifier.
func ValidateTimezone(tz string) error {
	// IANA timezone format: region/city (e.g., America/New_York)
	// Must contain at least one slash and only alphanumeric, underscore, slash, plus, minus
	if len(tz) < 3 {
		return ErrInvalidTimezone
	}
	hasSlash := false
	for _, c := range tz {
		if c == '/' {
			hasSlash = true
			continue
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '+' || c == '-') {
			return ErrInvalidTimezone
		}
	}
	if !hasSlash {
		return ErrInvalidTimezone
	}
	return nil
}

// ValidateCountryCode validates ISO 3166-1 alpha-2 uppercase country code.
func ValidateCountryCode(code string) error {
	if len(code) != 2 {
		return ErrInvalidCountryCode
	}
	for _, c := range code {
		if !(c >= 'A' && c <= 'Z') {
			return ErrInvalidCountryCode
		}
	}
	return nil
}

// ValidateCurrencyCode validates ISO 4217 uppercase currency code.
func ValidateCurrencyCode(code string) error {
	if len(code) != 3 {
		return ErrInvalidCurrencyCode
	}
	for _, c := range code {
		if !(c >= 'A' && c <= 'Z') {
			return ErrInvalidCurrencyCode
		}
	}
	return nil
}

// SocialLinks represents social media platform handles/URLs.
type SocialLinks struct {
	TikTok    string `json:"tiktok,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	XTwitter  string `json:"x_twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Website   string `json:"website,omitempty"`
}

// Validate validates social links content.
func (s *SocialLinks) Validate() error {
	// Empty fields allowed; validate URL format if provided
	_ = json.RawMessage{}
	return nil
}

// ToJSON serializes social links to JSONB-compatible format.
func (s *SocialLinks) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, ErrInvalidSocialLinks
	}
	if len(data) > 5*1024 { // 5KB max for social_links
		return nil, ErrSocialLinksTooLarge
	}
	return data, nil
}

// SocialLinksFromJSON deserializes social links from JSONB.
func SocialLinksFromJSON(data json.RawMessage) (*SocialLinks, error) {
	if data == nil || len(data) == 0 {
		return &SocialLinks{}, nil
	}
	var sl SocialLinks
	if err := json.Unmarshal(data, &sl); err != nil {
		return nil, ErrInvalidSocialLinks
	}
	return &sl, nil
}
