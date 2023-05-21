package extension

type Command struct {
	name        string
	command     string
	description string
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
