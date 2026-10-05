"""Teste Postgres pool infra (upsert + status em transação com ROLLBACK)."""
from __future__ import annotations

import os
import sys

import psycopg2

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ENVCAM = os.path.join(ROOT, ".env")
TEST_FRA = "__infra_pool_test__"


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
    conn.autocommit = False
    cur = conn.cursor()
    try:
        cur.execute(
            """
            SELECT 1 FROM information_schema.tables
            WHERE table_name = 'vis_franqueado_mediamtx_pool'
            """
        )
        if not cur.fetchone():
            print("SKIP: tabelas pool (migration 022) nao existem neste Postgres")
            conn.rollback()
            return 0

        cur.execute("SELECT id FROM vis_mediamtx_node WHERE COALESCE(status, 'ativo') = 'ativo' LIMIT 1")
        row = cur.fetchone()
        if not row:
            print("SKIP: nenhum vis_mediamtx_node ativo")
            conn.rollback()
            return 0
        node_id = row[0]

        cur.execute(
            """
            INSERT INTO vis_franqueado_mediamtx_pool
                (id_franqueado, vis_mediamtx_node_id, max_cameras, ordem, ativo, updated_at)
            VALUES (%s, %s, 5, 1, true, NOW())
            ON CONFLICT (id_franqueado, vis_mediamtx_node_id) DO UPDATE SET ativo = true
            RETURNING id
            """,
            (TEST_FRA, node_id),
        )
        cur.execute(
            """
            INSERT INTO vis_franqueado_worker_pool
                (id_franqueado, worker_id, worker_tipo, max_cameras, ordem, ativo, updated_at)
            VALUES (%s, 'worker-vps1', 'analitico', 30, 1, true, NOW())
            ON CONFLICT (id_franqueado, worker_id, worker_tipo) DO UPDATE SET ativo = true
            RETURNING id
            """,
            (TEST_FRA,),
        )
        cur.execute(
            "SELECT COUNT(*) FROM vis_franqueado_mediamtx_pool WHERE id_franqueado = %s AND ativo IS TRUE",
            (TEST_FRA,),
        )
        n_mtx = cur.fetchone()[0]
        cur.execute(
            """
            SELECT COUNT(*) FROM vis_franqueado_worker_pool
            WHERE id_franqueado = %s AND ativo IS TRUE AND worker_tipo = 'analitico'
            """,
            (TEST_FRA,),
        )
        n_wrk = cur.fetchone()[0]
        ok = n_mtx >= 1 and n_wrk >= 1
        print(f"upsert OK mtx={n_mtx} worker={n_wrk} infra_completa={ok}")
        if not ok:
            conn.rollback()
            return 1
        conn.rollback()
        print("ROLLBACK — nenhum dado de teste persistido")
        return 0
    except Exception as exc:
        conn.rollback()
        print(f"ERRO: {exc}")
        return 1
    finally:
        cur.close()
        conn.close()


if __name__ == "__main__":
    raise SystemExit(main())
