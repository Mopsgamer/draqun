import "./main.ts";

import "./chat-scroll.ts";
import "./chat-send.ts";
import "./chat-join-messages.ts";
import { domLoaded } from "./lib.ts";

import("htmx-ext-ws");

function closeAllBut(element: HTMLElement, secondaryViewList: HTMLElement[]) {
    element.classList.toggle("open");
    for (const secondary of secondaryViewList) {
        if (secondary !== element) {
            secondary.classList.remove("open");
        }
    }
}

domLoaded.then(() => {
    const membersToggler = document.getElementById("members-toggler")!;
    const secondaryViewList = Array.from(
        document.getElementsByClassName("secondary-view") as HTMLCollectionOf<
            HTMLElement
        >,
    );
    const membersView = document.getElementById("members-view");
    if (!membersView) {
        return
    }
    membersToggler.addEventListener(
        "click",
        () => closeAllBut(membersView, secondaryViewList),
    );
});
