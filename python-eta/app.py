from fastapi import FastAPI
from pydantic import BaseModel
from models import load_model, predict_eta
import datetime

app = FastAPI(title="MicroTransit ETA Brain")
load_model()

class Req(BaseModel):
    distance_km: float
    speed_kmh: float

@app.get("/health")
def health(): # type: ignore
    return {"status":"ok", "model_loaded": True} # type: ignore

@app.post("/predict")
def predict(r: Req): # type: ignore
    hour = datetime.datetime.now().hour
    eta = predict_eta(r.distance_km, r.speed_kmh, hour)
    return {"eta_minutes": round(max(eta,0.5),1), "hour": hour} # type: ignore