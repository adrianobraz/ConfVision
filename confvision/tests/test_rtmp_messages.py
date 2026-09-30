"""Testes de classificação RTMP (boot Guard / connection refused)."""

import unittest

from rtmp_messages import classificar_motivo


class TestRtmpAuthClassify(unittest.TestCase):
    def test_connection_refused_on_guard_is_not_client_auth(self):
        raw = (
            "failed to authenticate: HTTP request failed: Post "
            '"http://127.0.0.1:8100/auth": dial tcp 127.0.0.1:8100: connect: connection refused'
        )
        self.assertEqual(classificar_motivo(raw), "auth_guard_indisponivel")

    def test_closed_connection_refused_same(self):
        raw = (
            "closed: failed to authenticate: HTTP request failed: Post "
            '"http://127.0.0.1:8100/auth": dial tcp 127.0.0.1:8100: connect: connection refused'
        )
        self.assertEqual(classificar_motivo(raw), "auth_guard_indisponivel")

    def test_real_invalid_credentials_stays_auth_falhou(self):
        raw = "failed to authenticate: authentication failed: invalid credentials"
        self.assertEqual(classificar_motivo(raw), "auth_falhou")


if __name__ == "__main__":
    unittest.main()
