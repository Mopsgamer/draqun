import * as esbuild from "esbuild";
import { copy as copyPlugin } from "esbuild-plugin-copy";
import { tailwindPlugin } from "esbuild-plugin-tailwindcss";
import { Listr, type ListrTask } from "listr2";
import { denoPlugins } from "@luca/esbuild-deno-loader";
import { dirname } from "jsr:@std/path";
import { exists } from "jsr:@std/fs";

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

function buildTask(options: BuildOptions, title?: string): ListrTask {
    const { outdir, outfile } = options;
    title ??= outdir ?? outfile;
    return ({
        title: title,
        task: async () => {
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
                console.log("removing " + directory);
                await Deno.remove(directory, { recursive: true });
            }
            await esbuild.build(options);
        },
    } as ListrTask);
}

function copyTask(from: string, to: string): ListrTask {
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

const progress = new Listr(
    {
        title: "Generating ./web/static folder: js, css and assets",
        task: (_, task) =>
            task.newListr(
                [
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
                            tailwindPlugin()
                        ]
                    }),
                    copyTask(
                        "../node_modules/@shoelace-style/shoelace/dist/assets/**/*",
                        "./assets",
                    ),
                    copyTask(
                        "../src/assets/**/*",
                        "./static/assets",
                    ),
                ],
                {
                    collectErrors: "full",
                    concurrent: false,
                },
            ),
    },
);

await progress.run();
