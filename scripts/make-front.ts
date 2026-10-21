import fs from "node:fs/promises";
import fsSync from "node:fs";
import path from "node:path";
import postcss from "postcss";
import tailwindcss from "@tailwindcss/postcss";
import { transform } from "lightningcss";
import { format, type TaskRunnerReturn } from "@jsr/m234__logger";
import { distFolder, logClientComp } from "./tool/constants.ts";
import { showHelp } from "./tool/help.ts";

async function notifyHMR(type: "css" | "reload"): Promise<void> {
	const port = process.env.PORT || "3000";
	try {
		await fetch(`http://localhost:${port}/__hmr/notify?type=${type}`, {
			method: "POST",
		});
	} catch {
		// Ignore if server is offline or restarting
	}
}

async function buildJs(
	isWatch: boolean,
	minify: boolean,
): Promise<TaskRunnerReturn> {
	const entryPoints = [
		"client/src/ts/app.ts",
		"client/src/ts/docs.ts",
		"client/src/ts/homepage.ts",
		"client/src/ts/main.ts",
		"client/src/ts/hmr.ts",
	];

	for (const ep of entryPoints) {
		if (!fsSync.existsSync(ep)) {
			logClientComp.error(`File expected to exist: ${ep}`);
			return "failed";
		}
	}

	const outdir = path.join(distFolder, "static/js");

	try {
		const res = await Bun.build({
			entrypoints: entryPoints,
			outdir,
			minify,
			target: "browser",
			format: "esm",
			naming: "[name].[ext]",
		});

		if (!res.success) {
			for (const log of res.logs) {
				logClientComp.error(format(log));
			}
			return "failed";
		}
	} catch (error) {
		logClientComp.error(format(error));
		return "failed";
	}

	if (isWatch) {
		watchDir("client/src/ts", async (paths) => {
			const relevant = paths.filter(
				(p) => p.endsWith(".ts") || p.endsWith(".js"),
			);
			if (relevant.length === 0) return;

			let fileDesc: string;
			if (relevant.length === 1) {
				const p = relevant[0]!.replaceAll("\\", "/");
				fileDesc = ` '${p.slice(p.lastIndexOf("/") + 1)}'`;
			} else {
				fileDesc = ` (${relevant.length} files)`;
			}

			try {
				const start = performance.now();
				const res = await Bun.build({
					entrypoints: entryPoints,
					outdir,
					minify,
					target: "browser",
					format: "esm",
					naming: "[name].[ext]",
				});

				if (!res.success) {
					logClientComp.error(
						`Frontend build failed at ${new Date().toLocaleTimeString()}${fileDesc}`,
					);
					return;
				}

				const duration = (performance.now() - start).toFixed(0);
				logClientComp.info(
					`Frontend rebuilt in ${duration}ms at ${new Date().toLocaleTimeString()}${fileDesc}`,
				);
				await notifyHMR("reload");
			} catch (err) {
				logClientComp.error(
					`Frontend build failed at ${new Date().toLocaleTimeString()}${fileDesc}: ${
						(err as Error).message
					}`,
				);
			}
		});
	}

	return "completed";
}

async function compileCssFile(
	entryPath: string,
	outPath: string,
	minify: boolean,
): Promise<void> {
	const css = await fs.readFile(entryPath, "utf8");
	const result = await postcss([tailwindcss()]).process(css, {
		from: entryPath,
		to: outPath,
	});
	await fs.mkdir(path.dirname(outPath), { recursive: true });

	let code = result.css;
	if (minify) {
		const transformed = transform({
			filename: entryPath,
			code: Buffer.from(code),
			minify: true,
		});
		code = transformed.code.toString();
	}

	await fs.writeFile(outPath, code, "utf8");
}

async function buildCss(
	isWatch: boolean,
	minify: boolean,
): Promise<TaskRunnerReturn> {
	const cssEntries = [
		{
			in: "client/src/tailwindcss/main.css",
			out: path.join(distFolder, "static/css/main.css"),
		},
		{
			in: "client/src/tailwindcss/homepage.css",
			out: path.join(distFolder, "static/css/homepage.css"),
		},
		{
			in: "client/src/tailwindcss/docs.css",
			out: path.join(distFolder, "static/css/docs.css"),
		},
	];

	for (const entry of cssEntries) {
		if (!fsSync.existsSync(entry.in)) {
			logClientComp.error(`File expected to exist: ${entry.in}`);
			return "failed";
		}
	}

	try {
		for (const entry of cssEntries) {
			await compileCssFile(entry.in, entry.out, minify);
		}
	} catch (error) {
		logClientComp.error((error as Error).message);
		return "failed";
	}

	if (isWatch) {
		const watchDirs = ["client/templates", "client/src/tailwindcss"];
		for (const watchDirFolder of watchDirs) {
			watchDir(watchDirFolder, async (paths) => {
				const relevant = paths.filter(
					(p) => p.endsWith(".css") || p.endsWith(".tmpl"),
				);
				if (relevant.length === 0) return;

				let fileDesc: string;
				if (relevant.length === 1) {
					const p = relevant[0]!.replaceAll("\\", "/");
					fileDesc = ` '${p.slice(p.lastIndexOf("/") + 1)}'`;
				} else {
					fileDesc = ` (${relevant.length} files)`;
				}

				try {
					const start = performance.now();
					for (const entry of cssEntries) {
						await compileCssFile(entry.in, entry.out, minify);
					}
					const duration = (performance.now() - start).toFixed(0);
					logClientComp.info(
						`Frontend rebuilt in ${duration}ms at ${new Date().toLocaleTimeString()}${fileDesc}`,
					);

					const isCssOnly = relevant.every((p) => p.endsWith(".css"));
					await notifyHMR(isCssOnly ? "css" : "reload");
				} catch (err) {
					logClientComp.error(
						`Frontend build failed at ${new Date().toLocaleTimeString()}${fileDesc}: ${
							(err as Error).message
						}`,
					);
				}
			});
		}
	}

	return "completed";
}

function watchDir(
	dirPath: string,
	callback: (changedPaths: string[]) => void,
): void {
	if (!fsSync.existsSync(dirPath)) return;

	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	let pendingPaths: string[] = [];

	try {
		fsSync.watch(dirPath, { recursive: true }, (_eventType, filename) => {
			if (!filename) return;
			const fullPath = path.join(dirPath, filename);
			pendingPaths.push(fullPath);

			if (debounceTimer !== undefined) {
				clearTimeout(debounceTimer);
			}

			debounceTimer = setTimeout(() => {
				const currentPaths = [...new Set(pendingPaths)];
				pendingPaths = [];
				debounceTimer = undefined;
				callback(currentPaths);
			}, 150);
		});
	} catch (err) {
		logClientComp.error(`Watcher error for ${dirPath}: ${err}`);
	}
}

export async function runMakeFront(
	args: string[] = process.argv.slice(2),
): Promise<void> {
	const isWatch = args.includes("watch");
	const isQuiet = args.includes("quiet");
	const minify = args.includes("min");

	const calls: [() => Promise<TaskRunnerReturn>, string, string[]][] = [
		[
			async () => {
				const src = "client/src/assets";
				const dest = path.join(distFolder, "static/assets");
				if (fsSync.existsSync(src)) {
					await fs.mkdir(dest, { recursive: true });
					await fs.cp(src, dest, { recursive: true });
				}
				return "completed";
			},
			path.join(distFolder, "static/assets"),
			["assets"],
		],

		[
			() => buildJs(isWatch, minify),
			path.join(distFolder, "static/js"),
			["js"],
		],

		[
			() => buildCss(isWatch, minify),
			path.join(distFolder, "static/css"),
			["css"],
		],
	];

	const existingGroups = Array.from(new Set(calls.flatMap((c) => c[2])));
	const extraGroups = ["min", "watch", "quiet", "all"];
	const availableGroups = [...extraGroups, ...existingGroups];

	const unknownGroups = args.filter((a) => !availableGroups.includes(a));
	if (unknownGroups.length > 0) {
		logClientComp.warn(
			`Unknown groups: ${unknownGroups.join(", ")}\n` +
				`Available groups: ${availableGroups.join(", ")}.`,
		);
	}

	const existingGroupsUsed =
		!args.includes("all") && existingGroups.some((g) => args.includes(g));

	if (existingGroupsUsed) {
		calls.splice(
			0,
			calls.length,
			...calls.filter(([, , groups]) => groups.some((g) => args.includes(g))),
		);
	}

	const startTime = performance.now();
	await Promise.allSettled(
		calls.map(([builder, directory]) => {
			const text = `Bundling '${directory}'`;
			if (!isQuiet) return logClientComp.task({ text }).startRunner(builder);
			return builder()
				.then((result) => {
					if (result === "failed") logClientComp.error(`${text} failed.`);
				})
				.catch((error) =>
					logClientComp.error(`${text} failed: ${format(error)}`),
				);
		}),
	);

	const totalMs = (performance.now() - startTime).toFixed(0);
	if (!isQuiet) {
		logClientComp.info(`Frontend build completed in ${totalMs}ms.`);
	}

	if (isWatch && !isQuiet) {
		logClientComp.info("Frontend watching for changes.");
	}

	if (!isWatch) {
		process.exit(0);
	}
}

if (import.meta.main) {
	const extraGroups = ["min", "watch", "quiet", "all", "assets", "js", "css"];
	if (
		showHelp(
			{
				name: "front",
				description: "Build client assets.",
				usage: "bun run front [group ...] [options]",
				options: [
					"-h, --help  Show this help.",
					`Groups: ${extraGroups.join(", ")}.`,
					"min         Minify JavaScript and CSS.",
					"watch       Watch for client CSS changes and rebuild.",
					"quiet       Suppress progress output while building.",
					"all         Build all asset groups (the default).",
				],
			},
			true,
		)
	) {
		process.exit(0);
	}

	await runMakeFront();
}
