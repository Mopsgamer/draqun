import process from "node:process";
import kill from "tree-kill";
import { existsSync } from "node:fs";
import { logDevelopment } from "./tool/constants.ts";
import { compileTask } from "./tool/compile-binary.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "dev",
		description:
			"Build client assets and run the development server, restarting on changes.",
		usage: "deno task dev",
		options: ["-h, --help  Show this help."],
	})
) {
	Deno.exit(0);
}

const frontendWatcher = new Deno.Command("deno", {
	args: ["run", "-A", "scripts/make-front.ts", "watch", "quiet"],
	stdout: "inherit",
	stderr: "inherit",
}).spawn();
frontendWatcher.status.then((status) => {
	if (!status.success) {
		logDevelopment.error(
			`Frontend watcher exited with code ${status.code}.`,
		);
		process.exitCode = status.code || 1;
	}
});
logDevelopment.info("Frontend assets are building and watching for changes.");

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
let abortController = new AbortController();

async function start(signal: AbortSignal): Promise<boolean> {
	// 1. MANUALLY KILL BEFORE STARTING
	// We don't rely on Deno's signal to kill the old process.
	// We use tree-kill to nuke the group before we even attempt a new build.
	if (activePid) {
		await new Promise((resolve) => {
			kill(activePid!, "SIGKILL", () => {
				activePid = undefined;
				resolve(true);
			});
		});
	}

	if (signal.aborted) return true;

	try {
		const child = await compileTask(true, true);
		if (!child) return false;
		activePid = child.pid;
		// Ensure we clean up if the server crashes on its own
		child.status.then(() => {
			if (activePid === child.pid) activePid = undefined;
		});
		return true;
	} catch (e) {
		if ((e as Error).name === "AbortError") return true;
		throw e;
	}
}

async function watchAndRestart(): Promise<void> {
	abortController = new AbortController();
	if (!await start(abortController.signal)) {
		process.exitCode = 1;
		return;
	}

	const watcher = Deno.watchFs(paths, { recursive: true });
	let timeout: NodeJS.Timeout | undefined;
	for await (const event of watcher) {
		if (!["modify", "create", "remove"].includes(event.kind)) continue;

		clearTimeout(timeout);
		timeout = setTimeout(async () => {
			// Stop the ASYNC FLOW of the previous 'start' call
			abortController.abort();
			abortController = new AbortController();

			logDevelopment.info(
				"Refreshing (" + event.kind + " " + event.paths[0] + ")",
			);

			// Start the new cycle, which will manually tree-kill the old PID
			await start(abortController.signal);
		}, 150);
	}
}

await watchAndRestart();
