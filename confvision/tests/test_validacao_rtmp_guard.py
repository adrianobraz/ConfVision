"""Validação local: log rotate, rate-limit log ip_banido, watch DTS/track."""

from __future__ import annotations

import os
import tempfile
import time
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from log_rotate import log_rotate_config, maybe_rotate_log
from rtmp_ban import BanStore
from rtmp_guard import RtmpGuard
from rtmp_guard_main import _should_log_ban_deny
from rtmp_token import chave_rtmp, stream_path_from_chave
from rtmp_watch import RtmpLogParser, RtmpWatchStore
from stream_health_client import reset_pause_dedupe_for_tests


class TestLogRotate(unittest.TestCase):
    def test_small_file_no_rotate(self):
        with tempfile.TemporaryDirectory() as td:
            p = Path(td) / "mediamtx.log"
            p.write_text("x" * 100, encoding="utf-8")
            os.environ["RTMP_LOG_MAX_BYTES"] = "10000"
            self.assertFalse(maybe_rotate_log(str(p), min_interval_sec=0))

    def test_over_limit_rotates_and_truncates(self):
        with tempfile.TemporaryDirectory() as td:
            p = Path(td) / "mediamtx.log"
            content = "line\n" * 5000
            p.write_text(content, encoding="utf-8")
            os.environ["RTMP_LOG_MAX_BYTES"] = "500"
            os.environ["RTMP_LOG_KEEP_ROTATED"] = "2"
            self.assertTrue(maybe_rotate_log(str(p), min_interval_sec=0))
            self.assertLess(p.stat().st_size, 20)
            backup = p.with_name("mediamtx.log.1")
            self.assertTrue(backup.is_file())
            self.assertGreater(backup.stat().st_size, 400)

    def test_config_defaults(self):
        os.environ.pop("RTMP_LOG_MAX_BYTES", None)
        os.environ.pop("RTMP_LOG_KEEP_ROTATED", None)
        mx, keep = log_rotate_config()
        self.assertEqual(mx, 52_428_800)
        self.assertEqual(keep, 3)


class TestBanDenyLogRateLimit(unittest.TestCase):
    def test_ip_banido_throttles_log_only(self):
        ip = "203.0.113.50"
        t0 = 1_000_000.0
        with patch("rtmp_guard_main.time.time", return_value=t0):
            self.assertTrue(_should_log_ban_deny(ip, "ip_banido"))
            self.assertFalse(_should_log_ban_deny(ip, "ip_banido"))
        with patch("rtmp_guard_main.time.time", return_value=t0 + 61):
            self.assertTrue(_should_log_ban_deny(ip, "ip_banido"))

    def test_other_motivos_always_log(self):
        self.assertTrue(_should_log_ban_deny("1.2.3.4", "path_invalido"))

    def test_ban_still_blocks_after_throttle(self):
        os.environ.setdefault("RTMP_PUBLISH_SECRET", "test-secret-for-unit-tests")
        bans = BanStore("/tmp/rtmp_bans_unused_validation.json", max_fails=1, window_sec=60)
        bans.ban("203.0.113.99", motivo="manual", manual=True, ttl_sec=3600)
        guard = RtmpGuard(bans)
        code, motivo, _ = guard.authorize({"action": "publish", "ip": "203.0.113.99", "path": "live/1"})
        self.assertEqual(code, 403)
        self.assertEqual(motivo, "ip_banido")

    def test_rtmp_auth_get_sends_worker_key_to_go(self):
        os.environ["RTMP_PUBLISH_SECRET"] = "test-secret-for-unit-tests"
        secret = os.environ["RTMP_PUBLISH_SECRET"]
        camera_id = 42
        chave = chave_rtmp(camera_id, secret=secret)
        path = stream_path_from_chave(chave)
        os.environ["CONFVISION_API_URL"] = "http://127.0.0.1:59999"
        os.environ["VIS_WORKER_API_KEY"] = "unit-test-worker-key"
        bans = BanStore("/tmp/rtmp_bans_auth_headers.json", max_fails=99, window_sec=60)
        guard = RtmpGuard(bans)
        guard.cache.invalidate(camera_id)
        cam_payload = {
            "id": camera_id,
            "ativo": True,
            "plano": "analitico_24h_foto",
            "bloqueado": False,
            "stream_motivo_pausa": "",
            "analitico_pausado": False,
        }
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {"dados": cam_payload}
        mock_resp.raise_for_status = lambda: None
        with patch("rtmp_guard.requests.get", return_value=mock_resp) as mock_get:
            code, motivo, _ = guard.authorize(
                {"action": "publish", "ip": "198.51.100.3", "path": path}
            )
        self.assertEqual(code, 200)
        self.assertEqual(motivo, "publish_ok")
        mock_get.assert_called_once()
        headers = mock_get.call_args.kwargs.get("headers") or {}
        self.assertEqual(headers.get("X-Vis-Worker-Key"), "unit-test-worker-key")
        self.assertEqual(headers.get("Authorization"), "Bearer unit-test-worker-key")

    def test_rtmp_auth_401_does_not_increment_ip_fail(self):
        os.environ["RTMP_PUBLISH_SECRET"] = "test-secret-for-unit-tests"
        secret = os.environ["RTMP_PUBLISH_SECRET"]
        camera_id = 43
        chave = chave_rtmp(camera_id, secret=secret)
        path = stream_path_from_chave(chave)
        os.environ["CONFVISION_API_URL"] = "http://127.0.0.1:59999"
        os.environ["VIS_WORKER_API_KEY"] = "wrong-key"
        os.environ.pop("XANO_BASE_URL", None)
        bans = BanStore("/tmp/rtmp_bans_auth_401.json", max_fails=1, window_sec=60)
        guard = RtmpGuard(bans)
        guard.cache.invalidate(camera_id)
        mock_resp = MagicMock()
        mock_resp.status_code = 401
        with patch("rtmp_guard.requests.get", return_value=mock_resp):
            code, motivo, _ = guard.authorize(
                {"action": "publish", "ip": "198.51.100.4", "path": path}
            )
        self.assertEqual(code, 503)
        self.assertEqual(motivo, "rtmp_auth_nao_autorizado")
        self.assertFalse(bans.is_banned("198.51.100.4"))

    def test_banned_ip_skips_camera_fetch(self):
        os.environ["RTMP_PUBLISH_SECRET"] = "test-secret-for-unit-tests"
        secret = os.environ["RTMP_PUBLISH_SECRET"]
        chave = chave_rtmp(42, secret=secret)
        path = stream_path_from_chave(chave)
        bans = BanStore("/tmp/rtmp_bans_no_fetch.json", max_fails=1, window_sec=60)
        bans.ban("198.51.100.1", motivo="manual", manual=True, ttl_sec=3600)
        guard = RtmpGuard(bans)
        with patch("rtmp_guard.requests.get") as mock_get:
            code, motivo, meta = guard.authorize(
                {"action": "publish", "ip": "198.51.100.1", "path": path}
            )
        self.assertEqual(code, 403)
        self.assertEqual(motivo, "ip_banido")
        self.assertEqual(meta.get("camera_id"), "")
        mock_get.assert_not_called()


class TestRtmpWatchDiag(unittest.TestCase):
    def setUp(self):
        self.store = RtmpWatchStore(max_items=50, dedupe_sec=5)
        self.parser = RtmpLogParser(self.store)

    def test_dts_line_classified(self):
        line = (
            "2026/09/30 12:00:00 WAR [HLS] [muxer cam/abc123def456] "
            "DTS is not monotonically increasing"
        )
        falha = self.parser.feed(line)
        self.assertIsNotNone(falha)
        assert falha is not None
        self.assertEqual(falha.motivo_codigo, "hls_dts_nao_monotono")

    def test_video_track_not_set_up(self):
        line = (
            "2026/09/30 12:00:01 INF [RTMP] [conn 1.2.3.4:1234] "
            "received a packet for video track 0, but track is not set up"
        )
        falha = self.parser.feed(line)
        self.assertIsNotNone(falha)
        assert falha is not None
        self.assertEqual(falha.motivo_codigo, "rtmp_video_track_nao_configurado")

    def test_normal_rtmp_open_no_crash(self):
        line = "2026/09/30 12:00:02 INF [RTMP] [conn 1.2.3.4:5678] opened"
        self.assertIsNone(self.parser.feed(line))


class TestVideoTrackPauseIntegration(unittest.TestCase):
    def setUp(self):
        reset_pause_dedupe_for_tests()
        os.environ["RTMP_PUBLISH_SECRET"] = "test-secret-for-unit-tests"
        secret = os.environ["RTMP_PUBLISH_SECRET"]
        self.camera_id = 77
        chave = chave_rtmp(self.camera_id, secret=secret)
        self.path = stream_path_from_chave(chave)
        from rtmp_guard_main import PARSER, GUARD

        self.parser = PARSER
        GUARD.cache.invalidate(self.camera_id)

    def test_video_track_triggers_pause_once(self):
        from rtmp_guard_main import GUARD

        os.environ["CONFVISION_API_URL"] = "http://127.0.0.1:59999"
        os.environ["VIS_WORKER_API_KEY"] = "test-key"
        line = (
            f"2026/09/30 12:00:01 INF [path {self.path}] "
            "received a packet for video track 0, but track is not set up"
        )
        with patch("stream_health_client.requests.post") as mock_post:
            mock_post.return_value.status_code = 200
            falha = self.parser.feed(line)
            self.assertIsNotNone(falha)
            assert falha is not None
            self.assertEqual(falha.motivo_codigo, "rtmp_video_track_nao_configurado")
            self.assertEqual(mock_post.call_count, 1)
            body = mock_post.call_args.kwargs.get("json") or mock_post.call_args[1].get("json")
            health = body["camera_stream_health"][0]
            self.assertEqual(health["event"], "pause_analytic")
            self.assertEqual(health["camera_id"], self.camera_id)
            self.assertEqual(health["pause_reason"], "sistema_stream_video_track_not_set_up")
            self.assertEqual(health["error_class"], "VIDEO_TRACK_NOT_SET_UP")

            mock_post.reset_mock()
            self.parser.feed(line)
            mock_post.assert_not_called()

        self.assertIsNone(GUARD.cache.get(self.camera_id))

    def test_video_track_blocks_next_publish(self):
        from rtmp_guard_main import GUARD

        os.environ["CONFVISION_API_URL"] = "http://127.0.0.1:59999"
        os.environ["VIS_WORKER_API_KEY"] = "test-key"
        GUARD.cache.invalidate(self.camera_id)
        GUARD.cache.set(
            self.camera_id,
            {"id": self.camera_id, "ativo": True, "plano": "analitico_24h_foto", "bloqueado": False},
        )
        line = (
            f"2026/09/30 12:00:01 INF [path {self.path}] "
            "received a packet for video track 0, but track is not set up"
        )
        paused_cam = {
            "id": self.camera_id,
            "ativo": True,
            "plano": "analitico_24h_foto",
            "bloqueado": False,
            "analitico_pausado": True,
            "stream_motivo_pausa": "sistema_stream_video_track_not_set_up",
        }
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {"dados": paused_cam}
        mock_resp.raise_for_status = lambda: None
        with patch("stream_health_client.requests.post") as mock_post:
            mock_post.return_value.status_code = 200
            self.parser.feed(line)
        with patch("rtmp_guard.requests.get", return_value=mock_resp):
            code, motivo, _ = GUARD.authorize(
                {"action": "publish", "ip": "198.51.100.2", "path": self.path}
            )
        self.assertEqual(code, 403)
        self.assertEqual(motivo, "stream_pausado_sistema")

    def test_video_track_resolves_path_from_conn_ip(self):
        from rtmp_guard_main import PARSER

        ip = "203.0.113.10"
        PARSER.register_publish_context(ip, self.path, fonte="test")
        line = (
            f"2026/09/30 12:00:01 INF [RTMP] [conn {ip}:44001] closed: "
            "received a packet for video track 0, but track is not set up"
        )
        falha = PARSER.feed(line)
        self.assertIsNotNone(falha)
        assert falha is not None
        self.assertEqual(falha.path, self.path)
        self.assertEqual(falha.motivo_codigo, "rtmp_video_track_nao_configurado")

    def test_video_track_triggers_pause_from_auth_context_only(self):
        from rtmp_guard_main import GUARD, PARSER

        os.environ["CONFVISION_API_URL"] = "http://127.0.0.1:59999"
        os.environ["VIS_WORKER_API_KEY"] = "test-key"
        ip = "198.51.100.88"
        PARSER.register_publish_context(ip, self.path, fonte="auth")
        line = (
            f"2026/09/30 12:00:01 INF [RTMP] [conn {ip}:44002] "
            "received a packet for video track 0, but track is not set up"
        )
        with patch("stream_health_client.requests.post") as mock_post:
            mock_post.return_value.status_code = 200
            PARSER.feed(line)
            mock_post.assert_called_once()
        self.assertIsNone(GUARD.cache.get(self.camera_id))

    def test_dts_does_not_pause(self):
        line = (
            "2026/09/30 12:00:00 WAR [HLS] [muxer cam/abc123def456] "
            "DTS is not monotonically increasing"
        )
        with patch("stream_health_client.requests.post") as mock_post:
            self.parser.feed(line)
            mock_post.assert_not_called()


if __name__ == "__main__":
    unittest.main()
