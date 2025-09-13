import { distFolder, logServerComp } from "./constants.ts";

export async function writeGitJson(
    distination = distFolder + "/git.json",
): Promise<void> {
    const [hash, hashLong, branch] = await Promise.all([
        gitCommandOutput(["rev-parse", "--short", "HEAD"]),
        gitCommandOutput(["rev-parse", "HEAD"]),
        gitCommandOutput(["rev-parse", "--abbrev-ref", "HEAD"]),
    ]);

    const data = JSON.stringify({ hash, hashLong, branch });
    await Deno.writeTextFile(distination, data);
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
