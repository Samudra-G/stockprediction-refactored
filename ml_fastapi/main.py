from fastapi import FastAPI
from dotenv import load_dotenv
from fastapi.middleware.cors import CORSMiddleware
from api.v1.endpoints import router as api_router
import os

load_dotenv()

#origins = os.getenv("CLIENT_URL", "*").split(",")

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"], # Replace with frontend/client URL 
    allow_credentials=True,
    allow_methods=["*"],  
    allow_headers=["*"],  
)

app.include_router(api_router, prefix="/api/v1")