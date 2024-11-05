import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { tailwindPlugin } from "esbuild-plugin-tailwindcss";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { dirname } from "jsr:@std/path";
import { exists } from "jsr:@std/fs";

const isWatch = Deno.args.includes("--watch");

type BuildOptions = esbuild.SameShape<
    esbuild.BuildOptions,
    esbuild.BuildOptions
>;

const options: BuildOptions = {
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
    const { outdir, outfile } = options;
    title ??= outdir ?? outfile;

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
        console.log("Cleaning " + directory);
        await Deno.remove(directory, { recursive: true });
    }
    const ctx = await esbuild.context(options);
    await ctx.rebuild();
    if (isWatch) {
        await ctx.watch();
        console.log("Listening for changes: " + directory);
        return;
    }
    await ctx.dispose();
    console.log("Bundled: " + directory);
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
        outdir: "./static/js",
        entryPoints: ["./src/js/main.js"],
        plugins: [...denoPlugins()],
    }),
    buildTask({
        ...options,
        outfile: "./static/css/main.css",
        entryPoints: ["./src/css/main.css"],
        plugins: [
            tailwindPlugin(),
        ],
    }),
    copyTask(
        "../node_modules/@shoelace-style/shoelace/dist/assets/**/*",
        "./assets",
    ),
    copyTask(
        "../src/assets/**/*",
        "./static/assets",
    ),
];

await Promise.all(taskList);

if (isWatch) {
    console.log("Done: Watching for file changes...");
} else {
    console.log("Done: Bundled all files.");
}
