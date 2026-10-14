import type { TaskStateEnd } from "@m234/logger";
import { binaryInfo, compile, machineInfo } from "./tool/compile-binary.ts";
import { logProd } from "./tool/constants.ts";
import kill from "tree-kill";
import { compileDist } from "./tool/compile-dist.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "prod",
		description: "Build and run the production server with periodic updates.",
		usage: "deno task prod",
		options: ["-h, --help  Show this help."],
	})
) {
	Deno.exit(0);
}

const [os, arch] = machineInfo();
const { filePath } = binaryInfo(os, arch);

await compileDist(true, "min");

if (!await compile(os, arch)) {
	Deno.exit(1);
}

let pid = new Deno.Command(filePath, {
	stdout: "inherit",
	stderr: "inherit",
}).spawn().pid;

setInterval(() => {
	logProd.task({ text: "Refreshing" })
		.startRunner(start);
}, 3 * 60 * 1000);

async function start(): Promise<TaskStateEnd> {
	const fetch = await new Deno.Command("git", {
		args: ["fetch"],
		stdout: "piped",
		stderr: "piped",
	}).output();

	if (fetch.stdout.toString().length === 0) return "skipped";

	await new Deno.Command("git", {
		args: ["pull"],
		stdout: "piped",
		stderr: "piped",
	}).output();

	await compileDist(true, "min");
	if (!await compile(os, arch)) {
		await logProd.warn("Compilation failed, keeping the old version running.");
		return "aborted";
	}

	if (pid > 0) kill(pid);

	pid = new Deno.Command(filePath, {
		stdout: "inherit",
		stderr: "inherit",
	}).spawn().pid;

	return "completed";
}
