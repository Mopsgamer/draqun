import type { TaskStateEnd } from "@m234/logger";
import { binaryInfo, compile, machineInfo } from "./tool/compile-binary.ts";
import { distFolder, logProd } from "./tool/constants.ts";
import kill from "tree-kill";
import { compileDist } from "./tool/compile-dist.ts";
import { existsSync } from "@std/fs/exists";

const [os, arch] = machineInfo();
const { filePath } = binaryInfo(os, arch);

if (!existsSync(distFolder) || !existsSync(distFolder + "/static")) {
    await compileDist(true);
}

if (!await compile(os, arch)) {
    Deno.exit(1);
}

let pid = new Deno.Command(filePath, {
    stdout: "inherit",
    stderr: "inherit",
}).spawn().pid;

setInterval(() => {
    logProd.task({ text: "Refreshing" })
        .startRunner(start);
}, 3 * 60 * 1000);

async function start(): Promise<TaskStateEnd> {
    const fetch = await new Deno.Command("git", {
        args: ["fetch"],
        stdout: "piped",
        stderr: "piped",
    }).output();

    if (fetch.stdout.toString() == "") return "skipped";

    await new Deno.Command("git", {
        args: ["pull"],
        stdout: "piped",
        stderr: "piped",
    }).output();

    await compileDist(true);
    if (!await compile(os, arch)) {
        logProd.warn("Compilation failed, keeping the old version running.");
        return "aborted";
    }

    if (pid > 0) kill(pid);

    pid = new Deno.Command(filePath, {
        stdout: "piped",
        stderr: "piped",
    }).spawn().pid;

    return "completed";
}
