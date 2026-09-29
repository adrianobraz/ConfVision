const fs = require("fs");
const path = require("path");

const root =
  process.env.TRANSCRIPT_ROOT ||
  "C:/Users/Administrator/.cursor/projects/c-sistemaconfmonit-core4/agent-transcripts";
const outDir = path.join(__dirname, "../src/modulos/confvision");

// Do not overwrite route tables / main proxy from partial transcript snapshots.
const skip = new Set(["rotas.go", "proxy.go", "integracao_proxy.go"]);

function walkJsonl(dir, acc = []) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, ent.name);
    if (ent.isDirectory()) walkJsonl(p, acc);
    else if (ent.name.endsWith(".jsonl")) acc.push(p);
  }
  return acc;
}

const byBase = new Map();

for (const file of walkJsonl(root)) {
  for (const line of fs.readFileSync(file, "utf8").split("\n")) {
    if (!line.includes("confvision") || !line.includes(".go")) continue;
    let o;
    try {
      o = JSON.parse(line);
    } catch {
      continue;
    }
    for (const c of o.message?.content || []) {
      if (c.type !== "tool_use" || c.name !== "Write") continue;
      const pth = c.input?.path || "";
      if (!pth.includes("modulos") || !pth.includes("confvision")) continue;
      const base = path.basename(pth);
      if (!base.endsWith(".go") || base.endsWith("_test.go") || skip.has(base)) continue;
      const content = c.input?.contents || "";
      if (!content.startsWith("package confvision")) continue;
      const prev = byBase.get(base);
      if (!prev || content.length >= prev.len) {
        byBase.set(base, { content, len: content.length, from: pth });
      }
    }
  }
}

for (const [base, meta] of byBase) {
  fs.writeFileSync(path.join(outDir, base), meta.content);
  console.log(`${base}\t${meta.len}`);
}
console.log(`Total: ${byBase.size}`);
