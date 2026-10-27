// Клиент API. Все запросы идут на тот же адрес, что и сама страница
// (/api/...), а nginx-шлюз раскладывает их по сервисам. Поэтому CORS не нужен.
//
// Access-токен хранится только в памяти страницы, refresh-токен — в
// HttpOnly-cookie, которую JavaScript не видит. После перезагрузки страницы
// сессия восстанавливается запросом /auth/refresh (cookie браузер отправит сам).

const API = '/api';

export class ApiError extends Error {
  constructor(status, message, retryAfter = 0) {
    super(message);
    this.status = status;
    this.retryAfter = retryAfter;
  }
}

const session = {
  accessToken: null,
  expiresAt: 0,
  user: null,
};

let refreshInFlight = null;
const listeners = new Set();

/** Подписка на изменение сессии (вход, выход, истечение). */
export function onSessionChange(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

function setSession(data) {
  session.accessToken = data ? data.access_token : null;
  session.expiresAt = data ? Date.parse(data.expires_at) : 0;
  session.user = data ? data.user : null;
  listeners.forEach((fn) => fn(session.user));
}

export const currentUser = () => session.user;

async function toError(res) {
  let message = res.statusText || 'Request failed';
  try {
    const body = await res.json();
    if (body && body.error) message = body.error;
  } catch { /* тело не JSON */ }
  return new ApiError(res.status, message, Number(res.headers.get('Retry-After')) || 0);
}

/**
 * Обновляет пару токенов. Если несколько запросов одновременно получили 401,
 * refresh выполняется один раз: refresh-токен одноразовый, и второй обмен
 * того же токена сервер принял бы за кражу и закрыл бы все сессии.
 */
function refresh() {
  if (!refreshInFlight) {
    const call = () => fetch(`${API}/auth/refresh`, { method: 'POST', credentials: 'same-origin' });
    refreshInFlight = call()
      .then(async (res) => {
        // Если другая вкладка только что обменяла тот же refresh-токен,
        // в cookie уже лежит новый — достаточно повторить запрос.
        if (res.status === 401) {
          const body = await res.clone().json().catch(() => ({}));
          if (String(body.error || '').includes('just rotated')) {
            await new Promise((r) => setTimeout(r, 300));
            return call();
          }
        }
        return res;
      })
      .then(async (res) => {
        if (!res.ok) {
          const wasSignedIn = session.user !== null;
          setSession(null);
          const err = await toError(res);
          err.sessionLost = wasSignedIn;
          throw err;
        }
        setSession(await res.json());
      })
      .finally(() => { refreshInFlight = null; });
  }
  return refreshInFlight;
}

/** Пробует восстановить сессию по cookie. Возвращает пользователя или null. */
export async function restoreSession() {
  try {
    await refresh();
    return session.user;
  } catch {
    return null;
  }
}

async function request(path, { method = 'GET', body, auth = false } = {}) {
  if (auth) {
    if (!session.accessToken) throw new ApiError(401, 'Sign in to continue');
    // Обновляем заранее, если токен истекает в ближайшие 30 секунд.
    if (Date.now() > session.expiresAt - 30_000) await refresh();
  }

  const send = () => fetch(API + path, {
    method,
    credentials: 'same-origin',
    headers: {
      ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
      ...(auth ? { Authorization: `Bearer ${session.accessToken}` } : {}),
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  let res = await send();
  if (res.status === 401 && auth) {
    await refresh(); // если и это не удалось — сессия закончилась
    res = await send();
  }
  if (!res.ok) throw await toError(res);
  return res.status === 204 ? null : res.json();
}

// ---------- Авторизация ----------

export async function signIn(username, password) {
  const res = await fetch(`${API}/auth/sign-in`, {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw await toError(res);
  setSession(await res.json());
  return session.user;
}

export const signUp = (fullName, username, password) =>
  request('/auth/sign-up', { method: 'POST', body: { full_name: fullName, username, password } });

export async function signOut() {
  try {
    await fetch(`${API}/auth/logout`, { method: 'POST', credentials: 'same-origin' });
  } finally {
    setSession(null);
  }
}

// ---------- Каталог ----------

export const listSongs = (afterId = 0, limit = 20) =>
  request(`/songs?after_id=${afterId}&limit=${limit}`);

export const searchSongs = (q, limit = 30) =>
  request(`/songs/search?q=${encodeURIComponent(q)}&limit=${limit}`);

export const songsByIds = (ids) =>
  ids.length ? request(`/songs/batch?ids=${ids.join(',')}`) : Promise.resolve({ items: [] });

// ---------- Лайки ----------

export const getLikes = () => request('/likes', { auth: true });
export const like = (songId) => request(`/likes/${songId}`, { method: 'PUT', auth: true });
export const unlike = (songId) => request(`/likes/${songId}`, { method: 'DELETE', auth: true });

// ---------- Стриминг ----------

/** Билет — подписанная ссылка на трек для <audio src>. */
export const getStreamTicket = (songId) =>
  request('/stream/tickets', { method: 'POST', body: { song_id: songId }, auth: true });

/**
 * Узнаёт, почему <audio> не смог загрузить ссылку: сам элемент код ответа
 * не сообщает. Запрашиваем один байт и смотрим статус.
 */
export async function probeStream(url) {
  try {
    const res = await fetch(url, { headers: { Range: 'bytes=0-0' } });
    return { status: res.status, retryAfter: Number(res.headers.get('Retry-After')) || 0 };
  } catch {
    return { status: 0, retryAfter: 0 };
  }
}
