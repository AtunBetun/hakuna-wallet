package wallet

// Artifact represents a generated wallet asset that is ready to persist or distribute.
type Artifact struct {
	Platform    string
	FileName    string
	ContentType string
	Data        []byte
}
