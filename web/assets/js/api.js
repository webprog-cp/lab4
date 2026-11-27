// Клієнт REST API. Кожна функція відповідає одному ендпоінту сервера.
// Усі запити виконуються через Fetch API у функції request().

const API_BASE = "/api/v1";

// Ключ API для змінюючих запитів (POST, PUT, DELETE). Задається з форми на сторінці.
let apiKey = "";

export function setApiKey(key) {
  apiKey = key;
}

async function request(method, path, body) {
  const options = { method, headers: {} };

  if (method !== "GET") {
    options.headers["X-API-Key"] = apiKey;
  }
  if (body !== undefined) {
    options.headers["Content-Type"] = "application/json";
    options.body = JSON.stringify(body);
  }

  const response = await fetch(API_BASE + path, options);

  // Тіло може бути порожнім (204 No Content), тому спочатку читаємо його як текст.
  const text = await response.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = null;
    }
  }

  if (!response.ok) {
    // Сервер повертає помилки у форматі { "error": "..." }
    throw new Error(data?.error ?? `HTTP ${response.status}`);
  }

  // X-Total-Count — загальна кількість записів, навіть якщо сервер обрізав список (maxListSize).
  const total = Number(response.headers.get("X-Total-Count")) || null;
  return { data, total };
}

// ---------- Коти ----------

export const getCats = () => request("GET", "/cats");
export const getCat = (id) => request("GET", `/cats/${id}`);
export const createCat = (cat) => request("POST", "/cats", cat);
export const updateCat = (id, cat) => request("PUT", `/cats/${id}`, cat);
export const deleteCat = (id) => request("DELETE", `/cats/${id}`);
export const deleteAllCats = () => request("DELETE", "/cats");

// ---------- Притулки ----------

export const getShelters = () => request("GET", "/shelters");
export const getShelter = (id) => request("GET", `/shelters/${id}`);
export const getShelterCats = (id) => request("GET", `/shelters/${id}/cats`);
export const createShelter = (shelter) => request("POST", "/shelters", shelter);
export const updateShelter = (id, shelter) => request("PUT", `/shelters/${id}`, shelter);
export const deleteShelter = (id) => request("DELETE", `/shelters/${id}`);
export const deleteAllShelters = () => request("DELETE", "/shelters");
