# FlashTUI ─ Terminal Spaced Repetition System

FlashTUI is a local-first, keyboard-driven, exceptionally aesthetic flashcard application optimized for standard terminal environments.

## Core Design Pillars

- **Keyboard-Driven modal flow**: Modal Vim-style movement keys (`h`/`j`/`k`/`l`, `gg`, `G`, `:`) remove mouse friction.
- **Harmony color palettes**: Uses curated violet/cyan accents and Unicode elements (e.g. `🔥` for streak tracks).
- **Zero dynamic compile ties**: SQLite is embedded via a pure-Go driver with Write-Ahead Logging (WAL) and recursive CTE trees.
- ** Leeches and Undo rollbacks**: Sessions dynamically mark difficult "leech" cards and register snapshots for transactional undos.

## Installation Instructions

### Prerequisites
- **Go compiler**: Go 1.21 or higher installed on your path.
- **Terminal Emulator**: Any modern terminal with UTF-8 character and 256-color support.

### Compilation
From the project workspace root directory, compile the binary:
```bash
go build -o flashtui
```
This builds a self-contained, statically linked executable with no dynamically linked SQLite libraries.

### Execution
Run the compiled binary:
```bash
./flashtui
```
The application will automatically initialize the base configuration at `~/.config/flashtui/config.yaml` and standard sqlite database at `~/.config/flashtui/data.db` on first boot.

## Modal Mappings Cheat Sheet

### Dashboard Normal Mode
- `h` / `l` : Switch between Left Panel (Decks tree) and Right Panel (Cards list).
- `j` / `k` : Scroll list selections.
- `gg` / `G` : Jump to top / bottom of list.
- `ctrl+u` / `ctrl+d` : Half-page scroll up / down.
- `a` : Create new Deck (when Left Panel focused) or Card (when Right Panel focused).
- `e` : Edit selected Deck name or Card contents.
- `d d` : Delete/purge selected Deck (recursive cascade) or Card.
- `/` : Focus search input block.
- `:` : Focus console command input.
- `Enter` : Start review study session on selected Deck.

### Review Study Mode
- `Space` : Flip card to reveal answer.
- `h` : Peek hint.
- `s` : Shuffle remaining session queue.
- `u` : Rollback last graded card (undos database stats and log entry).
- `1` - `5` : Rate active card confidence score.
- `Esc` : Exit study session and return to dashboard.

### Console Commands (`:`)
- `:w` : Force database sync update.
- `:q` : Safely close SQLite and quit.
- `:import <path>` : Parse cards from external Markdown file (headers = front, comment blocks = hints/tags).
- `:export <deck_name> <path>` : Compile deck tree and cards recursively to JSON file.
