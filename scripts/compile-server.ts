import { binaryInfo, compile } from "./tool/compile-binary.ts";
import { logServerComp } from "./tool/constants.ts";
import { writeGitJson } from "./tool/generate-git.ts";

await writeGitJson();

const { stdout } = new Deno.Command("go", {
    args: ["env", "GOOS", "GOARCH"],
}).outputSync();
const output = new TextDecoder().decode(stdout);
const [os, arch] = output.trim().split("\n");

const { filePath } = binaryInfo(os, arch);
const task = logServerComp.task({ text: "Compiling " + filePath }).start();
const success = await compile(os, arch);
if (!success) Deno.exit(1);
task.end(success ? "completed" : "failed");
