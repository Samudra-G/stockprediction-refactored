from fastapi import APIRouter, UploadFile, File, HTTPException
from services.ml_service import predict_stock
from api.v1.schemas import PredictResponse

router = APIRouter()

@router.post("/predict", response_model=PredictResponse)
async def predict(file: UploadFile = File(...)):
    contents = await file.read()
    result = predict_stock(contents)

    if "error" in result:
        raise HTTPException(status_code=400, detail=result["error"])

    return result
