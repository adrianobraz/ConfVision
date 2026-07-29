"""Parseia logs do MediaMTX e grava falhas RTMP com mensagens amigáveis."""

from __future__ import annotations

import json
import os
import re
import threading
import time
from collections import deque
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Deque, Optional

from rtmp_messages import classificar_motivo, mensagem_amigavel
from rtmp_token import parse_chave_rtmp

RE_LINE = re.compile(
    r"^(?P<ts>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+"
    r"(?P<level>\w+)\s+"
    r"(?:\[(?P<proto>[^\]]+)\]\s+)?"
    r"(?:\[(?P<kind>conn|session|muxer|path) (?P<sub>[^\]]+)\]\s+)?"
    r"(?P<msg>.*)$"
)

RE_CONN = re.compile(r"^(?P<ip>\d{1,3}(?:\.\d{1,3}){3}):(?P<port>\d+)$")
RE_PUBLISH = re.compile(r"is publishing to path '([^']+)'")
RE_PATH_ONLINE = re.compile(r"^\[path (?P<path>[^\]]+)\] stream is available")
RE_MUXER_DESTROY = re.compile(r"^\[HLS\] \[muxer (?P<path>[^\]]+)\] destroyed")
RE_CLOSED = re.compile(r"^closed:\s*(?P<reason>.+)$")
RE_STREAM_PATH = re.compile(r"([0-9a-z]{12,})")


def path_label(path: str) -> str:
    """Path legível: inclui id da câmera quando for Hashids válido."""
    p = (path or "").strip().rstrip("/")
    if not p:
        return ""
    nome = p.rsplit("/", 1)[-1]
    cam = parse_chave_rtmp(nome)
    if cam:
        return f"{p} (câmera {cam})"
    if nome.isdigit():
        return f"{p} (câmera {int(nome)})"
    return p


@dataclass
class ConnState:
    ip: str
    port: str
    opened_at: str
    published: bool = False
    path: str = ""


@dataclass
class Falha:
    id: str
    ts: str
    ip: str
    porta: str
    protocolo: str
    path: str
    motivo_codigo: str
    titulo: str
    dica: str
    gravidade: str
    motivo_raw: str
    vezes: int = 1
    ts_ultimo: str = ""

    def to_dict(self) -> dict:
        d = asdict(self)
        if not d.get("ts_ultimo"):
            d["ts_ultimo"] = d["ts"]
        d["path_label"] = path_label(d.get("path") or "")
        return d


class RtmpWatchStore:
    def __init__(self, max_items: int = 500, dedupe_sec: int = 60):
        self.max_items = max(50, max_items)
        self.dedupe_sec = max(5, dedupe_sec)
        self._items: Deque[Falha] = deque(maxlen=self.max_items)
        self._lock = threading.Lock()
        self._seq = 0

    def add(
        self,
        *,
        ts: str,
        ip: str,
        porta: str,
        protocolo: str,
        path: str,
        motivo_raw: str,
        forcar_codigo: str = "",
    ) -> Optional[Falha]:
        codigo = forcar_codigo or classificar_motivo(motivo_raw)
        # Shutdown do servidor não é falha do cliente
        if codigo == "terminated":
            return None

        titulo, dica, gravidade = mensagem_amigavel(codigo)
        with self._lock:
            if self._items:
                last = self._items[0]
                if (
                    last.ip == ip
                    and last.motivo_codigo == codigo
                    and last.path == (path or last.path)
                    and self._mesmo_minuto(last.ts_ultimo or last.ts, ts)
                ):
                    last.vezes += 1
                    last.ts_ultimo = ts
                    last.motivo_raw = motivo_raw
                    return last

            self._seq += 1
            falha = Falha(
                id=f"{int(time.time())}-{self._seq}",
                ts=ts,
                ip=ip or "?",
                porta=porta or "",
                protocolo=protocolo or "RTMP",
                path=path or "",
                motivo_codigo=codigo,
                titulo=titulo,
                dica=dica,
                gravidade=gravidade,
                motivo_raw=motivo_raw,
                vezes=1,
                ts_ultimo=ts,
            )
            self._items.appendleft(falha)
            return falha

    @staticmethod
    def _mesmo_minuto(a: str, b: str) -> bool:
        # dedupe por janela aproximada (mesmo minuto no log MTX)
        return (a or "")[:16] == (b or "")[:16]

    def listar(self, limit: int = 100, ip: str = "", path: str = "") -> list[dict]:
        ip = (ip or "").strip()
        path = (path or "").strip()
        with self._lock:
            out = []
            for item in self._items:
                if ip and item.ip != ip:
                    continue
                if path and path not in (item.path or ""):
                    continue
                out.append(item.to_dict())
                if len(out) >= max(1, limit):
                    break
            return out

    def _resumo_unlocked(self) -> dict:
        por_codigo: dict[str, int] = {}
        por_ip: dict[str, int] = {}
        for item in self._items:
            por_codigo[item.motivo_codigo] = por_codigo.get(item.motivo_codigo, 0) + item.vezes
            por_ip[item.ip] = por_ip.get(item.ip, 0) + item.vezes
        top_ips = sorted(por_ip.items(), key=lambda x: x[1], reverse=True)[:20]
        return {
            "total_eventos": sum(i.vezes for i in self._items),
            "total_registros": len(self._items),
            "por_motivo": por_codigo,
            "top_ips": [{"ip": k, "vezes": v} for k, v in top_ips],
        }

    def resumo(self) -> dict:
        with self._lock:
            return self._resumo_unlocked()

    def salvar_json(self, path: str) -> None:
        p = Path(path)
        p.parent.mkdir(parents=True, exist_ok=True)
        with self._lock:
            payload = {
                "atualizado_em": datetime.now(timezone.utc).isoformat(),
                "falhas": [i.to_dict() for i in list(self._items)[:200]],
                "resumo": self._resumo_unlocked(),
            }
        tmp = p.with_suffix(".tmp")
        tmp.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
        tmp.replace(p)


class RtmpLogParser:
    def __init__(self, store: RtmpWatchStore):
        self.store = store
        self.conns: dict[str, ConnState] = {}
        # Último path visto por IP (publish) — útil em EOF sem path no log
        self.last_path_by_ip: dict[str, str] = {}
        # path -> {ip, since} para fallback da lista online
        self.publishers: dict[str, dict] = {}

    def feed(self, line: str) -> Optional[Falha]:
        line = (line or "").rstrip("\n")
        if not line.strip():
            return None

        # Formato alternativo: path online fora do RE_LINE padrão de conn
        m_online = re.search(
            r"^(?P<ts>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}).*\[path (?P<path>[^\]]+)\] stream is available",
            line,
        )
        if m_online:
            return None

        m = RE_LINE.match(line)
        if not m:
            return None

        ts = m.group("ts")
        proto = (m.group("proto") or "").strip()
        kind = (m.group("kind") or "").strip()
        sub = (m.group("sub") or "").strip()
        msg = (m.group("msg") or "").strip()

        if proto == "RTMP" and kind == "conn":
            return self._handle_rtmp_conn(ts, sub, msg)

        return None

    def _handle_rtmp_conn(self, ts: str, sub: str, msg: str) -> Optional[Falha]:
        cm = RE_CONN.match(sub)
        if not cm:
            return None
        ip = cm.group("ip")
        port = cm.group("port")
        key = f"{ip}:{port}"

        if msg == "opened":
            self.conns[key] = ConnState(ip=ip, port=port, opened_at=ts)
            return None

        pub = RE_PUBLISH.search(msg)
        if pub:
            st = self.conns.get(key) or ConnState(ip=ip, port=port, opened_at=ts)
            st.published = True
            st.path = pub.group(1).rstrip("/")
            self.conns[key] = st
            self.last_path_by_ip[ip] = st.path
            self.publishers[st.path] = {"ip": ip, "since": ts}
            print(
                f"[RTMP-WATCH] OK ip={ip} path={st.path} — publicando",
                flush=True,
            )
            return None

        closed = RE_CLOSED.match(msg)
        if closed or msg.startswith("closed"):
            reason = closed.group("reason") if closed else msg.replace("closed:", "").strip()
            st = self.conns.pop(key, None)
            path = (
                (st.path if st else "")
                or self._path_from_reason(reason)
                or self.last_path_by_ip.get(ip, "")
            )
            if st and st.published and st.path:
                cur = self.publishers.get(st.path)
                if cur and cur.get("ip") == ip:
                    self.publishers.pop(st.path, None)

            # Se publicou com sucesso e caiu depois, ainda é útil registrar (queda)
            if st and st.published and "terminated" not in reason.lower():
                falha = self.store.add(
                    ts=ts,
                    ip=ip,
                    porta=port,
                    protocolo="RTMP",
                    path=path or st.path,
                    motivo_raw=f"closed: {reason} (após publicar)",
                    forcar_codigo="muxer_destroyed"
                    if "eof" in reason.lower()
                    else "closed_outro",
                )
                if falha:
                    self._log_falha(falha)
                return falha

            # Nunca publicou → falha de conexão/publicação
            if not st or not st.published:
                codigo = classificar_motivo(f"closed: {reason}")
                if codigo == "eof_sem_publish" or "eof" in reason.lower():
                    codigo = "eof_sem_publish"
                falha = self.store.add(
                    ts=ts,
                    ip=ip,
                    porta=port,
                    protocolo="RTMP",
                    path=path,
                    motivo_raw=f"closed: {reason}",
                    forcar_codigo=codigo,
                )
                if falha:
                    self._log_falha(falha)
                return falha

        # Erros sem closed explícito na mesma linha (ex.: invalid path)
        if "invalid path" in msg.lower() or "authentication" in msg.lower():
            path = self._path_from_reason(msg) or self.last_path_by_ip.get(ip, "")
            if path:
                self.last_path_by_ip[ip] = path
            falha = self.store.add(
                ts=ts,
                ip=ip,
                porta=port,
                protocolo="RTMP",
                path=path,
                motivo_raw=msg,
            )
            if falha:
                self._log_falha(falha)
                self.conns.pop(key, None)
            return falha

        return None

    @staticmethod
    def _path_from_reason(reason: str) -> str:
        m = RE_STREAM_PATH.search(reason or "")
        if m:
            return m.group(1).rstrip("/")
        return ""

    @staticmethod
    def _log_falha(falha: Falha) -> None:
        print(
            f"[RTMP-WATCH] FALHA ip={falha.ip} path={falha.path or '-'} "
            f"vezes={falha.vezes} | {falha.titulo} | {falha.motivo_raw}",
            flush=True,
        )


def follow_file(path: str, parser: RtmpLogParser, store: RtmpWatchStore, json_out: str):
    """Tail -F do arquivo de log do MediaMTX."""
    p = Path(path)
    print(f"[RTMP-WATCH] aguardando log em {p}", flush=True)
    pos = 0
    inode = None
    last_save = 0.0
    # Primeira abertura: começa no fim (evita replay/auto-ban do histórico)
    started = False

    while True:
        try:
            if not p.exists():
                time.sleep(1)
                continue

            st = p.stat()
            if inode is None or st.st_ino != inode or st.st_size < pos:
                # arquivo novo / truncado / rotacionado
                inode = st.st_ino
                if not started:
                    pos = st.st_size
                    started = True
                    print(
                        f"[RTMP-WATCH] tail a partir do fim {p} (size={pos})",
                        flush=True,
                    )
                else:
                    pos = 0
                    print(f"[RTMP-WATCH] lendo {p} (size={st.st_size})", flush=True)

            with p.open("r", encoding="utf-8", errors="replace") as f:
                f.seek(pos)
                while True:
                    line = f.readline()
                    if not line:
                        break
                    parser.feed(line)
                pos = f.tell()

            now = time.time()
            if json_out and now - last_save >= 5:
                store.salvar_json(json_out)
                last_save = now

            time.sleep(0.4)
        except Exception as exc:
            print(f"[RTMP-WATCH] erro no follow: {exc}", flush=True)
            time.sleep(2)
