from pydantic import BaseModel
from typing import List

class PredictResponse(BaseModel):
    predictions: List[float]
    y_test: List[float]
    dates: List[str]