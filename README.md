<div align="center">

# 🐒 Desktop Monkey

**A tiny pixel-art monkey that lives on your desktop.**

<img src="https://desktop-monkey.my-monkey.fr/suivre.gif" width="440" alt="The monkey walking across the desktop, following the cursor">

*Free, open source, no account, no dependencies — Windows & macOS.*

</div>

---

Your cursor is his **best friend**: he follows it around, and when the mouse
sits still too long he gets on with his own life — wandering, climbing,
napping, and leaving the odd surprise on your desktop. Click him and he
complains; click him enough and he keels over (you can revive him). Hold the
click instead and you can pick him up and put him down anywhere.

## What he does

| | |
|:---:|:---:|
| <img src="https://desktop-monkey.my-monkey.fr/suivre.gif" width="380"><br>**Follows your cursor** everywhere | <img src="https://desktop-monkey.my-monkey.fr/chasse.gif" width="380"><br>**Hunts it down** and whacks it with bananas |
| <img src="https://desktop-monkey.my-monkey.fr/degats.gif" width="380"><br>**Click him** — three hearts, then he keels over | <img src="https://desktop-monkey.my-monkey.fr/grimpe.gif" width="380"><br>**Climbs the screen edge**, then drops off |
| <img src="https://desktop-monkey.my-monkey.fr/pond.gif" width="380"><br>**Leaves a steaming poop** — click it to pop it | |

And that's not all:

- **He has moods.** Four gauges — energy, boredom, happiness, fear — drift
  with how you treat him and weight everything he does: a bored monkey does
  mischief, a scared one keeps his distance, a drained one eats and naps.
  Peek at them in his tray menu.
- **He throws bananas.** When he's after your cursor he lobs one at it, from
  anywhere on screen — two different throws, and a hit knocks your pointer
  aside. Miss, and the banana arcs on and falls off the bottom.
- **Tickle him**: shake the mouse right next to him and he giggles.
- **Pick him up**: press and hold on him and he dangles from your cursor —
  drop him wherever you like, he carries on from there. A quick click still
  just bonks him.
- **He cleans up** — old poops get picked up and flung off the screen… or
  thrown at your cursor. Hit him mid-carry and he drops everything.
- **He digests**: no poop without a meal first, and never right after.
- **He naps for real** — lies down, little Z's floating up. Any mouse move
  startles him awake.

To bring a dead monkey back: **grab the body and shake it**. The rest is yours
to discover — small talk, stolen cursors, post-resurrection jitters…

## Install

### macOS

```bash
brew install --cask my-monkeys/tap/desktop-monkey
```

Launch **Desktop Monkey** from Spotlight. A monkey icon appears in the menu
bar — *Launch at startup*, *Open settings*, *Quit*. Signed and notarized by
Apple, so it opens without a warning.

### Windows

Download **`monkey.exe`** from the
[latest release](https://github.com/my-monkeys/desktop-monkey/releases/latest)
and double-click it. A monkey icon appears in the notification area with the
same little menu (*Launch at startup*, *Open settings*, *Quit*).

That's it — he has no window, he just lives on the desktop.

## Good to know

- **Languages** — he speaks English, or French if your system is set to
  French. Force it with `"langue": "en"` / `"langue": "fr"` in the settings.
- **Settings** — pick *Open settings* from his menu. On macOS a native
  settings window opens (Appearance / Life / Character tabs) with sliders for
  his size, speed, language and moods; on Windows a small settings page opens
  in your browser. Save, and he restarts with the new settings. (Power users
  can still edit `config.json` directly.)
- **Uninstall** — `brew uninstall --cask desktop-monkey` on macOS; just delete
  `monkey.exe` on Windows.

## He isn't an only child

Two other free, open-source things from [My-Monkey](https://my-monkey.fr) — no
accounts there either:

- 🎙️ **[OpenSuperWhisper](https://opensuperwhisper.com)** — speak, it types.
  Voice dictation in any app on your Mac, with four transcription engines (three
  of them entirely on-device). No subscription.
  ([source](https://github.com/my-monkeys/OpenSuperWhisper))
- 📊 **[Claude Monitor](https://github.com/my-monkeys/claude-monitor)** — a
  menu-bar app that counts your Claude Code instances and MCP servers, and kills
  a runaway swarm before it takes the whole Mac down.

## Credits

- Monkey sprites: [Pixel Art Monkey](https://tiki-ted.itch.io/pixel-art-monkey) by tiki-ted
- Heart explosion: [Reactorcore](https://reactorcore.itch.io/explosion-spriteanim-sheet-minipack)

Sprites belong to their authors and aren't covered by the code's MIT license.

<sub>Built in Go, no game engine — the source (in French 🥖) builds with `./build.sh windows` / `./build.sh mac`.</sub>
