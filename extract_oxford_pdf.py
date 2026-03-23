#!/usr/bin/env python3
"""Extract American Oxford 5000 word list from PDF to JSON."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path

from pypdf import PdfReader

POS_UNIT = r"(?:n\.|v\.|adj\.|adv\.|prep\.|pron\.|conj\.|interj\.|num\.|determiner|exclamation|adj\./adv\.|number)"
ORPHAN_TAIL = re.compile(rf"^({POS_UNIT})\s+(B2|C1)\s*$", re.I)


def pdf_to_lines(reader: PdfReader) -> list[str]:
    text = "\n".join((p.extract_text() or "") for p in reader.pages)
    text = re.sub(r"© Oxford University Press\s+\d+\s*/\s*\d+", "\n", text)
    text = re.sub(r"The Oxford 5000™\s*\(American English\)", "\n", text)
    raw = [ln.strip() for ln in text.splitlines() if ln.strip()]
    merged: list[str] = []
    for ln in raw:
        if (
            merged
            and ORPHAN_TAIL.match(ln)
            and not re.search(r"\b(B2|C1)\s*$", merged[-1])
        ):
            merged[-1] = f"{merged[-1].rstrip()} {ln}"
        else:
            merged.append(ln)
    return merged


def normalize_line(ln: str) -> str:
    ln = re.sub(r"\b([a-z]+\.)\s*,\s*(B2|C1)\b", r"\1 \2", ln, flags=re.I)
    return re.sub(r"\s+", " ", ln).strip()


tail_re = re.compile(
    rf"^(.+?)\s+(({POS_UNIT})\s*(?:B2|C1)?(?:\s*,\s*{POS_UNIT}\s*(?:B2|C1)?)*)\s*$",
    re.I,
)
clause_re = re.compile(rf"^\s*({POS_UNIT})\s*(B2|C1)?\s*$", re.I)


def normalize_pos(p: str) -> str:
    p = p.rstrip(".")
    if p.lower() == "number":
        return "num"
    return p


def parse_tail(tail: str) -> list[dict[str, str]] | None:
    parts = [p.strip() for p in tail.split(",")]
    senses: list[dict[str, str | None]] = []
    for p in parts:
        m = clause_re.match(p)
        if not m:
            return None
        pos_raw, lev = m.group(1), m.group(2)
        senses.append({"part_of_speech": normalize_pos(pos_raw), "level": lev})
    last: str | None = None
    for s in reversed(senses):
        if s["level"]:
            last = str(s["level"])
        elif last:
            s["level"] = last
    if any(s["level"] is None for s in senses):
        return None
    return [dict(s) for s in senses]  # type: ignore[arg-type]


def parse_entries(lines: list[str]) -> tuple[list[dict], list[str]]:
    entries: list[dict] = []
    unparsed: list[str] = []
    for ln in lines:
        ln = normalize_line(ln)
        if ln.startswith("The Oxford 5000 is") or ln.startswith("As well as"):
            continue
        if not re.search(r"\b(B2|C1)\s*$", ln):
            continue
        m = tail_re.match(ln)
        if not m:
            unparsed.append(ln)
            continue
        lemma, tail = m.group(1).strip(), m.group(2).strip()
        senses = parse_tail(tail)
        if senses is None:
            unparsed.append(ln)
            continue
        entries.append({"word": lemma, "senses": senses})
    return entries, unparsed


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "pdf",
        type=Path,
        nargs="?",
        default=Path.home() / "Downloads" / "American_Oxford_5000.pdf",
        help="Path to American Oxford 5000 PDF",
    )
    ap.add_argument(
        "-o",
        "--output",
        type=Path,
        default=Path(__file__).resolve().parent / "american_oxford_5000.json",
        help="Output JSON path",
    )
    args = ap.parse_args()

    reader = PdfReader(str(args.pdf))
    lines = pdf_to_lines(reader)
    entries, unparsed = parse_entries(lines)
    for i, e in enumerate(entries):
        e["index"] = i

    doc = {
        "source": str(args.pdf.resolve()),
        "title": "The Oxford 5000 (American English)",
        "note": "Additional 2000 words (B2–C1) beyond the Oxford 3000; extracted from PDF.",
        "entry_count": len(entries),
        "entries": entries,
    }
    if unparsed:
        doc["unparsed_lines"] = unparsed

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {len(entries)} entries to {args.output}")
    if unparsed:
        print(f"Warning: {len(unparsed)} unparsed lines")


if __name__ == "__main__":
    main()
