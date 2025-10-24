/**
 * @param {HTMLInputElement} inputElement
 */
async function searchBarOninput(inputElement) {
  let input = inputElement.value.trim();
  if (input === "") {
    return;
  }
  let res = await fetch(`/api/search_bar?input=${encodeURIComponent(input)}`);
  let data = await res.json();

  searchResultsDiv = document.querySelector("#search-results");
  searchResultsDiv.innerHTML = JSON.stringify(data);
  for (const key in data) {
    let div = document.createElement("div");
    title = document.createElement("h3");
    title.textContent = key;
    searchResultsDiv.appendChild(title);
  }
}

async function searchBarOnfocus() {
  let inputElement = document.querySelector("#search-results");
  inputElement.innerHTML = "<div>search...</div>";
  inputElement.removeAttribute("hidden");
}

async function searchBarOnblur() {
  let inputElement = document.querySelector("#search-results");
  inputElement.setAttribute("hidden", "true");
}

document.addEventListener("DOMContentLoaded", function () {
  const asideImages = document.querySelectorAll(".image-viewer .aside img");
  const selectedImage = document.querySelector(".image-viewer .selected img");
  asideImages.forEach((img) => {
    img.addEventListener("mouseover", function () {
      selectedImage.src = this.src;
      selectedImage.alt = this.alt;
    });
  });
});
