#!/usr/bin/env node
/** Testes online D3: sidecar + pilots /health */
const SIDECAR = "https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host";
const PILOTS = [
  "https://foxpro-rust-pilot.rkr351.easypanel.host",
  "https://foxpro-rust-pilot-b.rkr351.easypanel.host",
];

const JPEG_1X1 =
  "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/2wBDAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wAARCAABAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAAAf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA//2Q==";

let fail = 0;

async function postJson(url, body, label) {
  const t0 = Date.now();
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
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
  const res = await fetch(url);
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
  console.log("D3 online tests", new Date().toISOString());

  const root = await fetch(SIDECAR + "/");
  console.log(`\n=== Sidecar GET / ===\n  STATUS: ${root.status}`);

  const empty = await postJson(`${SIDECAR}/v1/detect`, { jpeg_base64: "" }, "Sidecar empty jpeg");
  if (empty.status !== 400) {
    console.log("  FAIL: esperado 400 para jpeg vazio");
    fail++;
  }

  const infer = await postJson(
    `${SIDECAR}/v1/detect`,
    { jpeg_base64: JPEG_1X1 },
    "Sidecar 1x1 jpeg infer"
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

  for (const p of PILOTS) {
    await getHealth(p);
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
