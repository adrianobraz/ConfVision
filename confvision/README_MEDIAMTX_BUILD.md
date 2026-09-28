# Build MediaMTX no branch rust-pilot

O EasyPanel costuma usar **subpasta `confvision/`** + `Dockerfile.mediamtx`.

Neste branch, o fix de boot (`start_mediamtx_guard.py` aguarda `:8100/health`) está em **`confvision/start_mediamtx_guard.py`** (espelho da raiz).

Para imagem completa MediaMTX+Guard, use no GitHub **branch `main`** (tree `confvision/`) **ou** mescle `confvision/` da main com este branch antes do build.
