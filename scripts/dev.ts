import process from "node:process";
import kill from "tree-kill";
import fsSync, { existsSync } from "node:fs";
import path from "node:path";
import { logDevelopment } from "./tool/constants.ts";
import { compileTask } from "./tool/compile-binary.ts";
import { showHelp } from "./tool/help.ts";
import { runMakeFront } from "./make-front.ts";
import type { ChildProcess } from "node:child_process";

if (
	showHelp({
		name: "dev",
		description:
			"Build client assets and run the development server, restarting on changes.",
		usage: "bun run dev [options]",
		options: [
			"-h, --help  Show this help.",
			"--no-front  Skip frontend build and watcher.",
		],
	})
) {
	process.exit(0);
}

const args = process.argv.slice(2);
const noFront = args.includes("--no-front");

if (!noFront) {
	logDevelopment.info("Frontend assets are building and watching for changes.");
	runMakeFront(["watch"]).catch((err) => {
		logDevelopment.error(`Frontend watcher error: ${err}`);
	});
} else {
	logDevelopment.info("Frontend build skipped (--no-front).");
}

const requiredPaths = ["server/", "main.go"];
const optionalPaths = ["client-lite.go", ".git/ORIG_HEAD", ".env"];

const watchPaths: string[] = [];

for (const p of requiredPaths) {
	if (existsSync(p)) {
		watchPaths.push(p);
	} else {
		logDevelopment.error(`Required path '${p}' not found.`);
		process.exit(1);
	}
}

for (const p of optionalPaths) {
	if (existsSync(p)) {
		watchPaths.push(p);
	} else {
		logDevelopment.warn(`Optional path '${p}' not found, skipping watch.`);
	}
}

logDevelopment.info("Watching paths: " + watchPaths.join(", "));

let activePid: number | undefined;
let isBuildingOrStarting = false;
let buildPending = false;
let pendingReason = "";

async function killActiveProcess(): Promise<void> {
	if (activePid) {
		const pidToKill = activePid;
		activePid = undefined;
		await new Promise((resolve) => {
			kill(pidToKill, "SIGKILL", () => {
				resolve(true);
			});
		});
	}
}

async function rebuildAndRun(reason?: string): Promise<boolean> {
	if (reason) {
		pendingReason = reason;
	}

	if (isBuildingOrStarting) {
		buildPending = true;
		return true;
	}

	isBuildingOrStarting = true;

	try {
		while (true) {
			buildPending = false;
			await killActiveProcess();

			const child = (await compileTask(true, true)) as ChildProcess | false;
			if (buildPending) {
				if (child && child.pid) {
					activePid = child.pid;
					await killActiveProcess();
				}
				logDevelopment.info(
					`Refreshing due to changes during build (${pendingReason})`,
				);
				continue;
			}

			if (!child) {
				if (buildPending) {
					continue;
				}
				return false;
			}

			activePid = child.pid;
			child.on("close", () => {
				if (activePid === child.pid) activePid = undefined;
			});

			break;
		}
	} finally {
		isBuildingOrStarting = false;
	}

	return true;
}

async function watchAndRestart(): Promise<void> {
	if (!(await rebuildAndRun())) {
		process.exitCode = 1;
		return;
	}

	let timeout: NodeJS.Timeout | undefined;

	for (const watchPath of watchPaths) {
		if (!existsSync(watchPath)) continue;

		fsSync.watch(watchPath, { recursive: true }, (_eventType, filename) => {
			if (!filename) return;

			clearTimeout(timeout);
			timeout = setTimeout(async () => {
				const reason = `change in ${path.join(watchPath, filename)}`;
				logDevelopment.info("Refreshing (" + reason + ")");
				await rebuildAndRun(reason);
			}, 150);
		});
	}
}

await watchAndRestart();
