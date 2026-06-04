from pathlib import Path

from ragas_runner import load_jsonl


def test_load_jsonl(tmp_path: Path) -> None:
    path = tmp_path / "data.jsonl"
    path.write_text('{"question":"q","answer":"a"}\n', encoding="utf-8")
    assert load_jsonl(path) == [{"question": "q", "answer": "a"}]

