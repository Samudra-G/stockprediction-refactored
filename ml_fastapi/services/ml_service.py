import io
import pandas as pd
import tensorflow as tf
from .helpers import prepare_test_data, make_predictions
from sklearn.preprocessing import MinMaxScaler
import math

model_path = "model/lstm_model.keras"
_model = tf.keras.models.load_model(model_path)

def predict_stock(file_bytes: bytes) -> dict:
    try:
        df = pd.read_csv(io.BytesIO(file_bytes))
    except Exception as e:
        return {"error": f"Failed to read CSV: {str(e)}"}

    if df.empty or "Close" not in df.columns:
        return {"error": "CSV must contain 'Close' column"}

    df['Date'] = pd.to_datetime(df['Date'])
    df.set_index('Date', inplace=True)

    data = df.filter(['Close'])
    dataset = data.values

    train_data_len = math.ceil(len(dataset) * 0.8)

    scaler = MinMaxScaler(feature_range=(0, 1))
    scaler = scaler.fit(dataset) 

    X_test, y_test = prepare_test_data(df, dataset, scaler, train_data_len)

    if len(X_test) == 0:
        return {"error": "Not enough data for prediction"}

    predictions, y_test_true, date_range = make_predictions(_model, X_test, scaler, df, train_data_len)
    dates = pd.to_datetime(date_range, utc=True)
    return {
        "predictions": predictions.flatten().tolist(),
        "y_test": y_test_true.flatten().tolist(),
        "dates": dates.strftime("%Y-%m-%d").tolist()
    }
