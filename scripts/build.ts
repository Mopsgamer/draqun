import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { existsSync } from "@std/fs";
import { environment, envKeys, logBuild } from "./tool.ts";
import dotenv from "dotenv";
import tailwindcssPlugin from "esbuild-plugin-tailwindcss";
import { dirname } from "@std/path/dirname";

const folder = "client";
dotenv.config();
const isWatch = Deno.args.includes("--watch");

type BuildOptions = esbuild.BuildOptions & {
    whenChange?: string[];
};

const minify = environment > 1;
logBuild.info(`${envKeys.ENVIRONMENT} = ${environment}`);
logBuild.info(
    `Starting bundling ${folder}${isWatch ? " in watch mode" : ""}...`,
);

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

let buildCalls = 0;
async function build(
    options: BuildOptions,
): Promise<void> {
    const { outdir, outfile, entryPoints = [], whenChange = [] } = options;
    buildCalls++;

    const directory = outdir || dirname(outfile!);
    logBuild.info(`${directory} ${buildCalls}/${calls.length}`);

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

    if (badEntryPoints.length > 0) {
        logBuild.fatal(`File expected to exist: ${badEntryPoints.join(", ")}`);
        return;
    }

    if (!outfile && !outdir) {
        logBuild.fatal(`Provide outdir or outfile.`);
        return;
    }

    if (outfile && outdir) {
        logBuild.fatal(`Can not use outdir and outfile at the same time.`);
        return;
    }

    const safeOptions = options;
    delete safeOptions.whenChange;
    const ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);

    async function rebuild() {
        try {
            const result = await ctx.rebuild();
            for (const warn of result.warnings) {
                logBuild.warn(warn);
            }
            for (const error of result.errors) {
                logBuild.error(error);
            }
        } catch (error) {
            logBuild.fatal(error);
        }
    }

    await rebuild();
    if (!isWatch) {
        await ctx!.dispose();
        return;
    }

    try {
        await ctx.watch();
    } catch (error) {
        logBuild.fatal(error);
        return;
    }

    if (whenChange.length === 0) return;

    let watcher: Deno.FsWatcher;
    try {
        watcher = Deno.watchFs(whenChange, { recursive: true });
    } catch (error) {
        logBuild.error(error);
        logBuild.fatal(
            "Bad paths, can not add watcher: " + whenChange.join(", ") + ".",
        );
        return;
    }

    // this callback won't block the process.
    // buildTask will return while ignoring loop
    (async () => {
        for await (const event of watcher) {
            if (
                event.kind === "modify" || event.kind === "create" ||
                event.kind === "remove"
            ) return;

            await rebuild();
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

const calls: unknown[][] = [
    [
        copy,
        "./node_modules/@shoelace-style/shoelace/dist/assets/**/*",
        `./${folder}/static/shoelace/assets`,
    ],

    [copy, `./${folder}/src/assets/**/*`, `./${folder}/static/assets`],

    [build, {
        ...options,
        outdir: `./${folder}/static/js`,
        entryPoints: [`./${folder}/src/ts/**/*`],
        plugins: [...denoPlugins()],
    }],

    [build, {
        ...options,
        outdir: `./${folder}/static/css`,
        entryPoints: [`./${folder}/src/tailwindcss/**/*.css`],
        whenChange: [
            `./${folder}/templates`,
            `./${folder}/src/tailwindcss`,
        ],
        external: ["/static/assets/*"],
        plugins: [
            tailwindcssPlugin(),
        ],
    }],
];

for (const [fn, ...args] of calls) {
    // deno-lint-ignore ban-types
    await (fn as Function)(...args);
}

logBuild.success("Bundled successfully");
if (isWatch) {
    logBuild.success("Watching for file changes...");
}
