import { compileTask } from "./tool/compile-binary.ts";
import { showHelp } from "./tool/help.ts";

if (
	showHelp({
		name: "back",
		description: "Compile the backend for the current platform.",
		usage: "deno task back [dev]",
		options: [
			"-h, --help  Show this help.",
			"dev         Build with the development (lite) tag.",
		],
	})
) {
	Deno.exit(0);
}

if (!await compileTask(Deno.args.includes("dev"))) Deno.exit(1);
