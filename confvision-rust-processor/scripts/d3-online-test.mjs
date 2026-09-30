#!/usr/bin/env node
/** Testes online D3: sidecar + pilots /health */
const SIDECAR = "https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host";
const PILOTS = [
  "https://foxpro-rust-pilot.rkr351.easypanel.host",
  "https://foxpro-rust-pilot-b.rkr351.easypanel.host",
];

const JPEG_1X1 =
  "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/2wBDAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wAARCAABAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAAAf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA//2Q==";

const args = new Set(process.argv.slice(2));
const QUICK = args.has("--quick");
const FULL = args.has("--full") || !QUICK;

const SIDECAR_TIMEOUT_MS = FULL ? 180_000 : 30_000;

let fail = 0;

function fetchWithTimeout(url, opts = {}, ms = 30_000) {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), ms);
  return fetch(url, { ...opts, signal: ctrl.signal }).finally(() => clearTimeout(t));
}

async function postJson(url, body, label, timeoutMs) {
  const t0 = Date.now();
  let res;
  try {
    res = await fetchWithTimeout(
      url,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      },
      timeoutMs
    );
  } catch (e) {
    console.log(`\n=== ${label} ===`);
    console.log(`  URL: ${url}`);
    console.log(`  FAIL: ${e.message || e} (${Date.now() - t0}ms)`);
    fail++;
    return { status: 0, parsed: null };
  }
  const text = await res.text();
  let parsed;
  try {
    parsed = JSON.parse(text);
  } catch {
    parsed = text;
  }
  console.log(`\n=== ${label} ===`);
  console.log(`  URL: ${url}`);
  console.log(`  STATUS: ${res.status} (${Date.now() - t0}ms)`);
  console.log(`  BODY: ${typeof parsed === "string" ? parsed.slice(0, 400) : JSON.stringify(parsed).slice(0, 400)}`);
  return { status: res.status, parsed };
}

async function getHealth(base) {
  const url = `${base.replace(/\/$/, "")}/health`;
  const res = await fetchWithTimeout(url, {}, 30_000);
  const h = await res.json();
  console.log(`\n=== ${base} /health ===`);
  console.log(`  STATUS: ${res.status}`);
  const keys = [
    "status",
    "yolo_enabled",
    "yolo_backend",
    "yolo_device",
    "capture_enabled",
    "capture_workers",
    "queue_backend",
    "event_queue_redis_ok",
    "event_queue_depth",
    "events_published",
    "events_captured",
    "events_queue_full",
    "cameras_online",
    "cameras_total",
    "capacity_state",
    "load_advisory",
  ];
  for (const k of keys) {
    if (h[k] !== undefined) console.log(`  ${k}=${JSON.stringify(h[k])}`);
  }
  if (h.yolo_enabled === true && (h.yolo_backend === "off" || !h.yolo_backend)) {
    console.log("  FAIL: yolo_enabled mas yolo_backend=off");
    fail++;
  }
  if (h.capture_enabled === true && h.event_queue_redis_ok !== true) {
    console.log("  WARN: capture_enabled com event_queue_redis_ok=false");
  }
  if (res.status !== 200 || h.status !== "ok") fail++;
  return h;
}

async function main() {
  console.log("D3 online tests", new Date().toISOString(), QUICK ? "(--quick)" : "(full)");

  for (const p of PILOTS) {
    await getHealth(p);
  }

  if (QUICK) {
    if (fail) {
      console.log(`\nRESULT: FAIL (${fail} checks)`);
      process.exit(1);
    }
    console.log("\nRESULT: OK d3-online-test --quick");
    return;
  }

  const health = await fetchWithTimeout(`${SIDECAR}/health`, {}, 15_000).catch(() => null);
  if (health) {
    console.log(`\n=== Sidecar GET /health ===\n  STATUS: ${health.status}`);
    try {
      console.log(`  BODY: ${JSON.stringify(await health.json())}`);
    } catch {
      /* ignore */
    }
  } else {
    const root = await fetchWithTimeout(SIDECAR + "/", {}, 10_000).catch(() => null);
    console.log(`\n=== Sidecar GET / ===\n  STATUS: ${root ? root.status : "timeout"}`);
  }

  const empty = await postJson(
    `${SIDECAR}/v1/detect`,
    { jpeg_base64: "" },
    "Sidecar empty jpeg",
    60_000
  );
  if (empty.status !== 400) {
    console.log("  FAIL: esperado 400 para jpeg vazio");
    fail++;
  }

  const infer = await postJson(
    `${SIDECAR}/v1/detect`,
    { jpeg_base64: JPEG_1X1 },
    "Sidecar 1x1 jpeg infer",
    SIDECAR_TIMEOUT_MS
  );
  if (infer.status !== 200) {
    console.log("  FAIL: esperado 200 na inferência mínima");
    fail++;
  } else if (infer.parsed && typeof infer.parsed === "object") {
    const { width, height, detections } = infer.parsed;
    if (typeof width !== "number" || typeof height !== "number" || !Array.isArray(detections)) {
      console.log("  FAIL: resposta sem width/height/detections");
      fail++;
    }
  }

  if (fail) {
    console.log(`\nRESULT: FAIL (${fail} checks)`);
    process.exit(1);
  }
  console.log("\nRESULT: OK d3-online-test");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
