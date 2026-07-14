#!/usr/bin/env python3
"""Hook MediaMTX runOnRecordSegmentComplete — processa um segmento pronto."""
import os
import sys
from pathlib import Path

from dvr_segment import process_segment_file
from xano_client import get_cameras_gravacao_ativas


def _camera_id_from_path(path: str) -> int | None:
    parts = path.strip("/").split("/")
    if len(parts) >= 2 and parts[0] == "live":
        try:
            return int(parts[1])
        except ValueError:
            return None
    return None


def main():
    segment_path = os.environ.get("MTX_SEGMENT_PATH") or (
        sys.argv[1] if len(sys.argv) > 1 else ""
    )
    mtx_path = os.environ.get("MTX_PATH") or (
        sys.argv[2] if len(sys.argv) > 2 else ""
    )

    if not segment_path:
        print("[DVR-HOOK] MTX_SEGMENT_PATH nao informado")
        sys.exit(1)

    file_path = Path(segment_path)
    camera_id = _camera_id_from_path(mtx_path)
    if camera_id is None and file_path.parent.name.isdigit():
        parent = file_path.parent
        if parent.parent.name == "live":
            camera_id = int(parent.name)
    if camera_id is None:
        print(f"[DVR-HOOK] camera_id nao identificado path={mtx_path} file={file_path}")
        sys.exit(1)

    cameras = {int(c["id"]): c for c in get_cameras_gravacao_ativas()}
    camera = cameras.get(camera_id)
    if not camera:
        print(f"[DVR-HOOK] camera={camera_id} sem gravacao ativa no Xano")
        sys.exit(0)

    segmento = int(camera.get("segmento_minutos") or 5)
    ok = process_segment_file(file_path, camera, segmento_minutos=segmento)
    sys.exit(0 if ok else 1)


if __name__ == "__main__":
    main()
