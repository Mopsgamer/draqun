import {
	type DefaultTaskOptions,
	Logger,
	type LoggerOptions,
} from "@jsr/m234__logger";
import fs from "node:fs/promises";

const defaultTaskOptions: DefaultTaskOptions = {
	suffixDuration: 0n,
	interactive: false,
};
const lopts: LoggerOptions = { prefix: "🐉", defaultTaskOptions };

lopts.prefix = "🧑‍💻 Dev";
export const logDevelopment = new Logger(lopts);
lopts.prefix = "🚚 Prod";
export const logProd = new Logger(lopts);
lopts.prefix = "🚚 Release";
export const logRelease = new Logger(lopts);
lopts.prefix = "📦 Back";
export const logServerComp = new Logger(lopts);
lopts.prefix = "📦 Front";
export const logClientComp = new Logger(lopts);
lopts.prefix = "🔨 DB";
export const logInitDb = new Logger(lopts);
lopts.prefix = "🔨 Files";
export const logInitFiles = new Logger(lopts);

/**
 * Consider using same value in the environment/config.go and package.json.
 */
export const distFolder = "dist";
await fs.mkdir(distFolder, { recursive: true });

export const encoder = new TextEncoder();
export const decoder = new TextDecoder("utf-8");

export enum envKeys {
	JWT_KEY = "JWT_KEY",
	USER_AUTH_TOKEN_EXPIRATION = "USER_AUTH_TOKEN_EXPIRATION",
	CHAT_MESSAGE_MAX_LENGTH = "CHAT_MESSAGE_MAX_LENGTH",

	PORT = "PORT",
	DB_PATH = "DB_PATH",
}
