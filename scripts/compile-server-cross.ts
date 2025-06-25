import { ensureDir } from "@std/fs";
import { resolve } from "@std/path";
import { logServerComp } from "./tool/constants.ts";
import { underline } from "@std/fmt/colors";

const isParallel = !Deno.args.includes("--queue");

const OS_LIST = ["windows", "linux", "darwin"];
const ARCH_LIST = ["amd64", "arm64"];
const releaseDir = "dist";
await ensureDir(releaseDir);

const total = OS_LIST.length * ARCH_LIST.length;
let current = 0;

async function compile(OS: string, ARCH: string): Promise<void> {
    current++;
    let binName = `server-${OS}-${ARCH}`;
    if (OS === "windows") binName += ".exe";
    if (!isParallel) {
        logServerComp.start(`Compiling ${current}/${total} ${binName}`);
    }

    const outPath = `${releaseDir}/${binName}`;
    const cmd = [
        "go",
        "build",
        "-tags",
        "prod",
        "-o",
        outPath,
        ".",
    ];
    const env = {
        ...Deno.env.toObject(),
        GOOS: OS,
        GOARCH: ARCH,
        GOCACHE: resolve(`${releaseDir}/cache`),
    };
    const command = new Deno.Command(cmd[0], {
        args: cmd.slice(1),
        env,
        stdout: "inherit",
        stderr: "inherit",
    });
    const { success } = await command.output();
    if (!success) {
        logServerComp.error(`Failed to build ${underline(binName)}.`);
        Deno.exit(1);
    }
    if (!isParallel) {
        logServerComp.end("completed");
    }
}

if (isParallel) {
    logServerComp.info(
        "You can use --queue flag to avoid parallel compilations.",
    );
    logServerComp.start(`Compiling for ${total} targets`);
}

const routineList: Promise<void>[] = [];
for (const OS of OS_LIST) {
    for (const ARCH of ARCH_LIST) {
        if (!isParallel) {
            await compile(OS, ARCH);
            continue;
        }
        routineList.push(compile(OS, ARCH));
    }
}

await Promise.all(routineList);
if (isParallel) {
    logServerComp.end();
}
logServerComp.success("Cross-compilation task completed.");
