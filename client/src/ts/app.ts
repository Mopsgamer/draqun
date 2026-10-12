import "./main.ts";

import "./chat-mutate.ts";
import "./chat-send.ts";
import { domLoaded } from "./lib.ts";

import("htmx-ext-ws");

function closeAllBut(
	element: HTMLElement,
	secondaryViewList: HTMLElement[],
): void {
	element.classList.toggle("open");
	for (const secondary of secondaryViewList) {
		if (secondary !== element) {
			secondary.classList.remove("open");
		}
	}
}

domLoaded.then(() => {
	document.addEventListener("click", (event) => {
		const target = event.target as HTMLElement | null;
		if (!target) return;

		const channelsBtn = target.closest(".group-channels-toggle");
		if (channelsBtn) {
			const channelsView = document.getElementById("channels-view");
			if (channelsView) {
				const secondaryViewList = Array.from(
					document.getElementsByClassName("secondary-view"),
				) as HTMLElement[];
				closeAllBut(channelsView, secondaryViewList);
				const isOpen = channelsView.classList.contains("open");
				document.querySelectorAll(".group-channels-toggle").forEach((btn) => {
					btn.setAttribute("aria-expanded", String(isOpen));
				});
			}
			return;
		}

		const membersBtn = target.closest(".group-members-toggle");
		if (membersBtn) {
			const membersView = document.getElementById("members-view");
			if (membersView) {
				const secondaryViewList = Array.from(
					document.getElementsByClassName("secondary-view"),
				) as HTMLElement[];
				closeAllBut(membersView, secondaryViewList);
				const isOpen = membersView.classList.contains("open");
				document.querySelectorAll(".group-members-toggle").forEach((btn) => {
					btn.setAttribute("aria-expanded", String(isOpen));
				});
			}
			return;
		}
	});
});
