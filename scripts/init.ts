import fs from "node:fs/promises";
import fsSync from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { logInitFiles } from "./tool/constants.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "init",
		description: "Configure the Git pre-commit hook.",
		usage: "bun run init [options]",
		options: [
			"-h, --help  Show this help.",
			"--no-git-hook  Skip installing, or remove an existing managed pre-commit hook.",
		],
	})
) {
	process.exit(0);
}

async function initGitHook(disabled: boolean): Promise<void> {
	const res = spawnSync("git", ["rev-parse", "--git-path", "hooks"], {
		encoding: "utf-8",
	});

	if (res.status !== 0 || !res.stdout) {
		if (disabled) {
			logInitFiles.info("Git hook setup skipped.");
			return;
		}
		throw new Error(
			"Unable to locate the Git hooks directory: " + (res.stderr || "").trim(),
		);
	}

	let hookDir = res.stdout.trim();
	if (hookDir === "/dev/null" || hookDir.startsWith("/dev/null/")) {
		hookDir = ".git/hooks";
	}

	const hookPath = path.resolve(hookDir, "pre-commit");
	const hookMarker = "# draqun managed pre-commit hook";
	const hookContents = `#!/bin/sh
${hookMarker}
repo_root=$(git rev-parse --show-toplevel) || exit 1
cd "$repo_root" || exit 1
bun lint --no-script && bun fmt --no-script --staged
`;

	if (fsSync.existsSync(hookPath)) {
		const existingHook = await fs.readFile(hookPath, "utf-8");
		if (!existingHook.includes(hookMarker)) {
			if (disabled) {
				logInitFiles.warn(
					`Leaving existing unmanaged pre-commit hook at '${hookPath}' unchanged.`,
				);
				return;
			}
			throw new Error(
				`A pre-commit hook already exists at '${hookPath}'. ` +
					"Preserve it and add 'bun lint --no-script && bun fmt --no-script --staged' to that hook.",
			);
		}
		if (disabled) {
			await fs.unlink(hookPath);
			return;
		}
	} else if (disabled) {
		return;
	}

	await fs.mkdir(path.dirname(hookPath), { recursive: true });
	await fs.writeFile(hookPath, hookContents, "utf-8");
	await fs.chmod(hookPath, 0o755);
}

const isNoGitHook = process.argv.slice(2).includes("--no-git-hook");

const hook = await logInitFiles
	.task({
		text: isNoGitHook
			? "Disabling Git pre-commit hook"
			: "Configuring Git pre-commit hook (disable with --no-git-hook)",
	})
	.startRunner(() => initGitHook(isNoGitHook));

if (hook.state === "failed") process.exit(1);
