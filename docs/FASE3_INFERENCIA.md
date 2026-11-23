# Fase 3 — captura, concorrência e inferência

Implementação local. Não houve deploy, commit nem push. A validação comparativa da Fase 2 continua pendente e não entra nestes números.

## Arquitetura

Antes: a thread da câmera lia o RTSP, rodava o MOG2 e, no frame elegível, esperava o lock global para chamar `model()` no `yolov8n.pt`. Durante essa espera a câmera não lia o próximo frame.

Agora, com `YOLO_INFER_DECOUPLED=1` (padrão):

1. A thread da câmera continua dona do `VideoCapture`, do MOG2, do gate e do `FRAME_SKIP`.
2. O frame elegível é copiado e entregue a um scheduler.
3. Uma thread só, `yolo-infer`, chama a mesma instância do modelo, com o mesmo lock. Não há duas inferências ao mesmo tempo.
4. O resultado volta para a câmera que originou o frame. Classe, confiança, área, cooldown e evento usam o mesmo código de antes.

`YOLO_INFER_DECOUPLED=0` devolve o lock para dentro do ciclo da câmera. As regras de detecção não mudam de lugar: ficam em `PersonDetector._handle_detection`.

O modelo não foi trocado. Não há ONNX, TensorRT nem segundo YOLO. O MOG2, as URLs, o MediaMTX, o relay, o banco e as APIs não foram alterados.

## Política de frames

Há no máximo um frame pendente por câmera. Se outro frame elegível chega antes de a inferência começar, o anterior é descartado e fica o mais novo. O horário de entrada na fila não muda nessa troca, para uma câmera não passar na frente das outras.

Consequência: com contenção, frames intermediários deixam de ser analisados. Sem contenção, o `FRAME_SKIP` (5 frames) é mais longo que uma inferência típica e o descarte não ocorre.

O snapshot e o evento usam a cópia do frame que foi inferido, não o buffer da leitura seguinte. `miss_streak` conta resultados de inferência, como já acontecia. Frame descartado não entra nessa conta.

`FRAME_SKIP` e o MOG2 não foram alterados.

## Concorrência

- Um modelo, um lock, uma thread de inferência.
- A leitura não pega esse lock.
- A fila de espera não cresce com o tempo: o tamanho é o número de câmeras com um frame ainda não iniciado.
- Ao reconectar ou encerrar a câmera, o pendente dela é retirado e o resultado que ainda estava no ar é ignorado (geração da conexão).
- Exceção numa inferência não encerra a thread nem a outra câmera.
- `cap.release()` também roda quando a abertura falha.

## Aquecimento

Antes das threads de câmera, `PersonDetector()` chama o modelo já carregado com um frame preto de 640×640. Não abre RTSP, não publica evento, não grava snapshot e não mexe em câmera. Se o aquecimento falha, o worker segue e a primeira inferência real paga o custo.

Neste ambiente de teste não há `torch` nem `ultralytics`. A medição com `yolov8n.pt` fica pendente. O teste local usa um modelo falso que dorme 60 ms na primeira chamada e 5 ms na segunda, e confirma que a segunda não repete esse custo.

## Métricas

O coletor da F0/F2 permanece. Campos novos, sem log por frame:

| Campo | Significado |
|---|---|
| `yolo_ms_*` | Igual à F0: tempo dentro de `model()` |
| `yolo_inference_ms_*` | O mesmo tempo, com média, p95 e máximo |
| `yolo_lock_wait_ms_*` | Só a espera do lock |
| `yolo_fila_ms_*` | Tempo no slot, antes de pedir o lock |
| `yolo_total_ms_*` | Fila + lock + `model()` |
| `cap_read_interval_ms_*` | Intervalo entre leituras boas, não a duração de `cap.read()` |
| `cap_read_interval_max_ms` | O máximo desse intervalo |
| `infer_descartados` | Frames elegíveis substituídos por um mais novo |
| `cameras_abertas` | Câmeras com captura aberta agora |
| `cameras_em_backoff` | Câmeras no backoff de abertura |
| `inferencias_aguardando` | Câmeras com frame no slot, inferência ainda não iniciada |
| `inferencias` | Quantidade na janela, como na F0 |

Falha do coletor não interrompe a captura.

## Benchmark local

Duas câmeras, frame sintético 24×24, leitura a cada 10 ms, inferência falsa de 80 ms, 0,45 s, `FRAME_SKIP=1`. Não é câmera real e não serve para comparar com a Fase 2.

| | Leituras | Intervalo máximo | Espera máxima do lock |
|---|---|---|---|
| Lock dentro da leitura | 7 | 172 ms | 79,7 ms |
| Leitura desacoplada | 87 | 16 ms | 0,0 ms |

CPU, RSS e o efeito em `yolov8n.pt` com RTSP real não foram medidos. Ficam pendentes para a homologação.

## Testes

`python -m pytest tests -q` em `confvision`: 61 testes passaram.

Cobertos aqui: separação das métricas, falha isolada do coletor, backoff no gauge, aquecimento que falha e que não repete o custo inicial, classe, confiança, cooldown, área dentro/fora, abertura e fechamento do gate, resultado atrasado ignorado, um slot por câmera, isolamento entre câmeras, exceção que não derruba a outra, cópia do frame, `FRAME_SKIP` inalterado, leitura que não espera o lock.

Não reexecutados contra serviço real: Redis, upload, clipe de 20 s, estado armado. O contrato de `on_person(conf, frame, area, bbox_json)` e os workers de clipe não foram alterados.

## Configuração

`YOLO_INFER_DECOUPLED=1` no `.env.example`. `0` restaura o lock no ciclo da câmera.

## Limitações

- A decodificação de todo frame continua. O ocioso de CPU não foi reduzido.
- Com várias câmeras em pessoa ao mesmo tempo, a inferência continua serial. O que mudou é que a leitura não para.
- Sob contenção, o evento sai do frame mais novo que ainda não tinha sido inferido, não de cada frame elegível.
- O aquecimento real do `yolov8n.pt` não foi medido nesta máquina.

## Arquivos

- `detector.py` — aquecimento, despacho e o mesmo pós-processamento.
- `inference_scheduler.py` — slot único por câmera e a thread do modelo.
- `detection_metrics.py` — campos novos.
- `rtsp_backoff.py` — contagem de câmeras em backoff, sem log extra.
- `config.py` e `.env.example` — a flag.
- `tests/test_fase3_inferencia.py`
- este documento.

## Rollback

1. `YOLO_INFER_DECOUPLED=0` e reiniciar só quando houver autorização.
2. Ou copiar de volta `.backups/2026-11-23_1521/confvision/` sobre `confvision/` (`detector.py`, `detection_metrics.py`, `config.py`, `rtsp_backoff.py`, `.env.example`). O arquivo novo `inference_scheduler.py` deixa de ser importado quando `detector.py` volta ao backup. Não apagar o backup nem o arquivo novo.
