import { sprintf } from "@std/fmt/printf";
import { blue, green, red, yellow } from "@std/fmt/colors";

export class Logger {
    private prefix: string;

    constructor(prefix: string = "") {
        this.prefix = prefix ? `[${prefix}]` : "";
    }

    private format(...args: unknown[]): string {
        const [message, ...other] = args;
        if (typeof message == "string") {
            return sprintf(message, ...other);
        }

        return args.map((a) => sprintf("%o", a)).join(" ");
    }

    private print(message: string) {
        Deno.stdout.write(new TextEncoder().encode(message));
    }

    info(...args: unknown[]) {
        this.print(`${blue("ⓘ")} ${this.prefix} ${this.format(...args)}\n`);
    }

    error(...args: unknown[]) {
        this.print(`${red("✖")} ${this.prefix} ${this.format(...args)}\n`);
    }

    warn(...args: unknown[]) {
        this.print(`${yellow("⚠")} ${this.prefix} ${this.format(...args)}\n`);
    }

    success(...args: unknown[]) {
        this.print(`${green("✔")} ${this.prefix} ${this.format(...args)}\n`);
    }

    inline(...args: unknown[]) {
        this.print(this.format(...args));
    }
}
