import { parse } from "dotenv"

function initEnv(): void {
    type EnvKeyEntry = {
        value?: string,
        comment?: string
    }
    const defaultEnv = new Map<string, EnvKeyEntry>()
    defaultEnv.set("ENVIRONMENT", { value: "1", comment: "can be test, dev or prod" })
    defaultEnv.set("JWT_KEY", { comment: "use any online jwt generator" })
    defaultEnv.set("DB_PASSWORD", {})
    defaultEnv.set("DB_NAME", { value: "restapp" })
    defaultEnv.set("DB_USER", { value: "root" })
    defaultEnv.set("DB_HOST", { value: "localhost" })
    defaultEnv.set("DB_PORT", { value: "3306" })

    const path = ".env";
    const decoder = new TextDecoder();
    const env = parse(decoder.decode(Deno.readFileSync(path)));

    const encoder = new TextEncoder();
    Deno.writeFileSync(
        path,
        encoder.encode(
            Array.from(defaultEnv.entries()).map(
                ([key, { value, comment }]) =>
                    `${comment ? "# " + comment + "\n" : ""}${key}=${env[key] || value || ""}\n`
            ).join("")
        ),
    );

    console.log("Writed " + path);
}

initEnv();
