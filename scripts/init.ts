import dotenv from "dotenv";
import mysql from "mysql2/promise";
import { existsSync } from "@std/fs";
import {
    decoder,
    encoder,
    envKeys,
    logInitDb,
    logInitFiles,
} from "./tool/constants.ts";
import { parse } from "@std/path/parse";
import { blue } from "@std/fmt/colors";
import { format } from "@m234/logger";

function logError(e: unknown) {
    logInitDb.error(format(e));
    throw e;
}

async function initMysqlTables(): Promise<void> {
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
    logInitDb.info(
        `We want to create tables:\n${blue(" - ")}` +
            sqlFileList.map((p) => parse(p).base).join("\n" + blue(" - ")),
    );
    logInitDb.info("You can pass 'nodb' to ignore DB initialization step.");

    const con = await mysql.createConnection({
        password: Deno.env.get(envKeys.DB_PASSWORD),
        database: Deno.env.get(envKeys.DB_NAME),
        user: Deno.env.get(envKeys.DB_USER),
        host: Deno.env.get(envKeys.DB_HOST),
        port: Number(Deno.env.get(envKeys.DB_PORT)),
    });
    con.catch(logError);
    const connection = await con;
    const decoder = new TextDecoder("utf-8");
    const connect = async (): Promise<void> => {
        await connection.connect().catch(logError);
    };
    const execQuery = connection.query;
    execQuery.bind(connection);
    const disconnect = async (): Promise<void> => {
        await connection.end().catch(logError);
    };

    await logInitDb.task({ text: "Connecting" }).startRunner(connect).catch(
        logError,
    );
    logInitDb.warn(
        "If you are trying to reinitialize the database, this have not changeed existing tables. Delete or change them manually.",
    );
    for (const sqlFile of sqlFileList) {
        await logInitDb.task({ text: `Executing '${sqlFile}'` })
            .startRunner(async () => {
                const sqlString = decoder.decode(Deno.readFileSync(sqlFile));
                await execQuery(sqlString);
            }).catch((e) => {
                logInitDb.warn(
                    "If the initialization fails because of references,\n" +
                        "we are supposed to CHANGE THE ORDER: './scripts/init.ts'.",
                );
                logError(e);
            });
    }

    logInitDb.success(
        "All queries have been executed.",
    );
    await logInitDb.task({ text: "Disconnecting from the database" })
        .startRunner(disconnect).catch(logError);
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

    defaultEnv.set(envKeys.DB_PASSWORD, { comment: "connection password" });
    defaultEnv.set(envKeys.DB_NAME, {
        value: "mysql",
        comment: "connection name",
    });
    defaultEnv.set(envKeys.DB_USER, {
        value: "admin",
        comment: "connection user\nroot user is not recommended",
    });
    defaultEnv.set(envKeys.DB_HOST, {
        value: "localhost",
        comment: "connection host",
    });
    defaultEnv.set(envKeys.DB_PORT, {
        value: 3306,
        comment: "database port",
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
    logInitFiles.task({ text: `Initializing '${path}'` }).startRunner(() => {
        initEnvFile(path);
    });
}

if (!Deno.args.includes("nodb")) {
    await logInitDb.task({ text: "Initializing DB" })
        .startRunner(initMysqlTables)
        .catch(() => {});
}
