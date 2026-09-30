#!/usr/bin/env python3
"""
Compara JSONL do shadow mode (Rust luma vs Python MOG2).

Uso:
  python scripts/motion_shadow_compare.py /tmp/confvision/motion-shadow --window-ms 2000
"""

from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path


def load_lines(path: Path) -> list[dict]:
    rows = []
    if not path.exists():
        return rows
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return rows


def main() -> None:
    p = argparse.ArgumentParser(description="Compare motion shadow JSONL")
    p.add_argument("log_dir", type=Path)
    p.add_argument("--window-ms", type=int, default=2000)
    args = p.parse_args()

    by_cam: dict[int, dict[str, list[dict]]] = defaultdict(lambda: {"rust": [], "python": []})
    for f in args.log_dir.glob("*.jsonl"):
        name = f.name
        if name.startswith("rust_cam_"):
            cam = int(name.removeprefix("rust_cam_").removesuffix(".jsonl"))
            by_cam[cam]["rust"] = load_lines(f)
        elif name.startswith("python_cam_"):
            cam = int(name.removeprefix("python_cam_").removesuffix(".jsonl"))
            by_cam[cam]["python"] = load_lines(f)

    if not by_cam:
        print(f"Nenhum JSONL em {args.log_dir}")
        return

    for cam_id in sorted(by_cam):
        rust = [r for r in by_cam[cam_id]["rust"] if r.get("session_event") is None]
        py = [r for r in by_cam[cam_id]["python"] if r.get("session_event") is None]
        agree = 0
        compared = 0
        for r in rust:
            ts = r.get("ts_unix_ms")
            if ts is None:
                continue
            near = [
                x
                for x in py
                if abs(x.get("ts_unix_ms", 0) - ts) <= args.window_ms
            ]
            if not near:
                continue
            compared += 1
            if any(x.get("detected") == r.get("detected") for x in near):
                agree += 1
        print(
            f"camera={cam_id} samples_rust={len(rust)} samples_python={len(py)} "
            f"paired={compared} agree_detected={agree} "
            f"rate={agree / compared:.2%}" if compared else f"camera={cam_id} paired=0"
        )


if __name__ == "__main__":
    main()
