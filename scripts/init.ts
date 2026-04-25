import dotenv from "dotenv";
import { DatabaseSync } from "node:sqlite"; // Deno 2.x native
import { existsSync } from "@std/fs";
import {
    decoder,
    encoder,
    envKeys,
    logInitDb,
    logInitFiles,
    taskDotenv,
} from "./tool/constants.ts";

taskDotenv(logInitFiles);

function initSqliteTables(): void {
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
    logInitDb.info("You can pass 'nodb' to ignore DB initialization step.");

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

    logInitDb.warn(
        "Rerunning 'init' won't change existing tables. Delete 'app_data.db' if you need a full reset.",
    );

    for (const sqlFile of sqlFileList) {
        const execution = logInitDb.task({
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

    logInitDb.success(
        "All queries have been executed against app_data.db.",
    );
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
        ? dotenv.parse(decoder.decode(Deno.readFileSync(path)))
        : {};

    Deno.writeFileSync(
        path,
        encoder.encode(
            Array.from(defaultEnv.entries()).map(
                ([key, { value, comment }]) => {
                    env[key] ||= value === undefined ? "" : String(value);
                    Deno.env.set(key, env[key]);
                    if (value == undefined) {
                        comment += "\ndefault: <empty>";
                    } else {
                        comment += "\ndefault: " + value;
                    }
                    return `${
                        comment
                            ? "# " + comment.replaceAll("\n", "\n# ") + "\n"
                            : ""
                    }${key}=${env[key]}\n\n`;
                },
            ).join(""),
        ),
    );
}

if (!Deno.args.includes("noenv")) {
    const path = ".env";
    logInitFiles.task({ text: `Initializing '${path}'` })
        .startRunner(() => initEnvFile(path));
}

if (!Deno.args.includes("nodb")) {
    logInitDb.task({ text: "Initializing DB" })
        .startRunner(initSqliteTables);
}
