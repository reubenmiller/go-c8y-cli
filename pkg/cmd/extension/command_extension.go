package extension

type Command struct {
	name        string `json:"name,omitempty"`
	command     string `json:"command,omitempty"`
	description string `json:"description,omitempty"`
}

func (c *Command) Command() string {
	return c.command
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Description() string {
	return c.description
}
