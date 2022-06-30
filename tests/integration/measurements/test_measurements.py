"""Measurement tests"""
from c8y_api import CumulocityApi
from c8y_api.model import Device, Measurement
from utils import command
from pytest_c8y.device_management import DeviceManagement
from pytest_c8y.task import BackgroundTask

def test_measurement(device01: Device):
    assert device01.id

    output = command.execute(
        command.prepare(
            f"""
            c8y devices get --id device01
            """,
            device01=device01.id
        )
    )

    assert output.exit_code == 0

class TestSubscribe:

    def test_measurement_subscribe_duration(self, device_with_measurements: Device):
        output = command.execute(
            command.prepare(
                f"""
                c8y measurements subscribe --device device01 --duration 10s
                """,
                device01=device_with_measurements.id
            )
        )

        assert len(output.jsonlines) > 0
        assert output.exit_code == 0
        assert output.duration > 10


    def test_measurement_subscribe_count(self, device_with_measurements: Device):
        """Subscribe to measurements and stop when 2 measurements have been received
        """
        output = command.execute(
            command.prepare(
                f"""
                c8y measurements subscribe --device device01 --count 2
                """,
                device01=device_with_measurements.id
            )
        )

        assert len(output.jsonlines) == 2
        assert output.exit_code == 0
        assert output.duration < 10


    def test_measurement_subscribe_all_count(self, device_with_measurements: Device):
        output = command.execute(
            command.prepare(
                f"""
                c8y measurements subscribe --count 2
                """,
                device01=device_with_measurements.id
            )
        )

        assert len(output.jsonlines) == 2
        assert output.exit_code == 0
        assert output.duration < 10




def test_measurement_filter_count(background_task: BackgroundTask, live_c8y: CumulocityApi, sample_device: Device):
    measurement = Measurement(c8y=live_c8y, type="cicd_TestMeasurement", source=sample_device.id, c8y_Temp={"c8y_T1":{"unit":"degC", "value": 1.23}})
    background_task.start(measurement.create, interval=5)

    output = command.execute(
        command.prepare(
            f"""
            c8y measurements subscribe --device device01 --count 1 --duration 20s
            """,
            device01=sample_device.id
        )
    )

    assert output.exit_code == 0
    assert output.duration < 20
    assert output.jsonlines[0]["source"]["id"] == sample_device.id
    # assert output.stdout == "test"
