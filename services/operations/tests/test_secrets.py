from pathlib import Path
import pytest
from operations.foundation.secrets import secret


def test_secret_files_fail_closed(monkeypatch: pytest.MonkeyPatch, tmp_path: Path) -> None:
    path = tmp_path / "credential"
    monkeypatch.setenv("PROVIDER_KEY_FILE", str(path))
    path.write_text("sensitive-value\n")
    assert secret("PROVIDER_KEY") == "sensitive-value"
    monkeypatch.setenv("PROVIDER_KEY", "")
    with pytest.raises(RuntimeError, match="mutually exclusive"):
        secret("PROVIDER_KEY")
    monkeypatch.delenv("PROVIDER_KEY")
    for value in ("", " \n"):
        path.write_text(value)
        with pytest.raises(RuntimeError, match="must not be empty"):
            secret("PROVIDER_KEY")
    path.unlink()
    with pytest.raises(RuntimeError) as failure:
        secret("PROVIDER_KEY")
    assert str(failure.value) == "PROVIDER_KEY_FILE cannot be read"


def test_direct_secret_preserves_spaces(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROVIDER_KEY", " value ")
    assert secret("PROVIDER_KEY") == " value "
