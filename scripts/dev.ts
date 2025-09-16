import kill from "tree-kill";
import { existsSync } from "@std/fs";
import { writeGitJson } from "./tool/generate-git.ts";
import { logDevelopment, taskDotenv } from "./tool/constants.ts";

taskDotenv(logDevelopment);

const paths = ["server", "lite.go"];
if (existsSync(".env")) paths.push(".env");
const serverCommand = new Deno.Command("go", {
    args: ["run", "-tags", "lite", "."],
});
let goRunProcess: Deno.ChildProcess | undefined = undefined;

async function start() {
    await writeGitJson();
    goRunProcess = serverCommand.spawn();
    await goRunProcess.status;
}

async function watchAndRestart() {
    const watcher = Deno.watchFs(paths, { recursive: true });
    await start();
    for await (const event of watcher) {
        if (
            !(
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            )
        ) continue;

        tryToKill();
        await start();
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
