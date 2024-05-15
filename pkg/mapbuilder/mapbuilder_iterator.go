package mapbuilder

func NewMapBuilderIterator(b *MapBuilder) *MapBuilderIterator {
	return &MapBuilderIterator{
		MapBuilder: b,
	}
}

type MapBuilderIterator struct {
	MapBuilder *MapBuilder
	Index      int64
}

func (i *MapBuilderIterator) GetNext() (line []byte, input interface{}, err error) {
	out, err := i.MapBuilder.MarshalJSON()
	return out, out, err
}

func (i *MapBuilderIterator) IsBound() bool {
	if i.MapBuilder.TemplateIterator != nil {
		return false
	}
	return i.MapBuilder.TemplateIterator.IsBound()
}
