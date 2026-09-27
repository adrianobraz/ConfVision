#!/usr/bin/env node
/** Conferência D3: sidecar, health, metrics (motion/eventos) */
const SIDECAR = "https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host";
const PILOTS = [
  { name: "A", base: "https://foxpro-rust-pilot.rkr351.easypanel.host" },
  { name: "B", base: "https://foxpro-rust-pilot-b.rkr351.easypanel.host" },
];

async function main() {
  console.log("=== D3 stack check ===", new Date().toISOString());

  const sidecarGet = await fetch(SIDECAR + "/");
  console.log("\nSidecar GET /:", sidecarGet.status);

  const infer = await fetch(`${SIDECAR}/v1/detect`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      jpeg_base64:
        "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/2wBDAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQH/wAARCAABAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAAAAAAAAAAAAf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA//2Q==",
    }),
  });
  const inferBody = await infer.json();
  console.log("Sidecar infer 1x1:", infer.status, `${JSON.stringify(inferBody).slice(0, 120)}...`);

  for (const { name, base } of PILOTS) {
    const h = await (await fetch(`${base}/health`)).json();
    const m = await (await fetch(`${base}/metrics`)).json();
    console.log(`\n--- Piloto ${name} (${h.processor_id}) ---`);
    console.log("  health: status=%s uptime=%s cameras_online=%s fps=%s",
      h.status, h.uptime_secs, h.cameras_online, h.fps_total?.toFixed?.(2) ?? h.fps_total);
    console.log("  D2: queue=%s redis_ok=%s depth=%s dlq=%s",
      h.queue_backend, h.event_queue_redis_ok, h.event_queue_depth, h.event_queue_dlq_depth);
    console.log("  D3: yolo=%s backend=%s capture=%s workers=%s",
      h.yolo_enabled, h.yolo_backend, h.capture_enabled, h.capture_workers);
    console.log("  eventos: published=%s captured=%s queue_full=%s",
      h.events_published, h.events_captured, h.events_queue_full);
    const met = m.metrics || {};
    console.log("  pipeline: frames_received=%s decoded=%s motion_analyzed=%s motion_detected=%s motion_errors=%s",
      met.frames_received, met.frames_decoded, met.frames_motion_analyzed, met.motion_detected, met.motion_errors);
    console.log("  capacity: state=%s limiting=%s advisory=%s",
      m.capacity?.state, m.capacity?.limiting_resource, m.load_advisory);
    for (const cam of m.cameras || []) {
      console.log("  camera id=%s status=%s fps=%s motion_detected=%s last_motion_score=%s",
        cam.camera_id, cam.status, cam.fps?.toFixed?.(2) ?? cam.fps,
        cam.motion_detected, cam.last_motion_score);
    }
  }
  console.log("\n=== fim ===");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
