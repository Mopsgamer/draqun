import dotenv from "dotenv";
// @deno-types="npm:@types/mysql"
import mysql from "mysql";
import { existsSync } from "@std/fs";

enum envKeys {
    ENVIRONMENT = "ENVIRONMENT",
    JWT_KEY = "JWT_KEY",
    DB_PASSWORD = "DB_PASSWORD",
    DB_NAME = "DB_NAME",
    DB_USER = "DB_USER",
    DB_HOST = "DB_HOST",
    DB_PORT = "DB_PORT",
}

function initMysqlTables(): void {
    const queryList = [
        `CREATE TABLE IF NOT EXISTS users (
		id INT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
		nickname VARCHAR(255) NOT NULL COMMENT 'Search-friendly changable identificator',
		username VARCHAR(255) NOT NULL COMMENT 'Customizable name',
		email VARCHAR(255) NOT NULL,
		phone VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) NOT NULL,
		avatar VARCHAR(255) DEFAULT NULL,
		created_at DATETIME NOT NULL COMMENT 'Account create time',
		last_seen DATETIME NOT NULL COMMENT 'Last seen time',
		PRIMARY KEY (id)
	    ) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Users data'`,
    ];

    const connection = mysql.createConnection({
        password: Deno.env.get(envKeys.DB_PASSWORD),
        database: Deno.env.get(envKeys.DB_NAME),
        user: Deno.env.get(envKeys.DB_USER),
        host: Deno.env.get(envKeys.DB_HOST),
        port: Number(Deno.env.get(envKeys.DB_PORT)),
    });

    connection.connect((err) => {
        if (err) throw err;
        console.log("Connected to the database using .env confifuration.");
    });

    for (const query of queryList) {
        connection.query(
            query,
            (err) => {
                if (err) throw err;
                console.log("Created the 'users' table if not exists.");
            },
        );
    }

    connection.end((err) => {
        if (err) throw err;
        console.log("Disconnected from the database.");
    },);
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
        comment: "use any online jwt generator to fill this value",
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
        comment: "connection port\ndefault: 3306",
    });

    const path = ".env";
    const decoder = new TextDecoder();
    const env = existsSync(path)
        ? dotenv.parse(decoder.decode(Deno.readFileSync(path)))
        : {};

    const encoder = new TextEncoder();
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

    console.log("Writed " + path);
}

initEnvFile();
initMysqlTables();
