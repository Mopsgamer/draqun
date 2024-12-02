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
