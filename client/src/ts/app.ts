import "./main.ts";

import "./chat-mutate.ts";
import "./chat-send.ts";
import "./notifications.ts";
import { domLoaded } from "./lib.ts";

import("htmx-ext-ws");

domLoaded.then(() => {
	document.addEventListener("click", (event) => {
		const target = event.target as HTMLElement | null;
		if (!target) return;

		const channelsBtn = target.closest(".group-channels-toggle");
		if (channelsBtn) {
			event.preventDefault();
			event.stopPropagation();
			const channelsView = document.getElementById("channels-view");
			if (channelsView) {
				const isDesktop = globalThis.innerWidth >= 768;
				if (isDesktop) {
					channelsView.classList.toggle("closed");
				} else {
					const membersView = document.getElementById("members-view");
					if (membersView) membersView.classList.remove("open");
					channelsView.classList.toggle("open");
				}
				const isOpen = isDesktop
					? !channelsView.classList.contains("closed")
					: channelsView.classList.contains("open");
				document.querySelectorAll(".group-channels-toggle").forEach((btn) => {
					btn.setAttribute("aria-expanded", String(isOpen));
				});
			}
			return;
		}

		const membersBtn = target.closest(".group-members-toggle");
		if (membersBtn) {
			event.preventDefault();
			event.stopPropagation();
			const membersView = document.getElementById("members-view");
			if (membersView) {
				const isDesktop = globalThis.innerWidth >= 768;
				if (isDesktop) {
					membersView.classList.toggle("closed");
				} else {
					const channelsView = document.getElementById("channels-view");
					if (channelsView) channelsView.classList.remove("open");
					membersView.classList.toggle("open");
				}
				const isOpen = isDesktop
					? !membersView.classList.contains("closed")
					: membersView.classList.contains("open");
				document.querySelectorAll(".group-members-toggle").forEach((btn) => {
					btn.setAttribute("aria-expanded", String(isOpen));
				});
			}
			return;
		}
	});
});
