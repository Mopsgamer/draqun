import { environment, logBuild } from "./tool.ts";

const paths = ["server", "main.go"];

const serverCommand = new Deno.Command("go", {
    args: ["run", "."],
});
let serverProcess: Deno.ChildProcess | undefined = undefined;

function start() {
    serverProcess = serverCommand.spawn();
}

async function watchAndRestart() {
    start();
    const watcher = Deno.watchFs(paths, { recursive: true });
    for await (const event of watcher) {
        if (
            !(
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            )
        ) continue;

        tryToKill();
        logBuild.info("File change detected: %s. Restarting...", event.kind);
        start();
    }
}

function tryToKill() {
    if (serverProcess == undefined) {
        return;
    }
    try {
        serverProcess.kill("SIGTERM");
    } catch { /* empty */ }
    serverProcess = undefined;
}

if (environment < 2) {
    logBuild.info("Watching for server code changes...");
    watchAndRestart();
} else {
    start();
}
