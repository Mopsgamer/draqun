import kill from "tree-kill";
import { sep } from "node:path";
import { existsSync } from "node:fs";
import { logDevelopment, taskDotenv } from "./tool/constants.ts";
import { compileTask } from "./tool/compile-binary.ts";

taskDotenv(logDevelopment);

const paths = [
	"server" + sep,
	"client-lite.go",
	"main.go",
	".git" + sep + "ORIG_HEAD",
];
if (existsSync(".env")) paths.push(".env");

let activePid: number | undefined;
let abortController = new AbortController();

async function start(signal: AbortSignal): Promise<void> {
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

	if (signal.aborted) return;

	try {
		const child = await compileTask(true);
		if (!child) return;
		// Ensure we clean up if the server crashes on its own
		child.status.then(() => {
			if (activePid === child.pid) activePid = undefined;
		});
	} catch (e) {
		if ((e as Error).name === "AbortError") return;
		throw e;
	}
}

async function watchAndRestart(): Promise<void> {
	const watcher = Deno.watchFs(paths, { recursive: true });

	abortController = new AbortController();
	start(abortController.signal);

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

			taskDotenv(logDevelopment);

			// Start the new cycle, which will manually tree-kill the old PID
			await start(abortController.signal);
		}, 150);
	}
}

await watchAndRestart();
