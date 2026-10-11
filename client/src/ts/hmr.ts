if (typeof EventSource !== "undefined") {
	let isReloading = false;
	const safeReload = (): void => {
		if (isReloading) return;
		isReloading = true;
		globalThis.location.reload();
	};

	const es = new EventSource("/__hmr");

	es.addEventListener("connected", (e: MessageEvent) => {
		try {
			const data = JSON.parse(e.data);
			const prevSessionId = sessionStorage.getItem("__hmr_session_id");
			sessionStorage.setItem("__hmr_session_id", data.sessionId);
			if (prevSessionId && prevSessionId !== data.sessionId) {
				safeReload();
			}
		} catch {
			// ignore invalid JSON
		}
	});

	es.addEventListener("css", () => {
		const links = document.querySelectorAll<HTMLLinkElement>(
			'link[rel="stylesheet"]',
		);
		const timestamp = Date.now().toString();
		links.forEach((link) => {
			const url = new URL(link.href, globalThis.location.origin);
			url.searchParams.set("v", timestamp);
			link.href = url.toString();
		});
	});

	es.addEventListener("reload", () => {
		safeReload();
	});

	es.onerror = () => {
		// EventSource will automatically reconnect when server is back
	};

	globalThis.addEventListener("beforeunload", () => {
		es.close();
	});
}
