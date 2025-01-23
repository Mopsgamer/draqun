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
logBuild.info(`Starting bundling web${isWatch ? " in watch mode" : ""}...`);

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

async function build(
    options: BuildOptions,
): Promise<void> {
    const { outdir, outfile, entryPoints = [], whenChange = [] } = options;

    const directory = outdir || dirname(outfile!);
    logBuild.info(directory);

    const entryPointsNormalized = Array.isArray(entryPoints)
        ? entryPoints
        : Object.keys(entryPoints);

    const badEntryPoints = entryPointsNormalized.filter((entry) => {
        const pth = typeof entry === "string" ? entry : entry.in;
        try {
            return !existsSync(pth);
        } catch {
            return false;
        }
    });

    if (
        badEntryPoints.length > 0
    ) throw new Error(`File expected to exist: ${badEntryPoints.join(", ")}`);

    if (
        !outfile && !outdir
    ) throw new Error(`Provide outdir or outfile.`);

    if (
        outfile && outdir
    ) throw new Error(`Can not use outdir and outfile at the same time.`);

    if (await exists(directory)) {
        await Deno.remove(directory, { recursive: true });
    }

    const safeOptions = options;
    delete safeOptions.whenChange;
    const ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);
    await ctx.rebuild();
    if (!isWatch) {
        await ctx.dispose();
        logBuild.success(directory);
        return;
    }

    await ctx.watch();
    logBuild.success(directory);
    if (whenChange.length === 0) return;

    const watcher = Deno.watchFs(whenChange, { recursive: true });

    // this callback won't block the process.
    // buildTask will return while ignoring loop
    (async () => {
        for await (const event of watcher) {
            if (
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            ) return;

            try {
                await ctx.rebuild();
            } catch { /* empty */ }
        }
        await ctx.dispose();
    })();
}

function copy(from: string, to: string): Promise<void> {
    return build({
        ...options,
        outdir: to,
        entryPoints: [],
        plugins: [copyPlugin({
            once: isWatch,
            resolveFrom: "cwd",
            assets: { to, from },
            copyOnStart: true,
        })],
    });
}

await build({
    ...options,
    outdir: "./web/static/js",
    entryPoints: ["./web/src/ts/**/*"],
    plugins: [...denoPlugins()],
});

await build({
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
});

await copy(
    "./node_modules/@shoelace-style/shoelace/dist/**/*",
    "./web/static/shoelace",
);

await copy(
    "./web/src/assets/**/*",
    "./web/static/assets",
);

logBuild.success("Bundled successfully");
if (isWatch) {
    logBuild.success("Watching for file changes...");
}
