# Format Rules Contract

Rules apply to the rendered body in every mode (plain, `--format md`, `--html`) and every target. Clean plain or markdown input MUST produce zero problems (enforced by a fixture test for both targets).

## Allow-list

Elements: `p br h1 h2 h3 ul ol li pre code a strong b em i u s blockquote hr`. Void: `br`, `hr` (`<br>` and `<br/>` both accepted). Attributes: only `href` on `a`. Tag and attribute names are compared case-insensitively.

## Body rules

| # | Rule id | Fails when | Detail example |
|---|---------|-----------|----------------|
| 1 | `tag-not-allowed` | A start or end tag not on the allow-list | `<table>` |
| 2 | `attribute-not-allowed` | Any attribute other than `href` on `a` | `style on <p>` |
| 3 | `broken-html` | Unclosed non-void element at end, mismatched end tag, or end tag with no open element | `unclosed <p>`, `</em> closes <strong>`, `stray </p>` |
| 4 | `leftover-markdown` | Visible text (outside `code`/`pre`) contains `**x**`, a line starting with `#`, `##` or `###` then a space, `[label](url)`, or `` `x` `` | `**bold**` |
| 5 | `literal-escape` | Visible text (outside `code`/`pre`) contains a backslash followed by `n`, `t` or `"` | `\n in "line one\nline two"` |
| 6 | `extra-blank-lines` | A `p` with no non-whitespace content, or 3+ `br` with only whitespace between | `empty <p>`, `3 <br> in a row` |
| 7 | `link-scheme` | `href` scheme other than `http`, `https`, `mailto` (relative or empty `href` also fails) | `javascript:alert(1)` |

A line starting with `- ` is not a problem. Text excerpts in details are cut to 40 characters.

Not flagged (known client limits): Teams web leading space after `<br>`; `<pre>` flattening and `<hr>` gap on Teams for iOS.

## Subject rules (`mail send` only)

| Rule id | Fails when | Detail |
|---------|-----------|--------|
| `leftover-markdown` | Rule 4 applied to the subject text | `subject: **urgent**` |
| `literal-escape` | Rule 5 applied to the subject text | `subject: \n` |
| `newline-in-subject` | Subject contains `\r` or `\n` | `subject` |

Subject problems are listed before body problems.

## Blocked send error

stderr (one line, existing schema):

```json
{"class":"usage","message":"2 format problems: broken-html: unclosed <p>; leftover-markdown: **bold**","hint":"run with --preview to see them"}
```

One problem uses `1 format problem: …`. Exit 3. Nothing is sent. Through MCP the same error is returned as `IsError`.
