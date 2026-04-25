import kill from "tree-kill";
import { sep } from "node:path";
import { existsSync } from "node:fs";
import { logDevelopment, taskDotenv } from "./tool/constants.ts";

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
		// 2. GENERATE
		const gen = new Deno.Command("go", {
			args: ["generate", "./..."],
			signal, // Signal here is okay, it only stops the generate task
		});
		const genOutput = await gen.output();
		if (!genOutput.success || signal.aborted) return;

		// 3. RUN
		const server = new Deno.Command("go", {
			args: ["run", "-tags", "lite", "."],
			// We DO NOT pass signal here.
			// Passing signal to spawn/run is why Deno orphans the child.
		});

		const child = server.spawn();
		activePid = child.pid;

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

	let timeout = NaN;
	for await (const event of watcher) {
		if (!["modify", "create", "remove"].includes(event.kind)) continue;

		clearTimeout(timeout);
		timeout = setTimeout(() => {
			// Stop the ASYNC FLOW of the previous 'start' call
			abortController.abort();
			abortController = new AbortController();

			logDevelopment.info(
				"Refreshing (" + event.kind + " " + event.paths[0] + ")",
			);

			taskDotenv(logDevelopment);

			// Start the new cycle, which will manually tree-kill the old PID
			start(abortController.signal);
		}, 150);
	}
}

await watchAndRestart();
