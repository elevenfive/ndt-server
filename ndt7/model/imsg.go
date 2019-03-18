package model

// IMsg is an internal message returned by the pipeline stages implementing
// the ndt7 download and the ndt7 upload.
type IMsg struct {
	// Err is the error that occurred, if any.
	Err error

	// Measurement contains a measurement.
	Measurement Measurement
}
