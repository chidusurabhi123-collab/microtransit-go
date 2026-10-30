from pathlib import Path

import joblib  # type: ignore
import numpy as np
from sklearn.linear_model import LinearRegression

BASE_DIR = Path(__file__).resolve().parent
MODEL_PATH = BASE_DIR / "model.pkl"

# fake data: distance, speed, hour, traffic -> eta
np.random.seed(42)
n = 2000
dist = np.random.uniform(0.5, 15, n)
speed = np.random.uniform(10, 45, n)
hour = np.random.randint(0, 24, n)
traffic = np.where(((hour >= 8) & (hour <= 10)) | ((hour >= 17) & (hour <= 19)), 1.4, 1.0)

eta = dist / speed * 60 * traffic + np.random.normal(0, 1, n)  #type: ignore

X = np.column_stack([dist, speed, hour, traffic])
y = eta

model = LinearRegression()
model.fit(X, y)
joblib.dump(model, MODEL_PATH)  # type: ignore
print(f"model saved to {MODEL_PATH}, score {model.score(X, y):.6f}")