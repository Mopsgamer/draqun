import dotenv from "dotenv";
import { Logger } from "@m234/logger";
import { ensureDir } from "@std/fs/ensure-dir";

export const logDevelopment = new Logger({ prefix: "development" });
export const logRelease = new Logger({ prefix: "release" });
export const logServerComp = new Logger({ prefix: "server-compilation" });
export const logClientComp = new Logger({ prefix: "client-compilation" });
export const logInitDb = new Logger({ prefix: "init-database" });
export const logInitFiles = new Logger({ prefix: "init-files" });

const task = logInitFiles.task({ text: "Loading ./.env" }).start();
dotenv.config({ quiet: true });
task.end("completed");

/**
 * Consider using same value in the environment/config.go and deno.json.
 */
export const distFolder = "dist";
await ensureDir(distFolder);

export const encoder = new TextEncoder();
export const decoder = new TextDecoder("utf-8");

export enum envKeys {
    JWT_KEY = "JWT_KEY",
    USER_AUTH_TOKEN_EXPIRATION = "USER_AUTH_TOKEN_EXPIRATION",
    CHAT_MESSAGE_MAX_LENGTH = "CHAT_MESSAGE_MAX_LENGTH",

    PORT = "PORT",

    DB_PASSWORD = "DB_PASSWORD",
    DB_NAME = "DB_NAME",
    DB_USER = "DB_USER",
    DB_HOST = "DB_HOST",
    DB_PORT = "DB_PORT",
}
