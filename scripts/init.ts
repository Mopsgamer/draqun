import { existsSync } from "@std/fs";
import { resolve } from "@std/path";
import { decoder, logInitFiles } from "./tool/constants.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "init",
		description: "Configure the Git pre-commit hook.",
		usage: "deno task init [options]",
		options: [
			"-h, --help  Show this help.",
			"--no-git-hook  Skip installing, or remove an existing managed pre-commit hook.",
		],
	})
) {
	Deno.exit(0);
}

async function initGitHook(disabled: boolean): Promise<void> {
	const result = await new Deno.Command("git", {
		args: ["rev-parse", "--git-path", "hooks"],
		stdout: "piped",
		stderr: "piped",
	}).output();

	if (!result.success) {
		if (disabled) {
			logInitFiles.info("Git hook setup skipped.");
			return;
		}
		throw new Error(
			"Unable to locate the Git hooks directory: " +
				decoder.decode(result.stderr).trim(),
		);
	}

	const hookPath = resolve(
		decoder.decode(result.stdout).trim(),
		"pre-commit",
	);
	const hookMarker = "# draqun managed pre-commit hook";
	const hookContents = `#!/bin/sh
${hookMarker}
repo_root=$(git rev-parse --show-toplevel) || exit 1
cd "$repo_root" || exit 1
deno task prepare
`;

	if (existsSync(hookPath)) {
		const existingHook = decoder.decode(await Deno.readFile(hookPath));
		if (!existingHook.includes(hookMarker)) {
			if (disabled) {
				logInitFiles.warn(
					`Leaving existing unmanaged pre-commit hook at '${hookPath}' unchanged.`,
				);
				return;
			}
			throw new Error(
				`A pre-commit hook already exists at '${hookPath}'. ` +
					"Preserve it and add 'deno task prepare' to that hook.",
			);
		}
		if (disabled) {
			await Deno.remove(hookPath);
			return;
		}
	} else if (disabled) {
		return;
	}

	await Deno.writeTextFile(hookPath, hookContents);
	await Deno.chmod(hookPath, 0o755);
}

const isNoGitHook = Deno.args.includes("--no-git-hook");

const hook = await logInitFiles.task({
	text: isNoGitHook
		? "Disabling Git pre-commit hook"
		: "Configuring Git pre-commit hook (disable with --no-git-hook)",
}).startRunner(() => initGitHook(isNoGitHook));

if (hook.state === "failed") Deno.exit(1);
