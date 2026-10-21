import { rm } from "node:fs/promises";
import { distFolder } from "./constants.ts";
import { spawnSync } from "node:child_process";

export async function compileDist(
	clean = true,
	...args: ("css" | "js" | "assets" | "min" | "watch" | "quiet" | "all")[]
): Promise<void> {
	if (clean) {
		await rm(`${distFolder}/static`, { recursive: true, force: true });
	}
	spawnSync("bun", ["run", "front", ...(args as string[])], {
		stdio: "inherit",
	});
}
