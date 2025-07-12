import io
import pytest
from fastapi.testclient import TestClient

from main import app

client = TestClient(app)


def test_health():
    response = client.get("/health")
    assert response.status_code == 200
    assert "FastAPI backend running" in response.text


from datetime import datetime, timedelta

def test_predict_success():
    csv_content = "Date,Close,Other\n"
    start_date = datetime(2021, 1, 1)

    for i in range(1000):
        day = start_date + timedelta(days=i)
        csv_content += f"{day.strftime('%Y-%m-%d')},{100 + i},foo\n"

    files = {
        "file": ("test.csv", io.BytesIO(csv_content.encode()), "text/csv")
    }

    response = client.post("/api/v1/predict", files=files)
    assert response.status_code == 200, response.text

    data = response.json()
    assert "predictions" in data
    assert "y_test" in data
    assert "dates" in data

    assert isinstance(data["predictions"], list)
    assert isinstance(data["y_test"], list)
    assert isinstance(data["dates"], list)

    assert len(data["predictions"]) > 0
    assert len(data["y_test"]) > 0
    assert len(data["dates"]) > 0

def test_predict_missing_file():
    response = client.post("/api/v1/predict")
    assert response.status_code == 422  # Required UploadFile missing
    assert "file" in response.text


def test_predict_invalid_header():
    csv_content = "Date,Price,Other\n2023-01-01,100,foo\n2023-01-02,101,foo\n"
    files = {
        "file": ("test.csv", io.BytesIO(csv_content.encode()), "text/csv")
    }

    response = client.post("/api/v1/predict", files=files)
    assert response.status_code == 400
    assert "Close" in response.text


def test_predict_not_enough_data():
    csv_content = "Date,Close,Other\n2023-01-01,100,foo\n"
    files = {
        "file": ("test.csv", io.BytesIO(csv_content.encode()), "text/csv")
    }

    response = client.post("/api/v1/predict", files=files)
    assert response.status_code == 400
    assert "Not enough data" in response.text
