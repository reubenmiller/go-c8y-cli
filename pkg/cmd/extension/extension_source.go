package extension

type ExtensionSource struct {
	Name  string
	Paths []string
}

type ExtensionItemCollection struct {
	Items []ExtensionSource
}

func (t *ExtensionItemCollection) Add(name string, paths []string) {
	if len(paths) > 0 {
		t.Items = append(t.Items, ExtensionSource{name, paths})
	}
}

func (t *ExtensionItemCollection) AddSources(sources []ExtensionSource) {
	t.Items = append(t.Items, sources...)
}

func NewExtensionItemCollection() *ExtensionItemCollection {
	collection := &ExtensionItemCollection{
		Items: []ExtensionSource{},
	}
	return collection
}
