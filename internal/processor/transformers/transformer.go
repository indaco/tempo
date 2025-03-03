package transformers

// Transformer defines the interface for data transformation operations.
type Transformer interface {
	// Transform processes the input data and returns transformed bytes or an error.
	Transform(cfg TransformationConfig) ([]byte, error)
}

type TransformationConfig struct {
	RawData    string
	Transform  func(string) (string, error)
	MarkerName string
}
