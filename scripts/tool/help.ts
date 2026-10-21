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
	const args = process.argv.slice(2);
	if (
		!args.includes("-h") &&
		!args.includes("--help") &&
		!(allowBareHelp && args.includes("help"))
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
Run "bun run ${name} --help" to show this help.`);
	return true;
}
