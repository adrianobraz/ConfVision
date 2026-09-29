const fs = require("fs");
const path = require("path");

const root =
  process.env.TRANSCRIPT_ROOT ||
  "C:/Users/Administrator/.cursor/projects/c-sistemaconfmonit-core4/agent-transcripts";
const outBase = path.join(__dirname, "../src/modulos");

const patterns = [
  [/confvision[\\/]grupos_visualizacao_proxy\.go$/i, "confvision/grupos_visualizacao_proxy.go"],
  [/confvision[\\/]escopo\.go$/i, "confvision/escopo.go"],
  [/login[\\/]dominio_login\.go$/i, "login/dominio_login.go"],
  [/confvision[\\/]xano_licenca_sync\.go$/i, "confvision/xano_licenca_sync.go"],
  [/confvision[\\/]cameras_resumo\.go$/i, "confvision/cameras_resumo.go"],
  [/confvision[\\/]rtmp_franqueado\.go$/i, "confvision/rtmp_franqueado.go"],
];

// integracao_proxy: prefer version that defines ProxyIntegracaoGet (UI integracao.js)
const integracaoPrefer = "ProxyIntegracaoGet";

function walkJsonl(dir, acc = []) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, ent.name);
    if (ent.isDirectory()) walkJsonl(p, acc);
    else if (ent.name.endsWith(".jsonl")) acc.push(p);
  }
  return acc;
}

const byKey = new Map();

for (const file of walkJsonl(root)) {
  const lines = fs.readFileSync(file, "utf8").split("\n");
  for (const line of lines) {
    if (!line.includes("Write")) continue;
    let o;
    try {
      o = JSON.parse(line);
    } catch {
      continue;
    }
    for (const c of o.message?.content || []) {
      if (c.type !== "tool_use" || c.name !== "Write") continue;
      const pth = c.input?.path || "";
      const content = c.input?.contents || "";
      for (const [re, key] of patterns) {
        if (!re.test(pth)) continue;
        const prev = byKey.get(key);
        if (!prev || content.length >= prev.len) {
          byKey.set(key, { content, len: content.length, from: pth });
        }
      }
      if (/confvision[\\/]integracao_proxy\.go$/i.test(pth)) {
        const key = "confvision/integracao_proxy.go";
        const prev = byKey.get(key);
        const score =
          (content.includes(integracaoPrefer) ? 1_000_000 : 0) + content.length;
        const prevScore =
          (prev?.content.includes(integracaoPrefer) ? 1_000_000 : 0) +
          (prev?.len || 0);
        if (!prev || score >= prevScore) {
          byKey.set(key, { content, len: content.length, from: pth });
        }
      }
    }
  }
}

for (const [key, meta] of byKey) {
  const dest = path.join(outBase, key);
  fs.mkdirSync(path.dirname(dest), { recursive: true });
  fs.writeFileSync(dest, meta.content);
  console.log(`wrote ${dest}\t${meta.len}\t${meta.from}`);
}

console.log(`Total: ${byKey.size}`);
