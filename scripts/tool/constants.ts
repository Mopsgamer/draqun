import dotenv from "dotenv";
import {
	type DefaultTaskOptions,
	Logger,
	type LoggerOptions,
} from "@m234/logger";
import { ensureDir } from "@std/fs/ensure-dir";

const defaultTaskOptions: DefaultTaskOptions = { suffixDuration: 0n };
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

export function taskDotenv(
	logger: Logger,
	destination = ".env",
): void {
	logger.task({ text: `Loading '${destination}'` }).startRunner(
		() => {
			dotenv.config({ quiet: true });
		},
	);
}

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
	DB_PATH = "DB_PATH",
}
