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

domLoaded.then(() => {
	if ("Notification" in globalThis) {
		updateNotificationButtonState(Notification.permission);
	}

	(globalThis as unknown as {
		requestBrowserNotificationsPermission:
			typeof requestBrowserNotificationsPermission;
	}).requestBrowserNotificationsPermission =
		requestBrowserNotificationsPermission;
});
