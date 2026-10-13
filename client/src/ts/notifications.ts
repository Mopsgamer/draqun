import { domLoaded } from "./lib.ts";

export function requestBrowserNotificationsPermission(): Promise<
	NotificationPermission
> {
	if (!("Notification" in globalThis)) {
		return Promise.resolve("denied");
	}
	if (Notification.permission === "granted") {
		return Promise.resolve("granted");
	}
	return Notification.requestPermission().then((permission) => {
		updateNotificationButtonState(permission);
		return permission;
	});
}

function updateNotificationButtonState(
	permission: NotificationPermission,
): void {
	const btn = document.getElementById("btn-enable-browser-notifications");
	if (!btn) return;

	if (permission === "granted") {
		btn.classList.add("text-success");
		btn.setAttribute("title", "Desktop notifications enabled");
	} else if (permission === "denied") {
		btn.classList.add("text-error");
		btn.setAttribute(
			"title",
			"Desktop notifications blocked by browser settings",
		);
	}
}

function showDesktopNotification(
	title: string,
	body: string,
	targetUrl?: string,
): void {
	if (!("Notification" in globalThis) || Notification.permission !== "granted") {
		return;
	}

	try {
		const notif = new Notification(title, {
			body: body,
			icon: "/static/js/logo.svg",
		});

		if (targetUrl) {
			notif.onclick = () => {
				globalThis.focus();
				globalThis.location.href = targetUrl;
			};
		}
	} catch (e) {
		console.error("Error showing browser notification:", e);
	}
}

function checkAndTriggerToastNotification(node: Element): void {
	let toastElem: HTMLElement | null = null;
	if (node instanceof HTMLElement && node.id === "desktop-notification-toast") {
		toastElem = node;
	} else if (node instanceof HTMLElement) {
		toastElem = node.querySelector("#desktop-notification-toast");
	}

	if (!toastElem) return;

	const title = toastElem.getAttribute("data-title") || "New Notification";
	const body = toastElem.getAttribute("data-body") || "";
	const targetUrl = toastElem.getAttribute("data-url") || undefined;

	showDesktopNotification(title, body, targetUrl);

	setTimeout(() => {
		toastElem?.remove();
	}, 5000);
}

domLoaded.then(() => {
	if ("Notification" in globalThis) {
		updateNotificationButtonState(Notification.permission);
	}

	(globalThis as unknown as {
		requestBrowserNotificationsPermission:
			typeof requestBrowserNotificationsPermission;
	}).requestBrowserNotificationsPermission =
		requestBrowserNotificationsPermission;

	const observer = new MutationObserver((mutations) => {
		for (const mutation of mutations) {
			for (const addedNode of mutation.addedNodes) {
				if (addedNode.nodeType === Node.ELEMENT_NODE) {
					checkAndTriggerToastNotification(addedNode as Element);
				}
			}
		}
	});

	observer.observe(document.body, { childList: true, subtree: true });
});
