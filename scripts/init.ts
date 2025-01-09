import dotenv from "dotenv";
// @deno-types="npm:@types/mysql"
import mysql from "mysql2";
import { existsSync } from "@std/fs";
import { decoder, encoder, envKeys, logInitDb, logInitFiles } from "./tool.ts";
import { promisify } from "node:util";
import { parse } from "@std/path/parse";

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
        "We want to create tables:\n%s",
        sqlFileList.map((p) => parse(p).base).join("\n -> "),
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

    await connect();
    logInitDb.success("Connected to the database using .env confifuration.");
    logInitDb.warn(
        "If you are trying to reinitialize the database, this will not change existing tables. Delete or change them manually.",
    );
    for (const sqlFile of sqlFileList) {
        logInitDb.info(`Executing '${sqlFile}'...`);
        const sqlString = decoder.decode(Deno.readFileSync(sqlFile));
        await execQuery(sqlString);
    }

    await disconnect();
    logInitDb.success(
        "Success. All queries executed. Disconnected from the database.",
    );
}

function initEnvFile(): void {
    type EnvKeyEntry = {
        value?: string;
        comment?: string;
    };
    const defaultEnv = new Map<string, EnvKeyEntry>();
    defaultEnv.set(envKeys.ENVIRONMENT, {
        value: "1",
        comment: "can be 0 (test), 1 (dev) or 2 (prod)\ndefault: 1",
    });
    defaultEnv.set(envKeys.JWT_KEY, {
        comment:
            "use any online jwt generator to fill this value: https://jwtsecret.com/generate",
    });
    defaultEnv.set(envKeys.PORT, {
        value: "3000",
        comment: "application port\ndefault: 3000",
    });
    defaultEnv.set(envKeys.DB_PASSWORD, { comment: "connection password" });
    defaultEnv.set(envKeys.DB_NAME, {
        value: "restapp",
        comment: "connection name\ndefault: restapp",
    });
    defaultEnv.set(envKeys.DB_USER, {
        value: "root",
        comment: "connection user\nroot user is not recommended\ndefault: root",
    });
    defaultEnv.set(envKeys.DB_HOST, {
        value: "localhost",
        comment: "connection host\ndefault: localhost",
    });
    defaultEnv.set(envKeys.DB_PORT, {
        value: "3306",
        comment: "database port\ndefault: 3306",
    });

    const path = ".env";
    const env = existsSync(path)
        ? dotenv.parse(decoder.decode(Deno.readFileSync(path)))
        : {};

    Deno.writeFileSync(
        path,
        encoder.encode(
            Array.from(defaultEnv.entries()).map(
                ([key, { value, comment }]) => {
                    env[key] ||= value ?? "";
                    Deno.env.set(key, env[key]);
                    return `${
                        comment
                            ? "# " + comment.replaceAll("\n", "\n# ") + "\n"
                            : ""
                    }${key}=${env[key]}\n\n`;
                },
            ).join(""),
        ),
    );

    logInitFiles.success("Writed " + path);
}

try {
    initEnvFile();
} catch (error) {
    logInitFiles.fatal(error);
    Deno.exit(1);
}

try {
    await initMysqlTables();
} catch (error) {
    logInitDb.fatal(error);
    logInitDb.warn(
        "If the initialization fails because of references,\n" +
            "we are supposed to CHANGE THE ORDER: './scripts/init.ts'.",
    );
    Deno.exit(1);
}
