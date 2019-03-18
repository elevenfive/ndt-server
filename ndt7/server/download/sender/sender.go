// Package sender implements the download sender.
package sender

import (
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt-server/logging"
	"github.com/m-lab/ndt-server/ndt7/model"
)

// makePreparedMessage generates a prepared message that should be sent
// over the network for generating network load.
func makePreparedMessage(size int) (*websocket.PreparedMessage, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	data := make([]byte, size)
	// This is not the fastest algorithm to generate a random string, yet it
	// is most likely good enough for our purposes. See [1] for a comprehensive
	// discussion regarding how to generate a random string in Golang.
	//
	// .. [1] https://stackoverflow.com/a/31832326/4354461
	for i := range data {
		data[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return websocket.NewPreparedMessage(websocket.BinaryMessage, data)
}

// Start starts the sender in a background goroutine. The sender is a filter
// that receives and processes internal messages coming from an upstream
// stage, typically the measurer. When the internal message contains a new
// measurement, that measurement is sent to the client as a textual message
// and also posted on the returned channel. If no new internal message is
// available, then sender will also send binary messages to generate network
// load. The input channel is always drained. If an error is received by a
// previous stage, the sender will leave early. If no error is received and
// the channel is closed, the sender will send a Close message to the client
// thereby initiating a clean shutdown of the websocket connection.
func Start(conn *websocket.Conn, in <-chan model.Measurement) <-chan model.Measurement {
	out := make(chan model.Measurement)
	go func() {
		defer close(out)
		defer func() {
			for range in {
				// make sure we drain the channel
			}
		}()
		logging.Logger.Debug("sender: start")
		defer logging.Logger.Debug("sender: stop")
		logging.Logger.Debug("sender: generating random buffer")
		const bulkMessageSize = 1 << 13
		preparedMessage, err := makePreparedMessage(bulkMessageSize)
		if err != nil {
			logging.Logger.WithError(err).Warn("makePreparedMessage failed")
			return
		}
		for {
			select {
			case measurement, ok := <-in:
				if !ok {
					// This means that the previous stage has terminated cleanly so
					// we can start closing the websocket connection.
					msg := websocket.FormatCloseMessage(
						websocket.CloseNormalClosure, "Done sending")
					err := conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
					if err != nil {
						logging.Logger.WithError(err).Warn("conn.WriteControl failed")
						return
					}
					return
				}
				if err := conn.WriteJSON(measurement); err != nil {
					logging.Logger.WithError(err).Warn("conn.WriteJSON failed")
					return
				}
				out <- measurement
			default:
				if err := conn.WritePreparedMessage(preparedMessage); err != nil {
					logging.Logger.WithError(err).Warn("conn.WritePreparedMessage failed")
					return
				}
			}
		}
	}()
	return out
}
