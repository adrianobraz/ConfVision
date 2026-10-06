"""Fase 3: sincroniza câmeras rtsp_pull → paths MediaMTX (sourceOnDemand)."""

from __future__ import annotations

import time
from typing import Any

import requests

from config import (
    MEDIAMTX_NODE_ID,
    RTSP_MTX_PATH_SYNC_ENABLED,
    RTSP_MTX_PATH_SYNC_POLL_SEC,
    XANO_BASE_URL,
)
from mtx_config import (
    mtx_path_config_add,
    mtx_path_config_delete,
    mtx_path_config_get,
    mtx_path_config_patch,
    rtsp_on_demand_path_conf,
)
from rtmp_token import chave_rtmp, stream_path_from_chave
from vis_api_auth import vis_api_headers


def _path_name(camera: dict[str, Any]) -> str:
    cid = int(camera.get("id") or 0)
    return stream_path_from_chave(chave_rtmp(cid))


def fetch_rtsp_pull_for_mtx_node(node_id: int) -> list[dict[str, Any]]:
    base = XANO_BASE_URL.rstrip("/")
    if not base:
        print("[RTSP-MTX] XANO_BASE_URL vazio", flush=True)
        return []
    if node_id < 1:
        print("[RTSP-MTX] MEDIAMTX_NODE_ID invalido (obrigatorio para sync)", flush=True)
        return []
    url = f"{base}/vis_camera_query_rtsp_pull"
    params = {"vis_mediamtx_node_id": str(node_id)}
    try:
        resp = requests.get(url, params=params, headers=vis_api_headers(), timeout=30)
        resp.raise_for_status()
        data = resp.json()
    except Exception as exc:
        print(f"[RTSP-MTX] listar cameras falhou: {exc}", flush=True)
        return []
    dados = data.get("dados") if isinstance(data, dict) else None
    if not isinstance(dados, list):
        return []
    return [c for c in dados if isinstance(c, dict)]


class RtspMtxPathSync:
    def __init__(self) -> None:
        self._node_id = MEDIAMTX_NODE_ID
        # path -> rtsp_url (para detectar mudança sem logar URL)
        self._synced: dict[str, str] = {}

    def _upsert(self, path: str, camera_id: int, rtsp_url: str) -> None:
        conf = rtsp_on_demand_path_conf(rtsp_url)
        prev = self._synced.get(path)
        if prev == rtsp_url and mtx_path_config_get(path):
            return

        existing = mtx_path_config_get(path)
        if existing:
            ok, err = mtx_path_config_patch(path, conf)
            action = "patch"
        else:
            ok, err = mtx_path_config_add(path, conf)
            action = "add"
        if ok:
            self._synced[path] = rtsp_url
            print(f"[RTSP-MTX] {action} path={path} camera={camera_id}", flush=True)
        else:
            print(f"[RTSP-MTX] {action} FALHOU path={path} camera={camera_id}: {err}", flush=True)

    def _remove(self, path: str, camera_id: int) -> None:
        ok, err = mtx_path_config_delete(path)
        self._synced.pop(path, None)
        if ok:
            print(f"[RTSP-MTX] delete path={path} camera={camera_id}", flush=True)
        else:
            print(f"[RTSP-MTX] delete FALHOU path={path}: {err}", flush=True)

    def tick(self) -> None:
        if not RTSP_MTX_PATH_SYNC_ENABLED:
            return
        cameras = fetch_rtsp_pull_for_mtx_node(self._node_id)
        want_paths: dict[str, tuple[int, str]] = {}
        for cam in cameras:
            cid = int(cam.get("id") or 0)
            rtsp = str(cam.get("rtsp_url_sec") or "").strip()
            path = _path_name(cam)
            if cid < 1 or not path or not rtsp.lower().startswith("rtsp"):
                continue
            want_paths[path] = (cid, rtsp)

        for path, (cid, rtsp) in want_paths.items():
            self._upsert(path, cid, rtsp)

        for path in list(self._synced.keys()):
            if path not in want_paths:
                cid = 0
                self._remove(path, cid)

    def run_forever(self) -> None:
        if not RTSP_MTX_PATH_SYNC_ENABLED:
            print("[RTSP-MTX] desabilitado (RTSP_MTX_PATH_SYNC_ENABLED=0)", flush=True)
            while True:
                time.sleep(60)
        print(
            f"[RTSP-MTX] ativo node_id={self._node_id} poll={RTSP_MTX_PATH_SYNC_POLL_SEC}s "
            f"(sourceOnDemand, sem ffmpeg relay)",
            flush=True,
        )
        while True:
            try:
                self.tick()
            except Exception as exc:
                print(f"[RTSP-MTX] tick erro: {exc}", flush=True)
            time.sleep(max(5, RTSP_MTX_PATH_SYNC_POLL_SEC))


def main() -> None:
    RtspMtxPathSync().run_forever()
