import { DatabaseSync } from "node:sqlite";
import { parse } from "@std/dotenv";
import { existsSync } from "@std/fs";
import { resolve } from "@std/path";
import {
	decoder,
	encoder,
	envKeys,
	logInitDb,
	logInitFiles,
} from "./tool/constants.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "init",
		description: "Initialize local environment files, database, and Git hook.",
		usage: "deno task init [options]",
		options: [
			"-h, --help  Show this help.",
			"--no-env    Skip creating or updating the .env file.",
			"--no-db     Skip initializing the SQLite database.",
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

async function initSqliteTables(): Promise<void> {
	const sqlFileList = [
		"./scripts/queries/create_users.sql",
		"./scripts/queries/create_groups.sql",
		"./scripts/queries/create_group_members.sql",
		"./scripts/queries/create_group_roles.sql",
		"./scripts/queries/create_group_role_assignees.sql",
		"./scripts/queries/create_group_messages.sql",
		"./scripts/queries/create_group_action_memberships.sql",
		"./scripts/queries/create_group_action_kicks.sql",
		"./scripts/queries/create_group_action_bans.sql",
	];

	// SQLite uses a local file instead of a network connection
	const dbPath = "app_data.db";

	const taskConnect = logInitDb.task({
		text: "Opening SQLite file",
		indent: 1,
	});

	// We use DatabaseSync for the initialization script because it's simpler for local file I/O
	let db: DatabaseSync;
	taskConnect.startRunner(() => {
		db = new DatabaseSync(dbPath);
	});

	if (taskConnect.state === "failed") {
		return;
	}

	// Ensure the file is closed when the script finishes
	using _ = {
		[Symbol.dispose](): void {
			db.close();
		},
	};

	for (const sqlFile of sqlFileList) {
		const execution = await logInitDb.task({
			text: "Executing " + sqlFile,
			indent: 1,
		})
			.startRunner(() => {
				const sqlString = decoder.decode(Deno.readFileSync(sqlFile));
				// db.exec runs the entire file content at once
				try {
					db.exec(sqlString);
				} catch (error) {
					logInitDb.error((error as Error).message);
					return "failed";
				}
			});

		if (execution.state === "failed") {
			logInitDb.warn(
				"If the initialization fails because of references, check the execution order of your .sql files.",
			);
			return;
		}
	}
}

function initEnvFile(path: string): void {
	type EnvKeyEntry = {
		value?: string | number | boolean;
		comment?: string;
	};
	const defaultEnv = new Map<string, EnvKeyEntry>();
	defaultEnv.set(envKeys.JWT_KEY, {
		comment: "use any online jwt generator to fill this value:\n" +
			"- https://randomfungenerator.com/generators/jwt-generator",
	});
	defaultEnv.set(envKeys.USER_AUTH_TOKEN_EXPIRATION, {
		value: 180,
		comment: "in minutes",
	});
	defaultEnv.set(envKeys.CHAT_MESSAGE_MAX_LENGTH, {
		value: 8000,
		comment: "max characters quantity after spaces are trimed",
	});

	defaultEnv.set(envKeys.PORT, {
		value: 3000,
		comment: "application port",
	});

	defaultEnv.set(envKeys.DB_PATH, {
		value: "app_data.db",
		comment: "local sqlite database file path",
	});

	const env = existsSync(path)
		? parse(decoder.decode(Deno.readFileSync(path)))
		: {};

	Deno.writeFileSync(
		path,
		encoder.encode(
			Array.from(defaultEnv.entries()).map(
				([key, { value, comment }]) => {
					env[key] ||= value === undefined ? "" : String(value);
					Deno.env.set(key, env[key]);
					if (value === undefined) {
						comment += "\ndefault: <empty>";
					} else {
						comment += "\ndefault: " + value;
					}
					return `${
						comment ? "# " + comment.replaceAll("\n", "\n# ") + "\n" : ""
					}${key}=${env[key]}\n\n`;
				},
			).join(""),
		),
	);
}

const isNoEnv = Deno.args.includes("--no-env");
const isNoDb = Deno.args.includes("--no-db");
const isNoGitHook = Deno.args.includes("--no-git-hook");

if (!isNoDb) {
	logInitDb.warn(
		"Existing tables are unchanged; delete 'app_data.db' to reset.",
	);
}

if (!isNoEnv) {
	const path = ".env";
	await logInitFiles.task({
		text: `Initializing '${path}' (disable with --no-env)`,
	})
		.startRunner(() => initEnvFile(path));
}

if (!isNoDb) {
	await logInitDb.task({
		text: "Initializing DB (disable with --no-db)",
	})
		.startRunner(initSqliteTables);
}

const hook = await logInitFiles.task({
	text: isNoGitHook
		? "Disabling Git pre-commit hook"
		: "Configuring Git pre-commit hook (disable with --no-git-hook)",
}).startRunner(() => initGitHook(isNoGitHook));

if (hook.state === "failed") Deno.exit(1);
