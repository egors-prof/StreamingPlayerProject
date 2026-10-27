// Мелкие общие помощники интерфейса.
import { icons } from './icons.js';

export const $ = (sel, root = document) => root.querySelector(sel);

export function escapeHtml(value) {
  return String(value ?? '').replace(/[&<>"']/g, (c) => (
    { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
  ));
}

/** Подставляет SVG вместо <span data-icon="name">. */
export function hydrateIcons(root = document) {
  root.querySelectorAll('[data-icon]').forEach((el) => {
    const icon = icons[el.dataset.icon];
    if (icon) el.outerHTML = icon;
  });
}

export function formatTime(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return '0:00';
  const s = Math.floor(seconds);
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
}

function hash(str) {
  let h = 0;
  for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) | 0;
  return Math.abs(h);
}

/** Обложка: картинка, если есть, иначе градиент с инициалами альбома. */
export function coverHTML(song) {
  if (song.cover_url) {
    return `<div class="cover"><img src="${escapeHtml(song.cover_url)}" alt="" loading="lazy"></div>`;
  }
  return `<div class="cover" style="${coverStyle(song)}">${escapeHtml(coverInitials(song))}</div>`;
}

export function coverStyle(song) {
  const h = hash(`${song.album}|${song.artist}`) % 360;
  return `background: linear-gradient(135deg, hsl(${h} 85% 58%), hsl(${(h + 70) % 360} 75% 42%))`;
}

export const coverInitials = (song) => initials(song.album);

function initials(text = '') {
  const words = text.trim().split(/\s+/).filter(Boolean);
  return (words.length > 1 ? words[0][0] + words[1][0] : (words[0] || '♪').slice(0, 2)).toUpperCase();
}

export function toast(message, kind = 'info', { html = false, timeout = 4000 } = {}) {
  const box = document.getElementById('toasts');
  if (!box) return;
  const el = document.createElement('div');
  el.className = `toast ${kind}`;
  if (html) el.innerHTML = message; else el.textContent = message;
  box.appendChild(el);
  setTimeout(() => el.remove(), timeout);
}

export function debounce(fn, ms) {
  let t;
  return (...args) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...args), ms);
  };
}

/** localStorage может быть недоступен (приватный режим) — не падаем. */
export const storage = {
  get(key) { try { return localStorage.getItem(key); } catch { return null; } },
  set(key, value) { try { localStorage.setItem(key, value); } catch { /* ignore */ } },
};
