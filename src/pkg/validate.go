package pkg

type Validatable interface {
	Validate() error
}

func New[T Validatable](v T) (T, error) {
	if err := v.Validate(); err != nil {
		var zero T
		return zero, err
	}
	return v, nil
}
