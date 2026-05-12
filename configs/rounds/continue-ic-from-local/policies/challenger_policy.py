def score(task):
    files = task.get("candidate_files", [])
    return files[:3]
