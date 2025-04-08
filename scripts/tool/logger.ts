import { sprintf } from "@std/fmt/printf";
import { blue, green, magenta, red, yellow } from "@std/fmt/colors";

export class Logger {
    private prefix: string;
    private started: boolean = false;
    private startedArgs: unknown[] = [];

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
        if (this.started) {
            this.end(true)
        }

        Deno.stdout.write(
            new TextEncoder().encode(message),
        );
    }

    info(...args: unknown[]) {
        this.print(`${blue("ⓘ")} ${blue(this.prefix)} ${this.format(...args)}\n`);
    }

    error(...args: unknown[]) {
        if (this.started) {
            this.end(false)
        }

        this.print(`${red("✖")} ${red(this.prefix)} ${this.format(...args)}\n`);
    }

    warn(...args: unknown[]) {
        this.print(`${yellow("⚠")} ${yellow(this.prefix)} ${this.format(...args)}\n`);
    }

    success(...args: unknown[]) {
        this.print(`${green("✔")} ${green(this.prefix)} ${this.format(...args)}\n`);
    }

    start(...args: unknown[]) {
        this.startedArgs = args;
        this.print(`${magenta("-")} ${magenta(this.prefix)} ${this.format(...args)}...`);
        this.started = true;
    }

    end(success: boolean) {
        this.started = false;
        const color = success ? green : red
        const message = success ? "done" : "fail"
        this.inline(`\r${color("-")} ${color(this.prefix)} ${this.format(...this.startedArgs)}...${color(message)}\n`);
    }

    inline(...args: unknown[]) {
        const message = this.format(...args);
        Deno.stdout.write(new TextEncoder().encode(message));
    }
}
