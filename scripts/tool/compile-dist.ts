import { rm } from "node:fs/promises";
import { distFolder } from "./constants.ts";

export async function compileDist(
    clean = true,
    ...args: ("css" | "js" | "sl" | "sl-assets")[]
): Promise<void> {
    if (clean) {
        await rm(distFolder + "/static", { recursive: true, force: true });
    }
    await new Deno.Command("deno", {
        args: ["task", "compile:client", ...(args as string[])],
    }).spawn().status;
}
