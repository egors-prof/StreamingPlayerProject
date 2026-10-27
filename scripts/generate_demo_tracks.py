#!/usr/bin/env python3
"""Генерирует демо-треки для каталога: простая синтезаторная музыка без прав.

Каждый трек собирается из пэда (аккорды), баса, арпеджио и драм-машины.
Результат кодируется в MP3 через ffmpeg и кладётся в services/stream/music/.

Запуск:  python3 scripts/generate_demo_tracks.py
Нужно:   numpy, scipy, ffmpeg
"""
import os
import subprocess
import sys
import tempfile
import wave

import numpy as np
from scipy.signal import lfilter

SR = 44100
OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "services", "stream", "music")

NOTE_INDEX = {"C": 0, "C#": 1, "Db": 1, "D": 2, "D#": 3, "Eb": 3, "E": 4, "F": 5,
              "F#": 6, "Gb": 6, "G": 7, "G#": 8, "Ab": 8, "A": 9, "A#": 10, "Bb": 10, "B": 11}

# file, bpm, progression (root, quality), seed, drums, feel
TRACKS = [
    ("demo-01-neon-rain.mp3", 100, [("A", "m"), ("F", "M"), ("C", "M"), ("G", "M")], 1, True, "drive"),
    ("demo-02-midnight-arcade.mp3", 118, [("E", "m"), ("C", "M"), ("G", "M"), ("D", "M")], 2, True, "bright"),
    ("demo-03-glass-horizon.mp3", 90, [("D", "m"), ("Bb", "M"), ("F", "M"), ("C", "M")], 3, True, "soft"),
    ("demo-04-pulse-runner.mp3", 124, [("F#", "m"), ("D", "M"), ("A", "M"), ("E", "M")], 4, True, "drive"),
    ("demo-05-low-orbit.mp3", 80, [("C", "M7"), ("A", "m7"), ("F", "M7"), ("G", "M")], 5, False, "ambient"),
    ("demo-06-coastline-fm.mp3", 108, [("G", "m"), ("Eb", "M"), ("Bb", "M"), ("F", "M")], 6, True, "bright"),
]

CHORD = {"M": [0, 4, 7], "m": [0, 3, 7], "M7": [0, 4, 7, 11], "m7": [0, 3, 7, 10]}


def freq(midi):
    return 440.0 * 2 ** ((midi - 69) / 12)


def lowpass(x, cutoff):
    """Однополюсный фильтр нижних частот: y[n] = (1-a)·x[n] + a·y[n-1]."""
    a = np.exp(-2 * np.pi * cutoff / SR)
    return lfilter([1 - a], [1, -a], x)


def saw(f, t):
    return 2 * ((f * t) % 1.0) - 1


def env_adsr(n, attack, release):
    e = np.ones(n)
    a = min(int(attack * SR), n)
    r = min(int(release * SR), n - a)
    if a:
        e[:a] = np.linspace(0, 1, a)
    if r:
        e[n - r:] *= np.linspace(1, 0, r)
    return e


def render(bpm, prog, seed, drums, feel, bars=32):
    rng = np.random.default_rng(seed)
    beat = 60 / bpm
    bar = 4 * beat
    total = int(bars * bar * SR) + SR * 2
    left = np.zeros(total)
    right = np.zeros(total)

    def add(sig, start_s, gain=1.0, pan=0.0):
        s = int(start_s * SR)
        e = min(s + len(sig), total)
        sig = sig[: e - s] * gain
        left[s:e] += sig * (1 - pan) * 0.5 ** 0.5 * 1.4
        right[s:e] += sig * (1 + pan) * 0.5 ** 0.5 * 1.4

    base_octave = 48  # C3
    cutoff = {"drive": 1800, "bright": 2600, "soft": 1200, "ambient": 900}[feel]

    for b in range(bars):
        root_name, quality = prog[b % len(prog)]
        root = base_octave + NOTE_INDEX[root_name]
        t0 = b * bar
        section = b // 8  # 0 intro, 1 main, 2 break, 3 outro

        # Пэд: детюненные пилы через фильтр, по аккорду на такт.
        n = int(bar * SR)
        t = np.arange(n) / SR
        pad = np.zeros(n)
        for iv in CHORD[quality]:
            f = freq(root + 12 + iv)
            pad += saw(f * 1.003, t) + saw(f * 0.997, t)
        pad = lowpass(pad / 8, cutoff * (0.6 if section in (0, 3) else 1.0))
        pad *= env_adsr(n, 0.25, 0.3)
        add(pad, t0, 0.32, pan=-0.2 if b % 2 else 0.2)

        # Бас: восьмые на тонике (кроме интро).
        if section in (1, 2) or (section == 3 and b % 8 < 4):
            step = beat / 2
            for i in range(8):
                m = int(step * SR)
                tt = np.arange(m) / SR
                f = freq(root - 12 + (7 if i in (3, 7) else 0))
                sig = np.sin(2 * np.pi * f * tt) + 0.3 * np.sin(4 * np.pi * f * tt)
                sig *= np.exp(-tt * 6)
                add(sig, t0 + i * step, 0.45)

        # Арпеджио: шестнадцатые по нотам аккорда в основной части.
        if section == 1 or (section == 2 and feel != "ambient"):
            notes = [root + 24 + iv for iv in CHORD[quality]]
            pattern = rng.permutation(len(notes)).tolist()
            step = beat / (4 if feel != "ambient" else 2)
            count = int(bar / step)
            for i in range(count):
                m = int(step * SR * 1.5)
                tt = np.arange(m) / SR
                f = freq(notes[pattern[i % len(pattern)]] + (12 if i % 8 == 7 else 0))
                sig = (np.sin(2 * np.pi * f * tt) + 0.25 * np.sin(6 * np.pi * f * tt)) * np.exp(-tt * 9)
                add(sig, t0 + i * step, 0.16, pan=0.5 * np.sin(i))

        # Ударные.
        if drums and section in (1, 2, 3):
            for i in range(4):  # бочка на каждую долю
                if section == 2 and i % 2:
                    continue
                m = int(0.35 * SR)
                tt = np.arange(m) / SR
                f = 120 * np.exp(-tt * 18) + 45
                kick = np.sin(2 * np.pi * np.cumsum(f) / SR) * np.exp(-tt * 9)
                add(kick, t0 + i * beat, 0.8)
            for i in (1, 3):  # малый барабан на 2 и 4
                m = int(0.2 * SR)
                tt = np.arange(m) / SR
                snare = (rng.standard_normal(m) * 0.6 + np.sin(2 * np.pi * 190 * tt)) * np.exp(-tt * 22)
                add(lowpass(snare, 5000), t0 + i * beat, 0.35)
            for i in range(8):  # хай-хэт на восьмые
                m = int(0.05 * SR)
                noise = rng.standard_normal(m)
                hat = (noise - lowpass(noise, 7000)) * np.exp(-np.arange(m) / SR * 80)
                add(hat, t0 + i * beat / 2, 0.12 if i % 2 else 0.07, pan=0.3)

    mix = np.stack([left, right], axis=1)
    fade = int(3 * SR)
    mix[:fade] *= np.linspace(0, 1, fade)[:, None]
    mix[-fade:] *= np.linspace(1, 0, fade)[:, None]
    mix /= np.max(np.abs(mix)) + 1e-9
    return (mix * 0.89 * 32767).astype(np.int16)


def main():
    os.makedirs(OUT_DIR, exist_ok=True)
    for name, bpm, prog, seed, drums, feel in TRACKS:
        pcm = render(bpm, prog, seed, drums, feel)
        with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
            with wave.open(tmp.name, "wb") as w:
                w.setnchannels(2)
                w.setsampwidth(2)
                w.setframerate(SR)
                w.writeframes(pcm.tobytes())
            out = os.path.join(OUT_DIR, name)
            subprocess.run(["ffmpeg", "-y", "-loglevel", "error", "-i", tmp.name,
                            "-codec:a", "libmp3lame", "-b:a", "128k", out], check=True)
            os.unlink(tmp.name)
        dur = len(pcm) / SR
        print(f"{name}: {dur:.0f} s, {os.path.getsize(out) / 1e6:.1f} MB")


if __name__ == "__main__":
    sys.exit(main())
