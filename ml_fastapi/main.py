from fastapi import FastAPI
from dotenv import load_dotenv
from fastapi.middleware.cors import CORSMiddleware
from api.v1.endpoints import router as api_router
import os

load_dotenv()

origins = os.getenv("GO_BACKEND", "*").split(",")

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins, # Replace with backend URL 
    allow_credentials=True,
    allow_methods=["*"],  
    allow_headers=["*"],  
)

app.include_router(api_router, prefix="/api/v1")

@app.get("/health")
def health():
    return {"message": "FastAPI backend running"}