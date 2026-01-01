import { html, LitElement } from "lit";
import { customElement, property, state } from "lit/decorators.js";

interface Closet {
    name: string;
    [key: string]: unknown;
}

@customElement("add-to-closet")
export class AddToClosetButton extends LitElement {
    @property({ type: String })
    brand = "";
    @property({ type: String })
    item = "";

    // internal reactive state
    @state()
    private open = false;
    @state()
    private loading = false;
    @state()
    private query = "";
    @state()
    private data: unknown = null;
    @state()
    private closets: Closet[] = [];

    // internal non-reactive helpers
    private _abort?: AbortController;
    private _debounceTimer?: number;

    // use arrow functions so `this` is preserved; no manual bind needed
    private _onDocumentPointerDown = (e: PointerEvent) => {
        const target = e.target as Node | null;
        if (!target || !this.contains(target)) {
            this.open = false;
        }
    };

    private _onClick = async () => {
        if (this.open) {
            this.open = false;
            return;
        }

        this.loading = true;
        try {
            const res = await fetch("/api/user/closets");
            if (!res.ok) throw new Error("Failed to load closets");
            const closets = (await res.json()) as Closet[];
            this.closets = closets;
            this.open = true;
        } catch (err) {
            console.error(err);
        } finally {
            this.loading = false;
        }
    };

    private _createNewCloset = async (e: Event) => {
        e.preventDefault();
        const form = e.target as HTMLFormElement;
        const formData = new FormData(form);
        const newClosetName = formData.get("newClosetName") as string;
        if (!newClosetName) return;

        this._abort?.abort();
        this._abort = new AbortController();

        try {
            const res = await fetch("/api/user/closets/create", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name: newClosetName,
                }),
                signal: this._abort.signal,
            });

            if (!res.ok) {
                console.error("Failed to create new closet", await res.text());
                return;
            }

            const newCloset = await res.json();
            this.closets = [...this.closets, newCloset];
            form.reset();
        } catch (err) {
            if ((err as DOMException)?.name === "AbortError") return;
            console.error(err);
        } finally {
            this._abort = undefined;
        }
    };

    private _addToCloset = async (e: Event) => {
        const target = (e.currentTarget ?? e.target) as HTMLElement | null;
        const closetName = target?.textContent?.trim() ?? "";
        if (!closetName) return;

        this._abort?.abort();
        this._abort = new AbortController();

        try {
            const res = await fetch("/api/user/closets/add_item", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    closet_name: closetName,
                    item: this.item,
                    brand: this.brand,
                }),
                signal: this._abort.signal,
            });

            if (!res.ok) {
                console.error("Failed to add item to closet", await res.text());
                return;
            }

            // success
            this.open = false;
        } catch (err) {
            if ((err as DOMException)?.name === "AbortError") return;
            console.error(err);
        } finally {
            this._abort = undefined;
        }
    };

    override connectedCallback() {
        super.connectedCallback();
        document.addEventListener(
            "pointerdown",
            this._onDocumentPointerDown,
            true,
        );
    }

    override disconnectedCallback() {
        document.removeEventListener(
            "pointerdown",
            this._onDocumentPointerDown,
            true,
        );
        this._abort?.abort();
        if (this._debounceTimer) {
            window.clearTimeout(this._debounceTimer);
            this._debounceTimer = undefined;
        }
        super.disconnectedCallback();
    }

    override createRenderRoot() {
        return this;
    }

    override render() {
        const btnSelector = ".add-to-closet-trigger";
        let menuStyle = "position:fixed; right:1rem; top:3rem; z-index:2000;"; // fallback

        if (this.open) {
            const btn = this.renderRoot.querySelector(btnSelector) as
                | HTMLElement
                | null;
            if (btn) {
                const r = btn.getBoundingClientRect();
                menuStyle = `position:fixed; left:${
                    Math.max(0, r.left)
                }px; top:${
                    Math.max(
                        0,
                        r.bottom,
                    )
                }px; min-width:${r.width}px; z-index:2000;`;
            }
        }

        return html`
            <button
                @click=${this._onClick}
                type="button"
                class="add-to-closet-trigger btn btn-link position-absolute top-0 end-0 m-2 z-3 p-2 d-flex align-items-center justify-content-center text-light"
                aria-label="Add to closet"
                aria-haspopup="true"
                aria-expanded=${this.open ? "true" : "false"}
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="1.25rem"
                    height="1.25rem"
                    fill="currentColor"
                    class="bi bi-heart"
                    viewBox="0 0 16 16"
                >
                    <path
                        d="m8 2.748-.717-.737C5.6.281 2.514.878 1.4 3.053c-.523 1.023-.641 2.5.314 4.385.92 1.815 2.834 3.989 6.286 6.357 3.452-2.368 5.365-4.542 6.286-6.357.955-1.886.838-3.362.314-4.385C13.486.878 10.4.28 8.717 2.01zM8 15C-7.333 4.868 3.279-3.04 7.824 1.143q.09.083.176.171a3 3 0 0 1 .176-.17C12.72-3.042 23.333 4.867 8 15"
                    />
                </svg>

                ${
            this.open
                ? html`<div
                            class="dropdown-menu show dropdown-search mt-1"
                            style=${menuStyle}
                        >
                            <h3 class="dropdown-header">Select Closet to Add</h3>
                            ${
                    this.closets.map(
                        (closet) =>
                            html` <button
                                    @click=${this._addToCloset}
                                    type="button"
                                    class="dropdown-item"
                                >
                                    ${closet.name}
                                </button>`,
                    )
                }
                            <h3 class="dropdown-header">Create New Closet</h3>
                            <form @submit=${this._createNewCloset}>
                                <input type="text" name="newClosetName" placeholder="New closet name" required />
                                <button type="submit" class="btn btn-primary">Create</button>
                            </form>
                        </div>`
                : null
        }
            </button>
        `;
    }
}
