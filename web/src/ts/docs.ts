import type { SlButton, SlTextarea } from "@shoelace-style/shoelace";
import { domLoaded } from "./lib.ts";

import "./main.ts";

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
    const editor = document.getElementById("gql-input")! as SlTextarea;
    const executeButton = document.getElementById("gql-send")! as SlButton;
    const resultView = document.getElementById("gql-output")! as HTMLDivElement;
    const example = (document.getElementById("gql-example")! as HTMLTemplateElement).innerHTML!;

    customElements.whenDefined("sl-textarea").then(() => {
        console.log("ex: %o", example)
        editor.value = example;
        gqlSend(editor.value, resultView)
    });

    executeButton.addEventListener("click", () => {
        gqlSend(editor.value, resultView);
    });

    editor.addEventListener("keydown", (event) => {
        if (event.key !== "Enter" || !event.ctrlKey) {
            return;
        }
        gqlSend(editor.value, resultView);
    });
});
