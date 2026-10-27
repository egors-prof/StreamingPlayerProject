// Плеер на стандартном <audio>. Браузер сам запрашивает файл кусками через
// HTTP Range: воспроизведение начинается сразу, а при перемотке загружается
// только нужный участок. Раньше файл целиком качался по WebSocket и
// декодировался только после полной загрузки.

import { getStreamTicket, probeStream, ApiError } from './api.js';

export const RepeatMode = { OFF: 'off', ALL: 'all', ONE: 'one' };

export class Player extends EventTarget {
  constructor() {
    super();
    this.audio = new Audio();
    this.audio.preload = 'auto';
    this.queue = [];      // песни в порядке воспроизведения
    this.order = [];      // индексы queue с учётом shuffle
    this.position = -1;   // позиция в order
    this.shuffle = false;
    this.repeat = RepeatMode.OFF;
    this.loading = false;
    this.requestId = 0;      // защищает от гонок при быстрой смене треков
    this.loadedSongId = null;
    this.busyRetries = 0;    // попытки после 503 (лимит стримов)
    this.ticketRetried = false;

    const a = this.audio;
    a.addEventListener('play', () => this.emit());
    a.addEventListener('pause', () => this.emit());
    a.addEventListener('playing', () => {
      this.loading = false;
      this.busyRetries = 0;
      this.ticketRetried = false;
      this.emit();
    });
    a.addEventListener('waiting', () => { this.loading = true; this.emit(); });
    a.addEventListener('canplay', () => { this.loading = false; this.emit(); });
    a.addEventListener('timeupdate', () => this.emit('progress'));
    a.addEventListener('progress', () => this.emit('progress'));
    a.addEventListener('durationchange', () => this.emit('progress'));
    a.addEventListener('volumechange', () => this.emit('volume'));
    a.addEventListener('ended', () => this.handleEnded());
    a.addEventListener('error', () => this.recover());

    this.setupMediaSession();
  }

  emit(type = 'change') {
    this.dispatchEvent(new Event(type));
  }

  get current() {
    const i = this.order[this.position];
    return i === undefined ? null : this.queue[i];
  }

  get isPlaying() {
    return !this.audio.paused && !this.audio.ended;
  }

  /** Запускает песню; queue — список, из которого её выбрали (для next/prev). */
  async playSong(song, queue = [song]) {
    this.queue = queue.slice();
    const startIndex = Math.max(0, this.queue.findIndex((s) => s.id === song.id));
    this.rebuildOrder(startIndex);
    await this.load(0);
  }

  async toggle() {
    if (!this.current) return;
    if (this.isPlaying) {
      this.audio.pause();
    } else {
      try { await this.audio.play(); } catch { /* браузер может запретить автозапуск */ }
    }
  }

  async next(auto = false) {
    if (!this.current) return;
    if (this.position < this.order.length - 1) {
      this.position += 1;
    } else if (this.repeat === RepeatMode.ALL || !auto) {
      if (this.shuffle) this.rebuildOrder(this.order[this.position], true);
      this.position = 0;
    } else {
      this.audio.pause();
      this.audio.currentTime = 0;
      return;
    }
    await this.load(0);
  }

  async prev() {
    if (!this.current) return;
    // Как в обычных плеерах: если трек играет дольше 3 секунд — в начало.
    if (this.audio.currentTime > 3 || this.position === 0) {
      this.audio.currentTime = 0;
      return;
    }
    this.position -= 1;
    await this.load(0);
  }

  seek(seconds) {
    if (!this.current || !Number.isFinite(this.audio.duration)) return;
    this.audio.currentTime = Math.max(0, Math.min(seconds, this.audio.duration));
  }

  setVolume(v) {
    this.audio.volume = Math.max(0, Math.min(1, v));
    this.audio.muted = this.audio.volume === 0;
  }

  toggleMute() {
    this.audio.muted = !this.audio.muted;
  }

  toggleShuffle() {
    this.shuffle = !this.shuffle;
    if (this.current) this.rebuildOrder(this.order[this.position]);
    this.emit();
  }

  cycleRepeat() {
    const modes = [RepeatMode.OFF, RepeatMode.ALL, RepeatMode.ONE];
    this.repeat = modes[(modes.indexOf(this.repeat) + 1) % modes.length];
    this.emit();
  }

  /** Обновляет данные песни (например, после лайка) без перезапуска. */
  replaceQueueItem(song) {
    const i = this.queue.findIndex((s) => s.id === song.id);
    if (i >= 0) this.queue[i] = song;
  }

  // ---------- внутреннее ----------

  rebuildOrder(currentIndex, excludeCurrentFirst = false) {
    const indices = this.queue.map((_, i) => i);
    if (!this.shuffle) {
      this.order = indices;
      this.position = currentIndex;
      return;
    }
    // Тасование Фишера–Йетса; текущая песня остаётся первой.
    const rest = indices.filter((i) => i !== currentIndex);
    for (let i = rest.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [rest[i], rest[j]] = [rest[j], rest[i]];
    }
    this.order = excludeCurrentFirst ? rest.concat(currentIndex) : [currentIndex, ...rest];
    this.position = 0;
  }

  async load(startAt) {
    const song = this.current;
    if (!song) return;
    const id = ++this.requestId;
    if (song.id !== this.loadedSongId) { // новый трек — счётчики попыток с нуля
      this.loadedSongId = song.id;
      this.busyRetries = 0;
      this.ticketRetried = false;
    }
    this.loading = true;
    this.emit();
    this.emit('track');

    try {
      const ticket = await getStreamTicket(song.id);
      if (id !== this.requestId) return; // пока ждали, выбрали другой трек
      this.audio.src = ticket.url;
      if (startAt > 0) {
        this.audio.addEventListener('loadedmetadata', () => { this.audio.currentTime = startAt; }, { once: true });
      }
      await this.audio.play();
    } catch (err) {
      if (id !== this.requestId) return;
      this.loading = false;
      this.emit();
      if (err instanceof ApiError || err.name === 'NotAllowedError') {
        this.dispatchEvent(new CustomEvent('failure', { detail: err }));
      }
      // Ошибки загрузки самого файла обработает recover() по событию error.
    }
  }

  /**
   * Восстановление после ошибки загрузки. <audio> не сообщает код ответа,
   * поэтому спрашиваем сервер сами:
   *  401 — истёк билет (пауза дольше часа): берём новый и продолжаем с того же места;
   *  503 — сервер упёрся в лимит одновременных стримов: ждём и пробуем снова.
   */
  async recover() {
    const song = this.current;
    if (!song || !this.audio.src) return;
    const at = this.audio.currentTime || 0;
    const { status, retryAfter } = await probeStream(this.audio.src);
    if (this.current !== song) return;

    if (status === 503 && this.busyRetries < 5) {
      this.busyRetries += 1;
      const wait = Math.max(retryAfter, 2);
      this.dispatchEvent(new CustomEvent('busy', { detail: { wait } }));
      setTimeout(() => { if (this.current === song) this.load(at); }, wait * 1000);
      return;
    }
    if (status === 401 && !this.ticketRetried) {
      this.ticketRetried = true; // одна попытка с новым билетом
      await this.load(at);
      return;
    }
    this.loading = false;
    this.emit();
    this.dispatchEvent(new CustomEvent('failure', { detail: new Error('Could not play this track') }));
  }

  handleEnded() {
    if (this.repeat === RepeatMode.ONE) {
      this.audio.currentTime = 0;
      this.audio.play();
      return;
    }
    this.next(true);
  }

  /** Управление с клавиатуры, гарнитуры и экрана блокировки. */
  setupMediaSession() {
    if (!('mediaSession' in navigator)) return;
    const ms = navigator.mediaSession;
    ms.setActionHandler('play', () => this.toggle());
    ms.setActionHandler('pause', () => this.toggle());
    ms.setActionHandler('previoustrack', () => this.prev());
    ms.setActionHandler('nexttrack', () => this.next());
    try {
      ms.setActionHandler('seekto', (e) => this.seek(e.seekTime));
    } catch { /* не везде поддерживается */ }
    this.addEventListener('track', () => {
      const s = this.current;
      if (s && 'MediaMetadata' in window) {
        ms.metadata = new MediaMetadata({ title: s.title, artist: s.artist, album: s.album });
      }
    });
  }
}
