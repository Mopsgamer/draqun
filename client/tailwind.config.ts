import type { Config } from "tailwindcss";
const config: Config = {
    content: {
        relative: true,
        files: [
            "./templates/**/*",
            "./src/**/*",
        ],
    },
};
export default config;
