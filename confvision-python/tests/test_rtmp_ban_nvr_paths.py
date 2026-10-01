"""NVR: vários paths no mesmo IP não devem somar falhas em um único contador."""

import tempfile
import unittest

from rtmp_ban import BanStore


class TestBanStoreNvrPerPath(unittest.TestCase):
    def test_varios_path_invalido_mesmo_ip_nao_bane_rapido(self):
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            ban_path = f.name
        bans = BanStore(ban_path, max_fails=3, window_sec=60)
        ip = "177.234.172.73"
        paths = [
            "cam/dkn59rq6yjvq",
            "cam/wa3oy4bb9g4k",
            "cam/boezyjkm9dlk",
            "cam/qzg7yo82powa",
        ]
        for p in paths:
            bans.registrar_falha(ip, motivo="path_invalido", path=p)
        self.assertIsNone(bans.get(ip), "1 falha por path distinto não deve auto-banir IP")

    def test_tres_falhas_mesmo_path_invalido_bane(self):
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            ban_path = f.name
        bans = BanStore(ban_path, max_fails=3, window_sec=60)
        ip = "177.206.111.2"
        p = "live/1"
        for _ in range(3):
            bans.registrar_falha(ip, motivo="path_invalido", path=p)
        self.assertIsNotNone(bans.get(ip))


if __name__ == "__main__":
    unittest.main()
