# 📈 Stock Price Prediction Microservice

This project is a **containerized, microservice-based stock price prediction system** combining **FastAPI**, **Go**, and **Streamlit**. It demonstrates a practical, production-like architecture for serving an **LSTM time series model** with a clean UI and asynchronous processing.

---

## 🚀 Features

* **LSTM Model**: Predicts future stock closing prices based on uploaded CSV data.
* **FastAPI Service**: Serves the trained model, handles data preprocessing, and returns predictions.
* **Go Backend**: Orchestrates requests, handles metric logging, and manages polling.
* **Streamlit Frontend**: Displays results interactively with clear, responsive visualizations (Plotly).
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

⚙️ What’s inside?

* ml_fastapi/ — FastAPI service for heavy LSTM predictions (TensorFlow).

* backend-go/ — Go service handling metrics & concurrency.

* frontend/ — Streamlit UI that polls the backend & shows results.

* docker-compose.yml — Orchestrates everything.

* Makefile — Run your stack with simple commands like:

* make up — Start all services.

* make down — Stop all services.

* make restart — Restart everything.

---

## 📂 Project Structure

```plaintext
.
├── Makefile
├── docker-compose.yml
├── backend-go/
│   ├── cmd/
│   ├── api/
│   ├── pkg/
│   └── utils/
├── ml_fastapi/
│   ├── api/v1/
│   ├── model/
│   ├── services/
│   ├── requirements.txt
│   └── main.py
├── frontend/
│   ├── main.py
│   ├── plots.py
│   ├── utils.py
│   ├── requirements.txt
└── LICENSE

```

---

## ✅ How It Works

1. **Put Stock Ticker** — Streamlit lets the user select a stock ticker.
2. **Go Backend Calls FastAPI** — The Go service computes metrics and forwards the data to FastAPI.
3. **FastAPI Predicts** — Loads the LSTM, preprocesses data, returns predictions.
4. **Polling** — Go polls FastAPI status while Streamlit shows a spinner.
5. **Results Rendered** — Streamlit displays predictions vs. actuals interactively with Plotly.

---

## 🚢 Deployment

This project runs **entirely in Docker containers**. Example commands:

```bash
make up
```

All secrets (DB, keys, endpoints) are managed via `.env` and injected with Docker Compose.

---

## 🗂️ Environment & Secrets

Your `.gitignore` excludes:

* Python caches
* `.env` with secrets
* Large models

Be sure to keep your `.env` and model files secure.

---

## 📊 Future Improvements

* Add database support for multi-user jobs.
* Add more metrics and logging.
* Include unit tests.
* Deploy with CI/CD.

---

## 🏆 Credits

Built with ❤️ by \Samudra-G.

---

**MIT License** — use, modify, and learn freely!
