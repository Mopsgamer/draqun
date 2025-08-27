import { logServerComp } from "./tool/constants.ts";
import { compile } from "./tool/compile-binary.ts";

const osList = ["windows", "linux", "darwin"];
const archList = ["amd64", "arm64"];

const total = osList.length * archList.length;

let current: number = 0;
let success = true;
for (const os of osList) {
    for (const arch of archList) {
        success &&= compile(os, arch, ({ filePath }) => {
            current++;
            logServerComp.start(
                `Compiling ${current}/${total} ${filePath}`,
            );
        });
        logServerComp.end();
    }
}

if (!success) {
    Deno.exit(1);
}
