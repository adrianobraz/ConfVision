"""Read-only: pré-requisitos cadastro câmera (licença, pool, capacidade)."""
from __future__ import annotations

import os
import sys

import psycopg2

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ENVCAM = os.path.join(ROOT, ".env")

FID = "2026082511344085584850570"
LIC = 390


def load_env(path: str) -> None:
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, v = line.split("=", 1)
            os.environ.setdefault(k.strip(), v.strip())


def main() -> int:
    load_env(ENVCAM)
    pg = os.environ.get("POSTGRES_URL")
    if not pg:
        print("POSTGRES_URL ausente")
        return 1
    conn = psycopg2.connect(pg)
    cur = conn.cursor()

    cur.execute(
        """
        SELECT id, status, id_franqueado, plano, unidade, valido_ate, vis_camera_id
        FROM vis_licenca WHERE id = %s
        """,
        (LIC,),
    )
    lic = cur.fetchone()
    print("=== licenca", LIC, "===")
    print(lic)

    cur.execute(
        """
        SELECT COUNT(1) FROM vis_franqueado_mediamtx_pool p
        JOIN vis_mediamtx_node n ON n.id = p.vis_mediamtx_node_id
        WHERE p.id_franqueado = %s AND p.ativo IS TRUE
          AND COALESCE(n.status, 'ativo') = 'ativo'
        """,
        (FID,),
    )
    print("pool_mtx_ativo", cur.fetchone()[0])

    cur.execute(
        """
        SELECT COUNT(1) FROM vis_franqueado_worker_pool
        WHERE id_franqueado = %s AND ativo IS TRUE AND worker_tipo = 'analitico'
        """,
        (FID,),
    )
    print("pool_worker_analitico", cur.fetchone()[0])

    cur.execute(
        """
        SELECT id, nome, vis_licenca_id, ativo FROM vis_camera
        WHERE id_franqueado = %s AND (nome ILIKE %s OR vis_licenca_id = %s)
        ORDER BY id
        """,
        (FID, "%teste adriano%", LIC),
    )
    print("cameras relacionadas", cur.fetchall())

    cur.execute(
        """
        SELECT id, status FROM vis_licenca
        WHERE id_franqueado = %s AND status = 'disponivel' AND unidade = 'camera'
        ORDER BY id LIMIT 5
        """,
        (FID,),
    )
    print("licencas disponiveis (5)", cur.fetchall())

    conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
