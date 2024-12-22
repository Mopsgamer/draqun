import { getFormControls } from "@shoelace-style/shoelace";

export function isMessageJoinElement(value: unknown): value is HTMLDivElement {
    return !!value && value instanceof HTMLDivElement &&
        value.classList.contains("join");
}

export function isMessageJoinDateElement(
    value: unknown,
): value is HTMLDivElement {
    return isMessageJoinElement(value) && value.classList.contains("date");
}

export function isMessageElement(value: unknown): value is HTMLDivElement {
    return !!value && value instanceof HTMLDivElement &&
        value.classList.contains("message");
}

export function findLastMessage(): Element | undefined {
    const chat = document.getElementById("chat");
    if (!chat) return;
    return Array.from(chat.getElementsByClassName("message")).reverse()[0];
}

export function findLastMessageVisibleDate(): Element | undefined {
    const chat = document.getElementById("chat");
    if (!chat) return;
    return Array.from(chat.getElementsByClassName("message")).reverse().find(
        (c) => {
            const previous = c.previousElementSibling;
            if (!isMessageJoinDateElement(previous)) {
                return true;
            }

            return false;
        },
    );
}

export function getFormPropData(form: HTMLFormElement, capital = false) {
    const data: Record<string, string> = {};
    for (
        const slElement of getFormControls(
            form,
        ) as (Element & { value: unknown; name: unknown })[]
    ) {
        let { name } = slElement;
        const { value } = slElement;

        if (
            typeof name !== "string" || typeof value !== "string" || !name ||
            !value
        ) {
            continue;
        }

        if (capital) {
            name = name[0].toUpperCase() + name.substring(1);
        }
        data[name as string] = value;
    }

    return data;
}

export function chatJoinMessages(): void {
    const chat = document.getElementById("chat");
    if (!chat) return;

    for (const element of chat.children) {
        if (!isMessageElement(element)) {
            continue;
        }

        if (!isMessageElement(element.nextElementSibling)) {
            continue;
        }

        const shouldJoin = element.getAttribute("data-author") ===
            element.nextElementSibling.getAttribute("data-author");

        const dateDiff = new Date(
            element.nextElementSibling.getAttribute("data-created-at")!,
        ).getTime() -
            new Date(
                element.getAttribute("data-created-at")!,
            ).getTime();
        const shouldJoinDate = dateDiff < 1000 * 60 * 5; // 5 minutes

        if (shouldJoin) {
            element.classList.add("join-end");
            element.nextElementSibling.classList.add("join-start");
            if (shouldJoinDate) {
                element.nextElementSibling.classList.add("hide-date");
            }
        }
    }
}
