import process from "node:process";
import kill from "tree-kill";
import { existsSync } from "node:fs";
import { logDevelopment } from "./tool/constants.ts";
import { compileTask } from "./tool/compile-binary.ts";
import { showHelp } from "./tool/help.ts";
import { runMakeFront } from "./make-front.ts";

if (
	showHelp({
		name: "dev",
		description:
			"Build client assets and run the development server, restarting on changes.",
		usage: "deno task dev [options]",
		options: [
			"-h, --help  Show this help.",
			"--no-front  Skip frontend build and watcher.",
		],
	})
) {
	Deno.exit(0);
}

const noFront = Deno.args.includes("--no-front");

if (!noFront) {
	logDevelopment.info("Frontend assets are building and watching for changes.");
	runMakeFront(["watch"]).catch((err) => {
		logDevelopment.error(`Frontend watcher error: ${err}`);
	});
} else {
	logDevelopment.info("Frontend build skipped (--no-front).");
}

const requiredPaths = [
	"server/",
	"main.go",
];

const optionalPaths = [
	"client-lite.go",
	".git/ORIG_HEAD",
	".env",
];

const paths: string[] = [];

for (const p of requiredPaths) {
	if (existsSync(p)) {
		paths.push(p);
	} else {
		logDevelopment.error(`Required path '${p}' not found.`);
		Deno.exit(1);
	}
}

for (const p of optionalPaths) {
	if (existsSync(p)) {
		paths.push(p);
	} else {
		logDevelopment.warn(`Optional path '${p}' not found, skipping watch.`);
	}
}

logDevelopment.info("Watching paths: " + paths.join(", "));

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

			const child = await compileTask(true, true);
			if (buildPending) {
				if (child) {
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
			child.status.then(() => {
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
	if (!await rebuildAndRun()) {
		process.exitCode = 1;
		return;
	}

	const watcher = Deno.watchFs(paths, { recursive: true });
	let timeout: NodeJS.Timeout | undefined;
	for await (const event of watcher) {
		if (!["modify", "create", "remove"].includes(event.kind)) continue;

		clearTimeout(timeout);
		timeout = setTimeout(async () => {
			const reason = event.kind + " " + event.paths[0];
			logDevelopment.info("Refreshing (" + reason + ")");
			await rebuildAndRun(reason);
		}, 150);
	}
}

await watchAndRestart();
