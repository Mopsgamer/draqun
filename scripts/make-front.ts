import * as esbuild from "esbuild";
import { denoPlugin } from "@deno/esbuild-plugin";
import { existsSync } from "@std/fs";
import { cp } from "node:fs/promises";
import { distFolder, logClientComp } from "./tool/constants.ts";
import tailwindcssPlugin from "esbuild-plugin-tailwindcss";
import { format, type TaskRunnerReturn } from "@m234/logger";
import { showHelp } from "./tool/help.ts";

type BuildOptions = esbuild.BuildOptions & {
	whenChange?: string[];
};

async function notifyHMR(type: "css" | "reload"): Promise<void> {
	const port = Deno.env.get("PORT") || "3000";
	try {
		await fetch(`http://localhost:${port}/__hmr/notify?type=${type}`, {
			method: "POST",
		});
	} catch {
		// Ignore if server is offline or restarting
	}
}

async function build(
	options: BuildOptions,
	isWatch: boolean,
): Promise<TaskRunnerReturn> {
	const { outdir, outfile, entryPoints = [], whenChange = [] } = options;

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

	(async () => {
		let debounceTimer: ReturnType<typeof setTimeout> | undefined;
		let pendingPaths: string[] = [];

		for await (const { paths, kind } of watcher) {
			const isTargetEvent = kind === "modify" ||
				kind === "create" ||
				kind === "remove";

			if (!isTargetEvent) continue;

			const relevantPaths = paths.filter((p) => {
				if (outdir && outdir.endsWith("/js")) {
					return p.endsWith(".ts") || p.endsWith(".js");
				}
				if (outdir && outdir.endsWith("/css")) {
					return p.endsWith(".css") || p.endsWith(".tmpl");
				}
				return false;
			});

			if (relevantPaths.length === 0) continue;

			pendingPaths.push(...relevantPaths);

			if (debounceTimer !== undefined) {
				clearTimeout(debounceTimer);
			}

			debounceTimer = setTimeout(async () => {
				const currentPaths = [...new Set(pendingPaths)];
				pendingPaths = [];
				debounceTimer = undefined;

				let x = " ";
				if (currentPaths.length === 1) {
					const path = currentPaths[0]!.replaceAll("\\", "/");
					x = " '" + path.slice(path.lastIndexOf("/") + 1) + "'";
				} else x = " (" + currentPaths.length + " files)";

				try {
					const start = performance.now();
					await ctx.rebuild();
					const duration = (performance.now() - start).toFixed(0);
					logClientComp.info(
						`Frontend rebuilt in ${duration}ms at ` +
							new Date().toLocaleTimeString() + x,
					);

					const isCssOnly = currentPaths.every((p) => p.endsWith(".css"));
					await notifyHMR(isCssOnly ? "css" : "reload");
				} catch (error) {
					const message = "Frontend build failed at " +
						new Date().toLocaleTimeString() + x + ": " +
						(error as Error).message;
					logClientComp.error(message);
				}
			}, 150);
		}

		await ctx.dispose();
	})();
	return "completed";
}

export async function runMakeFront(
	args: string[] = Deno.args,
): Promise<void> {
	const isWatch = args.includes("watch");
	const isQuiet = args.includes("quiet");
	const minify = args.includes("min");

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

	const calls: [() => Promise<TaskRunnerReturn>, string, string[]][] = [
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
					entryPoints: [
						"client/src/ts/app.ts",
						"client/src/ts/docs.ts",
						"client/src/ts/homepage.ts",
						"client/src/ts/main.ts",
						"client/src/ts/hmr.ts",
					],
					whenChange: [
						"client/src/ts",
					],
					plugins: [denoPlugin()],
				}, isWatch),
			distFolder + "/static/js",
			["js"],
		],

		[
			() =>
				build({
					...options,
					outdir: distFolder + "/static/css",
					entryPoints: [
						"client/src/tailwindcss/main.css",
						"client/src/tailwindcss/homepage.css",
						"client/src/tailwindcss/docs.css",
					],
					whenChange: [
						"client/templates",
						"client/src/tailwindcss",
					],
					external: ["/static/assets/*"],
					plugins: [
						tailwindcssPlugin(),
					],
				}, isWatch),
			distFolder + "/static/css",
			["css"],
		],
	];

	const existingGroups = Array.from(new Set(calls.flatMap((c) => c[2])));
	const extraGroups = ["min", "watch", "quiet", "all"];
	const availableGroups = [...extraGroups, ...existingGroups];

	const unknownGroups = args.filter(
		(a) => !availableGroups.includes(a),
	);
	if (unknownGroups.length > 0) {
		logClientComp.warn(
			`Unknown groups: ${unknownGroups.join(", ")}\n` +
				"Available groups: " +
				availableGroups.join(", ") + ".",
		);
	}

	const existingGroupsUsed = !args.includes("all") &&
		existingGroups.some((g) => args.includes(g));

	if (existingGroupsUsed) {
		calls.splice(
			0,
			calls.length,
			...calls.filter(
				([, , groups]) => {
					return groups.some((g) => {
						const includes = args.includes(g);
						return includes;
					});
				},
			),
		);
	}

	const startTime = performance.now();
	await Promise.allSettled(calls.map(([builder, directory]) => {
		const text = `Bundling '${directory}'`;
		if (!isQuiet) return logClientComp.task({ text }).startRunner(builder);
		return builder()
			.then((result) => {
				if (result === "failed") logClientComp.error(`${text} failed.`);
			})
			.catch((error) => logClientComp.error(`${text} failed: ${format(error)}`));
	}));

	const totalMs = (performance.now() - startTime).toFixed(0);
	if (!isQuiet) {
		logClientComp.info(`Frontend build completed in ${totalMs}ms.`);
	}

	if (isWatch && !isQuiet) {
		logClientComp.info("Frontend watching for changes.");
	}

	if (!isWatch) {
		Deno.exit(0);
	}
}

if (import.meta.main) {
	const extraGroups = ["min", "watch", "quiet", "all", "assets", "js", "css"];
	if (
		showHelp({
			name: "front",
			description: "Build client assets.",
			usage: "deno task front [group ...] [options]",
			options: [
				"-h, --help  Show this help.",
				`Groups: ${extraGroups.join(", ")}.`,
				"min         Minify JavaScript and CSS.",
				"watch       Watch for client CSS changes and rebuild.",
				"quiet       Suppress progress output while building.",
				"all         Build all asset groups (the default).",
			],
		}, true)
	) {
		Deno.exit();
	}

	await runMakeFront();
}
