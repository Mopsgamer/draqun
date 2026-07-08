import { compileTask } from "./tool/compile-binary.ts";
import { logServerComp, taskDotenv } from "./tool/constants.ts";

taskDotenv(logServerComp);

if (!await compileTask()) Deno.exit(1);
