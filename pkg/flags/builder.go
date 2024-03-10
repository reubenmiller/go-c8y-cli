package flags

func WithOptionBuilder() *OptionBuilder {
	opts := make([]GetOption, 0)
	return &OptionBuilder{
		Options: opts,
	}
}

type OptionBuilder struct {
	Options []GetOption
}

func (b *OptionBuilder) Append(opts ...GetOption) *OptionBuilder {
	b.Options = append(b.Options, opts...)
	return b
}

func (b *OptionBuilder) AppendSlice(opts []GetOption) *OptionBuilder {
	b.Options = append(b.Options, opts...)
	return b
}

func (b *OptionBuilder) Build() []GetOption {
	return b.Options
}
