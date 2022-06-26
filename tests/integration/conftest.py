"""Global fixtures"""
import pytest
from c8y_api import CumulocityApi
from c8y_api.model import Device, Measurement
from pytest_c8y.device_management import DeviceManagement
from pytest_c8y.task import BackgroundTask


@pytest.fixture(scope="function")
def device01(sample_device: Device):
    yield sample_device


@pytest.fixture(scope="function")
def asserter(device_mgmt: DeviceManagement):
    yield device_mgmt

@pytest.fixture(scope="function")
def device_with_measurements(background_task: BackgroundTask, live_c8y: CumulocityApi, sample_device: Device):

    measurement = Measurement(c8y=live_c8y, type="cicd_TestMeasurement", source=sample_device.id, c8y_Temp={"c8y_T1":{"unit":"degC", "value": 1.23}})
    background_task.start(measurement.create, interval=5)

    yield sample_device
    background_task.stop()
