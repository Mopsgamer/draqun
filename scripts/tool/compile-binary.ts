import { distFolder, logServerComp } from "./constants.ts";
import { resolve } from "@std/path";

export type BinaryInfo = {
	fileName: string;
	filePath: string;
};

export async function compileTask(dev?: boolean): Promise<boolean>;
export async function compileTask(
	dev: true,
	run: true,
): Promise<Deno.ChildProcess | false>;
export async function compileTask(
	dev = false,
	run = false,
): Promise<Deno.ChildProcess | boolean> {
	const [os, arch] = machineInfo();
	const { filePath } = binaryInfo(os, arch);
	const task = logServerComp.task({
		text: run ? "Compiling and starting" : "Compiling " + filePath,
	}).start();
	//@ts-expect-error we do not care about signature
	const result = await compile(os, arch, dev, run);
	task.end(result ? "completed" : "failed");
	return result;
}

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
	dev?: boolean,
): Promise<boolean>;
export async function compile(
	os: string,
	arch: string,
	dev: true,
	run: true,
): Promise<Deno.ChildProcess | false>;
export async function compile(
	os: string,
	arch: string,
	dev = false,
	run = false,
): Promise<Deno.ChildProcess | boolean> {
	const { filePath } = binaryInfo(os, arch);

	const env = {
		GOCACHE: resolve(`${distFolder}/cache`),
		...Deno.env.toObject(),
	};

	let child = await new Deno.Command("go", {
		args: [
			"generate",
			"./...",
		],
		env,
		stdout: "inherit",
		stderr: "inherit",
	}).output();

	if (!child.success) return false;

	const spawn = new Deno.Command("go", {
		args: [
			run ? "run" : "build",
			"-tags",
			dev ? "lite" : "prod",
			...(run ? [] : ["-o", filePath]),
			".",
		],
		env: {
			...env,
			GOOS: os,
			GOARCH: arch,
		},
	}).spawn();

	if (dev && run) return spawn;
	child = await spawn.output();

	return child.success;
}
