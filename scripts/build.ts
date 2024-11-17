import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { tailwindPlugin } from "esbuild-plugin-tailwindcss";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { dirname } from "@std/path";
import { exists, existsSync } from "@std/fs";
import { logBuild } from "./tool.ts";

const isWatch = Deno.args.includes("--watch");

type BuildOptions = esbuild.BuildOptions & {
    whenChange?: string | string[];
};

const options: esbuild.BuildOptions = {
    bundle: true,
    minify: false,
    platform: "browser",
    format: "esm",
    target: [
        "esnext",
        "chrome67",
        "edge79",
        "firefox68",
        "safari14",
    ],
};

async function buildTask(options: BuildOptions, title?: string): Promise<void> {
    const { outdir, outfile, entryPoints = [], whenChange = [] } = options;
    title ??= outdir ?? outfile;
    const badEntryPoints = (
        Array.isArray(entryPoints) ? entryPoints : Object.keys(entryPoints)
    ).filter(
        (entry) => !existsSync(typeof entry === "string" ? entry : entry.in),
    );
    if (badEntryPoints.length > 0) {
        throw new Error(`File expected to exist: ${badEntryPoints.join(", ")}`);
    }

    if (!outfile && !outdir) {
        throw new Error(
            `Provide outdir (${outdir}) or outfile (${outfile}).`,
        );
    }

    if (outfile && outdir) {
        throw new Error(
            `Expected or outdir (${outdir}) or outfile (${outfile}), not both.`,
        );
    }

    const directory = outdir || dirname(outfile!);
    if (await exists(directory)) {
        logBuild.info("Cleaning: " + directory);
        await Deno.remove(directory, { recursive: true });
    }
    const safeOptions = options;
    delete safeOptions.whenChange;
    const ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);
    await ctx.rebuild();
    logBuild.info("Bundled: " + directory);
    if (isWatch) {
        await ctx.watch();
        logBuild.info("Listening: " + directory);
        if (whenChange.length > 0) {
            const watcher = Deno.watchFs(whenChange, { recursive: true });
            (async () => {
                for await (const event of watcher) {
                    if (event.kind !== "modify") {
                        continue;
                    }
                    await ctx.rebuild();
                    logBuild.info("Bundled: " + directory);
                }
            })();
        }
        return;
    }
    await ctx.dispose();
}

function copyTask(from: string, to: string) {
    return buildTask({
        ...options,
        outdir: to,
        entryPoints: [],
        plugins: [copyPlugin({
            resolveFrom: "cwd",
            assets: { to, from },
        })],
    });
}

const taskList = [
    buildTask({
        ...options,
        outdir: "./web/static/js",
        entryPoints: ["./web/src/ts/main.ts"],
        plugins: [...denoPlugins()],
    }),
    buildTask({
        ...options,
        outfile: "./web/static/css/main.css",
        entryPoints: ["./web/src/tailwindcss/main.css"],
        whenChange: "./web/templates",
        plugins: [
            tailwindPlugin({ configPath: "./tailwind.config.ts" }),
        ],
    }),
    copyTask(
        "./node_modules/@shoelace-style/shoelace/dist/assets/**/*",
        "./web/assets",
    ),
    copyTask(
        "./web/src/assets/**/*",
        "./web/static/assets",
    ),
];

await Promise.all(taskList);

logBuild.success("Done: Bundled all files.");
if (isWatch) {
    logBuild.success("Watching for file changes...");
}
