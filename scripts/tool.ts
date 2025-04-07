import consola from "consola";
import dotenv from "dotenv";

dotenv.config();

consola.options.formatOptions.columns = 3;

export const logBuild = consola.withTag("build");
export const logCleanup = consola.withTag("cleanup");
export const logInitDb = consola.withTag("init-database");
export const logInitFiles = consola.withTag("init-files");

export const encoder = new TextEncoder();
export const decoder = new TextDecoder("utf-8");

export enum envKeys {
    JWT_KEY = "JWT_KEY",
    USER_AUTH_TOKEN_EXPIRATION = "USER_AUTH_TOKEN_EXPIRATION",
    CHAT_MESSAGE_MAX_LENGTH = "CHAT_MESSAGE_MAX_LENGTH",

    ENVIRONMENT = "ENVIRONMENT",
    FS_COMPILED = "FS_COMPILED",
    PORT = "PORT",

    DB_PASSWORD = "DB_PASSWORD",
    DB_NAME = "DB_NAME",
    DB_USER = "DB_USER",
    DB_HOST = "DB_HOST",
    DB_PORT = "DB_PORT",
}

export const environment = Number(Deno.env.get(envKeys.ENVIRONMENT));
