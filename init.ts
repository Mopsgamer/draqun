import { parse } from "dotenv"

function initEnv(): void {
    type EnvKeyEntry = {
        value?: string,
        comment?: string
    }
    const defaultEnv = new Map<string, EnvKeyEntry>()
    defaultEnv.set("ENVIRONMENT", { value: "1", comment: "can be 0 (test), 1 (dev) or 2 (prod)\ndefault: 1" })
    defaultEnv.set("JWT_KEY", { comment: "use any online jwt generator to fill this value" })
    defaultEnv.set("DB_PASSWORD", { comment: "connection password" })
    defaultEnv.set("DB_NAME", { value: "restapp", comment: "connection name\ndefault: restapp" })
    defaultEnv.set("DB_USER", { value: "root", comment: "connection user\nroot user is not recommended\ndefault: root" })
    defaultEnv.set("DB_HOST", { value: "localhost", comment: "connection host\ndefault: localhost" })
    defaultEnv.set("DB_PORT", { value: "3306", comment: "connection port\ndefault: 3306" })

    const path = ".env";
    const decoder = new TextDecoder();
    const env = parse(decoder.decode(Deno.readFileSync(path)));

    const encoder = new TextEncoder();
    Deno.writeFileSync(
        path,
        encoder.encode(
            Array.from(defaultEnv.entries()).map(
                ([key, { value, comment }]) =>
                    `${comment ? "# " + comment.replace('\n', '\n# ') + "\n" : ""}${key}=${env[key] || value || ""}\n\n`
            ).join("")
        ),
    );

    console.log("Writed " + path);
}

initEnv();
