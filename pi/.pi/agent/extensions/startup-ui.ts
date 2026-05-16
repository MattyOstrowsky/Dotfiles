import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { Container, Text, Spacer } from "@earendil-works/pi-tui";
import * as fs from "node:fs";
import * as path from "node:path";

export default function (pi: ExtensionAPI) {
    pi.registerMessageRenderer("dashboard", (data: any, theme: any) => {
        const container = new Container();
        container.addChild(new Text(theme.fg("toolTitle", theme.bold("🚀 Pi Multi-Agent Ecosystem Dashboard")), 0, 0));
        container.addChild(new Spacer(1));
        
        container.addChild(new Text(theme.fg("accent", "🤖 Available Subagents:"), 0, 0));
        if (data.agents && data.agents.length > 0) {
            data.agents.forEach((a: any) => {
                container.addChild(new Text(`  • ${theme.fg("success", a.name.padEnd(15))} ${theme.fg("muted", "->")} ${theme.fg("dim", a.model)}`, 0, 0));
            });
        } else {
            container.addChild(new Text(theme.fg("dim", "  No agents found in agents/ directory."), 0, 0));
        }
        
        container.addChild(new Spacer(1));
        const modeColor = data.mode.includes("Free") ? "success" : data.mode.includes("Premium") ? "error" : "warning";
        container.addChild(new Text(theme.fg("accent", "⚡ Current Financial Mode: ") + theme.fg(modeColor, theme.bold(data.mode)), 0, 0));
        container.addChild(new Text(theme.fg("dim", "Type `/budget free` to use 100% free models. Type `/start <task>` to run Orchestrator."), 0, 0));
        
        return container;
    });

    pi.on("session_start", async (_event, ctx) => {
        const agentsPath = path.join(process.env.HOME || "", "Dotfiles/pi/.pi/agent/agents");
        const agents: Array<{name: string, model: string}> = [];
        try {
            if (fs.existsSync(agentsPath)) {
                const files = fs.readdirSync(agentsPath);
                for (const f of files) {
                    if (f.endsWith(".md")) {
                        const content = fs.readFileSync(path.join(agentsPath, f), "utf-8");
                        const nameMatch = content.match(/^name:\s*(.+)$/m);
                        const modelMatch = content.match(/^model:\s*(.+)$/m);
                        if (nameMatch) {
                            agents.push({
                                name: nameMatch[1].trim(),
                                model: modelMatch ? modelMatch[1].trim() : "default"
                            });
                        }
                    }
                }
            }
        } catch(e) {}
        
        const mode = process.env.PI_BUDGET_MODE === "1" ? "Free (Budget 0$)" 
                   : process.env.PI_BUDGET_MODE === "0" ? "Premium (Flagship)" 
                   : "Smart (Hybrid DeepSeek/Free)";
                   
        pi.appendEntry("dashboard", { agents, mode });
    });
}
