import { getFormControls } from "@shoelace-style/shoelace";

export function findLastMessage(): Element | undefined {
    const chat = document.getElementById("chat")!;
    return Array.from(chat.getElementsByClassName("message")).reverse()[0];
}

export function findLastMessageVisibleDate(): Element | undefined {
    const chat = document.getElementById("chat")!;
    return Array.from(chat.getElementsByClassName("message")).reverse().find(
        (c) => {
            const { classList } = c.previousElementSibling ?? {};
            if (!classList) {
                // if message is first, found
                return true;
            }
            // if message is joined by date, ignore
            if (classList.contains("join") && classList.contains("date")) {
                return false;
            }

            return true;
        },
    );
}

export function getFormPropData(form: HTMLFormElement, capital = false) {
    const data: Record<string, string> = {};
    for (const slElement of getFormControls(form) as (Element & {value: unknown, name: unknown})[]) {
        let { name } = slElement;
        const { value } = slElement;

        if (typeof name !== 'string' || typeof value !== 'string') {
            continue;
        }

        if (capital) {
            name = (name[0].toUpperCase() + name.substring(1));
        }
        data[name as string] = value;
    }

    return data;
}
