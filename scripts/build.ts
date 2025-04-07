import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { existsSync } from "@std/fs";
import { logBuild } from "./tool.ts";
import tailwindcssPlugin from "esbuild-plugin-tailwindcss";
import { dirname } from "@std/path/dirname";

const folder = "client";
const isWatch = Deno.args.includes("wait");

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

// deno-lint-ignore no-explicit-any
type Call<Args extends (...args: any[]) => Promise<void>> = [
    fn: (...args: Parameters<Args>) => Promise<void>,
    params: Parameters<Args>,
    group: string[],
];

const slAlias = ["shoelace", "shoe", "sl"];

const calls: (Call<typeof copy> | Call<typeof build>)[] = [
    [copy, [
        "./node_modules/@shoelace-style/shoelace/dist/assets/**/*",
        `./${folder}/static/shoelace/assets`,
    ], [...slAlias]],

    [copy, [
        `./${folder}/src/assets/**/*`,
        `./${folder}/static/assets`,
    ], ["assets"]],

    [build, [{
        ...options,
        outdir: `./${folder}/static/js`,
        entryPoints: [`./${folder}/src/ts/**/*`],
        plugins: [...denoPlugins()],
    }], ["js", ...slAlias]],

    [build, [{
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
    }], ["css", ...slAlias]],
];

const existingGroups = Array.from(new Set(calls.flatMap((c) => c[2])));
const extraGroups = ["min", "watch", "all", "help"];
const availableGroups = [...extraGroups, ...existingGroups];

if (Deno.args.includes("help")) {
    logBuild.info(
        "Available options: %s.",
        availableGroups.join(", "),
    );
    logBuild.info(
        "Usage example:\n\n\tdeno task compile:client js css min watch\n",
    );
    Deno.exit();
}

const unknownGroups = Deno.args.filter(
    (a) => !availableGroups.includes(a),
);
if (unknownGroups.length > 0) {
    logBuild.warn(
        `Unknown groups: ${unknownGroups.join(", ")}\n` +
            "Available groups: %s.",
        availableGroups.join(", "),
    );
}

logBuild.info(
    `Starting bundling ${folder}${isWatch ? " in watch mode" : ""}...`,
);

const existingGroupsUsed = !Deno.args.includes("all") &&
    existingGroups.some((g) => Deno.args.includes(g));

if (existingGroupsUsed) {
    calls.length = 0;
    calls.push(
        ...calls.filter(([, a, groups]) => {
            logBuild.start(a, groups);
            return groups.some((g) => Deno.args.includes(g));
        }),
    );
}

for (const [fn, args] of calls) {
    // deno-lint-ignore no-explicit-any
    await fn(...args as any);
}

logBuild.success("Bundled successfully");
if (isWatch) {
    logBuild.success("Watching for file changes...");
}
