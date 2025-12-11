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
