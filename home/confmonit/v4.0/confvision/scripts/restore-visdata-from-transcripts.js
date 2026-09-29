// One-off: restore visdata/*.go from agent transcript Write records (latest/largest wins).
const fs = require("fs");
const path = require("path");

const transcriptRoot =
  process.env.TRANSCRIPT_ROOT ||
  "C:/Users/Administrator/.cursor/projects/c-sistemaconfmonit-core4/agent-transcripts";
const outDir = path.join(__dirname, "../src/modulos/visdata");

function walkJsonl(dir, acc = []) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, ent.name);
    if (ent.isDirectory()) walkJsonl(p, acc);
    else if (ent.name.endsWith(".jsonl")) acc.push(p);
  }
  return acc;
}

const byBase = new Map();

for (const file of walkJsonl(transcriptRoot)) {
  const lines = fs.readFileSync(file, "utf8").split("\n");
  for (const line of lines) {
    if (!line.includes("visdata") || !line.includes(".go")) continue;
    let o;
    try {
      o = JSON.parse(line);
    } catch {
      continue;
    }
    for (const c of o.message?.content || []) {
      if (c.type !== "tool_use" || c.name !== "Write") continue;
      const pth = c.input?.path || "";
      if (!/visdata[\\/].+\.go$/i.test(pth)) continue;
      const base = path.basename(pth);
      const content = c.input?.contents || "";
      if (!content.includes("package visdata")) continue;
      const prev = byBase.get(base);
      if (!prev || content.length >= prev.content.length) {
        byBase.set(base, { content, from: pth, len: content.length });
      }
    }
  }
}

const written = [];
for (const [base, meta] of byBase.entries()) {
  const dest = path.join(outDir, base);
  fs.writeFileSync(dest, meta.content);
  written.push(`${base}\t${meta.len}\t${meta.from}`);
}
written.sort();
console.log(written.join("\n"));
console.log(`\nTotal: ${written.length} files -> ${outDir}`);
