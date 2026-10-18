import { DatabaseSync } from "node:sqlite";
import { resolve } from "@std/path";
import { decoder, logInitDb } from "./tool/constants.ts";

export const MIGRATIONS = [
	{ version: 1, file: "alter_group_roles.sql" },
	{ version: 2, file: "alter_group_messages.sql" },
	{ version: 3, file: "alter_channels.sql" },
	{ version: 4, file: "alter_users.sql" },
];

export function runMigrations(
	dbPath = Deno.env.get("DB_PATH") || "app_data.db",
): void {
	const db = new DatabaseSync(dbPath);
	using _ = {
		[Symbol.dispose](): void {
			db.close();
		},
	};

	let currentVersion = 0;
	try {
		const row = db.prepare("PRAGMA user_version").get() as
			| { user_version?: number; [key: string]: unknown }
			| undefined;
		if (row) {
			currentVersion = Number(Object.values(row)[0] ?? 0);
		}
	} catch {
		currentVersion = 0;
	}

	for (const migration of MIGRATIONS) {
		if (migration.version > currentVersion) {
			const filePath = resolve("scripts/queries/migrate", migration.file);
			try {
				const sql = decoder.decode(Deno.readFileSync(filePath));
				const statements = sql
					.split(";")
					.map((s) => s.trim())
					.filter((s) => s.length > 0);

				for (const stmt of statements) {
					try {
						db.exec(stmt);
					} catch (err) {
						const msg = (err as Error).message.toLowerCase();
						if (
							msg.includes("duplicate column name") ||
							msg.includes("already exists") ||
							msg.includes("no such table")
						) {
							// Ignored defensively if column already exists or table does not exist yet
							continue;
						}
						throw err;
					}
				}

				db.exec(`PRAGMA user_version = ${migration.version}`);
				logInitDb.info(
					`Applied migration v${migration.version}: ${migration.file}`,
				);
			} catch (err) {
				logInitDb.error(
					`Failed to apply migration v${migration.version} (${migration.file}): ${
						(err as Error).message
					}`,
				);
				throw err;
			}
		}
	}
}

if (import.meta.main) {
	runMigrations();
}
