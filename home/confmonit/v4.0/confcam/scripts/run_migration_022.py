"""Aplica sql/022_franqueado_infra_pool.sql no POSTGRES_URL do .env ConfCam."""
from __future__ import annotations

import os
import sys

import psycopg2

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ENV = os.path.join(ROOT, ".env")
SQL = os.path.join(ROOT, "sql", "022_franqueado_infra_pool.sql")


def load_env() -> None:
    with open(ENV, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, v = line.split("=", 1)
            os.environ.setdefault(k.strip(), v.strip())


def main() -> int:
    load_env()
    pg = os.environ.get("POSTGRES_URL")
    if not pg:
        print("POSTGRES_URL ausente no .env", file=sys.stderr)
        return 2
    sql = open(SQL, encoding="utf-8").read()
    conn = psycopg2.connect(pg)
    conn.autocommit = True
    cur = conn.cursor()
    cur.execute(sql)
    cur.execute(
        """
        SELECT table_name FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name IN ('vis_franqueado_mediamtx_pool', 'vis_franqueado_worker_pool')
        ORDER BY 1
        """
    )
    tables = [r[0] for r in cur.fetchall()]
    cur.execute("SELECT COUNT(*) FROM vis_franqueado_mediamtx_pool")
    n_mtx = cur.fetchone()[0]
    cur.execute("SELECT COUNT(*) FROM vis_franqueado_worker_pool")
    n_wrk = cur.fetchone()[0]
    conn.close()
    print("Migration 022 executada com sucesso.")
    print("Tabelas presentes:", ", ".join(tables) if tables else "(nenhuma)")
    print(f"Registros: mediamtx_pool={n_mtx}, worker_pool={n_wrk}")
    return 0 if len(tables) == 2 else 1


if __name__ == "__main__":
    raise SystemExit(main())
