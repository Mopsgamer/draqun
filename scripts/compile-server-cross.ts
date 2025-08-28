import { logServerComp } from "./tool/constants.ts";
import { binaryInfo, compile } from "./tool/compile-binary.ts";
import { limit1 } from "./tool/limit1.ts";
import type { Task, TaskStateEnd } from "@m234/logger";

const osList = ["windows", "linux", "darwin"];
const archList = ["amd64", "arm64"];

let success = true;
const queue: Promise<Task>[] = []
for (const os of osList) {
    for (const arch of archList) {
        const { filePath } = binaryInfo(os, arch);
        queue.push(logServerComp.task({
            text: `Compiling ${filePath}`,
        }).startRunner(limit1(async (): Promise<TaskStateEnd> => {
            const result = await compile(os, arch);
            success &&= result;
            return result ? "completed" : "failed";
        })));
    }
}

await Promise.allSettled(queue)

if (!success) {
    Deno.exit(1);
}
