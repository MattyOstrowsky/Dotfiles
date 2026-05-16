import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

export default function (pi: ExtensionAPI) {
  pi.on("tool_call", async (event, ctx) => {
    // 1. Ochrona w poleceniach bash
    if (event.toolName === "bash") {
      const command = event.input.command || "";
      if (
        command.includes("rm -rf") ||
        command.includes("terraform destroy") ||
        command.includes(".tfstate") ||
        command.includes(".env") ||
        command.includes("secrets/")
      ) {
        const ok = await ctx.ui.confirm("Dangerous Command Detected!", `Allow command: ${command}?`);
        if (!ok) {
          return { block: true, reason: "Blocked by user for security" };
        }
      }
    } 
    // 2. Ochrona przed modyfikacją wrażliwych plików (edit/write)
    else if (event.toolName === "write" || event.toolName === "edit") {
      const file = event.input.file || event.input.targetFile || "";
      if (
        file.endsWith(".tfstate") ||
        file.endsWith(".env") ||
        file.includes("secrets/")
      ) {
        const ok = await ctx.ui.confirm("Protected File Detected!", `Allow modification of: ${file}?`);
        if (!ok) {
          return { block: true, reason: "Blocked by user. Protected path." };
        }
      }
    }
  });
}
