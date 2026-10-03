type HelpOptions = {
	name: string;
	description: string;
	usage: string;
	options?: string[];
};

export function showHelp(
	{ name, description, usage, options = [] }: HelpOptions,
	allowBareHelp = false,
): boolean {
	if (
		!Deno.args.includes("-h") &&
		!Deno.args.includes("--help") &&
		!(allowBareHelp && Deno.args.includes("help"))
	) {
		return false;
	}

	console.log(`${description}

Usage:
  ${usage}`);

	if (options.length > 0) {
		console.log(`
Options:
${options.map((option) => `  ${option}`).join("\n")}`);
	}

	console.log(`
Run "deno task ${name} --help" to show this help.`);
	return true;
}
