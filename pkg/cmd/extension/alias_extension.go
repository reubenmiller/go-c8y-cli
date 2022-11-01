package extension

type AliasExtension struct {
	Source      string `json:"-"`
	Name        string `json:"name,omitempty"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	Shell       bool   `json:"shell"`
}

func (a *AliasExtension) GetCommand() string {
	if a.Shell {
		return "!" + a.Command
	}
	return a.Command
}

func (a *AliasExtension) GetName() string {
	return a.Name
}

func (a *AliasExtension) GetDescription() string {
	return a.Description
}

func (a *AliasExtension) IsShell() bool {
	return a.Shell
}
