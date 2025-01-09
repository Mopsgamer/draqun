import { envKeys } from "./tool.ts";

const watchDir = "./internal";

let serverProcess: Deno.ChildProcess | undefined = undefined;
async function startDenoTask() {
    // If there is an existing process, kill it
    if (serverProcess) {
        console.log("Stopping existing Deno task...");
        serverProcess.kill("SIGTERM");
        await serverProcess.status;
    }

    const serverCommand = new Deno.Command("go", {
        args: ["run", "."],
    });
    serverProcess = serverCommand.spawn();
}

async function watchAndRestart() {
    const watcher = Deno.watchFs(watchDir, { recursive: true });
    for await (const event of watcher) {
        if (
            event.kind === "modify" || event.kind === "create" ||
            event.kind === "remove"
        ) continue;

        console.log("File change detected, restarting Deno task...");
        await startDenoTask(); // Restart the Deno task on change
    }
}

await startDenoTask();
if (Number(Deno.env.get(envKeys.ENVIRONMENT)) < 2) {
    await watchAndRestart();
}
