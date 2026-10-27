// Главная страница: каталог, поиск, лайки и плеер.
import * as api from './api.js';
import { Player, RepeatMode } from './player.js';
import { icons, eq } from './icons.js';
import {
  $, escapeHtml, hydrateIcons, formatTime, coverHTML, coverStyle, coverInitials, toast, debounce, storage,
} from './ui.js';

const PAGE_SIZE = 20;

const state = {
  route: 'home',     // 'home' | 'liked'
  query: '',
  songs: [],         // что сейчас показано
  nextAfterId: null, // курсор следующей страницы каталога
  likedIds: [],      // ID лайкнутых песен, новые первыми
  user: null,
  loadSeq: 0,        // защищает от гонок, когда ответы приходят не по порядку
};

const player = new Player();
const el = {
  tracks: $('#tracks'),
  empty: $('#empty'),
  more: $('#more'),
  title: $('#view-title'),
  meta: $('#view-meta'),
  search: $('#search'),
  greeting: $('#greeting'),
  likedCount: $('#liked-count'),
  avatar: $('#avatar'),
  profileName: $('#profile-name'),
  profileSub: $('#profile-sub'),
  signOut: $('#sign-out'),
  player: $('#player'),
  nowCover: $('#now-cover'),
  nowTitle: $('#now-title'),
  nowArtist: $('#now-artist'),
  nowLike: $('#now-like'),
  play: $('#play'),
  prev: $('#prev'),
  next: $('#next'),
  shuffle: $('#shuffle'),
  repeat: $('#repeat'),
  seek: $('#seek'),
  timeCurrent: $('#time-current'),
  timeTotal: $('#time-total'),
  volume: $('#volume'),
  mute: $('#mute'),
};

// ---------- Отрисовка списка ----------

const isLiked = (id) => state.likedIds.includes(id);

function trackHTML(song, index) {
  const liked = isLiked(song.id);
  return `
    <li class="track" data-id="${song.id}">
      <button class="track-play" type="button" aria-label="Play ${escapeHtml(song.title)}">
        <span class="track-num">${index + 1}</span><span class="when-paused">${icons.play}</span><span class="when-playing">${icons.pause}</span>${eq}
      </button>
      ${coverHTML(song)}
      <div class="track-main">
        <div class="track-title">${escapeHtml(song.title)}</div>
        <div class="track-artist">${escapeHtml(song.artist)}</div>
      </div>
      <div class="track-album">${escapeHtml(song.album)}</div>
      <button class="like-button" type="button" aria-pressed="${liked}" aria-label="${liked ? 'Remove from liked' : 'Like'}">
        ${liked ? icons.heartFilled : icons.heart}
      </button>
      <span class="track-duration">${formatTime(song.duration_sec)}</span>
    </li>`;
}

function renderList() {
  el.tracks.innerHTML = state.songs.map(trackHTML).join('');
  el.more.hidden = !(state.route === 'home' && !state.query && state.nextAfterId);
  markCurrent();

  if (state.songs.length) {
    el.empty.hidden = true;
    return;
  }
  el.empty.hidden = false;
  if (state.query) {
    el.empty.innerHTML = `${icons.search}<h3>Nothing found</h3><p>No songs, artists or albums match “${escapeHtml(state.query)}”.</p>`;
  } else if (state.route === 'liked' && !state.user) {
    el.empty.innerHTML = `${icons.heart}<h3>Your liked songs live here</h3><p>Sign in to save the tracks you love.</p><a class="button" href="sign-in.html">Sign in</a>`;
  } else if (state.route === 'liked') {
    el.empty.innerHTML = `${icons.heart}<h3>No liked songs yet</h3><p>Tap the heart next to a track and it will appear here.</p><a class="button secondary" href="#/">Browse tracks</a>`;
  } else {
    el.empty.innerHTML = `${icons.music}<h3>The catalog is empty</h3><p>Run the catalog migrations to add demo tracks.</p>`;
  }
}

function renderSkeleton() {
  el.empty.hidden = true;
  el.more.hidden = true;
  el.tracks.innerHTML = Array.from({ length: 5 }, () => '<li class="skeleton"></li>').join('');
}

function renderHeader() {
  const titles = { home: 'All tracks', liked: 'Liked songs' };
  el.title.textContent = state.query ? `Results for “${state.query}”` : titles[state.route];
  const n = state.songs.length;
  el.meta.textContent = n ? `${n}${state.nextAfterId && !state.query && state.route === 'home' ? '+' : ''} track${n === 1 ? '' : 's'}` : '';

  document.querySelectorAll('.nav-link').forEach((a) => {
    a.toggleAttribute('aria-current', !state.query && a.dataset.route === state.route);
    if (a.hasAttribute('aria-current')) a.setAttribute('aria-current', 'page');
  });
  el.likedCount.textContent = state.user && state.likedIds.length ? state.likedIds.length : '';
}

function renderUser() {
  const u = state.user;
  if (u) {
    const first = (u.full_name || u.username).split(/\s+/)[0];
    el.greeting.textContent = `Welcome back, ${first}`;
    el.profileName.textContent = u.full_name || u.username;
    el.profileSub.textContent = `@${u.username}`;
    el.avatar.className = 'avatar';
    el.avatar.textContent = (u.full_name || u.username).trim()[0].toUpperCase();
    el.signOut.hidden = false;
  } else {
    el.greeting.textContent = 'Welcome to Neon Player';
    el.profileName.textContent = 'Guest';
    el.profileSub.innerHTML = '<a href="sign-in.html">Sign in</a> to listen';
    el.avatar.className = 'avatar guest';
    el.avatar.innerHTML = icons.user;
    el.signOut.hidden = true;
  }
}

/** Подсвечивает играющий трек без перерисовки списка. */
function markCurrent() {
  const current = player.current;
  el.tracks.querySelectorAll('.track').forEach((row) => {
    const isCurrent = current && Number(row.dataset.id) === current.id;
    row.classList.toggle('is-current', Boolean(isCurrent));
    row.classList.toggle('is-playing', Boolean(isCurrent && player.isPlaying));
    const btn = row.querySelector('.track-play');
    if (btn) btn.setAttribute('aria-label', `${isCurrent && player.isPlaying ? 'Pause' : 'Play'} ${row.querySelector('.track-title').textContent}`);
  });
}

function updateLikeButtons(songId) {
  const liked = isLiked(songId);
  el.tracks.querySelectorAll(`.track[data-id="${songId}"] .like-button`).forEach((b) => {
    b.setAttribute('aria-pressed', String(liked));
    b.setAttribute('aria-label', liked ? 'Remove from liked' : 'Like');
    b.innerHTML = liked ? icons.heartFilled : icons.heart;
  });
  renderNowLike();
  renderHeader();
}

// ---------- Загрузка данных ----------

async function loadView({ append = false } = {}) {
  const seq = ++state.loadSeq;
  if (!append) renderSkeleton();
  try {
    let items;
    let next = null;
    if (state.query) {
      ({ items } = await api.searchSongs(state.query));
    } else if (state.route === 'liked') {
      items = state.user ? (await api.songsByIds(state.likedIds)).items : [];
    } else {
      const res = await api.listSongs(append ? state.nextAfterId : 0, PAGE_SIZE);
      items = res.items;
      next = res.next_after_id;
    }
    if (seq !== state.loadSeq) return; // пока ждали, пользователь ушёл в другой раздел
    state.songs = append ? state.songs.concat(items) : items;
    state.nextAfterId = next;
  } catch (err) {
    if (seq !== state.loadSeq) return;
    toast(`Could not load tracks: ${err.message}`, 'error');
    if (!append) state.songs = [];
  }
  renderHeader();
  renderList();
}

async function loadLikes() {
  if (!state.user) {
    state.likedIds = [];
    return;
  }
  try {
    state.likedIds = (await api.getLikes()).song_ids;
  } catch (err) {
    toast(`Could not load likes: ${err.message}`, 'error');
  }
}

// ---------- Действия ----------

function requireSignIn(action) {
  toast(`<a href="sign-in.html">Sign in</a> to ${action}.`, 'info', { html: true });
}

async function toggleLike(songId) {
  if (!state.user) return requireSignIn('save tracks you like');
  const wasLiked = isLiked(songId);
  // Оптимистичное обновление: сразу меняем интерфейс, откатываем при ошибке.
  state.likedIds = wasLiked ? state.likedIds.filter((id) => id !== songId) : [songId, ...state.likedIds];
  updateLikeButtons(songId);
  try {
    await (wasLiked ? api.unlike(songId) : api.like(songId));
    if (state.route === 'liked' && !state.query && wasLiked) {
      state.songs = state.songs.filter((s) => s.id !== songId);
      renderHeader();
      renderList();
    }
  } catch (err) {
    state.likedIds = wasLiked ? [songId, ...state.likedIds] : state.likedIds.filter((id) => id !== songId);
    updateLikeButtons(songId);
    toast(`Could not update like: ${err.message}`, 'error');
  }
}

function playFromList(songId) {
  if (!state.user) return requireSignIn('listen');
  const song = state.songs.find((s) => s.id === songId);
  if (!song) return;
  if (player.current && player.current.id === songId) {
    player.toggle();
    return;
  }
  player.playSong(song, state.songs);
}

// ---------- Плеер ----------

function renderNowLike() {
  const song = player.current;
  el.nowLike.hidden = !song || !state.user;
  if (!song) return;
  const liked = isLiked(song.id);
  el.nowLike.setAttribute('aria-pressed', String(liked));
  el.nowLike.setAttribute('aria-label', liked ? 'Remove from liked' : 'Like');
  el.nowLike.innerHTML = liked ? icons.heartFilled : icons.heart;
}

function renderPlayer() {
  const song = player.current;
  el.player.classList.toggle('is-idle', !song);
  if (song) {
    el.nowTitle.textContent = song.title;
    el.nowArtist.textContent = `${song.artist} · ${song.album}`;
    el.nowCover.setAttribute('style', coverStyle(song));
    el.nowCover.textContent = coverInitials(song);
    document.title = `${player.isPlaying ? '▶ ' : ''}${song.title} — ${song.artist}`;
  }
  const playing = player.isPlaying;
  el.play.disabled = !song;
  el.play.classList.toggle('is-loading', player.loading && Boolean(song));
  el.play.innerHTML = player.loading && song ? icons.loader : (playing ? icons.pause : icons.play);
  el.play.setAttribute('aria-label', playing ? 'Pause' : 'Play');
  el.prev.disabled = !song;
  el.next.disabled = !song;
  el.seek.disabled = !song;
  el.shuffle.setAttribute('aria-pressed', String(player.shuffle));
  el.repeat.setAttribute('aria-pressed', String(player.repeat !== RepeatMode.OFF));
  el.repeat.dataset.mode = player.repeat;
  el.repeat.setAttribute('aria-label', `Repeat: ${player.repeat}`);
  renderNowLike();
  markCurrent();
}

let seeking = false;

function renderProgress() {
  const a = player.audio;
  const duration = Number.isFinite(a.duration) ? a.duration : (player.current?.duration_sec || 0);
  if (!seeking) {
    el.seek.max = String(duration || 0);
    el.seek.value = String(a.currentTime || 0);
    el.timeCurrent.textContent = formatTime(a.currentTime);
  }
  el.timeTotal.textContent = formatTime(duration);
  el.seek.style.setProperty('--played', `${duration ? ((seeking ? Number(el.seek.value) : a.currentTime) / duration) * 100 : 0}%`);

  // Сколько уже загружено: видно, что файл приходит частями, а не целиком.
  let buffered = 0;
  for (let i = 0; i < a.buffered.length; i++) {
    if (a.buffered.start(i) <= a.currentTime + 0.5) buffered = Math.max(buffered, a.buffered.end(i));
  }
  el.seek.style.setProperty('--buffered', `${duration ? (buffered / duration) * 100 : 0}%`);
}

function renderVolume() {
  const a = player.audio;
  const v = a.muted ? 0 : a.volume;
  el.volume.value = String(v);
  el.volume.style.setProperty('--played', `${v * 100}%`);
  el.mute.innerHTML = v === 0 ? icons.mute : icons.volume;
  el.mute.setAttribute('aria-label', v === 0 ? 'Unmute' : 'Mute');
}

function bindPlayer() {
  player.addEventListener('change', renderPlayer);
  player.addEventListener('track', () => { renderPlayer(); renderProgress(); });
  player.addEventListener('progress', renderProgress);
  player.addEventListener('volume', () => {
    renderVolume();
    storage.set('volume', String(player.audio.volume));
  });
  player.addEventListener('busy', (e) => toast(`Too many listeners right now. Retrying in ${e.detail.wait} s…`));
  player.addEventListener('failure', (e) => {
    const err = e.detail;
    if (err.status === 401) return requireSignIn('listen');
    if (err.name === 'NotAllowedError') return; // браузер заблокировал автозапуск — пользователь нажмёт play
    toast(err.message || 'Playback failed', 'error');
  });

  el.play.addEventListener('click', () => player.toggle());
  el.prev.addEventListener('click', () => player.prev());
  el.next.addEventListener('click', () => player.next());
  el.shuffle.addEventListener('click', () => player.toggleShuffle());
  el.repeat.addEventListener('click', () => player.cycleRepeat());
  el.mute.addEventListener('click', () => player.toggleMute());
  el.volume.addEventListener('input', () => player.setVolume(Number(el.volume.value)));
  el.nowLike.addEventListener('click', () => player.current && toggleLike(player.current.id));

  el.seek.addEventListener('input', () => {
    seeking = true;
    el.timeCurrent.textContent = formatTime(Number(el.seek.value));
    renderProgress();
  });
  el.seek.addEventListener('change', () => {
    player.seek(Number(el.seek.value));
    seeking = false;
  });

  const savedVolume = Number(storage.get('volume'));
  player.setVolume(Number.isFinite(savedVolume) && storage.get('volume') !== null ? savedVolume : 0.8);
  renderVolume();
  renderPlayer();

  document.addEventListener('keydown', (e) => {
    const typing = e.target.closest('input, textarea, [contenteditable]') && e.target.type !== 'range';
    if (typing || e.metaKey || e.ctrlKey || e.altKey) return;
    if (e.code === 'Space' && player.current) {
      e.preventDefault();
      player.toggle();
    } else if (e.code === 'ArrowRight' && player.current && e.target.type !== 'range') {
      player.seek(player.audio.currentTime + 5);
    } else if (e.code === 'ArrowLeft' && player.current && e.target.type !== 'range') {
      player.seek(player.audio.currentTime - 5);
    }
  });
}

// ---------- Навигация и события списка ----------

function routeFromHash() {
  return location.hash === '#/liked' ? 'liked' : 'home';
}

function bindUI() {
  el.tracks.addEventListener('click', (e) => {
    const row = e.target.closest('.track');
    if (!row) return;
    const id = Number(row.dataset.id);
    if (e.target.closest('.like-button')) toggleLike(id);
    // Кнопка play, обложка или название: на телефоне отдельной кнопки нет.
    else if (e.target.closest('.track-play, .cover, .track-main')) playFromList(id);
  });

  el.more.addEventListener('click', async () => {
    el.more.disabled = true;
    await loadView({ append: true });
    el.more.disabled = false;
  });

  const runSearch = debounce(() => {
    state.query = el.search.value.trim();
    loadView();
  }, 300);
  el.search.addEventListener('input', runSearch);
  el.search.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      el.search.value = '';
      state.query = '';
      loadView();
    }
  });

  window.addEventListener('hashchange', () => {
    state.route = routeFromHash();
    state.query = '';
    el.search.value = '';
    loadView();
  });

  el.signOut.addEventListener('click', async () => {
    player.audio.pause();
    await api.signOut();
    toast('Signed out', 'success');
  });

  api.onSessionChange((user) => {
    const lostSession = state.user && !user;
    state.user = user;
    renderUser();
    if (lostSession) {
      player.audio.pause();
      state.likedIds = [];
      renderHeader();
      if (state.route === 'liked') loadView();
      else renderList();
      renderPlayer();
    }
  });
}

async function init() {
  hydrateIcons();
  bindPlayer();
  bindUI();
  state.route = routeFromHash();

  state.user = await api.restoreSession();
  renderUser();
  await loadLikes();
  await loadView();
}

init();
