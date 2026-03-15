package config

// CommandSettings contains the standard commonly used configuration settings
type CommandSettings struct {
	ActivityLog *ActivityLogSettings   `json:"activityLog,omitempty"`
	Encryption  *EncryptionSettings    `json:"encryption,omitempty"`
	IncludeAll  *IncludeAllSettings    `json:"includeAll,omitempty"`
	Session     *SessionSettings       `json:"session,omitempty"`
	Storage     *StorageSettings       `json:"storage,omitempty"`
	Template    *TemplateSettings      `json:"template,omitempty"`
	View        *ViewSettings          `json:"views,omitempty"`
	Login       *LoginSettings         `json:"login,omitempty"`
	SSO         *SSOSettings           `json:"sso,omitempty"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
}

// Bool returns a bool pointer set to the given value
func (s *CommandSettings) Bool(v bool) *bool { return &v }

// Int returns a int pointer set to the given value
func (s *CommandSettings) Int(v int) *int { return &v }

// ActivityLogSettings activity log settings
type ActivityLogSettings struct {
	CurrentPath  string `json:"currentPath,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	MethodFilter string `json:"methodFilter,omitempty"`
	Path         string `json:"path,omitempty"`
}

// SSOSettings SSO settings
type SSOSettings struct {
	DiscoveryURL       string   `json:"discoveryUrl,omitempty"`
	Scopes             []string `json:"scopes,omitempty"`
	Audience           string   `json:"audience,omitempty"`
	BrowserCallbackURL string   `json:"browserCallbackUrl,omitempty"`
}

// EncryptionSettings encryption settings
type EncryptionSettings struct {
	CachePassphrase *bool `json:"cachePassphrase,omitempty"`
	Enabled         *bool `json:"enabled,omitempty"`
}

// StorageSettings storage settings whether passwords and tokens are stored
type StorageSettings struct {
	StoreToken    *bool `json:"storeToken,omitempty"`
	StorePassword *bool `json:"storePassword,omitempty"`
}

// IncludeAllSettings controls the pagination parameters used when paging through all results
type IncludeAllSettings struct {
	DelayMS  *int `json:"delayMS,omitempty"`
	PageSize *int `json:"pageSize,omitempty"`
}

// TemplateSettings controls the jsonnet template settings
type TemplateSettings struct {
	Path string `json:"path,omitempty"`
}

// ViewSettings controls the console table view settings
type ViewSettings struct {
	CommonPaths string `json:"commonPaths,omitempty"`
	CustomPaths string `json:"customPaths,omitempty"`
}

type SessionSettings struct {
	DefaultUsername string `json:"defaultUsername,omitempty"`

	// Types of commands which require confirmation
	Confirmation string `json:"confirmation,omitempty"`

	// ModeSettings controls which types of commands are disabled or not
	Mode string `json:"mode,omitempty"`
}

type LoginSettings struct {
	Type string `json:"type,omitempty"`
}
