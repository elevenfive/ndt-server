// Package download implements the ndt7/server downloader.
package download

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/m-lab/go/warnonerror"
	"github.com/m-lab/ndt-server/logging"
	"github.com/m-lab/ndt-server/ndt7/model"
	"github.com/m-lab/ndt-server/ndt7/server/download/measurer"
	"github.com/m-lab/ndt-server/ndt7/server/download/receiver"
	"github.com/m-lab/ndt-server/ndt7/server/download/sender"
	"github.com/m-lab/ndt-server/ndt7/server/results"
	"github.com/m-lab/ndt-server/ndt7/spec"
)

// Handler handles a download subtest from the server side.
type Handler struct {
	Upgrader websocket.Upgrader
	DataDir  string
}

// warnAndClose emits a warning |message| and then closes the HTTP connection
// using the |writer| http.ResponseWriter.
func warnAndClose(writer http.ResponseWriter, message string) {
	logging.Logger.Warn(message)
	writer.Header().Set("Connection", "Close")
	writer.WriteHeader(http.StatusBadRequest)
}

// Merge merges the ouput from many Measurements channels into a single channel.
//
// See <https://medium.com/justforfunc/two-ways-of-merging-n-channels-in-go-43c0b57cd1de>
func merge(cs ...<-chan model.Measurement) <-chan model.Measurement {
	out := make(chan model.Measurement)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan model.Measurement) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Handle handles the download subtest.
func (dl Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logging.Logger.Debug("Upgrading to WebSockets")
	if request.Header.Get("Sec-WebSocket-Protocol") != spec.SecWebSocketProtocol {
		warnAndClose(writer, "Missing Sec-WebSocket-Protocol in request")
		return
	}
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", spec.SecWebSocketProtocol)
	conn, err := dl.Upgrader.Upgrade(writer, request, headers)
	if err != nil {
		warnAndClose(writer, "Cannnot UPGRADE to WebSocket")
		return
	}
	// TODO(bassosimone): an error before this point means that the *os.File
	// will stay in cache until the cache pruning mechanism is triggered. This
	// should be a small amount of seconds. If Golang does not call shutdown(2)
	// and close(2), we'll end up keeping sockets that caused an error in the
	// code above (e.g. because the handshake was not okay) alive for the time
	// in which the corresponding *os.File is kept in cache.
	defer warnonerror.Close(conn, "Ignoring close connection error")
	resultfp, err := results.OpenFor(request, conn, dl.DataDir, "download")
	if err != nil {
		return // error already printed
	}
	defer warnonerror.Close(resultfp, "Ignoring close results file error")
	sender := sender.Start(conn, measurer.Start(request.Context(), conn))
	receiver := receiver.Start(request.Context(), conn)
	for measurement := range(merge(sender, receiver)) {
		err := resultfp.WriteMeasurement(measurement, measurement.Origin)
		if err != nil {
			logging.Logger.WithError(err).Warn("Cannot save measurement on disk")
			return
		}
	}
}
