"""Auth RTMP: pausa sistema bloqueia publish sem auto-ban de IP."""

from __future__ import annotations

import os
import unittest
from unittest.mock import patch

from rtmp_ban import BanStore
from rtmp_guard import RtmpGuard

_PATCH = patch.multiple(
    "rtmp_guard",
    chave_valida=lambda *_a, **_k: True,
    parse_chave_rtmp=lambda *_a, **_k: 28,
)


class TestRtmpGuardStreamPolicy(unittest.TestCase):
    def setUp(self) -> None:
        self._patchers = _PATCH
        self._patchers.start()
        self.addCleanup(self._patchers.stop)
        os.environ.setdefault("RTMP_PUBLISH_SECRET", "test-secret-for-unit-tests")
        self.bans = BanStore("/tmp/rtmp_bans_test_unused.json", max_fails=3, window_sec=60)
        self.guard = RtmpGuard(self.bans)

    def _cam(self, **kwargs):
        base = {
            "id": 28,
            "ativo": True,
            "bloqueado": False,
            "plano": "analitico_armado_foto_video",
        }
        base.update(kwargs)
        return base

    def test_negado_stream_pausado_sistema(self) -> None:
        cam = self._cam(
            analitico_pausado=True,
            stream_motivo_pausa="sistema_stream_h264_nal",
        )
        self.guard.cache.set(28, cam)
        status, motivo, _meta = self.guard._authorize_camera_path(
            "177.0.0.1", "cam/boezyjkm9dlk", ok_motivo="publish_ok"
        )
        self.assertEqual(status, 403)
        self.assertEqual(motivo, "stream_pausado_sistema")
        self.assertFalse(self.bans.is_banned("177.0.0.1"))

    def test_ok_pausa_manual_sem_motivo_sistema(self) -> None:
        cam = self._cam(analitico_pausado=True, stream_motivo_pausa=None)
        self.guard.cache.set(28, cam)
        status, motivo, _meta = self.guard._authorize_camera_path(
            "177.0.0.2", "cam/boezyjkm9dlk", ok_motivo="publish_ok"
        )
        self.assertEqual(status, 200)
        self.assertEqual(motivo, "publish_ok")

    def test_camera_inativa_nao_ban_ip(self) -> None:
        with patch("rtmp_guard.parse_chave_rtmp", return_value=7):
            cam = self._cam(ativo=False, plano="analitico_24h_foto")
            self.guard.cache.set(7, cam)
            for _ in range(5):
                status, motivo, _ = self.guard._authorize_camera_path(
                    "179.0.0.1", "cam/bdm69e59ngae", ok_motivo="publish_ok"
                )
                self.assertEqual(status, 403)
                self.assertEqual(motivo, "camera_inativa")
            self.assertFalse(self.bans.is_banned("179.0.0.1"))


if __name__ == "__main__":
    unittest.main()
