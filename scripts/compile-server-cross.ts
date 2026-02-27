import { logServerComp, taskDotenv } from "./tool/constants.ts";
import { binaryInfo, compile } from "./tool/compile-binary.ts";
import { taskGitJson } from "./tool/generate-git.ts";

taskDotenv(logServerComp);
await taskGitJson(logServerComp);

const osList = ["windows", "linux", "darwin"] as const;
const archList = ["amd64", "arm64"] as const;

const targets: [os: string, arch: string][] = []
for (const os of osList) {
    for (const arch of archList) {
        targets.push([os, arch])
    }
}

let success = true;
logServerComp.info("Plan:")
for (const [os, arch] of targets) {
    const { filePath } = binaryInfo(os, arch);
    await logServerComp.info("\t" + filePath)
}
for (const [os, arch] of targets) {
    const { filePath } = binaryInfo(os, arch);
    const task = logServerComp.task({
        text: `Compiling ${filePath}`,
    });
    task.start();
    const result = await compile(os, arch);
    success &&= result;
    task.end(result ? "completed" : "failed");
}

if (!success) {
    Deno.exit(1);
}
