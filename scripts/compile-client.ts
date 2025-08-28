import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { existsSync } from "@std/fs";
import { distFolder, logClientComp } from "./tool/constants.ts";
import tailwindcssPlugin from "esbuild-plugin-tailwindcss";
import { dirname } from "@std/path/dirname";
import { format, type TaskStateEnd } from "@m234/logger";
import { limit1 } from "./tool/limit1.ts";

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
            `File expected to exist: ${badEntryPoints.join(", ")}`,
        );
        return;
    }

    if (!outfile && !outdir) {
        logClientComp.error(`Provide outdir or outfile.`);
        return;
    }

    if (outfile && outdir) {
        logClientComp.error(`Can not use outdir and outfile at the same time.`);
        return;
    }

    const safeOptions = options;
    delete safeOptions.whenChange;
    const ctx = await esbuild.context(safeOptions as esbuild.BuildOptions);

    async function rebuild() {
        try {
            await ctx.rebuild();
        } catch (error) {
            logClientComp.error(format(error));
        }
    }

    await rebuild();

    if (!isWatch) {
        await ctx.dispose();
        return;
    }

    try {
        await ctx.watch();
    } catch (error) {
        logClientComp.error(format(error));
        return;
    }

    if (whenChange.length === 0) {
        logClientComp.error("Nothing to watch: " + whenChange.join(", ") + ".");
        await ctx.dispose();
        return;
    }

    let watcher: Deno.FsWatcher;
    try {
        watcher = Deno.watchFs(whenChange, { recursive: true });
    } catch (error) {
        logClientComp.error(format(error));
        logClientComp.error(
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
        whenChange: [to],
        plugins: [copyPlugin({
            once: isWatch,
            resolveFrom: "cwd",
            assets: { to: to, from: from + "/**/*" },
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
        "./node_modules/@shoelace-style/shoelace/dist/assets",
        `./${distFolder}/static/shoelace/assets`,
    ], [...slAlias]],

    [copy, [
        `./client/src/assets`,
        `./${distFolder}/static/assets`,
    ], ["assets"]],

    [build, [{
        ...options,
        outdir: `./${distFolder}/static/js`,
        entryPoints: [`./client/src/ts/**/*`],
        whenChange: [
            `./${distFolder}/static/js`,
        ],
        plugins: [...denoPlugins()],
    }], ["js", ...slAlias]],

    [build, [{
        ...options,
        outdir: `./${distFolder}/static/css`,
        entryPoints: [`./client/src/tailwindcss/**/*.css`],
        whenChange: [
            `./client/templates`,
            `./client/src/tailwindcss`,
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
    logClientComp.info(
        "Available options: " +
            availableGroups.join(", ") + ".",
    );
    logClientComp.info(
        "Usage example:\n\n\tdeno task compile:client js css min watch\n",
    );
    Deno.exit();
}

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

await Promise.allSettled(calls.map(([builder, builderArgs]) => {
    const { outdir, outfile } = builderArgs.length === 1
        ? builderArgs[0]
        : { outdir: dirname(builderArgs[1]) };
    const directory = outdir || dirname(outfile!);

    return logClientComp.task({ text: "Bundling " + directory }).startRunner(
        limit1(async (): Promise<TaskStateEnd> => {
            // deno-lint-ignore no-explicit-any
            await builder(...builderArgs as any);
            return "completed";
        }),
    );
}))

if (isWatch) {
    logClientComp.info("Watching for file changes...");
}
