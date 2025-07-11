import pandas as pd
import numpy as np
import requests
import yfinance as yf
import os
from io import StringIO
from dotenv import load_dotenv

load_dotenv()

BACKEND_GO_URL = os.getenv("GO_BACKEND")
if not BACKEND_GO_URL:
    raise ValueError("GO_BACKEND environment variable is not set.")

def fetch_stock_data(ticker: str) -> pd.DataFrame | None:
    try:
        stock = yf.Ticker(ticker)
        df = stock.history(period='max')
        company_name = stock.info.get('longName', '')
        if df.empty or not company_name:
            return None
        df.reset_index(inplace=True)
        df['Ticker'] = ticker
        df['Name'] = company_name
        return df
    except (ValueError, KeyError):
        return None

def send_metrics_to_go(df: pd.DataFrame, ticker: str) -> dict:
    csv_buffer = StringIO()
    df.reset_index().to_csv(csv_buffer, index=False)
    csv_buffer.seek(0)

    files = {"file": ("data.csv", csv_buffer.read())}
    data = {"ticker": ticker}

    response = requests.post(f"{BACKEND_GO_URL}/metric", files=files, data=data)
    if response.ok:
        return response.json()
    else:
        return {"error": "Failed to get metrics from Go backend."}

def poll_prediction_status() -> dict:
    response = requests.get(f"{BACKEND_GO_URL}/poll")
    if response.ok:
        return response.json()
    else:
        return {"error": "Failed to get prediction status from Go backend."}