import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { tailwindPlugin } from "esbuild-plugin-tailwindcss";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { dirname } from "@std/path";
import { exists, existsSync } from "@std/fs";
import { envKeys, logBuild } from "./tool.ts";
import dotenv from "dotenv";

dotenv.config();
const isWatch = Deno.args.includes("--watch");

type BuildOptions = esbuild.BuildOptions & {
    whenChange?: string | string[];
};

const environment = Number(Deno.env.get(envKeys.ENVIRONMENT));
const minify = environment > 1;
logBuild.info(`${envKeys.ENVIRONMENT} = ${environment}`);

const options: esbuild.BuildOptions = {
    bundle: true,
    minify: minify,
    minifyIdentifiers: minify,
    minifySyntax: minify,
    minifyWhitespace: minify,
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
        (entry) => {
            const pth = typeof entry === "string" ? entry : entry.in;
            try {
                return !existsSync(pth);
            } catch {
                return false;
            }
        },
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
        logBuild.start("Cleaning: " + directory);
        await Deno.remove(directory, { recursive: true });
    }
    const safeOptions = options;
    delete safeOptions.whenChange;
    const ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);
    logBuild.info("Bundling: " + directory);
    await ctx.rebuild();
    logBuild.success("Bundled: " + directory);
    if (isWatch) {
        await ctx.watch();
        // logBuild.success("Watching for changes: " + directory);
        if (!(whenChange.length > 0)) {
            return;
        }
        const watcher = Deno.watchFs(whenChange, { recursive: true });
        (async () => {
            for await (const _ of watcher) {
                await ctx.rebuild();
                logBuild.success("Bundled: " + directory);
            }
        })();
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
        entryPoints: ["./web/src/ts/**/*"],
        plugins: [...denoPlugins()],
    }),
    buildTask({
        ...options,
        outdir: "./web/static/css",
        entryPoints: ["./web/src/tailwindcss/**/*"],
        whenChange: [
            "./web/templates",
            "./web/src/tailwindcss",
            // "./tailwind.config.ts", // should reload process, anyway won't work
        ],
        external: ["/static/assets/*"],
        plugins: [
            tailwindPlugin({ configPath: "./tailwind.config.ts" }),
        ],
    }),
    copyTask(
        "./node_modules/@shoelace-style/shoelace/dist/**/*",
        "./web/static/shoelace",
    ),
    copyTask(
        "./web/src/assets/**/*",
        "./web/static/assets",
    ),
];

for (const task of taskList) {
    await task;
}

logBuild.success("Done: Bundled all files.");
if (isWatch) {
    logBuild.success("Watching for file changes...");
}
