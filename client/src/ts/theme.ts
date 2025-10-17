import { SlMenu, type SlMenuItem, SlSelect } from "@shoelace-style/shoelace";
import { domLoaded } from "./lib.ts";

enum Theme {
    dark = "sl-theme-dark",
    light = "sl-theme-light",
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

    const target = document.documentElement;
    target.classList.remove(...themeList);

    if (theme === Theme.system) {
        const prefersDark = matchMedia("(prefers-color-scheme: dark)").matches;
        target.classList.add(prefersDark ? Theme.dark : Theme.light);
    } else {
        target.classList.add(theme);
    }
}

function updateThemeMenuElements(): void {
    const menuList = Array.from(document.querySelectorAll<SlMenu | SlSelect>(".theme-menu"));
    for (const menu of menuList) {
        if (menu instanceof SlSelect) {
            menu.value = getTheme();
            continue;
        }
        if (menu instanceof SlMenu) {
            const allItemList = menuList.flatMap(
                (menu) => [
                    ...menu.querySelectorAll<SlMenuItem>(
                        "sl-menu-item[type=checkbox][value]",
                    ),
                ],
            );

            const theme = getTheme();
            for (const child of allItemList) {
                child.checked = child.value === theme;
            }
            continue;
        }
    }
}

function initThemeMenuElements(): void {
    const menuList = Array.from(
        document.querySelectorAll<SlMenu | SlSelect>(".theme-menu"),
    );

    for (const menu of menuList) {
        if (menu instanceof SlSelect) {
            menu.addEventListener("sl-change", () => {
                const theme = menu.value as Theme
                setTheme(theme)
            })
            continue;
        }
        if (menu instanceof SlMenu) {
            menu.addEventListener("sl-select", (event) => {
                const item = event.detail.item;
                if (item.type !== "checkbox") {
                    return;
                }

                const theme = item.value;
                if (!isTheme(theme)) {
                    console.error(
                        `Unknown theme ${theme}, can not change: %o.`,
                        item,
                    );
                    item.checked = !item.checked;
                    return;
                }

                setTheme(theme);
                updateThemeMenuElements();
            });
            continue;
        }
    }
}

/**
 * Initialize the theme by reading from localStorage and applying it.
 */
function initTheme(): void {
    const theme = getTheme();
    setTheme(theme);

    if (theme === Theme.system) {
        // Add a listener to react to system theme changes
        matchMedia("(prefers-color-scheme: dark)").addEventListener(
            "change",
            () => {
                setTheme(Theme.system);
            },
        );
    }
}

initTheme();
domLoaded.then(() => {
    updateThemeMenuElements();
    initThemeMenuElements();
});
