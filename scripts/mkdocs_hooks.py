"""MkDocs hooks so docs/ renders well both on GitHub and on the Pages site.

1. Links that leave docs/ (../README.md, ../examples/, ../agent-skills/...) are
   not part of the site, so they are rewritten to the file on GitHub.
2. GitHub alert blockquotes (> [!NOTE], > [!TIP], ...) become MkDocs
   admonitions (!!! note), so callouts look right on the site too.
"""

import posixpath
import re

REPO_URL = "https://github.com/parvez3019/go-swagger3"
BRANCH = "main"

# [text](target) or [text](target "title")
_LINK = re.compile(r"\]\((?P<target>[^)\s]+)(?P<rest>[^)]*)\)")

# > [!NOTE]\n> line\n> line
_ALERT = re.compile(
    r"^> \[!(?P<kind>NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\n(?P<body>(?:>.*(?:\n|$))*)",
    re.MULTILINE,
)
_ALERT_KIND = {
    "NOTE": "note",
    "TIP": "tip",
    "IMPORTANT": "info",
    "WARNING": "warning",
    "CAUTION": "danger",
}


def _rewrite(target: str, src_uri: str) -> str | None:
    if re.match(r"^[a-z][a-z0-9+.-]*:", target) or target.startswith(("#", "/")):
        return None
    path, sep, anchor = target.partition("#")
    page_dir = posixpath.dirname(src_uri)
    resolved = posixpath.normpath(posixpath.join("docs", page_dir, path))
    if resolved.startswith("docs/") or resolved == "docs":
        return None
    kind = "tree" if path.endswith("/") else "blob"
    return f"{REPO_URL}/{kind}/{BRANCH}/{resolved.rstrip('/')}" + (sep + anchor if sep else "")


def _alert(match: re.Match) -> str:
    lines = [re.sub(r"^> ?", "", line) for line in match.group("body").splitlines()]
    body = "\n".join(("    " + line) if line.strip() else "" for line in lines)
    return f"!!! {_ALERT_KIND[match.group('kind')]}\n\n{body}\n"


def on_page_markdown(markdown, page, config, files):
    src_uri = page.file.src_uri

    def repl(match):
        new = _rewrite(match.group("target"), src_uri)
        if new is None:
            return match.group(0)
        return f"]({new}{match.group('rest')})"

    markdown = _ALERT.sub(_alert, markdown)
    return _LINK.sub(repl, markdown)
