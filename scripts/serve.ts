import { logClientComp } from "./tool/constants.ts";
import kill from "tree-kill";

const paths = ["server", "lite.go"];

const serverCommand = new Deno.Command("go", {
    args: ["run", "-tags", "lite", "."],
});
let goRunProcess: Deno.ChildProcess | undefined = undefined;

function start() {
    goRunProcess = serverCommand.spawn();
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
        logClientComp.info(
            "File change detected: %s. Restarting...",
            event.kind,
        );
        start();
    }
}

function tryToKill() {
    if (goRunProcess == undefined) {
        return;
    }
    try {
        kill(goRunProcess.pid, "SIGTERM");
    } catch { /* empty */ }
    goRunProcess = undefined;
}

watchAndRestart();
