import { distFolder, logServerComp } from "./constants.ts";
import { resolve } from "@std/path";

export type BinaryInfo = {
	fileName: string;
	filePath: string;
};

export async function compileTask(run?: false): Promise<boolean>;
export async function compileTask(
	run: true,
): Promise<Deno.ChildProcess | false>;
export async function compileTask(
	run = false,
): Promise<Deno.ChildProcess | boolean> {
	const [os, arch] = machineInfo();
	const { filePath } = binaryInfo(os, arch);
	const task = logServerComp.task({
		text: run ? "Compiling and starting" : "Compiling " + filePath,
	}).start();
	if (!run) {
		const result = await compile(os, arch, run);
		task.end(result ? "completed" : "failed");
		return result;
	}
	const result = await compile(os, arch, run);
	if (result) task.end("completed");
	else task.end("failed");
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
	run?: false,
): Promise<boolean>;
export async function compile(
	os: string,
	arch: string,
	run: true,
): Promise<Deno.ChildProcess | false>;
export async function compile(
	os: string,
	arch: string,
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
			run ? "lite" : "prod",
			...(run ? [] : ["-o", filePath]),
			".",
		],
		env: {
			...env,
			GOOS: os,
			GOARCH: arch,
		},
	}).spawn();

	if (run) return spawn;
	child = await spawn.output();

	return child.success;
}
