"""Armazenamento edge de eventos em Postgres + sync assincrono para Xano."""

from __future__ import annotations

import threading
import time
import traceback
from pathlib import Path
from typing import Any, Optional

from config import (
    EVENT_STORE,
    EVENT_SYNC_BATCH_SIZE,
    EVENT_SYNC_INTERVAL_SEC,
    POSTGRES_URL,
)

_conn = None
_lock = threading.Lock()
_sync_thread: threading.Thread | None = None


def available() -> bool:
    return bool(POSTGRES_URL) and EVENT_STORE in ("postgres", "dual")


def _connect():
    global _conn
    if not POSTGRES_URL:
        return None
    import psycopg2
    import psycopg2.extras

    if _conn is None or _conn.closed:
        _conn = psycopg2.connect(POSTGRES_URL)
        _conn.autocommit = True
    return _conn


def init_schema() -> bool:
    if not available():
        return False
    schema_path = Path(__file__).resolve().parent / "sql" / "001_edge_schema.sql"
    sql = schema_path.read_text(encoding="utf-8")
    conn = _connect()
    if conn is None:
        return False
    with conn.cursor() as cur:
        cur.execute(sql)
    print("[POSTGRES] schema edge ok")
    return True


def insert_evento(camera: dict, confianca: float, tipo: str = "humano", status: str = "capturando") -> dict:
    conn = _connect()
    assert conn is not None
    import psycopg2.extras

    payload = {
        "vis_camera_id": camera["id"],
        "id_franqueado": camera.get("id_franqueado"),
        "id_cliente": camera.get("id_cliente"),
        "id_dispositivo": camera.get("id_dispositivo"),
        "conta": camera.get("conta"),
        "particao": camera.get("particao"),
        "canal": camera.get("canal"),
        "tipo_deteccao": tipo,
        "confianca": confianca,
        "status": status,
        "clip_count": 0,
        "processado": False,
    }
    with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
        cur.execute(
            """
            INSERT INTO vis_evento_edge (
                vis_camera_id, id_franqueado, id_cliente, id_dispositivo,
                conta, particao, canal, tipo_deteccao, confianca, status,
                clip_count, processado
            ) VALUES (
                %(vis_camera_id)s, %(id_franqueado)s, %(id_cliente)s, %(id_dispositivo)s,
                %(conta)s, %(particao)s, %(canal)s, %(tipo_deteccao)s, %(confianca)s, %(status)s,
                %(clip_count)s, %(processado)s
            )
            RETURNING *
            """,
            payload,
        )
        row = dict(cur.fetchone())
    return row


def update_evento(edge_id: int, campos: dict[str, Any]) -> dict:
    conn = _connect()
    assert conn is not None
    import psycopg2.extras

    allowed = {
        "snapshot_url",
        "video_url",
        "status",
        "clip_count",
        "processado",
        "id_evento",
        "id_processo",
        "xano_evento_id",
        "synced_at",
        "sync_error",
    }
    sets = []
    values: dict[str, Any] = {"id": edge_id}
    for key, val in campos.items():
        if key in allowed:
            sets.append(f"{key} = %({key})s")
            values[key] = val
    if not sets:
        return {"id": edge_id}
    sets.append("updated_at = NOW()")
    sql = f"UPDATE vis_evento_edge SET {', '.join(sets)} WHERE id = %(id)s RETURNING *"
    with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
        cur.execute(sql, values)
        row = cur.fetchone()
    return dict(row) if row else {"id": edge_id}


def insert_clip(edge_evento_id: int, seq: int, video_url: str, duracao_seg: int, snapshot_url: Optional[str]):
    conn = _connect()
    assert conn is not None
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO vis_evento_clip_edge (
                edge_evento_id, seq, video_url, duracao_seg, snapshot_url
            ) VALUES (%s, %s, %s, %s, %s)
            """,
            (edge_evento_id, seq, video_url, duracao_seg, snapshot_url),
        )


def fetch_unsynced(limit: int = 20) -> list[dict]:
    conn = _connect()
    assert conn is not None
    import psycopg2.extras

    with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
        cur.execute(
            """
            SELECT * FROM vis_evento_edge
            WHERE synced_at IS NULL AND status IN ('pronto', 'erro')
            ORDER BY created_at ASC
            LIMIT %s
            """,
            (limit,),
        )
        return [dict(r) for r in cur.fetchall()]


def _sync_one(row: dict) -> bool:
    from xano_client import create_evento, finalizar_evento, put_evento

    xano_id = row.get("xano_evento_id")
    if not xano_id:
        camera = {
            "id": row["vis_camera_id"],
            "id_franqueado": row.get("id_franqueado"),
            "id_cliente": row.get("id_cliente"),
            "id_dispositivo": row.get("id_dispositivo"),
            "conta": row.get("conta"),
            "particao": row.get("particao"),
            "canal": row.get("canal"),
        }
        created = create_evento(
            camera,
            float(row.get("confianca") or 0),
            tipo=row.get("tipo_deteccao") or "humano",
            status=row.get("status") or "capturando",
        )
        xano_id = created.get("id") or (created.get("dados") or {}).get("id")
        if not xano_id:
            update_evento(row["id"], {"sync_error": "xano_sem_id"})
            return False
        update_evento(row["id"], {"xano_evento_id": int(xano_id)})

    status = row.get("status") or "pronto"
    if status == "erro":
        put_evento(int(xano_id), {"status": "erro", "clip_count": 0})
    elif row.get("video_url"):
        finalizar_evento(
            int(xano_id),
            snapshot_url=row.get("snapshot_url"),
            video_url=row.get("video_url"),
            status=status,
            clip_count=int(row.get("clip_count") or 1),
            clip_duracao_seg=None,
        )
    else:
        put_evento(
            int(xano_id),
            {
                "snapshot_url": row.get("snapshot_url"),
                "status": status,
                "clip_count": int(row.get("clip_count") or 0),
                "processado": bool(row.get("processado")),
            },
        )

    update_evento(row["id"], {"synced_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())})
    return True


def sync_loop():
    if not available():
        return
    print(f"[POSTGRES-SYNC] iniciado batch={EVENT_SYNC_BATCH_SIZE} interval={EVENT_SYNC_INTERVAL_SEC}s")
    while True:
        try:
            rows = fetch_unsynced(EVENT_SYNC_BATCH_SIZE)
            for row in rows:
                try:
                    _sync_one(row)
                except Exception as exc:
                    update_evento(row["id"], {"sync_error": str(exc)[:500]})
                    print(f"[POSTGRES-SYNC] erro evento edge={row['id']}: {exc}")
        except Exception as exc:
            print(f"[POSTGRES-SYNC] loop erro: {exc}")
            traceback.print_exc()
        time.sleep(EVENT_SYNC_INTERVAL_SEC)


def start_sync_worker():
    global _sync_thread
    if not available() or EVENT_STORE != "dual":
        return
    with _lock:
        if _sync_thread is not None and _sync_thread.is_alive():
            return
        _sync_thread = threading.Thread(target=sync_loop, daemon=True, name="postgres-xano-sync")
        _sync_thread.start()
