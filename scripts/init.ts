import dotenv from "dotenv";
// @deno-types="npm:@types/mysql"
import mysql from "mysql2";
import { existsSync } from "@std/fs";
import {
    decoder,
    encoder,
    envKeys,
    logInitDb,
    logInitFiles,
} from "./tool/index.ts";
import { promisify } from "node:util";
import { parse } from "@std/path/parse";
import { blue } from "@std/fmt/colors";

async function initMysqlTables(): Promise<void> {
    const sqlFileList = [
        "./scripts/queries/create_users.sql",
        "./scripts/queries/create_groups.sql",
        "./scripts/queries/create_group_members.sql",
        "./scripts/queries/create_group_roles.sql",
        "./scripts/queries/create_group_role_assigns.sql",
        "./scripts/queries/create_group_messages.sql",
    ];
    logInitDb.info(
        `We want to create tables:\n${blue(" - ")}%s`,
        sqlFileList.map((p) => parse(p).base).join("\n" + blue(" - ")),
    );
    const connection = mysql.createConnection({
        password: Deno.env.get(envKeys.DB_PASSWORD),
        database: Deno.env.get(envKeys.DB_NAME),
        user: Deno.env.get(envKeys.DB_USER),
        host: Deno.env.get(envKeys.DB_HOST),
        port: Number(Deno.env.get(envKeys.DB_PORT)),
    });

    const decoder = new TextDecoder("utf-8");
    const connect = promisify(connection.connect.bind(connection));
    const execQuery = promisify(connection.query.bind(connection));
    const disconnect = promisify(connection.end.bind(connection));

    logInitDb.start("Connecting to the database using .env confifuration")
    await connect();
    logInitDb.end(true)
    logInitDb.warn(
        "If you are trying to reinitialize the database, this have not changeed existing tables. Delete or change them manually.",
    );
    for (const sqlFile of sqlFileList) {
        logInitDb.start(`Executing '${sqlFile}'`);
        const sqlString = decoder.decode(Deno.readFileSync(sqlFile));
        await execQuery(sqlString);
        logInitDb.end(true)
    }

    logInitDb.success(
        "All queries have been executed.",
    );
    logInitDb.start("Disconnecting from the database")
    await disconnect();
    logInitDb.end(true)
}

function initEnvFile(path: string): void {
    type EnvKeyEntry = {
        value?: string | number | boolean;
        comment?: string;
    };
    const defaultEnv = new Map<string, EnvKeyEntry>();
    defaultEnv.set(envKeys.JWT_KEY, {
        comment: "use any online jwt generator to fill this value:\n" +
            "- https://jwtsecret.com/generate",
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

try {
    const path = ".env"
    logInitFiles.start(`Initializing '${path}'`);
    initEnvFile(path);
    logInitFiles.end(true)
} catch (error) {
    logInitFiles.error(error);
    Deno.exit(1);
}

if (!Deno.args.includes("nodb")) {
    try {
        await initMysqlTables();
    } catch (error) {
        logInitDb.error(error);
        logInitDb.warn(
            "If the initialization fails because of references,\n" +
                "we are supposed to CHANGE THE ORDER: './scripts/init.ts'.",
        );
        Deno.exit(1);
    }
}
