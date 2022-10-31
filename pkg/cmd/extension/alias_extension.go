package extension

type AliasExtension struct {
	source      string `json:"-"`
	name        string `json:"name,omitempty"`
	command     string `json:"command,omitempty"`
	description string `json:"description,omitempty"`
	shell       bool   `json:"shell"`
}

func (a *AliasExtension) Command() string {
	if a.shell {
		return "!" + a.command
	}
	return a.command
}

func (a *AliasExtension) Name() string {
	return a.name
}

func (a *AliasExtension) Description() string {
	return a.description
}

func (a *AliasExtension) IsShell() bool {
	return a.shell
}
