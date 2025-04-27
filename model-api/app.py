from flask import Flask, request, jsonify
import time
import random

app = Flask(__name__)

# Simulated model: returns "positive" or "negative" randomly
@app.route("/predict", methods=["POST"])
def predict():
    start_time = time.time()

    # Simulate model processing time
    time.sleep(random.uniform(0.1, 0.5))

    # Simple fake prediction
    output = random.choice(["positive", "negative"])

    response_time = round(time.time() - start_time, 4)

    return jsonify({
        "prediction": output,
        "response_time": response_time
    })

@app.route("/", methods=["GET"])
def health_check():
    return "Model API is running!"

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
