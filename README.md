# 📈 Stock Price Prediction Microservice

This project is a **containerized, microservice-based stock price prediction system** combining **FastAPI**, **Go**, and **Streamlit**. It demonstrates a practical, production-like architecture for serving an **LSTM time series model** with a clean UI and asynchronous processing.

---

## 🚀 Features

* **LSTM Model**: Predicts future stock closing prices automatically.
* **FastAPI Service**: Serves the trained model, handles data preprocessing, and returns predictions.
* **Go Backend**: Orchestrates requests, handles metric logging, and manages polling.
* **Streamlit Frontend**: User only needs to select a stock ticker — no CSV upload required!
* **Asynchronous Polling**: Go polls FastAPI in the background so the UI remains responsive.
* **Fully Dockerized**: Each component runs in its own container.

---

## ⚙️ Tech Stack

* **FastAPI** — Python backend for ML inference.
* **TensorFlow/Keras** — LSTM model.
* **Go (Gin)** — Fast backend for metrics and coordination.
* **Streamlit** — Interactive frontend.
* **Docker & Docker Compose** — Container orchestration.

---

## ⚙️ What’s inside?

* ml\_fastapi/ — FastAPI service for LSTM predictions.
* backend-go/ — Go service handling metrics & async requests.
* frontend/ — Streamlit UI for users to pick tickers and view results.
* docker-compose.yml — Orchestrates everything.
* Makefile — Run your stack with simple commands like `make up`.

---

## 📂 Project Structure

```plaintext
.
├── Makefile
├── docker-compose.yml
├── backend-go/
├── ml_fastapi/
│   ├── model/   # Place your downloaded .keras model here!
├── frontend/
└── LICENSE
```

---

## ✅ How It Works

1. **Pick Stock Ticker** — The Streamlit app lets the user select a stock.
2. **Go Backend Calls FastAPI** — The Go service prepares metrics and forwards the ticker to FastAPI.
3. **FastAPI Predicts** — Loads the LSTM, fetches historical data, returns predictions.
4. **Polling** — Go polls FastAPI while Streamlit shows progress.
5. **Results Rendered** — Streamlit displays predictions vs. actuals with interactive Plotly charts.

---

## 🚢 Deployment

Runs fully in Docker. Example:

```bash
make up
```

Just make sure to place the shared [model file](https://drive.google.com/file/d/1LSpZ__JbnbioPlqMHCfirp37NvJT3ld4/view?usp=sharing) in `ml_fastapi/model/` before starting.

---

## 📊 Future Improvements

* Add database support for multi-user jobs.
* Add more advanced metrics and logs.
* Expand tests.
* Automate with CI/CD.

---

## 📺 Live Demo

🚀 [**Try it live → Click here!**](https://stocksamudra.onrender.com)

Predict stock prices in real time — pick a ticker, run the LSTM model, and see interactive Plotly charts comparing predictions vs. actuals.

[![Live Demo](https://img.shields.io/badge/Live%20App-Open%20Now-brightgreen?style=for-the-badge&logo=render)](https://stocksamudra.onrender.com)

---

## 🏆 Credits

Built with ❤️ by Samudra-G.

**MIT License** — use, modify, and learn freely! Just don't forget to credit me!
