import os

try:
    import joblib  # type: ignore
except ImportError:
    joblib = None  # type: ignore

MODEL_PATH = os.path.join(os.path.dirname(__file__), "eta_model.pkl")
model = None  # type: ignore

def load_model():
    global model
    if not os.path.exists(MODEL_PATH):
        print(f"[ETA] No model file, using fallback")
        return None
    if joblib is None:
        return None
    try:
        model = joblib.load(MODEL_PATH)  # type: ignore
        print("[ETA] Model loaded")
        return model
    except Exception as e:
        print(f"[ETA] Failed to load ({e}), fallback")
        model = None
        return None

def predict_eta(distance_km: float, speed_kmh: float) -> float:
    if speed_kmh < 1:
        return 0
    if model is not None:
        try:
            return float(model.predict([[distance_km, speed_kmh]])[0])  # type: ignore
        except:
            pass
    return distance_km / speed_kmh * 60