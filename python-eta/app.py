from fastapi import FastAPI
from pydantic import BaseModel
from models import load_model, predict_eta

app = FastAPI(title="MicroTransit ETA")

class Req(BaseModel):
    distance: float
    speed: float

# don't crash on startup
load_model()

@app.get("/")
def root():
    return {"status": "ok", "service": "python-eta"}

@app.post("/predict")
def predict(req: Req):
    eta_min = predict_eta(req.distance, req.speed)
    return {"eta_minutes": eta_min, "distance": req.distance, "speed": req.speed}