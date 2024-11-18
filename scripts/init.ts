import dotenv from "dotenv";
// @deno-types="npm:@types/mysql"
import mysql from "mysql";
import { existsSync } from "@std/fs";
import { logInitDb, logInitFiles } from "./tool.ts";
import {promisify} from "node:util";

enum envKeys {
    ENVIRONMENT = "ENVIRONMENT",
    JWT_KEY = "JWT_KEY",
    DB_PASSWORD = "DB_PASSWORD",
    DB_NAME = "DB_NAME",
    DB_USER = "DB_USER",
    DB_HOST = "DB_HOST",
    DB_PORT = "DB_PORT",
}

async function initMysqlTables(): Promise<void> {
    const queryList = [
        `CREATE TABLE IF NOT EXISTS app_users (
		id UNSIGNED BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
		nickname VARCHAR(255) NOT NULL COMMENT 'Search-friendly changable identificator',
		username VARCHAR(255) NOT NULL COMMENT 'Customizable name',
		email VARCHAR(255) NOT NULL,
		phone VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) NOT NULL,
		avatar VARCHAR(255) DEFAULT NULL,
		created_at DATETIME NOT NULL COMMENT 'Account create time',
		last_seen DATETIME NOT NULL COMMENT 'Last seen time',
		PRIMARY KEY (id)
	    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Restapp users'`,

        `CREATE TABLE IF NOT EXISTS app_groups (
		id UNSIGNED BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
		nickname VARCHAR(255) NOT NULL COMMENT 'Search-friendly changable identificator',
		groupname VARCHAR(255) NOT NULL COMMENT 'Customizable name',
        groupmode ENUM('dm', 'private', 'public') NOT NULL,
		password VARCHAR(255) DEFAULT NULL,
		avatar VARCHAR(255) DEFAULT NULL,
		created_at DATETIME NOT NULL COMMENT 'Group create time',
		PRIMARY KEY (id)
	    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Restapp groups'`,

        `CREATE TABLE IF NOT EXISTS app_group_roles (
		group_id UNSIGNED BIGINT NOT NULL,
        id UNSIGNED MEDIUMINT NOT NULL,
        perm_chat_read BIT NOT NULL,
        perm_chat_write BIT NOT NULL,
        perm_chat_delete BIT NOT NULL,
        perm_kick BIT NOT NULL,
        perm_ban BIT NOT NULL,
        perm_change_group BIT NOT NULL,
        perm_change_member BIT NOT NULL,
        PRIMARY KEY (group_id, id),
        FOREIGN KEY (group_id) REFERENCES app_groups(id),
	    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Restapp all groups roles'`,

        `CREATE TABLE IF NOT EXISTS app_group_members (
		group_id UNSIGNED BIGINT NOT NULL,
        user_id UNSIGNED BIGINT NOT NULL,
        role_id UNSIGNED MEDIUMINT NOT NULL,
        is_owner BIT NOT NULL,
        is_creator BIT NOT NULL,
        is_banned BIT NOT NULL,
        membername VARCHAR(255),
        PRIMARY KEY (group_id, user_id),
        FOREIGN KEY (group_id) REFERENCES app_groups(id),
        FOREIGN KEY (user_id) REFERENCES app_users(id),
        FOREIGN KEY (role_id) REFERENCES app_group_roles(id)
	    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Restapp all groups members'`,

        `CREATE TABLE IF NOT EXISTS app_messages (
		group_id UNSIGNED BIGINT NOT NULL,
        author_id UNSIGNED BIGINT NOT NULL,
        content VARCHAR(40000) NOT NULL,
        PRIMARY KEY (group_id, user_id),
        FOREIGN KEY (group_id) REFERENCES app_groups(id),
        FOREIGN KEY (author_id) REFERENCES app_users(id),
	    ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Restapp messages'`,
    ];

    const connection = mysql.createConnection({
        password: Deno.env.get(envKeys.DB_PASSWORD),
        database: Deno.env.get(envKeys.DB_NAME),
        user: Deno.env.get(envKeys.DB_USER),
        host: Deno.env.get(envKeys.DB_HOST),
        port: Number(Deno.env.get(envKeys.DB_PORT)),
    });

    const connect = promisify(connection.connect.bind(connection))
    const execQuery = promisify(connection.query.bind(connection))
    const disconnect = promisify(connection.end.bind(connection))

    await connect()
    logInitDb.info("Connected to the database using .env confifuration.");

    for (const query of queryList) {
        await execQuery(query)
    }

    logInitDb.info("Initialized.");

    await disconnect()
    logInitDb.success("Disconnected from the database.");
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

    logInitFiles.success("Writed " + path);
}

try {
    initEnvFile();
} catch (error) {
    logInitFiles.fatal(error);
    Deno.exit(1)
}

try {
    await initMysqlTables();
} catch (error) {
    logInitDb.fatal(error);
    Deno.exit(1)
}
