// Package measurer contains the downloader measurer
package measurer

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m-lab/ndt-server/bbr"
	"github.com/m-lab/ndt-server/fdcache"
	"github.com/m-lab/ndt-server/logging"
	"github.com/m-lab/ndt-server/ndt7/model"
	"github.com/m-lab/ndt-server/ndt7/spec"
	"github.com/m-lab/ndt-server/tcpinfox"
)

// getSocketAndPossiblyEnableBBR returns the socket bound to the connection
// and possibly enables BBR. You own the returned socket. Failure in enabling
// BBR for the socket does not generate an error.
func getSocketAndPossiblyEnableBBR(conn *websocket.Conn) (*os.File, error) {
	fp := fdcache.GetAndForgetFile(conn.UnderlyingConn())
	// Implementation note: in theory fp SHOULD always be non-nil because
	// now we always register the fp bound to a net.TCPConn. However, in
	// some weird cases it MAY happen that the cache pruning mechanism will
	// remove the fp BEFORE we can steal it. In case we cannot get a file
	// we just abort the test, as this should not happen (TM).
	if fp == nil {
		err := errors.New("cannot get file bound to websocket conn")
		logging.Logger.WithError(err).Warn("fdcache.GetAndForgetFile failed")
		return nil, err
	}
	err := bbr.Enable(fp)
	if err != nil {
		logging.Logger.WithError(err).Warn("Cannot enable BBR")
		// FALLTHROUGH
	}
	return fp, nil
}

// measure performs a TCPInfo and BBR measurement using |sockfp|.
func measure(measurement *model.Measurement, sockfp *os.File) error {
	bbrinfo, err := bbr.GetMaxBandwidthAndMinRTT(sockfp)
	if err == nil {
		measurement.BBRInfo = &bbrinfo
	}
	metrics, err := tcpinfox.GetTCPInfo(sockfp)
	if err == nil {
		measurement.TCPInfo = &metrics
	}
	return nil
}

// maxDownloadTime is the maximum download time in nanosecond
const maxDownloadTime = 10 * time.Second

// Start starts the measurer in a background gorutine. The measurer will
// perform measurements and return them as internal messages on the returned
// channel. On error, the error is returned on the output channel, then the
// goroutine closes the channel and terminates. When the context is done,
// the channel is simply closed without returning any message.
func Start(ctx context.Context, conn *websocket.Conn) <-chan model.Measurement {
	dst := make(chan model.Measurement)
	go func() {
		defer close(dst)
		logging.Logger.Debug("measurer: start")
		defer logging.Logger.Debug("measurer: stop")
		sockfp, err := getSocketAndPossiblyEnableBBR(conn)
		if err != nil {
			return // error already printed
		}
		defer sockfp.Close()
		t0 := time.Now()
		ticker := time.NewTicker(spec.MinMeasurementInterval)
		defer ticker.Stop()
		dctx, cancel := context.WithTimeout(ctx, maxDownloadTime)
		defer cancel()
		for {
			select {
			case <-dctx.Done():
				logging.Logger.Debug("measurer: context done")
				return
			case now := <-ticker.C:
				elapsed := now.Sub(t0)
				measurement := model.Measurement{
					Elapsed: elapsed.Seconds(),
				}
				err = measure(&measurement, sockfp)
				if err != nil {
					return // error already printed
				}
				measurement.Origin = "server"
				dst <- measurement
			}
		}
	}()
	return dst
}
