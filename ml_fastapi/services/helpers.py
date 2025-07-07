import pandas as pd
import numpy as np
import math
from sklearn.preprocessing import MinMaxScaler

def preprocess_data(df):
    df['Date'] = pd.to_datetime(df['Date'], utc=True)
    df.set_index('Date', inplace=True)

    data = df.filter(['Close'])
    dataset = data.values

    train_data_len = math.ceil(len(dataset) * 0.8)

    scaler = MinMaxScaler(feature_range=(0, 1))
    scaled_data = scaler.fit_transform(dataset)

    train_data = scaled_data[0:train_data_len, :]

    X_train = []
    y_train = []

    for i in range(60, len(train_data)):
        X_train.append(train_data[i-60:i, 0])
        y_train.append(train_data[i, 0])

    X_train, y_train = np.array(X_train), np.array(y_train)
    X_train = np.reshape(X_train, (X_train.shape[0], X_train.shape[1], 1))

    return X_train, y_train, scaler, train_data_len, dataset

def prepare_test_data(df, dataset, scaler, train_data_len, timesteps=60):
    data = df.filter(['Close']).values
    scaled_data = scaler.transform(data)

    X_test = []
    for i in range(train_data_len, len(scaled_data)):
        X_test.append(scaled_data[i-timesteps:i, 0])

    X_test = np.array(X_test)
    if X_test.size == 0:
        return np.array([]), np.array([])

    X_test = np.reshape(X_test, (X_test.shape[0], X_test.shape[1], 1))
    y_test = dataset[train_data_len:, :]

    return X_test, y_test

def make_predictions(model, X_test, scaler, df, train_data_len, timesteps=60):
    predictions = model.predict(X_test)
    predictions = scaler.inverse_transform(predictions)

    scaled_data = scaler.transform(df.filter(['Close']).values)
    y_test_scaled = scaled_data[train_data_len:, :]
    y_test_scaled = y_test_scaled[timesteps:]
    y_test = scaler.inverse_transform(y_test_scaled)

    if len(predictions) != len(y_test):
        min_len = min(len(predictions), len(y_test))
        predictions = predictions[:min_len]
        y_test = y_test[:min_len]

    date_range = df.index[train_data_len + timesteps : train_data_len + timesteps + len(predictions)]
    print(f"Predictions: {len(predictions)}, Y_test: {len(y_test)}, Dates: {len(date_range)}")

    return predictions, y_test, date_range
