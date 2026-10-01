import tempfile
import unittest

from rtmp_ban import BanStore


class TestBanStoreSoftPath(unittest.TestCase):
    def test_sucesso_outro_path_nao_zera_eof(self):
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        bans = BanStore(
            path,
            soft_max_fails=8,
            soft_window_sec=300,
            soft_ban_ttl_sec=600,
        )
        ip = "45.235.195.118"
        p3 = "cam/dkn59r69jvqr"
        p4 = "cam/boezyjmydlk4"

        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p3)
        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p3)
        self.assertIsNone(bans.get(ip))

        bans.registrar_sucesso(ip, path=p4)
        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p3)
        # Ainda na 3ª falha do path problemático (não voltou a 1).
        with bans._lock:
            times = bans._fails[ip]["eof_sem_publish"][p3]
        self.assertEqual(len(times), 3)

    def test_sucesso_mesmo_path_zera_eof(self):
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        bans = BanStore(path, soft_max_fails=8, soft_window_sec=300)
        ip = "1.2.3.4"
        p = "cam/abc"

        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p)
        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p)
        bans.registrar_sucesso(ip, path=p)
        bans.registrar_falha(ip, motivo="eof_sem_publish", path=p)
        with bans._lock:
            times = bans._fails[ip]["eof_sem_publish"][p]
        self.assertEqual(len(times), 1)

    def test_auth_hard_ainda_limpa_no_publish(self):
        with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
            path = f.name
        bans = BanStore(path, max_fails=3, window_sec=60)
        ip = "9.9.9.9"
        bans.registrar_falha(ip, motivo="chave_invalida")
        bans.registrar_falha(ip, motivo="chave_invalida")
        bans.registrar_sucesso(ip, path="cam/x")
        self.assertNotIn(ip, bans._fails)


if __name__ == "__main__":
    unittest.main()
