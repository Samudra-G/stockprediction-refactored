import streamlit as st
import pandas as pd
import numpy as np
import plotly.graph_objects as go
import plotly.express as px
import utils
import time

PREDICTION_CACHE = {}

def plot_moving_averages(metrics):
    if "ma100" in metrics and "ma200" in metrics:
        ma100 = pd.Series(metrics["ma100"])
        ma200 = pd.Series(metrics["ma200"])
        x_vals = list(range(len(ma100)))

        fig = go.Figure()
        fig.add_trace(go.Scatter(x=x_vals, y=ma100, mode='lines', line=dict(color='blue'), name='MA100'))
        fig.add_trace(go.Scatter(x=x_vals, y=ma200, mode='lines', line=dict(color='orange'), name='MA200'))

        fig.update_layout(
            autosize=True,
            width=None,
            height=400,
            margin=dict(l=20, r=20, t=40, b=20),
            title="Moving Averages",
            xaxis_title="Index",
            yaxis_title="Value",
            xaxis=dict(showgrid=True, gridcolor='lightgrey'),
            yaxis=dict(showgrid=True, gridcolor='lightgrey'),
            legend=dict(x=0, y=1, orientation="h")
        )

        with st.container():
            st.plotly_chart(fig, use_container_width=True)

def plot_rsi(metrics):
    if "rsi" in metrics and isinstance(metrics["rsi"], list):
        rsi_period = st.slider(
            "Select RSI display period (days)",
            min_value=5, max_value=50, value=14, step=1
        )
        rsi = pd.Series(metrics["rsi"][-rsi_period:])
        x_vals = list(range(len(rsi)))

        fig = go.Figure()
        fig.add_trace(go.Scatter(x=x_vals, y=rsi, mode='lines', line=dict(color='purple'), name='RSI'))
        fig.add_hline(y=70, line=dict(color='red', dash='dash'), annotation_text='Overbought')
        fig.add_hline(y=30, line=dict(color='green', dash='dash'), annotation_text='Oversold')

        fig.update_layout(
            title=f"RSI (Last {rsi_period} Days)",
            xaxis_title="Index",
            yaxis_title="RSI Value",
            xaxis=dict(showgrid=True, gridcolor='lightgrey'),
            yaxis=dict(showgrid=True, gridcolor='lightgrey'),
            legend=dict(x=0, y=1)
        )
        st.plotly_chart(fig, use_container_width=True)


def show_volatility(metrics):
    if "volatility" in metrics and isinstance(metrics["volatility"], (int, float)):
        baseline_vol = 0.02
        delta = metrics["volatility"] - baseline_vol
        st.metric(
            label="Volatility",
            value=round(metrics["volatility"], 4),
            delta=round(delta, 4)
        )


def plot_macd(metrics):
    if all(k in metrics for k in ["macd", "signal", "histogram"]):
        macd_period = st.slider(
            "Select MACD display period (days)",
            min_value=10, max_value=52, value=26, step=1
        )

        macd_line = metrics["macd"][-macd_period:]
        signal_line = metrics["signal"][-macd_period:]
        histogram = metrics["histogram"][-macd_period:]
        x_vals = list(range(len(macd_line)))

        fig = go.Figure()
        fig.add_trace(go.Scatter(x=x_vals, y=macd_line, mode='lines', line=dict(color='blue'), name='MACD Line'))
        fig.add_trace(go.Scatter(x=x_vals, y=signal_line, mode='lines', line=dict(color='orange', dash='dash'), name='Signal Line'))
        fig.add_trace(go.Bar(x=x_vals, y=histogram, marker_color=['green' if val >= 0 else 'red' for val in histogram], name='Histogram', opacity=0.6))

        fig.update_layout(
            title=f"MACD (Last {macd_period} Days)",
            xaxis_title="Index",
            yaxis_title="Value",
            xaxis=dict(showgrid=True, gridcolor='lightgrey'),
            yaxis=dict(showgrid=True, gridcolor='lightgrey'),
            legend=dict(x=0, y=1)
        )
        st.plotly_chart(fig, use_container_width=True)


def plot_prediction_vs_actual(ticker: str):
    st.subheader("📈 Prediction vs Actual")

    with st.spinner("Waiting for prediction to complete..."):
        status = ""
        result = {}

        while status not in ["success", "failed"]:
            result = utils.poll_prediction_status(ticker)
            status = result.get("status", "")
            if status in ["success", "failed"]:
                break
            time.sleep(2)

    if status == "success":
        inner = result["predictions"]
        preds = np.array(inner["predictions"])
        y_true = np.array(inner["y_test"])
        dates = pd.to_datetime(inner["dates"])

        PREDICTION_CACHE["preds"] = preds
        PREDICTION_CACHE["y_true"] = y_true

        df = pd.DataFrame({
            "Date": dates,
            "Predicted": preds.flatten(),
            "Actual": y_true.flatten()
        })

        df.sort_values("Date", inplace=True)

        years = st.slider("Select how many years to display:", 1, 5, 3)
        cutoff_date = df["Date"].max() - pd.DateOffset(years=years)
        df_recent = df[df["Date"] >= cutoff_date]

        fig = px.line(
            df_recent,
            x="Date",
            y=["Actual", "Predicted"],
            labels={"value": "Stock Price", "Date": "Date"},
            title="Predicted vs Actual Stock Prices",
        )

        for trace in fig.data:
            if trace.name == "Actual":
                trace.line.color = "green"
            elif trace.name == "Predicted":
                trace.line.color = "orange"

        fig.update_layout(
            legend=dict(title="Legend"),
            xaxis_title="Date",
            yaxis_title="Price",
            template="plotly_white",
            hovermode="x unified"
        )

        st.plotly_chart(fig, use_container_width=True)

    else:
        st.error("Prediction failed or not available.")
