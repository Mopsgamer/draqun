import * as HTMX from "htmx.org";
import { getFormControls, SlButton } from "@shoelace-style/shoelace";

declare global {
    namespace globalThis {
        // deno-lint-ignore no-var
        var htmx: typeof HTMX;
    }
}
globalThis.htmx = HTMX;

HTMX.on("htmx:wsConfigSend", (
    event: Event | CustomEvent,
) => {
    const form = event.target;

    if (!(form instanceof HTMLFormElement) || !(event instanceof CustomEvent)) {
        return;
    }

    const { detail } = event as CustomEvent<
        { elt: HTMLElement; parameters: Record<string, string> }
    >;

    Object.assign(detail.parameters, getFormPropData(form, true));
});

let intervalHandle: number | undefined = undefined;
HTMX.on("htmx:wsOpen", (ev: Event | CustomEvent) => {
    intervalHandle = setInterval(() => {
        if (!(ev instanceof CustomEvent)) {
            return;
        }
        ev.detail.socketWrapper.send(JSON.stringify({ type: "ping" }));
    }, 2000);
});
HTMX.on("htmx:wsClose", () => {
    if (intervalHandle) {
        clearInterval(intervalHandle);
        intervalHandle = undefined;
    }
});

// TODO: htmx websocket does not work
// fixes websocket, idk why, it does not replaces the dom
// we are sending only the string content as examples and docs saying
HTMX.on("htmx:wsAfterMessage", (
    event: Event | CustomEvent,
) => {
    if (!(event instanceof CustomEvent)) {
        return;
    }

    const { detail } = event as CustomEvent<
        { elt: HTMLElement; message: string }
    >;
    const { elt, message } = detail;

    if (elt.innerHTML === message) {
        return
    }

    elt.innerHTML = message;
});

HTMX.defineExtension("shoelace", {
    onEvent(
        name: string,
        event: Event | CustomEvent,
    ) {
        if (name === "htmx:beforeSend" || name === "htmx:afterRequest") {
            const form = event.target;
            let button: SlButton | undefined;
            if (form instanceof SlButton) {
                button = form;
            } else if (form instanceof HTMLFormElement) {
                button =
                    form.querySelector<SlButton>("sl-button[type=submit]") ??
                        undefined;
            }

            if (!button) {
                return true;
            }
            button.loading = name === "htmx:beforeSend";
            return true;
        }
        if (name !== "htmx:configRequest") {
            return true;
        }

        if (!(event instanceof CustomEvent)) {
            console.groupEnd();
            return true;
        }
        const { detail } = event as CustomEvent<
            { elt: HTMLElement; parameters: Record<string, string> }
        >;
        const form = detail.elt;
        if (!(form instanceof HTMLFormElement)) {
            console.groupEnd();
            return true;
        }

        Object.assign(detail.parameters, getFormPropData(form));

        // Prevent form submission if one or more fields are invalid.
        // form is always a form as per the main if statement
        if (!form.checkValidity()) {
            console.error("Form is invalid: %o", form);
            console.groupEnd();
            return false;
        }
        console.groupEnd();
        return true;
    },
});

function getFormPropData(
    form: HTMLFormElement,
    capital = false,
): Record<string, string> {
    const data: Record<string, string> = {};
    for (const slElement of getFormControls(form) as HTMLFormElement[]) {
        let { name } = slElement;
        const { value } = slElement;

        if (!name) {
            continue;
        }

        if (capital) {
            name = name[0].toUpperCase() + name.substring(1);
        }
        data[name] = value;
    }

    return data;
}
