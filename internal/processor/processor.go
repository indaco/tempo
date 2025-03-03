package processor

// FileProcessor is an interface for processing files.
type FileProcessor interface {
	Process(inputFilePath, outputFilePath, markerName string) error
}
