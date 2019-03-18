// Package receiver implements the counter-flow messages receiver.
package receiver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt-server/logging"
	"github.com/m-lab/ndt-server/ndt7/model"
	"github.com/m-lab/ndt-server/ndt7/spec"
)

// Start starts the counter-flow messages receiver in a background
// goroutine. The receiver will run until the context is not done
// and it receives valid counter-flow messages. These messages will
// be emitted on the returned channel. Also errors will be emitted
// there. An error causes the receiver to terminate. The clean final
// state is reached when the client sends us a Close message.
func Start(ctx context.Context, conn *websocket.Conn) <-chan model.IMsg {
	out := make(chan model.IMsg)
	go func() {
		defer close(out)
		logging.Logger.Debug("receiver: start")
		defer logging.Logger.Debug("receiver: stop")
		conn.SetReadLimit(spec.MinMaxMessageSize)
		for {
			select {
			case <-ctx.Done():
				logging.Logger.Debug("receiver: context done")
				return
			default:
			}
			mtype, mdata, err := conn.ReadMessage()
			if err != nil {
				// A normal closure is what we'd like to see here. The receiver should
				// issue a normal closure when it has finished uploading all the
				// pending counter-flow measurements. If the receiver is a simple
				// receiver that doesn't upload counter-flow measurements, we'll
				// be blocked in conn.ReadMessage until the next stage of the pipeline
				// will timeout and close the connection.
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					out <- model.IMsg{Err: err}
					return
				}
				return
			}
			if mtype != websocket.TextMessage {
				out <- model.IMsg{Err: errors.New("received non textual message")}
				return
			}
			var measurement model.Measurement
			err = json.Unmarshal(mdata, &measurement)
			if err != nil {
				out <- model.IMsg{Err: err}
				return
			}
			out <- model.IMsg{Measurement: measurement}
		}
	}()
	return out
}
