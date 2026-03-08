import kill from "tree-kill";
import { sep } from "node:path";
import { existsSync } from "@std/fs";
import { logDevelopment, taskDotenv } from "./tool/constants.ts";

taskDotenv(logDevelopment);

const paths = [
    "server" + sep,
    "client-lite.go",
    "main.go",
    ".git" + sep + "ORIG_HEAD",
];
if (existsSync(".env")) paths.push(".env");

async function start(signal?: AbortSignal): Promise<number> {
    const generateCommand = new Deno.Command("go", {
        args: ["generate", "./..."],
        signal,
    });
    const { success } = await generateCommand.output();
    if (!success) return NaN;

    const serverCommand = new Deno.Command("go", {
        args: ["run", "-tags", "lite", "."],
        signal,
    });
    return serverCommand.spawn().pid;
}

async function watchAndRestart(): Promise<void> {
    const watcher = Deno.watchFs(paths, { recursive: true });
    await start();
    let pid = NaN;
    let controller: AbortController | undefined;
    let timeout = NaN;
    for await (const event of watcher) {
        if (
            !(
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            )
        ) continue;

        clearTimeout(timeout);

        timeout = setTimeout(async () => {
            if (!isNaN(pid)) {
                controller!.abort();
                tryToKill(pid);
            }
            controller = new AbortController();
            await logDevelopment.info(
                "Refreshing (" + event.kind + " " + event.paths + ")",
            );
            pid = await start(controller.signal);
        }, 50);
    }
}

function tryToKill(pid: number): void {
    try {
        kill(pid, "SIGTERM");
    } catch { /* empty */ }
}

watchAndRestart();
