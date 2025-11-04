import { printErrors, type TaskStateEnd } from "@m234/logger";
import { binaryInfo, machineInfo } from "./tool/compile-binary.ts";
import { logProd } from "./tool/constants.ts";
import kill from "tree-kill";

const [os, arch] = machineInfo();
const { filePath } = binaryInfo(os, arch);

let pid = new Deno.Command(filePath, {
    stdout: "inherit",
    stderr: "inherit",
}).spawn().pid;

setInterval(() => {
    logProd.task({ text: "Refreshing" })
        .startRunner(printErrors(logProd, start));
}, 60 * 1000);

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

    if (pid > 0) kill(pid);

    pid = new Deno.Command(filePath, {
        stdout: "piped",
        stderr: "piped",
    }).spawn().pid;

    return "completed";
}
