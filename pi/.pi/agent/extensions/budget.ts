import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

export default function (pi: ExtensionAPI) {
    pi.registerCommand("budget", {
        description: "Toggle between Free (Budget), Premium, and Smart modes for subagents",
        handler: async (args, ctx) => {
            const mode = args?.trim().toLowerCase();
            
            if (mode === "free") {
                process.env.PI_BUDGET_MODE = "1";
                // Set main model to a free one so the top-level agent is also free
                pi.setModel("openrouter/owl-alpha");
                ctx.ui.notify("💸 Budget Mode ON: All subagents will use 100% FREE models (Laguna/Owl).", "success");
            } else if (mode === "premium") {
                process.env.PI_BUDGET_MODE = "0";
                // Set main model to Sonnet 3.7
                pi.setModel("anthropic/claude-3.7-sonnet");
                ctx.ui.notify("💎 Premium Mode ON: All subagents will use Flagship models (Claude 3.7).", "success");
            } else if (mode === "smart") {
                delete process.env.PI_BUDGET_MODE;
                // Set main model back to default smart
                pi.setModel("deepseek/deepseek-v4-pro");
                ctx.ui.notify("🧠 Smart Mode ON: Subagents will use the optimized mix of DeepSeek and free models.", "success");
            } else {
                ctx.ui.notify("Usage: /budget [free|premium|smart]\n\nfree - Forces all agents to use $0 models\npremium - Forces all agents to use Claude 3.7\nsmart - Uses default optimized mix (DeepSeek + Free)", "error");
            }
        }
    });
}
