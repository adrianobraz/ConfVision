"""Testes de backoff RTSP (sem rede)."""

import time

import rtsp_backoff as rb


def test_backoff_after_max_failures(monkeypatch):
    monkeypatch.setattr(rb, "RTSP_BACKOFF_ENABLED", True)
    monkeypatch.setattr(rb, "RTSP_OPEN_FAIL_MAX", 3)
    monkeypatch.setattr(rb, "RTSP_OPEN_BACKOFF_SEC", 60)
    monkeypatch.setattr(rb, "RTSP_OPEN_RETRY_SEC", 1)

    cid = 99991
    for i in range(1, 3):
        sec = rb.after_open_failure(cid, "rtsp://test/cam/x")
        assert sec == 1.0
    rb.after_open_failure(cid, "rtsp://test/cam/x")

    assert rb.wait_if_backoff(cid, "rtsp://test/cam/x", lambda: False) is True

    rb.after_open_success(cid)
    assert rb.wait_if_backoff(cid, "rtsp://test/cam/x") is False
