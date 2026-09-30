package visdata

import (
	"confvision/src/config"
	"database/sql"
	"testing"
)

func TestDefaultMediamtxRtspURLSec(t *testing.T) {
	config.RtmpPublishSecret = "test-salt-confvision"
	config.MediamtxRtspBase = "rtsp://srv1.example:8554"
	t.Cleanup(func() {
		config.RtmpPublishSecret = ""
		config.MediamtxRtspBase = ""
	})

	node := &MediamtxNode{
		RtspInternal: sql.NullString{String: "rtsp://mtx-internal:8554", Valid: true},
	}
	url := DefaultMediamtxRtspURLSecForNode(21, node)
	if url == "" {
		t.Fatal("expected url")
	}
	if want := "rtsp://mtx-internal:8554/"; url[:len(want)] != want {
		t.Fatalf("base node: got %q", url)
	}
	if !shouldAutoFillRtspURLSec("") {
		t.Fatal("empty should auto fill")
	}
	if shouldAutoFillRtspURLSec("rtsp://dvr/cam/1") {
		t.Fatal("explicit rtsp should not auto")
	}
	if !shouldAutoFillRtspURLSec("rtsp://x/live/1") {
		t.Fatal("legacy live should auto replace")
	}
}
