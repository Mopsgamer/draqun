import { distFolder } from "./constants.ts";
import { resolve } from "@std/path";

const isDev = Deno.args.includes("dev");

export type BinaryInfo = {
    fileName: string;
    filePath: string;
};

export function binaryInfo(os: string, arch: string): BinaryInfo {
    let fileName = `server-${os}-${arch}`;
    if (os === "windows") fileName += ".exe";
    const filePath = `${distFolder}/${fileName}`;
    return { fileName, filePath };
}

export function machineInfo(): [os: string, arch: string] {
    const { stdout } = new Deno.Command("go", {
        args: ["env", "GOOS", "GOARCH"],
    }).outputSync();
    const output = new TextDecoder().decode(stdout);
    const [os, arch] = output.trim().split("\n");
    return [os, arch];
}

export async function compile(
    os: string,
    arch: string,
): Promise<boolean> {
    const { filePath } = binaryInfo(os, arch);

    const { success } = await new Deno.Command("go", {
        args: [
            "build",
            "-tags",
            isDev ? "lite" : "prod",
            "-o",
            filePath,
            ".",
        ],
        env: {
            GOOS: os,
            GOARCH: arch,
            GOCACHE: resolve(`${distFolder}/cache`),
            ...Deno.env.toObject(),
        },
        stdout: "inherit",
        stderr: "inherit",
    }).output();

    return success;
}
