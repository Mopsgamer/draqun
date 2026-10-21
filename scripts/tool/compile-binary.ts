import { distFolder, logServerComp } from "./constants.ts";
import { resolve } from "@jsr/std__path/resolve";
import { spawn, spawnSync, type ChildProcess } from "node:child_process";
import fs from "node:fs/promises";

export type BinaryInfo = {
	fileName: string;
	filePath: string;
};

export async function compileTask(dev?: boolean): Promise<boolean>;
export async function compileTask(
	dev: true,
	run: true,
): Promise<ChildProcess | false>;
export async function compileTask(
	dev = false,
	run = false,
): Promise<ChildProcess | boolean> {
	const [os, arch] = machineInfo();
	const { filePath } = binaryInfo(os, arch);
	const task = logServerComp
		.task({
			text: run ? "Compiling and starting" : `Compiling ${filePath}`,
		})
		.start();
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
	const proc = spawnSync("go", ["env", "GOOS", "GOARCH"], {
		encoding: "utf-8",
	});
	const output = proc.stdout ? proc.stdout.trim() : "";
	const [os, arch] = output.split("\n");
	return [os || "linux", arch || "amd64"];
}

async function runGo(
	args: string[],
	env: Record<string, string>,
): Promise<boolean> {
	return new Promise((resolvePromise) => {
		const child = spawn("go", args, {
			env,
			stdio: "inherit",
		});
		child.on("close", (code) => {
			resolvePromise(code === 0);
		});
		child.on("error", () => {
			resolvePromise(false);
		});
	});
}

async function generateGitJson(): Promise<void> {
	try {
		const git = (args: string[]) => {
			const res = spawnSync("git", args, { encoding: "utf-8" });
			return res.stdout ? res.stdout.trim() : "";
		};
		const data = {
			hash: git(["rev-parse", "--short", "HEAD"]),
			hashLong: git(["rev-parse", "HEAD"]),
			branch: git(["rev-parse", "--abbrev-ref", "HEAD"]),
		};
		await fs.mkdir(distFolder, { recursive: true });
		await fs.writeFile(`${distFolder}/git.json`, JSON.stringify(data), "utf-8");
	} catch {
		// Ignore if git is not available
	}
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
): Promise<ChildProcess | false>;
export async function compile(
	os: string,
	arch: string,
	dev = false,
	run = false,
): Promise<ChildProcess | boolean> {
	const { filePath } = binaryInfo(os, arch);

	const env = {
		GOCACHE: resolve(`${distFolder}/cache`),
		...process.env,
	};

	await generateGitJson();

	const generated = await runGo(
		["generate", "./..."],
		env as Record<string, string>,
	);
	if (!generated) return false;

	const built = await runGo(
		["build", "-tags", dev ? "lite" : "prod", "-o", filePath, "."],
		{
			...env,
			GOOS: os,
			GOARCH: arch,
		} as Record<string, string>,
	);
	if (!built) return false;

	if (run) {
		const child = spawn(filePath, [], {
			env: {
				...env,
				GOOS: os,
				GOARCH: arch,
			},
			stdio: "inherit",
		});
		return child;
	}

	return true;
}
