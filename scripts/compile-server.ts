import { binaryInfo, compile, machineInfo } from "./tool/compile-binary.ts";
import { logServerComp, taskDotenv } from "./tool/constants.ts";

taskDotenv(logServerComp);

const [os, arch] = machineInfo();
const { filePath } = binaryInfo(os, arch);
const task = logServerComp.task({ text: "Compiling " + filePath }).start();
const success = await compile(os, arch);
if (!success) Deno.exit(1);
task.end(success ? "completed" : "failed");
