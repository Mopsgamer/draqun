import { logDevelopment } from "./tool/constants.ts";
import kill from "tree-kill";

const paths = ["server", "lite.go", ".env"];

const serverCommand = new Deno.Command("go", {
    args: ["run", "-tags", "lite", "."],
});
let goRunProcess: Deno.ChildProcess | undefined = undefined;

function start() {
    goRunProcess = serverCommand.spawn();
}

async function watchAndRestart() {
    logDevelopment.task({
        text: "Starting",
    }).startRunner(start);
    const watcher = Deno.watchFs(paths, { recursive: true });
    for await (const event of watcher) {
        if (
            !(
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            )
        ) continue;

        tryToKill();
        logDevelopment.task({ text: "Restarting" }).startRunner(start);
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
