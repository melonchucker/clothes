import { LitElement, html, TemplateResult } from "lit";
import { customElement, property, state } from "lit/decorators.js";

interface SearchData {
    tags: string[];
    items: string[];
    brands: string[];
    _error?: boolean;
}

/**
 * <search-bar>
 */
@customElement("search-bar")
export class SearchBarElement extends LitElement {
    // Public attributes
    @property({ type: String }) action = "/search";
    @property({ type: String }) method = "get";
    @property({ type: String }) api = "/api/search_bar";
    @property({ type: String }) placeholder = "Search";
    @property({ type: String }) name = "q";
    @property({ type: Number }) minChars = 1;
    @property({ type: Number }) debounceMs = 150;

    // Internal reactive state
    @state() open = false;
    @state() loading = false;
    @state() query = "";
    @state() data: SearchData | null = null;

    // Private fields
    private _cache = new Map<string, SearchData>();
    private _abort: AbortController | null = null;
    private _debounceTimer: number | null = null;

    constructor() {
        super();
        this._onDocumentPointerDown = this._onDocumentPointerDown.bind(this);
    }

    // Use light DOM so external CSS (e.g. Bootstrap) can style dropdowns
    override createRenderRoot() {
        return this;
    }

    override connectedCallback() {
        super.connectedCallback();
        document.addEventListener("pointerdown", this._onDocumentPointerDown, true);
    }

    override disconnectedCallback() {
        document.removeEventListener(
            "pointerdown",
            this._onDocumentPointerDown,
            true
        );
        this._abort?.abort();
        if (this._debounceTimer !== null) {
            clearTimeout(this._debounceTimer);
            this._debounceTimer = null;
        }
        super.disconnectedCallback();
    }

    private _onDocumentPointerDown(e: PointerEvent) {
        const target = e.target as Node | null;
        if (!target) return;
        if (!this.contains(target)) this.open = false;
    }

    private _onFocusIn() {
        this.open = true;
        const input = this.renderRoot.querySelector('input[type="search"]') as
            | HTMLInputElement
            | null;
        const q = input?.value?.trim() ?? "";
        if (q.length >= this.minChars) this._scheduleFetch(q);
    }

    private _onInput(e: InputEvent) {
        const target = e.target as HTMLInputElement | null;
        const q = target?.value.trim() ?? "";
        this.query = q;

        if (q.length < this.minChars) {
            this.loading = false;
            this.data = null;
            return;
        }

        this.open = true;
        this._scheduleFetch(q);
    }

    private _scheduleFetch(q: string) {
        if (this._debounceTimer !== null) {
            clearTimeout(this._debounceTimer);
        }
        this._debounceTimer = window.setTimeout(() => this._fetch(q), this.debounceMs);
    }

    private async _fetch(q: string) {
        // cache hit
        if (this._cache.has(q)) {
            this.data = this._cache.get(q) ?? null;
            this.loading = false;
            return;
        }

        // abort any in-flight request
        this._abort?.abort();
        this._abort = new AbortController();

        this.loading = true;
        this.data = null;

        const url = new URL(this.api, window.location.origin);
        url.searchParams.set("input", q);

        try {
            const res = await fetch(url.toString(), { signal: this._abort.signal });
            if (!res.ok) throw new Error(`search api failed: ${res.status}`);

            const data = (await res.json()) as SearchData;
            this._cache.set(q, data);

            // ignore stale responses
            if (this.query !== q) return;

            this.data = data;
        } catch (err: unknown) {
            if ((err as any)?.name === "AbortError") return;
            this.data = { tags: [], items: [], brands: [], _error: true };
        } finally {
            if (this.query === q) this.loading = false;
        }
    }

    private _renderSection(
        title: string,
        items: string[] | undefined,
        hrefPrefix: string
    ): TemplateResult | null {
        if (!items || items.length === 0) return null;

        return html`
            <div class="mb-2">
                <h6 class="dropdown-header">${title}</h6>
                <ul class="list-unstyled mb-0">
                    ${items.map(
                        (v) => html`
                            <li>
                                <a class="dropdown-item" href="${hrefPrefix}${encodeURIComponent(v)}">
                                    ${v}
                                </a>
                            </li>
                        `
                    )}
                </ul>
            </div>
        `;
    }

    override render() {
        return html`
            <form
                class="topbar-search d-none d-lg-flex position-relative"
                role="search"
                action="${this.action}"
                method="${this.method}"
                @focusin=${() => this._onFocusIn()}
            >
                <div class="input-group nav-search">
                    <span class="input-group-text">
                        <svg
                            class="icon"
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 640 640"
                            aria-hidden="true"
                            focusable="false"
                        >
                            <path
                                d="M480 272C480 317.9 465.1 360.3 440 394.7L566.6 521.4C579.1 533.9 579.1 554.2 566.6 566.7C554.1 579.2 533.8 579.2 521.3 566.7L394.7 440C360.3 465.1 317.9 480 272 480C157.1 480 64 386.9 64 272C64 157.1 157.1 64 272 64C386.9 64 480 157.1 480 272zM272 416C351.5 416 416 351.5 416 272C416 192.5 351.5 128 272 128C192.5 128 128 192.5 128 272C128 351.5 192.5 416 272 416z"
                            />
                        </svg>
                        <span class="visually-hidden">Search</span>
                    </span>

                    <input
                        name="${this.name}"
                        type="search"
                        class="form-control"
                        placeholder="${this.placeholder}"
                        autocomplete="off"
                        @input=${(e: InputEvent) => this._onInput(e)}
                        @keydown=${(e: KeyboardEvent) => {
                            if (e.key === "Escape") this.open = false;
                        }}
                    />
                </div>

                ${this.open
                    ? html`
                            <div
                                class="dropdown-menu show dropdown-search w-100 mt-1"
                                style="position:absolute; left:0; top:100%; z-index:2000;"
                            >
                                ${this.loading
                                    ? html`<div class="px-3 py-2 text-muted small">Searching…</div>`
                                    : this.data
                                    ? html`
                                            ${this._renderSection("Tags", this.data.tags, "/tag/")}
                                            ${this._renderSection("Items", this.data.items, "/item/")}
                                            ${this._renderSection("Brands", this.data.brands, "/brand/")}
                                            ${!this.data.tags?.length &&
                                            !this.data.items?.length &&
                                            !this.data.brands?.length
                                                ? html`<div class="px-3 py-2 text-muted small">No results</div>`
                                                : null}
                                        `
                                    : html`<div class="px-3 py-2 text-muted small">Type to search</div>`}
                            </div>
                        `
                    : null}
            </form>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        "search-bar": SearchBarElement;
    }
}