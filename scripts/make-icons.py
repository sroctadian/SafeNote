#!/usr/bin/env python3
"""Generate correctly sized/padded Linux desktop icons for SafeNote.

Why this exists
---------------
Desktop shells (GNOME Alt+Tab/dash, KDE taskbar, ...) scale an app's icon
file to a fixed slot. Artwork that fills the canvas edge-to-edge renders
visibly LARGER than other apps, because the freedesktop/GNOME convention is
that app artwork occupies only ~80-88% of the canvas, with transparent
padding around it (see the GNOME HIG "app icons" grid).

Inputs/outputs
--------------
Source (full-bleed artwork, any size):
    build/linux/safenote.raw.png

Outputs:
    build/linux/safenote.png                       512x512 master
    build/linux/safenote-128.png                   128x128 legacy helper
    build/linux/icons/hicolor/<N>x<N>/apps/safenote.png   full hicolor set

Usage:
    python3 scripts/make-icons.py            # regenerate everything

Requires: Python 3 + Pillow (pip install Pillow)
"""
import shutil
import sys
from pathlib import Path

from PIL import Image

PROJECT = Path(__file__).resolve().parent.parent
SRC = PROJECT / "build" / "linux" / "safenote.raw.png"
OUT_DIR = PROJECT / "build" / "linux"

MASTER = 512          # standard hicolor size for the master file
FILL = 0.88           # visible artwork max-dim / canvas (HIG grid: tall=91%, square=75%)
ALPHA_CUT = 16        # zero out alpha below this (invisible artifacts/halos)
SIZES = (16, 24, 32, 48, 64, 72, 96, 128, 256, 512)


def cut_invisible(img: Image.Image) -> Image.Image:
    """Zero alpha < ALPHA_CUT (invisible) - removes resize ringing/halos."""
    r, g, b, a = img.split()
    a = a.point(lambda v: 0 if v < ALPHA_CUT else v)
    return Image.merge("RGBA", (r, g, b, a))


def scaled(master: Image.Image, size: int) -> Image.Image:
    out = master.resize((size, size), Image.LANCZOS)
    return cut_invisible(out) if size != MASTER else out


def main() -> None:
    if not SRC.exists():
        # First run: adopt the current full-bleed icon as the raw source.
        legacy = OUT_DIR / "safenote.png"
        if not legacy.exists():
            sys.exit(f"error: source artwork not found: {SRC}")
        shutil.copy2(legacy, SRC)
        print(f"adopted existing {legacy.name} as raw source -> {SRC.name}")

    art = Image.open(SRC).convert("RGBA")
    # Drop invisible near-transparent pixels (stray halos/shadows) so the
    # artwork is measured by what is actually visible on screen.
    r, g, b, a = art.split()
    a = a.point(lambda v: 0 if v < ALPHA_CUT else v)
    art = Image.merge("RGBA", (r, g, b, a))

    bbox = a.getbbox()
    if bbox is None:
        sys.exit("error: source artwork is fully transparent")
    art = art.crop(bbox)

    # Scale artwork so its larger dimension becomes FILL of the canvas.
    target = round(MASTER * FILL)
    scale = target / max(art.size)
    art = art.resize(
        (max(1, round(art.width * scale)), max(1, round(art.height * scale))),
        Image.LANCZOS,
    )

    master = Image.new("RGBA", (MASTER, MASTER), (0, 0, 0, 0))
    master.paste(art, ((MASTER - art.width) // 2, (MASTER - art.height) // 2))

    # Master + legacy 128 helper.
    master.save(OUT_DIR / "safenote.png", optimize=True)
    scaled(master, 128).save(OUT_DIR / "safenote-128.png", optimize=True)

    # Full hicolor tree for packaging.
    hicolor = OUT_DIR / "icons" / "hicolor"
    if hicolor.exists():
        shutil.rmtree(hicolor)
    for size in SIZES:
        d = hicolor / f"{size}x{size}" / "apps"
        d.mkdir(parents=True)
        scaled(master, size).save(d / "safenote.png", optimize=True)

    w, h = art.size
    print(f"artwork {w}x{h} = {w / MASTER:.0%}x{h / MASTER:.0%} of {MASTER}px canvas")
    print(f"wrote master, safenote-128.png and hicolor set {SIZES[0]}..{SIZES[-1]}")


if __name__ == "__main__":
    main()
