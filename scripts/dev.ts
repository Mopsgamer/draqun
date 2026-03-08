import kill from "tree-kill";
import { existsSync } from "@std/fs";
import { logDevelopment, taskDotenv } from "./tool/constants.ts";

taskDotenv(logDevelopment);

const paths = ["server/", "client-lite.go", "main.go", ".git/ORIG_HEAD"];
if (existsSync(".env")) paths.push(".env");
const generateCommand = new Deno.Command("go", {
    args: ["generate", "./..."],
});
const serverCommand = new Deno.Command("go", {
    args: ["run", "-tags", "lite", "."],
});
let goRunProcess: Deno.ChildProcess | undefined = undefined;

async function start(): Promise<void> {
    goRunProcess = generateCommand.spawn();
    const { success } = await goRunProcess.output();
    if (!success) return;
    goRunProcess = serverCommand.spawn();
}

async function watchAndRestart(): Promise<void> {
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

function tryToKill(): void {
    if (goRunProcess == undefined) {
        return;
    }
    try {
        kill(goRunProcess.pid, "SIGTERM");
    } catch { /* empty */ }
    goRunProcess = undefined;
}

watchAndRestart();
