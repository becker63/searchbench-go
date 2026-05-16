def score_fn(node, graph, depth):
    data = getattr(node, "data", None)
    if data is None and isinstance(node, dict):
        data = node.get("data", {})
    if data is None:
        data = {}

    score = 0.0

    kind = str(data.get("kind", "")).lower()
    path = str(data.get("file", data.get("path", ""))).lower()
    symbol = str(data.get("symbol", data.get("name", ""))).lower()

    if kind in {"function", "method", "class", "symbol", "file"}:
        score += 2.0

    if path.endswith(".py"):
        score += 1.5

    if "/test" in path or path.startswith("test") or "_test." in path:
        score -= 1.0

    if any(part in path for part in [".venv", "__pycache__", "vendor", "generated"]):
        score -= 3.0

    if symbol:
        score += 0.5

    try:
        score -= 0.25 * float(depth)
    except Exception:
        pass

    return float(score)
