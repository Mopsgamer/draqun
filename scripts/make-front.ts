import * as esbuild from "esbuild";
import { denoPlugin } from "@deno/esbuild-plugin";
import { existsSync } from "@std/fs";
import { cp } from "node:fs/promises";
import { distFolder, logClientComp, taskDotenv } from "./tool/constants.ts";
import tailwindcssPlugin from "esbuild-plugin-tailwindcss";
import { format, type TaskRunnerReturn } from "@m234/logger";

const isWatch = Deno.args.includes("watch");

type BuildOptions = esbuild.BuildOptions & {
	whenChange?: string[];
};

const minify = Deno.args.includes("min");

const options: esbuild.BuildOptions = {
	bundle: true,
	minify: minify,
	minifyIdentifiers: minify,
	minifySyntax: minify,
	minifyWhitespace: minify,
	platform: "browser",
	format: "esm",
	target: "esnext",
};

let buildCalls = 0;
async function build(
	options: BuildOptions,
): Promise<TaskRunnerReturn> {
	const { outdir, outfile, entryPoints = [], whenChange = [] } = options;
	buildCalls++;

	const entryPointsNormalized = Array.isArray(entryPoints)
		? entryPoints
		: Object.keys(entryPoints);

	const badEntryPoints = entryPointsNormalized.filter((entry) => {
		const pth = typeof entry === "string" ? entry : entry.in;

		if (pth.includes("*")) {
			return false;
		}

		try {
			return !existsSync(pth);
		} catch {
			return true;
		}
	});

	if (badEntryPoints.length > 0) {
		logClientComp.error(
			format("File expected to exist: %o", badEntryPoints),
		);
		return "failed";
	}

	if (!outfile && !outdir) {
		logClientComp.error("Provide outdir or outfile.");
		return "failed";
	}

	if (outfile && outdir) {
		logClientComp.error("Can not use outdir and outfile at the same time.");
		return "failed";
	}

	const safeOptions = options;
	delete safeOptions.whenChange;
	let ctx: esbuild.BuildContext<esbuild.BuildOptions>;
	try {
		ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);
	} catch (error) {
		logClientComp.error(format(error));
		return "failed";
	}

	try {
		await ctx.rebuild();
	} catch (error) {
		logClientComp.error((error as Error).message);
		return "failed";
	}

	if (!isWatch) {
		await ctx.dispose();
		return "completed";
	}

	try {
		await ctx.watch();
	} catch (error) {
		logClientComp.error(format(error));
		return "failed";
	}

	if (whenChange.length === 0) {
		logClientComp.error("Nothing to watch: " + whenChange.join(", ") + ".");
		await ctx.dispose();
		return "failed";
	}

	let watcher: Deno.FsWatcher;
	try {
		watcher = Deno.watchFs(whenChange, { recursive: true });
	} catch (error) {
		logClientComp.error(format(error));
		logClientComp.error(
			"Bad paths, can not add watcher: " + whenChange.join(", ") + ".",
		);
		await ctx.dispose();
		return "failed";
	}

	// this callback won't block the process.
	// buildTask will return while ignoring loop
	(async () => {
		const task = logClientComp.task({ text: "Watching for file changes" })
			.start();

		for await (const { paths, kind } of watcher) {
			const isTargetEvent = kind === "modify" ||
				kind === "create" ||
				kind === "remove";

			if (!isTargetEvent) continue;

			const hasRelevantChanges = paths.some((p) => p.endsWith(".css"));

			if (!hasRelevantChanges) continue;

			let x = " ";
			if (paths.length === 1) {
				const path = paths[0]!.replaceAll("\\", "/");
				x = " '" + path.slice(path.lastIndexOf("/") + 1) + "'";
			} else x = " (" + paths.length + " files)";
			try {
				task.text = "Building at " + new Date().toLocaleTimeString() + x;
				await ctx.rebuild();
				task.text = "Updated at " + new Date().toLocaleTimeString() + x;
			} catch (error) {
				task.text = "Failed at " + new Date().toLocaleTimeString() + x + ": " +
					(error as Error).message;
			}
		}

		await ctx.dispose();
	})();
	return "completed";
}

const slAlias = ["shoelace", "shoe", "sl"];
const slAssets = slAlias.map((a) => (a + "-assets"));

const calls: [() => Promise<TaskRunnerReturn>, string, string[]][] = [
	[
		() =>
			cp(
				"node_modules/@shoelace-style/shoelace/dist/assets",
				distFolder + "/static/shoelace/assets",
				{ recursive: true },
			),
		distFolder + "/static/shoelace/assets",
		[...slAssets, ...slAlias],
	],

	[
		() =>
			cp(
				"client/src/assets",
				distFolder + "/static/assets",
				{ recursive: true },
			),
		distFolder + "/static/assets",
		["assets"],
	],

	[
		() =>
			build({
				...options,
				outdir: distFolder + "/static/js",
				entryPoints: ["client/src/ts/**/*"],
				whenChange: [
					distFolder + "/static/js",
				],
				plugins: [denoPlugin()],
			}),
		distFolder + "/static/js",
		["js", ...slAlias],
	],

	[
		() =>
			build({
				...options,
				outdir: distFolder + "/static/css",
				entryPoints: ["client/src/tailwindcss/**/*.css"],
				whenChange: [
					"client/templates",
					"client/src/tailwindcss",
				],
				external: ["/static/assets/*"],
				plugins: [
					tailwindcssPlugin(),
				],
			}),
		distFolder + "/static/css",
		["css", ...slAlias],
	],
];

const existingGroups = Array.from(new Set(calls.flatMap((c) => c[2])));
const extraGroups = ["min", "watch", "all", "help"];
const availableGroups = [...extraGroups, ...existingGroups];

if (
	Deno.args.includes("help") || Deno.args.includes("--help") ||
	Deno.args.includes("-h")
) {
	logClientComp.info(
		"Available options: " +
			availableGroups.join(", ") + ".",
	);
	logClientComp.info(
		"Usage example:\n\n\tdeno task front js css min watch\n",
	);
	Deno.exit();
}

taskDotenv(logClientComp);

const unknownGroups = Deno.args.filter(
	(a) => !availableGroups.includes(a),
);
if (unknownGroups.length > 0) {
	logClientComp.warn(
		`Unknown groups: ${unknownGroups.join(", ")}\n` +
			"Available groups: " +
			availableGroups.join(", ") + ".",
	);
}

const existingGroupsUsed = !Deno.args.includes("all") &&
	existingGroups.some((g) => Deno.args.includes(g));

if (existingGroupsUsed) {
	calls.splice(
		0,
		calls.length,
		...calls.filter(
			([, , groups]) => {
				return groups.some((g) => {
					const includes = Deno.args.includes(g);
					return includes;
				});
			},
		),
	);
}

await Promise.allSettled(calls.map(([builder, directory]) => {
	const text = `Bundling '${directory}'`;
	return logClientComp.task({ text }).startRunner(builder);
}));
