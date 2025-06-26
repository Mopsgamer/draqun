import { compile } from "./tool/compile-binary.ts";
import { logServerComp } from "./tool/constants.ts";

const { stdout } = new Deno.Command("go", {
    args: ["env", "GOOS", "GOARCH"],
}).outputSync();
const output = new TextDecoder().decode(stdout);
const [os, arch] = output.trim().split("\n");

if (
    !compile(os, arch, ({ filePath }) => {
        logServerComp.start(`Compiling ${filePath}`);
    })
) Deno.exit(1);
logServerComp.end();
