// Страницы входа и регистрации.
import * as api from './api.js';
import { $, hydrateIcons } from './ui.js';

const form = $('#auth-form');
const mode = form.dataset.mode;
const submit = $('#submit');
const formError = $('#form-error');

const USERNAME_RE = /^[a-zA-Z0-9_.-]{3,32}$/;
const utf8Length = (s) => new TextEncoder().encode(s).length;

function showFormError(message) {
  formError.textContent = message;
  formError.hidden = !message;
}

function setFieldError(name, message) {
  const input = form.elements[name];
  const box = $(`#${name}-error`);
  input.setAttribute('aria-invalid', String(Boolean(message)));
  if (box) box.textContent = message || '';
}

// Проверки дублируют серверные, чтобы сразу подсказать пользователю.
// Сервер всё равно проверяет всё сам: клиенту доверять нельзя.
function validateSignUp(values) {
  const errors = {};
  if (!values.full_name.trim()) errors.full_name = 'Enter your name';
  if (!USERNAME_RE.test(values.username)) errors.username = 'Use 3–32 letters, digits, “_”, “.” or “-”';
  if (values.password.length < 8) errors.password = 'Use at least 8 characters';
  else if (utf8Length(values.password) > 72) errors.password = 'Password is too long (max 72 bytes)';
  if (values.confirm !== values.password) errors.confirm = 'Passwords do not match';
  return errors;
}

function strength(password) {
  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
  if (/\d/.test(password) && /[^a-zA-Z0-9]/.test(password)) score++;
  return score;
}

function friendlyError(err) {
  if (err.status === 429) return `Too many attempts. Try again in ${err.retryAfter || 60} seconds.`;
  if (err.status === 401) return 'Wrong username or password.';
  if (err.status === 409) return 'This username is already taken.';
  if (err.status >= 500 || err.status === undefined) return 'Server is unavailable. Please try again later.';
  return err.message.replace(/^invalid input: /, '');
}

async function handleSubmit(e) {
  e.preventDefault();
  showFormError('');
  const values = Object.fromEntries(new FormData(form));

  if (mode === 'sign-up') {
    const errors = validateSignUp(values);
    ['full_name', 'username', 'password', 'confirm'].forEach((f) => setFieldError(f, errors[f]));
    const first = Object.keys(errors)[0];
    if (first) {
      form.elements[first].focus();
      return;
    }
  } else if (!values.username || !values.password) {
    showFormError('Enter your username and password.');
    return;
  }

  submit.disabled = true;
  submit.textContent = mode === 'sign-up' ? 'Creating account…' : 'Signing in…';
  try {
    if (mode === 'sign-up') {
      await api.signUp(values.full_name.trim(), values.username.trim(), values.password);
    }
    await api.signIn(values.username.trim(), values.password);
    location.href = './';
  } catch (err) {
    if (mode === 'sign-up' && err.status === 409) {
      setFieldError('username', 'This username is already taken');
      form.elements.username.focus();
    } else {
      showFormError(friendlyError(err));
    }
    submit.disabled = false;
    submit.textContent = mode === 'sign-up' ? 'Create account' : 'Sign in';
  }
}

hydrateIcons();
form.addEventListener('submit', handleSubmit);
// Ошибка поля исчезает, как только пользователь начинает его исправлять.
form.addEventListener('input', (e) => {
  if (e.target.name && $(`#${e.target.name}-error`)) setFieldError(e.target.name, '');
  showFormError('');
});

if (mode === 'sign-up') {
  const bar = $('#strength-bar');
  const colors = ['#f87171', '#f87171', '#fbbf24', '#60a5fa', '#34d399'];
  form.elements.password.addEventListener('input', (e) => {
    const s = strength(e.target.value);
    bar.style.width = e.target.value ? `${Math.max(s, 1) * 25}%` : '0';
    bar.style.background = colors[s];
  });
}

// Уже вошли — сразу на главную.
api.restoreSession().then((user) => { if (user) location.replace('./'); });
