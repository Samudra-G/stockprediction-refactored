import streamlit as st
import pandas as pd
import numpy as np
import utils
from plots import plot_moving_averages, plot_rsi, plot_macd, plot_prediction_vs_actual, show_volatility


def main():
    st.title("Stock Analysis and Prediction App")

    ticker = st.text_input("Enter a stock ticker (e.g., AAPL):", "AAPL").upper()

    if not ticker:
        return

    df = utils.fetch_stock_data(ticker)
    if df is None or df.empty:
        st.error(f"Invalid or no data found for ticker {ticker}")
        return

    st.subheader("Descriptive Statistics")
    st.write(df.describe())

    metrics = utils.send_metrics_to_go(df, ticker)
    if "error" in metrics:
        st.error(metrics["error"])
        return

    st.subheader(f"Metrics for {ticker}")
    # st.json(metrics)  # Debug if needed

    plot_moving_averages(metrics)
    plot_rsi(metrics)
    show_volatility(metrics)
    plot_macd(metrics)
    plot_prediction_vs_actual(ticker)


if __name__ == "__main__":
    main()
