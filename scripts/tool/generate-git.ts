import type { Logger, TaskStateEnd } from "@m234/logger";
import { distFolder, logServerComp } from "./constants.ts";

export async function taskGitJson(
    logger: Logger,
    distination = distFolder + "/git.json",
): Promise<void> {
    await logger.task({ text: "Creating " + distination }).startRunner(
        async () => {
            const success = await writeGitJson(distination);
            if (!success) {
                return "skipped";
            }
        },
    );
}

export async function writeGitJson(
    distination = distFolder + "/git.json",
): Promise<TaskStateEnd> {
    const [hash, hashLong, branch] = await Promise.all([
        gitCommandOutput(["rev-parse", "--short", "HEAD"]),
        gitCommandOutput(["rev-parse", "HEAD"]),
        gitCommandOutput(["rev-parse", "--abbrev-ref", "HEAD"]),
    ]);

    const data = JSON.stringify({ hash, hashLong, branch });
    const same = data === await Deno.readTextFile(distination);
    if (same) return "skipped";
    await Deno.writeTextFile(distination, data);
    return "completed";
}

export async function gitCommandOutput(args: string[]): Promise<string> {
    const { success, stdout } = await new Deno.Command("git", { args })
        .output();

    if (!success) {
        logServerComp.error(`Failed to get git information.`);
        return "unknown";
    }
    return new TextDecoder().decode(stdout).trim();
}
