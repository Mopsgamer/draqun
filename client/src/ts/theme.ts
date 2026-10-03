import { domLoaded } from "./lib.ts";

enum Theme {
	dark = "dark",
	light = "light",
	system = "system",
}

const themeList = new Set(Object.values(Theme));

function isTheme(theme: unknown): theme is Theme {
	return themeList.has(theme as Theme);
}

/**
 * Get the current theme from localStorage.
 * If no theme is set, defaults to 'system'.
 */
function getTheme(): Theme {
	if (location.pathname === "/docs") {
		return Theme.light;
	}

	const theme = localStorage.getItem("theme") ?? Theme.system;
	if (!isTheme(theme)) {
		return Theme.system;
	}

	return theme;
}

/**
 * Set the theme in localStorage and apply it to the document body.
 */
function setTheme(theme: Theme): void {
	if (location.pathname !== "/docs") {
		localStorage.setItem("theme", theme);
	}

	if (theme === Theme.system) {
		const prefersDark = matchMedia("(prefers-color-scheme: dark)").matches;
		document.documentElement.dataset.theme = prefersDark
			? Theme.dark
			: Theme.light;
	} else {
		document.documentElement.dataset.theme = theme;
	}
}

function updateThemeMenuElements(): void {
	const theme = getTheme();
	for (
		const select of document.querySelectorAll<HTMLSelectElement>(
			"select.theme-menu",
		)
	) {
		select.value = theme;
	}
	for (
		const choice of document.querySelectorAll<HTMLInputElement>(
			"input[name=theme-choice]",
		)
	) {
		choice.checked = choice.value === theme;
	}
}

function initThemeMenuElements(): void {
	document.addEventListener("change", (event) => {
		const target = event.target;
		if (
			target instanceof HTMLSelectElement &&
			target.matches("select.theme-menu")
		) {
			if (!isTheme(target.value)) {
				console.error(`Unknown theme ${target.value}, can not change.`);
				updateThemeMenuElements();
				return;
			}
			setTheme(target.value);
			updateThemeMenuElements();
			target.closest(".dropdown")?.removeAttribute("open");
			return;
		}

		if (
			target instanceof HTMLInputElement &&
			target.matches("input[name=theme-choice]")
		) {
			if (!isTheme(target.value)) {
				console.error(`Unknown theme ${target.value}, can not change.`);
				updateThemeMenuElements();
				return;
			}
			setTheme(target.value);
			updateThemeMenuElements();
		}
	});
}

/**
 * Initialize the theme by reading from localStorage and applying it.
 */
function initTheme(): void {
	const theme = getTheme();
	setTheme(theme);

	matchMedia("(prefers-color-scheme: dark)").addEventListener(
		"change",
		() => {
			if (getTheme() === Theme.system) setTheme(Theme.system);
		},
	);
}

initTheme();
domLoaded.then(() => {
	updateThemeMenuElements();
	initThemeMenuElements();
});
