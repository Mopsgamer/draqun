import { existsSync } from "@std/fs";

const encoder = new TextEncoder();

function initEnv(): void {
    const path = ".env";
    const optionForce = "--force";
    const isForce = Deno.args.includes(optionForce);
    const exists = existsSync(path);

    if (exists && !isForce) {
        console.log(
            "Failed to write " + path + " - already exists, use " + optionForce,
        );
        Deno.exit(1);
    }

    Deno.writeFileSync(
        path,
        encoder.encode(
            "DB_PASSWORD=\n" +
                "JWT_KEY=\n" +
                "DB_NAME=restapp\n" +
                "DB_USER=root\n" +
                "DB_HOST=localhost\n" +
                "DB_PORT=3306\n",
        ),
    );

    console.log("Writed " + path);
}

initEnv();
