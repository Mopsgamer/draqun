import { compileTask } from "./tool/compile-binary.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "back",
		description: "Compile the backend for the current platform.",
		usage: "bun run back [dev]",
		options: [
			"-h, --help  Show this help.",
			"dev         Build with the development (lite) tag.",
		],
	})
) {
	process.exit(0);
}

const isDev = process.argv.slice(2).includes("dev");
if (!(await compileTask(isDev))) process.exit(1);
