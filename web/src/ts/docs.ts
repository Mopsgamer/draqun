import type { SlButton, SlTextarea } from "@shoelace-style/shoelace";
import { domLoaded } from "./lib.ts";

async function gqlSend(query: string, outp: HTMLDivElement) {
    try {
        const response = await fetch("/gql", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                query, // GraphQL expects a `query` property
            }),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        outp.textContent = JSON.stringify(result, null, "  ");
    } catch (error) {
        outp.textContent = error instanceof Error
            ? error.message
            : String(error);
    }
}

domLoaded.then(() => {
    const inp = document.getElementById("gql-input")! as SlTextarea;
    const btn = document.getElementById("gql-send")! as SlButton;
    const outp = document.getElementById("gql-output")! as HTMLDivElement;

    customElements.whenDefined("sl-textarea").then(() =>
        gqlSend(inp.value, outp)
    );

    btn.addEventListener("click", () => {
        gqlSend(inp.value, outp);
    });

    inp.addEventListener("keydown", (event) => {
        if (event.key !== "Enter" || !event.ctrlKey) {
            return;
        }
        gqlSend(inp.value, outp);
    });
});
