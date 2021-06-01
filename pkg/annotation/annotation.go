package annotation

// Annotations is a list of annotations to describe a command
type Annotations map[string]string

// Option adds meta information to the Annotation
type Option func(Annotations) Annotations

// NewAnnotation creaete a new annotation to describe a command and configure it using a list of options
func NewAnnotation(a Annotations, opts ...Option) Annotations {
	if a == nil {
		a = make(Annotations)
	}

	for _, opt := range opts {
		opt(a)
	}
	return a
}
