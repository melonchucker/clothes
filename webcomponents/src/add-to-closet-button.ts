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
    private closets: Closet[] = [];

    private svg = html`
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
    </svg>`;

    override connectedCallback() {
        super.connectedCallback();
        // document.addEventListener(
        //     "pointerdown",
        //     this._onDocumentPointerDown,
        //     true,
        // );
    }

    override disconnectedCallback() {
        // document.removeEventListener(
        //     "pointerdown",
        //     this._onDocumentPointerDown,
        //     true,
        // );
        super.disconnectedCallback();
    }

    override createRenderRoot() {
        return this;
    }

    private async _buttonClick(e: MouseEvent) {
        e.stopPropagation();
        if (this.open) {
            this.open = false;
            return;
        }

        let req = await fetch("/api/user/closets");
        let closets = await req.json();
        this.closets = closets;
        this.open = true;
    }

    private async _addToCloset(closetName: string) {
        console.log(
            `Adding ${this.item} by ${this.brand} to closet ${closetName}`,
        );
        let req = await fetch("/api/user/closets/add_item", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                closet_name: closetName,
                item: this.item,
                brand: this.brand,
            }),
        });
        if (!req.ok) {
            console.error("Failed to add item to closet");
            return;
        }

        this.open = false;
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
                    Math.max(0, Math.round(r.left))
                }px; top:${
                    Math.max(
                        0,
                        Math.round(r.bottom),
                    )
                }px; min-width:${Math.round(r.width)}px; z-index:2000;`;
            }
        }

        // BEGIN KLUDGE
        // This is a workaround to use the bootstrap modals within a Lit component
        // if they share the same ID, they all refer to the first one in the DOM
        const uid = `${this.brand}-${this.item}`;
        const id = `add-to-closet-${
            uid.replace(/[^a-z0-9]+/gi, "-").toLowerCase()
        }`;
        const labelId = `${id}-label`;
        // END KLUDGE

        return html`
            <!-- Button trigger modal -->
            <button 
            type="button" 
            class="add-to-closet-trigger btn btn-link position-absolute top-0 end-0 m-2 z-3 p-2 d-flex align-items-center justify-content-center text-light"
            aria-label="Add to closet" 
            data-bs-toggle="modal" 
            data-bs-target="#${id}"
            @click=${(e: MouseEvent) => this._buttonClick(e)}
            >
            ${this.svg}
            </button>

            <!-- Modal -->
            <div class="modal fade" id="${id}" tabindex="-1" aria-labelledby="${labelId}" aria-hidden="true">
            <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
            <div class="modal-header">
            <h1 class="modal-title fs-5" id="${labelId}">Add <em>${this.brand} ${this.item}</em> to closet</h1>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
            </div>
            <div class="modal-body">
            ${
            this.closets.length
                ? html`<div class="list-group">
                ${
                    this.closets.map(
                        (closet) =>
                            html`<button
                type="button"
                class="list-group-item list-group-item-action"
                @click=${() => this._addToCloset(closet.name)}
                data-bs-dismiss="modal"
                >${closet.name}</button>`,
                    )
                }
            </div>`
                : html`<div class="dropdown-item-text p-2">No closets</div>`
        }
            </div>
            </div>
            </div>
            </div>          
            `;
    }
}
