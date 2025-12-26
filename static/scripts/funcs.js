/**
 * @param {HTMLInputElement} inputElement
 */
// async function searchBarOninput(inputElement) {
//   let input = inputElement.value.trim();
//   if (input === "") {
//     return;
//   }
//   let res = await fetch(`/api/search_bar?input=${encodeURIComponent(input)}`);
//   let data = await res.json();

//   searchResultsDiv = document.querySelector("#search-results");
//   searchResultsDiv.innerHTML = JSON.stringify(data);
//   for (const key in data) {
//     let div = document.createElement("div");
//     title = document.createElement("h3");
//     title.textContent = key;
//     searchResultsDiv.appendChild(title);
//   }
// }

class SearchBar {
  static dropdown_template = `
    <div class="dropdown-menu show" style="
      position: absolute;
      left: 0;
      top: calc(100% + 6px);
      z-index: 2000;
      width: 100%;
      background: #fff;
      color: #212529;
      box-shadow: 0 .5rem 1rem rgba(0,0,0,.15);
      border-radius: .25rem;
      overflow: hidden;
    ">
    </div>
  `;

  /**
   * @param {string} searchBarId
   */
  constructor(searchBarId) {
    this.searchBar = /** @type {HTMLInputElement} */ (
      document.getElementById(searchBarId)
    );
    this.input = this.input.bind(this);
    this.focus = this.focus.bind(this);
    this.blur = this.blur.bind(this);

    this.searchBar.addEventListener("input", this.input);
    this.searchBar.addEventListener("focus", this.focus);
    this.searchBar.addEventListener("blur", this.blur);
  }

  /**
   * @param {Event} e
   */
  async input(e) {
    let res = await fetch(`/api/search_bar?input=${encodeURIComponent(e.target.value)}`);
    let data = await res.json();
    console.log("querying:", e.target.value);
    console.log(data);
    document.querySelector("#search-dropdown").innerHTML = JSON.stringify(data);
  }

  async focus() {
    // add div below search bar
    console.log("dropdown");
    let div = document.createElement("div");
    div.id = "search-dropdown";
    div.classList.add("dropdown-menu", "show");
    div.style.position = "absolute";
    div.style.left = "0";
    div.style.top = "calc(100% + 6px)";
    div.style.zIndex = "2000";
    div.style.width = "100%";
    div.style.background = "#fff";
    div.style.color = "#212529";
    div.style.boxShadow = "0 .5rem 1rem rgba(0,0,0,.15)";
    div.style.borderRadius = ".25rem";
    div.style.overflow = "hidden";
    div.textContent = "Searching...";
    this.searchBar.parentNode.appendChild(div);
  }

  async blur() {
    console.log("remove dropdown");
    let dropdown = this.searchBar.parentNode.querySelector(".dropdown-menu");
    if (dropdown) {
      dropdown.remove();
    }
  }
}

search = new SearchBar("desktop-search");

/**
 *
 * @param {BigInt} closetId
 * @param {BigInt} itemId
 */
async function addToCloset(closetId, itemId) {
  let res = await fetch(`/api/closet/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      closet_id: closetId,
      item_id: itemId,
    }),
  });
  if (res.ok) {
    alert("Item added to closet!");
  } else {
    alert("Failed to add item to closet.");
  }
}
