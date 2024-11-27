import { defineExtension } from "htmx.org";
import { getFormControls, SlButton } from "@shoelace-style/shoelace";

defineExtension("shoelace", {
    onEvent(
        name: string,
        event:
            | Event
            | CustomEvent<
                { elt: HTMLElement; parameters: Record<string, string> }
            >,
    ) {
        console.log(name);
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
        const form = event.detail.elt;
        if (!(form instanceof HTMLFormElement)) {
            console.groupEnd();
            return true;
        }

        for (const slElement of getFormControls(form) as HTMLFormElement[]) {
            const { name, value } = slElement;

            console.log("Form data set: %o %o", name, value);
            event.detail.parameters[name] = value;
        }
        console.log(
            "Event detail parameters (form data):",
            event.detail.parameters,
        );

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
