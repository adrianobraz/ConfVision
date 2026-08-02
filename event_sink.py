"""Camada unificada de persistencia de eventos: Xano | Postgres | dual."""

from __future__ import annotations

from typing import Any, Optional

from config import EVENT_STORE


def _use_postgres() -> bool:
    return EVENT_STORE in ("postgres", "dual")


def _use_xano_immediate() -> bool:
    return EVENT_STORE in ("xano", "dual")


def create_evento(camera, confianca: float, tipo: str = "humano", status: str = "capturando", extra=None):
    edge_row = None
    xano_evento = None

    if _use_postgres():
        import postgres_store

        if postgres_store.available():
            edge_row = postgres_store.insert_evento(camera, confianca, tipo=tipo, status=status)

    if _use_xano_immediate():
        from xano_client import create_evento as xano_create

        xano_evento = xano_create(camera, confianca, tipo=tipo, status=status, extra=extra)
        xid = None
        if isinstance(xano_evento, dict):
            xid = xano_evento.get("id") or (xano_evento.get("dados") or {}).get("id")
        if edge_row and xid:
            import postgres_store

            postgres_store.update_evento(edge_row["id"], {"xano_evento_id": int(xid)})
        return xano_evento if xano_evento else {"id": edge_row["id"], "edge": True, **edge_row}

    if edge_row:
        return {"id": edge_row["id"], "edge": True, **edge_row}
    from xano_client import create_evento as xano_create

    return xano_create(camera, confianca, tipo=tipo, status=status, extra=extra)


def put_evento(evento_id: int, campos: dict[str, Any], base: Optional[dict] = None, edge: bool = False):
    if edge and _use_postgres():
        import postgres_store

        if postgres_store.available():
            return postgres_store.update_evento(evento_id, campos)

    if _use_xano_immediate() and not edge:
        from xano_client import put_evento as xano_put

        return xano_put(evento_id, campos, base=base)

    if edge and _use_postgres():
        import postgres_store

        return postgres_store.update_evento(evento_id, campos)

    from xano_client import put_evento as xano_put

    return xano_put(evento_id, campos, base=base)


def finalizar_evento(
    evento_id: int,
    *,
    snapshot_url: Optional[str],
    video_url: Optional[str],
    status: str = "pronto",
    clip_count: int = 0,
    clip_duracao_seg: Optional[int] = None,
    processado: bool = False,
    edge: bool = False,
):
    campos = {
        "snapshot_url": snapshot_url,
        "video_url": video_url,
        "status": status,
        "clip_count": clip_count,
        "processado": processado,
    }

    if edge and _use_postgres():
        import postgres_store

        if postgres_store.available():
            postgres_store.update_evento(evento_id, campos)
            if video_url:
                postgres_store.insert_clip(
                    evento_id, 1, video_url, clip_duracao_seg or 0, snapshot_url
                )

    if _use_xano_immediate() and not edge:
        from xano_client import finalizar_evento as xano_fin

        return xano_fin(
            evento_id,
            snapshot_url=snapshot_url,
            video_url=video_url,
            status=status,
            clip_count=clip_count,
            clip_duracao_seg=clip_duracao_seg,
            processado=processado,
        )

    if edge and EVENT_STORE == "dual":
        import postgres_store

        row = postgres_store.update_evento(evento_id, campos) if postgres_store.available() else {}
        xano_id = row.get("xano_evento_id") if isinstance(row, dict) else None
        if xano_id:
            from xano_client import finalizar_evento as xano_fin

            return xano_fin(
                int(xano_id),
                snapshot_url=snapshot_url,
                video_url=video_url,
                status=status,
                clip_count=clip_count,
                clip_duracao_seg=clip_duracao_seg,
                processado=processado,
            )
        return row

    if edge and EVENT_STORE == "postgres":
        return campos

    from xano_client import finalizar_evento as xano_fin

    return xano_fin(
        evento_id,
        snapshot_url=snapshot_url,
        video_url=video_url,
        status=status,
        clip_count=clip_count,
        clip_duracao_seg=clip_duracao_seg,
        processado=processado,
    )


def evento_id_from_response(evento) -> int | None:
    if not evento:
        return None
    if isinstance(evento, dict):
        if evento.get("edge") and evento.get("id"):
            return int(evento["id"])
        if evento.get("id"):
            return int(evento["id"])
        dados = evento.get("dados")
        if isinstance(dados, dict) and dados.get("id"):
            return int(dados["id"])
    return None


def is_edge_evento(evento) -> bool:
    return bool(isinstance(evento, dict) and evento.get("edge") and EVENT_STORE == "postgres")
