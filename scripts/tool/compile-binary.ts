import { distFolder, logServerComp } from "./constants.ts";
import { resolve } from "@std/path";
import { underline } from "@std/fmt/colors";

const isDev = Deno.args.includes("dev");

export type BinaryInfo = {
    fileName: string;
    filePath: string;
};
export function compile(
    os: string,
    arch: string,
    onStart?: (info: BinaryInfo) => void,
): boolean {
    let fileName = `server-${os}-${arch}`;
    if (os === "windows") fileName += ".exe";
    const filePath = `${distFolder}/${fileName}`;
    onStart?.({ fileName, filePath });

    const { success } = new Deno.Command("go", {
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
    }).outputSync();

    if (!success) {
        logServerComp.error(`Failed to build ${underline(fileName)}.`);
    }
    return success;
}
