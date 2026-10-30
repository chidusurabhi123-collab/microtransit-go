from pathlib import Path

import joblib
import numpy as np

BASE_DIR = Path(__file__).resolve().parent
MODEL_PATH = BASE_DIR / "model.pkl"
model = None


def load_model():
    global model
    if MODEL_PATH.exists():
        model = joblib.load(MODEL_PATH)
    return model


def predict_eta(distance_km: float, speed_kmh: float, hour: int):
    global model
    if model is None:
        load_model()
    traffic = 1.4 if hour in [8, 9, 10, 17, 18, 19] else 1.0
    # fallback if no model
    if model is None:
        return distance_km / max(speed_kmh, 5) * 60 * traffic
    X = np.array([[distance_km, speed_kmh, hour, traffic]], dtype=float)
    return float(model.predict(X)[0])