import { compileTask } from "./tool/compile-binary.ts";

if (!await compileTask(Deno.args.includes("dev"))) Deno.exit(1);
