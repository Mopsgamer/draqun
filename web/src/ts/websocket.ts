import * as HTMX from "htmx.org";
import { getFormPropData } from "./lib.ts";

HTMX.on("htmx:wsConfigSend", (
    event,
) => {
    const form = event.target;

    if (!(form instanceof HTMLFormElement) || !(event instanceof CustomEvent)) {
        return;
    }

    const { detail } = event;

    Object.assign(detail.parameters, getFormPropData(form, true), {
        Type: form.id,
    });
});
