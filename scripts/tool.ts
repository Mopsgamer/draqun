import consola from "consola";

consola.options.formatOptions.columns = 0;

export const logBuild = consola.withTag("esbuild");
export const logInitDb = consola.withTag("init-database");
export const logInitFiles = consola.withTag("init-files");
