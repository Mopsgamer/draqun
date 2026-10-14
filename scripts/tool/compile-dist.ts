import { rm } from "node:fs/promises";
import { distFolder } from "./constants.ts";

export async function compileDist(
	clean = true,
	...args: ("css" | "js" | "assets" | "min" | "watch" | "quiet" | "all")[]
): Promise<void> {
	if (clean) {
		await rm(distFolder + "/static", { recursive: true, force: true });
	}
	await new Deno.Command("deno", {
		args: ["task", "front", ...(args as string[])],
	}).spawn().status;
}
