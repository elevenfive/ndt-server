package model

// The Measurement struct contains measurement results. This structure is
// meant to be serialised as JSON as sent as a textual message.
type Measurement struct {
	// Elapsed is the number of seconds elapsed since the beginning.
	Elapsed float64 `json:"elapsed"`

	// AppInfo contains application level measurements.
	AppInfo *AppInfo `json:"app_info,omitempty"`

	// BBRInfo is the data measured using TCP BBR instrumentation.
	BBRInfo *BBRInfo `json:"bbr_info,omitempty"`

	// TCPInfo contains metrics measured using TCP_INFO instrumentation.
	TCPInfo *TCPInfo `json:"tcp_info,omitempty"`

	// Origin contains information on who generated this measurement. For very
	// obvious reasons this field is not transferred via JSON but rather it's
	// assigned by the server itself.
	//
	// FIXME: this is sent actually. We probably need an internal message
	// wrapper for moving this information around.
	Origin string
}
