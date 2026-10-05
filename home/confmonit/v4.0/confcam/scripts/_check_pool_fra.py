import os
import psycopg2

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
with open(os.path.join(ROOT, ".env"), encoding="utf-8") as f:
    for line in f:
        line = line.strip()
        if line and not line.startswith("#") and "=" in line:
            k, v = line.split("=", 1)
            os.environ.setdefault(k.strip(), v.strip())

fid = "2026072204185539042285696"
conn = psycopg2.connect(os.environ["POSTGRES_URL"])
cur = conn.cursor()
cur.execute(
    "SELECT COUNT(*) FROM vis_franqueado_mediamtx_pool WHERE id_franqueado = %s AND ativo IS TRUE",
    (fid,),
)
print("mtx", cur.fetchone()[0])
cur.execute(
    """
    SELECT COUNT(*) FROM vis_franqueado_worker_pool
    WHERE id_franqueado = %s AND ativo IS TRUE AND worker_tipo = 'analitico'
    """,
    (fid,),
)
print("worker", cur.fetchone()[0])
conn.close()
