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
    const data: Record<string, string | number | boolean> = {};
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
        ) continue;

        if (capital) {
            name = name[0].toUpperCase() + name.substring(1);
        }

        const convert = slElement.getAttribute("data-convert");
        data[name as string] = convert == "number"
            ? Number(value)
            : convert == "boolean"
            ? Boolean(value)
            : value;
    }

    return data;
}

export function chatJoinMessages(): void {
    const chat = document.getElementById("chat");
    if (!chat) return;

    for (const element of chat.children) {
        if (!isMessageElement(element)) continue;

        const next = element.nextElementSibling;
        if (!isMessageElement(next)) continue;

        const elementAuthorId = element.getAttribute("data-author");
        const nextAuthorId = next.getAttribute("data-author");
        const sameAuthor = elementAuthorId === nextAuthorId;

        const elementDate = new Date(element.getAttribute("data-created-at")!);
        const nextDate = new Date(next.getAttribute("data-created-at")!);
        const dateDiff = nextDate.getTime() - elementDate.getTime();
        const sameDate = dateDiff < 1000 * 60 * 5; // 5 minutes

        // join same author messages always

        // if (sameAuthor) {
        //     element.classList.add("join-end");
        //     next.classList.add("join-start");
        //     if (sameDate) {
        //         next.classList.add("same-date");
        //     }
        // }

        // join same author messages only if same date

        if (sameAuthor && sameDate) {
            element.classList.add("join-end");
            next.classList.add("join-start");
            next.classList.add("same-date");
        }
    }
}

export function chatScrollDownIfNoScroll(): void {
    const chat = document.getElementById("chat");
    if (!chat) return;

    const isScrolledToBottom =
        chat.scrollHeight - chat.scrollTop - chat.clientHeight < 450;
    if (isScrolledToBottom) {
        chat.scrollTop = chat.scrollHeight;
    }
}

export function chatScrollDown(): void {
    const chat = document.getElementById("chat");
    if (!chat) return;
    chat.scrollTop = chat.scrollHeight;
}

export const domLoaded = new Promise<void>((resolve) => {
    addEventListener("DOMContentLoaded", () => {
        resolve();
    });
});

export function toggleFocus(element: HTMLElement): void {
    if (element.matches(":focus")) {
        element.blur();
    } else {
        element.focus();
    }
}

const usedAnchorIds = new Set<string>();
export function initAnchorHeadersFor(target: HTMLElement): void {
    const hAnchorList = Array.from(
        target.getElementsByClassName("anchor-header"),
    )
        .filter((el) => /H[1-6]/.test(el.tagName)) as HTMLHeadingElement[];

    for (const hAnchor of hAnchorList) {
        initAnchorHeader(hAnchor);
    }
}

export function initAnchorHeader(hAnchor: HTMLHeadingElement): void {
    const text = Array.from(hAnchor.childNodes)
        .filter((el) => el.nodeType === el.TEXT_NODE)
        .map((el) => el.textContent).join("");

    if (!text) {
        console.error("Invalid anchor-header text content\n%o", hAnchor);
        return;
    }

    const hashId = text.trim().replaceAll(/[^a-zA-Z]/g, "_");
    if (usedAnchorIds.has(hashId)) {
        console.warn(
            "Repeating anchor-header id: %o\nElement: %o",
            hashId,
            hAnchor,
        );
        const element = document.getElementById(hashId);
        if (element) {
            uninitAnchorHeader(element);
        }
        return;
    }

    hAnchor.id = hashId;
    usedAnchorIds.add(hashId);

    const a = document.createElement("a");
    a.href = "#" + hashId;
    a.classList.add("anchor-header-link");
    a.append("#");
    hAnchor.append(a);
}

export function uninitAnchorHeader(hAnchor: HTMLElement): void {
    hAnchor.removeAttribute("id");
    for (
        const element of hAnchor.getElementsByClassName("anchor-header-link")
    ) {
        element.remove();
    }
}
