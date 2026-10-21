import type { TaskStateEnd } from "@jsr/m234__logger";
import { binaryInfo, compile, machineInfo } from "./tool/compile-binary.ts";
import { logProd } from "./tool/constants.ts";
import kill from "tree-kill";
import { compileDist } from "./tool/compile-dist.ts";
import { showHelp } from "./tool/help.ts";
import { spawn, spawnSync } from "node:child_process";

if (
	showHelp({
		name: "prod",
		description: "Build and run the production server with periodic updates.",
		usage: "bun run prod",
		options: ["-h, --help  Show this help."],
	})
) {
	process.exit(0);
}

const [os, arch] = machineInfo();
const { filePath } = binaryInfo(os, arch);

await compileDist(true, "min");

if (!(await compile(os, arch))) {
	process.exit(1);
}

let child = spawn(filePath, [], { stdio: "inherit" });
let pid = child.pid || 0;

setInterval(
	() => {
		logProd.task({ text: "Refreshing" }).startRunner(start);
	},
	3 * 60 * 1000,
);

async function start(): Promise<TaskStateEnd> {
	const fetch = spawnSync("git", ["fetch"], { encoding: "utf-8" });

	if (!fetch.stdout || fetch.stdout.length === 0) return "skipped";

	spawnSync("git", ["pull"], { encoding: "utf-8" });

	await compileDist(true, "min");
	if (!(await compile(os, arch))) {
		logProd.warn("Compilation failed, keeping the old version running.");
		return "aborted";
	}

	if (pid > 0) kill(pid);

	child = spawn(filePath, [], { stdio: "inherit" });
	pid = child.pid || 0;

	return "completed";
}
