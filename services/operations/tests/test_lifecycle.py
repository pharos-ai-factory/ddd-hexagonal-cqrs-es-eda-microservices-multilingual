"""Composition failures release acquired resources before propagating the error."""

from unittest.mock import MagicMock, patch
import pytest
from operations.apps import service
from operations.apps.composition import preparation as preparation_di, collection as collection_di


def test_registration_failure_closes_both_resources() -> None:
    preparation, collection = MagicMock(), MagicMock()
    with (
        patch.dict("os.environ", {"APP_ENV": "development"}),
        patch.object(service, "secret", return_value="test-configuration"),
        patch.object(preparation_di, "PreparationContainer", return_value=preparation),
        patch.object(collection_di, "CollectionContainer", return_value=collection),
        patch.object(preparation_di, "subscriptions", side_effect=ValueError("registration")),
    ):
        with pytest.raises(ValueError):
            service.create_app()
    preparation.shutdown_resources.assert_called_once()
    collection.shutdown_resources.assert_called_once()


def test_partial_resource_startup_is_closed() -> None:
    preparation, collection = MagicMock(), MagicMock()
    collection.init_resources.side_effect = ValueError("pool")
    with (
        patch.dict("os.environ", {"APP_ENV": "development"}),
        patch.object(service, "secret", return_value="test-configuration"),
        patch.object(preparation_di, "PreparationContainer", return_value=preparation),
        patch.object(collection_di, "CollectionContainer", return_value=collection),
    ):
        with pytest.raises(ValueError):
            service.create_app()
    preparation.shutdown_resources.assert_called_once()
    collection.shutdown_resources.assert_called_once()


def test_partial_worker_startup_drains_before_resource_cleanup() -> None:
    import asyncio
    from collections.abc import Callable
    from threading import Event, Thread

    preparation, collection = MagicMock(), MagicMock()
    started, release, finished = Event(), Event(), Event()
    failures: list[BaseException] = []
    calls = 0

    def thread_factory(target: Callable[[], None], daemon: bool) -> Thread:
        nonlocal calls
        calls += 1
        if calls == 2:
            raise RuntimeError("worker startup")

        def work() -> None:
            started.set()
            release.wait()
            finished.set()

        return Thread(target=work, daemon=daemon)

    def shutdown() -> None:
        assert finished.is_set(), "Pool closed before in-flight work drained"

    preparation.shutdown_resources.side_effect = shutdown
    with (
        patch.dict("os.environ", {"APP_ENV": "development"}),
        patch.object(service, "secret", return_value="test-configuration"),
        patch.object(preparation_di, "PreparationContainer", return_value=preparation),
        patch.object(collection_di, "CollectionContainer", return_value=collection),
        patch.object(preparation_di, "subscriptions", return_value=()),
        patch.object(collection_di, "subscriptions", return_value=()),
        patch.object(service, "authenticate"),
        patch.object(service, "mount_paged_queries"),
        patch.object(service, "mount_command"),
        patch.object(service, "Thread", side_effect=thread_factory),
    ):
        app = service.create_app()

        async def start() -> None:
            async with app.router.lifespan_context(app):
                raise AssertionError("Startup should have failed")

        def run() -> None:
            try:
                asyncio.run(start())
            except BaseException as error:
                failures.append(error)

        runner = Thread(target=run)
        runner.start()
        try:
            assert started.wait(2)
            preparation.shutdown_resources.assert_not_called()
        finally:
            release.set()
            runner.join(3)
        assert not runner.is_alive()
    assert len(failures) == 1 and str(failures[0]) == "worker startup"
    preparation.shutdown_resources.assert_called_once()
    collection.shutdown_resources.assert_called_once()
