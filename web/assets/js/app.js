import * as api from "./api.js";

// ---------- Стан сторінки ----------

let allCats = [];           // усі коти з GET /cats
let totalCats = 0;          // загальна кількість котів із заголовка X-Total-Count
let shelters = new Map();   // притулки за id
let shownCats = [];         // коти, які зараз показані: усі або одного притулку
let selected = 0;           // індекс обраного кота в shownCats
let shelterFilter = null;   // id притулку, якщо показуємо лише його котів

// ---------- Елементи сторінки ----------

const byId = (id) => document.getElementById(id);

const catForm = byId("cat-form");
const shelterForm = byId("shelter-form");
const apiKeyInput = byId("api-key");

// ---------- Фото (у API фото немає, тому підбираємо за id) ----------

const CAT_PHOTOS = ["f194866408", "f217417", "f73471816", "f169774582", "f79162873",
  "f138888175", "f1400204", "f117884392", "f11390310"];
const SHELTER_PHOTOS = ["f6739852", "f128755638", "f189448641", "f143501550",
  "f84441832", "f31074768", "f6607296"];

function catPhoto(id, size = 960) {
  return `assets/img/photo-${CAT_PHOTOS[id % CAT_PHOTOS.length]}-${size}.webp`;
}

function shelterPhoto(id) {
  return `assets/img/photo-${SHELTER_PHOTOS[id % SHELTER_PHOTOS.length]}-960.webp`;
}

// ---------- Допоміжні функції ----------

function plural(count, word) {
  return `${count} ${word}${count === 1 ? "" : "s"}`;
}

function formatAge(months) {
  if (months < 12) return plural(months, "month");
  const years = plural(Math.floor(months / 12), "year");
  const rest = months % 12;
  return rest ? `${years} ${plural(rest, "month")}` : years;
}

function showStatus(message, isError = false) {
  const status = byId("status");
  status.textContent = message;
  status.className = isError ? "status is-error" : "status is-ok";
}

// ---------- Завантаження даних ----------

async function loadData() {
  try {
    const [cats, shelterList] = await Promise.all([api.getCats(), api.getShelters()]);
    allCats = cats.data ?? [];
    totalCats = cats.total ?? allCats.length;
    shelters = new Map((shelterList.data ?? []).map((s) => [s.id, s]));
  } catch (error) {
    const message = `Could not load data (${error.message}). Check that the API is running.`;
    byId("cats-caption").textContent = message;
    byId("shelter-list").innerHTML = `<li class="empty">${message}</li>`;
    return;
  }

  renderStats();
  renderShelters();
  renderShelterOptions();

  if (shelterFilter !== null && shelters.has(shelterFilter)) {
    await showShelterCats(shelterFilter);
  } else {
    showAllCats();
  }
}

function renderStats() {
  const waiting = allCats.filter((cat) => !cat.adopted);
  const cities = new Set([...shelters.values()].map((s) => s.city));

  byId("stat-waiting").textContent = waiting.length;
  byId("stat-shelters").textContent = shelters.size;
  byId("stat-cities").textContent = [...cities].join(", ") || "—";
  byId("stat-youngest").textContent = waiting.length
    ? formatAge(Math.min(...waiting.map((cat) => cat.ageMonths)))
    : "—";
  byId("hero-text").textContent =
    `${plural(waiting.length, "cat")} waiting in ${plural(shelters.size, "shelter")} right now.`;
}

// ---------- Коти ----------

function showAllCats() {
  shelterFilter = null;
  const waiting = allCats.filter((cat) => !cat.adopted).length;
  const caption = allCats.length
    ? `${plural(waiting, "cat")} need a home. Pick one to see its age and shelter.`
    : "No cats are listed yet. Add the first one under Manage.";
  setShownCats(allCats, caption, totalCats);
}

async function showShelterCats(shelterId) {
  const shelter = shelters.get(shelterId);
  try {
    const { data, total } = await api.getShelterCats(shelterId);
    shelterFilter = shelterId;
    setShownCats(data, `${plural(data.length, "cat")} at ${shelter.name}, ${shelter.city}.`, total);
    location.hash = "#cats";
  } catch (error) {
    showStatus(`Could not load the cats of ${shelter.name}: ${error.message}`, true);
  }
}

function setShownCats(cats, caption, total) {
  const selectedId = shownCats[selected]?.id;

  // Спочатку коти, яким потрібен дім, потім усиновлені
  shownCats = [...cats].sort((a, b) => a.adopted - b.adopted || a.id - b.id);
  selected = Math.max(0, shownCats.findIndex((cat) => cat.id === selectedId));

  byId("cats-caption").textContent = caption;
  byId("cats-count").textContent = total > cats.length
    ? `Showing ${cats.length} of ${total}`
    : plural(cats.length, "cat");
  byId("show-all-cats").hidden = shelterFilter === null;

  renderCatList();
  renderSelectedCat();
}

function renderCatList() {
  const list = byId("cat-list");
  const template = byId("cat-item-template");
  list.replaceChildren();

  shownCats.forEach((cat, index) => {
    const item = template.content.cloneNode(true);
    item.querySelector("img").src = catPhoto(cat.id, 500);
    item.querySelector(".cat-item__name").textContent = cat.name;
    item.querySelector(".cat-item__info").textContent =
      formatAge(cat.ageMonths) + (cat.adopted ? " · adopted" : "");
    item.querySelector("button").addEventListener("click", () => selectCat(index));
    list.append(item);
  });
}

function selectCat(index) {
  selected = (index + shownCats.length) % shownCats.length;
  renderSelectedCat();
}

function renderSelectedCat() {
  const cat = shownCats[selected];

  for (const id of ["toggle-adopted", "edit-cat", "delete-cat", "prev-cat", "next-cat"]) {
    byId(id).disabled = !cat;
  }

  byId("cat-list").querySelectorAll("button").forEach((button, index) => {
    button.classList.toggle("is-selected", index === selected);
  });

  if (!cat) {
    for (const id of ["cat-name-text", "cat-age-text", "cat-gender-text", "cat-status-text", "cat-shelter-text"]) {
      byId(id).textContent = "—";
    }
    byId("cat-about").textContent = shelterFilter === null ? "No cats yet." : "This shelter has no cats yet.";
    byId("pager-count").textContent = "0 / 0";
    return;
  }

  const shelter = shelters.get(cat.shelterID);
  const shelterName = shelter ? shelter.name : `Shelter #${cat.shelterID}`;
  const pronoun = cat.gender === "female" ? "She" : "He";

  byId("cat-name-text").textContent = cat.name;
  byId("cat-age-text").textContent = formatAge(cat.ageMonths);
  byId("cat-gender-text").textContent = cat.gender === "female" ? "Female" : "Male";
  byId("cat-status-text").textContent = cat.adopted ? "Adopted" : "Needs a home";
  byId("cat-shelter-text").textContent = shelter ? `${shelter.name}, ${shelter.city}` : shelterName;
  byId("cat-photo").src = catPhoto(cat.id);
  byId("cat-photo").alt = cat.name;
  byId("cat-about").textContent = cat.adopted
    ? `${cat.name} already found a home.`
    : `${cat.name} is ${formatAge(cat.ageMonths)} old and lives at ${shelterName}. ${pronoun} is waiting for a family.`;
  byId("toggle-adopted").textContent = cat.adopted ? "Mark as needing a home" : "Mark as adopted";
  byId("pager-count").textContent = `${selected + 1} / ${shownCats.length}`;
}

async function toggleAdopted() {
  const cat = shownCats[selected];
  const changed = { ...cat, adopted: !cat.adopted };
  delete changed.id;

  try {
    const { data } = await api.updateCat(cat.id, changed);
    showStatus(`${data.name} is now ${data.adopted ? "adopted" : "looking for a home"}.`);
    await loadData();
  } catch (error) {
    showStatus(`${cat.name} was not changed: ${error.message}`, true);
    location.hash = "#manage";
  }
}

async function editCat() {
  const cat = shownCats[selected];
  try {
    const { data } = await api.getCat(cat.id);
    const fields = catForm.elements;
    fields.id.value = data.id;
    fields.name.value = data.name;
    fields.ageMonths.value = data.ageMonths;
    fields.gender.value = data.gender;
    fields.shelterID.value = data.shelterID;
    fields.adopted.checked = data.adopted;
    setCatFormMode(data.id);
    location.hash = "#manage";
  } catch (error) {
    showStatus(`Could not open ${cat.name}: ${error.message}`, true);
  }
}

async function deleteSelectedCat() {
  const cat = shownCats[selected];
  if (!confirm(`Delete ${cat.name}? This cannot be undone.`)) return;

  try {
    await api.deleteCat(cat.id);
    showStatus(`Deleted ${cat.name}.`);
    await loadData();
  } catch (error) {
    showStatus(`${cat.name} was not deleted: ${error.message}`, true);
    location.hash = "#manage";
  }
}

// ---------- Притулки ----------

function renderShelters() {
  const list = byId("shelter-list");
  const template = byId("shelter-template");
  list.replaceChildren();

  if (shelters.size === 0) {
    list.innerHTML = `<li class="empty">No shelters yet. Add the first one under Manage.</li>`;
    return;
  }

  for (const shelter of shelters.values()) {
    const catCount = allCats.filter((cat) => cat.shelterID === shelter.id).length;
    const percent = Math.min(100, Math.round((catCount / shelter.capacity) * 100));

    const card = template.content.cloneNode(true);
    card.querySelector("li").dataset.id = shelter.id;
    card.querySelector("img").src = shelterPhoto(shelter.id);
    card.querySelector(".shelter__name").textContent = shelter.name;
    card.querySelector(".shelter__where").textContent = `${shelter.address}, ${shelter.city}`;
    card.querySelector(".meter__cats").textContent = plural(catCount, "cat");
    card.querySelector(".meter__places").textContent = plural(shelter.capacity, "place");
    card.querySelector(".meter__bar span").style.width = `${percent}%`;
    list.append(card);
  }
}

// Список притулків у формі кота
function renderShelterOptions() {
  const select = catForm.elements.shelterID;
  select.replaceChildren();

  for (const shelter of shelters.values()) {
    select.append(new Option(`${shelter.name}, ${shelter.city}`, shelter.id));
  }
  if (shelters.size === 0) {
    select.append(new Option("Add a shelter first", ""));
  }
}

async function editShelter(id) {
  try {
    const { data } = await api.getShelter(id);
    const fields = shelterForm.elements;
    fields.id.value = data.id;
    fields.name.value = data.name;
    fields.city.value = data.city;
    fields.address.value = data.address;
    fields.capacity.value = data.capacity;
    setShelterFormMode(data.id);
    location.hash = "#manage";
  } catch (error) {
    showStatus(`Could not open the shelter: ${error.message}`, true);
  }
}

async function deleteShelterById(id) {
  const shelter = shelters.get(id);
  if (!confirm(`Delete ${shelter.name}? A shelter that still has cats cannot be deleted.`)) return;

  try {
    await api.deleteShelter(id);
    showStatus(`Deleted ${shelter.name}.`);
    if (shelterFilter === id) shelterFilter = null;
    await loadData();
  } catch (error) {
    showStatus(`${shelter.name} was not deleted: ${error.message}`, true);
    location.hash = "#manage";
  }
}

// ---------- Форми ----------

function setCatFormMode(editingId) {
  byId("cat-form-title").textContent = editingId ? `Edit cat #${editingId}` : "Add a cat";
  byId("cat-submit").textContent = editingId ? "Save changes" : "Add the cat";
}

function setShelterFormMode(editingId) {
  byId("shelter-form-title").textContent = editingId ? `Edit shelter #${editingId}` : "Add a shelter";
  byId("shelter-submit").textContent = editingId ? "Save changes" : "Add the shelter";
}

// Форма не надсилається, доки браузер не перевірить атрибути required/min/max.
// Після перевірки дані відправляються через Fetch API без перезавантаження сторінки.
catForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const fields = catForm.elements;
  const id = fields.id.value;
  const cat = {
    name: fields.name.value.trim(),
    ageMonths: Number(fields.ageMonths.value),
    gender: fields.gender.value,
    shelterID: Number(fields.shelterID.value),
    adopted: fields.adopted.checked,
  };

  try {
    const { data } = id ? await api.updateCat(id, cat) : await api.createCat(cat);
    showStatus(id ? `Saved ${data.name}.` : `Added ${data.name} (cat #${data.id}).`);
    catForm.reset();
    await loadData();
    selectCat(Math.max(0, shownCats.findIndex((c) => c.id === data.id)));
  } catch (error) {
    showStatus(`The cat was not saved: ${error.message}`, true);
  }
});

shelterForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const fields = shelterForm.elements;
  const id = fields.id.value;
  const shelter = {
    name: fields.name.value.trim(),
    city: fields.city.value.trim(),
    address: fields.address.value.trim(),
    capacity: Number(fields.capacity.value),
  };

  try {
    const { data } = id ? await api.updateShelter(id, shelter) : await api.createShelter(shelter);
    showStatus(id ? `Saved ${data.name}.` : `Added ${data.name} (shelter #${data.id}).`);
    shelterForm.reset();
    await loadData();
  } catch (error) {
    showStatus(`The shelter was not saved: ${error.message}`, true);
  }
});

// Кнопка Clear: форма знову працює в режимі додавання
catForm.addEventListener("reset", () => {
  catForm.elements.id.value = "";
  setCatFormMode(null);
});

shelterForm.addEventListener("reset", () => {
  shelterForm.elements.id.value = "";
  setShelterFormMode(null);
});

// ---------- Обробники кнопок ----------

byId("prev-cat").addEventListener("click", () => selectCat(selected - 1));
byId("next-cat").addEventListener("click", () => selectCat(selected + 1));
byId("toggle-adopted").addEventListener("click", toggleAdopted);
byId("edit-cat").addEventListener("click", editCat);
byId("delete-cat").addEventListener("click", deleteSelectedCat);
byId("show-all-cats").addEventListener("click", showAllCats);

byId("add-cat").addEventListener("click", () => {
  catForm.reset();
  location.hash = "#manage";
  byId("cat-name").focus({ preventScroll: true });
});

byId("add-shelter").addEventListener("click", () => {
  shelterForm.reset();
  location.hash = "#manage";
  byId("sh-name").focus({ preventScroll: true });
});

// Картки притулків створюються динамічно, тому слухаємо кліки на всьому списку
byId("shelter-list").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-action]");
  if (!button) return;

  const id = Number(button.closest("li").dataset.id);
  if (button.dataset.action === "show-cats") showShelterCats(id);
  if (button.dataset.action === "edit") editShelter(id);
  if (button.dataset.action === "delete") deleteShelterById(id);
});

byId("delete-all-cats").addEventListener("click", async () => {
  if (!confirm(`Delete all ${plural(allCats.length, "cat")}? This cannot be undone.`)) return;
  try {
    await api.deleteAllCats();
    showStatus("Deleted every cat.");
    await loadData();
  } catch (error) {
    showStatus(`The cats were not deleted: ${error.message}`, true);
  }
});

byId("delete-all-shelters").addEventListener("click", async () => {
  if (!confirm(`Delete all ${plural(shelters.size, "shelter")}? Shelters with cats cannot be deleted.`)) return;
  try {
    await api.deleteAllShelters();
    showStatus("Deleted every shelter.");
    shelterFilter = null;
    await loadData();
  } catch (error) {
    showStatus(`The shelters were not deleted: ${error.message}`, true);
  }
});

// Ключ API зберігається лише до закриття вкладки
apiKeyInput.value = sessionStorage.getItem("api-key") ?? "";
api.setApiKey(apiKeyInput.value);
apiKeyInput.addEventListener("input", () => {
  api.setApiKey(apiKeyInput.value);
  sessionStorage.setItem("api-key", apiKeyInput.value);
});

loadData();
