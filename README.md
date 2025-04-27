# 🚀 CloudPulse-AI

![Build](https://img.shields.io/badge/build-passing-brightgreen)

> Real-time AI model monitoring agent with system metrics tracking, anomaly detection, AWS CloudWatch integration, and a dynamic dashboard.

---

## 🌟 Overview

CloudPulse-AI is a lightweight Go-based monitoring agent designed to track AI model performance, detect anomalies (drift), monitor system health (CPU, memory, uptime), and display everything in a live dashboard.  
Logs and alerts are pushed to AWS CloudWatch for centralized cloud monitoring.

---

## 📈 Features

- **Real-time Monitoring**
  - Model prediction monitoring
  - Response time tracking
  - Confidence score logging
  - Model version tracking
- **Drift Detection**
  - Sudden prediction flipping
  - Confidence score drops
  - Response time spikes
- **System Metrics**
  - CPU usage
  - Memory usage
  - Server uptime
- **AWS Cloud Integration**
  - Pushes logs and drift alerts to **AWS CloudWatch Logs**
- **Responsive Dashboard**
  - View predictions, response times, drift warnings
  - Beautiful interactive charts using Chart.js and Bootstrap 5
- **Error Handling**
  - Logs detailed errors to CloudWatch
  - Shows system and drift warnings in dashboard
- **Dockerized Deployment**
  - Runs easily in containers with Docker and Docker Compose
- **Unit Tests + GitHub Actions CI/CD**
  - Includes basic unit tests
  - Automated build/test pipeline ready

---

## ⚙️ Tech Stack

| Technology    | Usage                             |
|---------------|-----------------------------------|
| **Go**        | Backend server, metrics collection |
| **HTML/CSS/JS** | Dashboard frontend (Bootstrap, Chart.js) |
| **AWS CloudWatch** | Centralized logs and alerts |
| **gopsutil**  | Real system metrics (CPU, memory) |
| **Docker**    | Containerization of the agent |
| **GitHub Actions** | CI/CD pipelines (build + test) |

---

## ⚙️ Project Structure

```
cloudpulse-ai/
├── model-api/
│   └── app.py (mock ML API)
│   └── Dockerfile
├── monitor-agent/
│   ├── main.go (Monitoring agent)
│   ├── monitor-agent_test.go (Unit tests)
│   ├── templates/
│   │   └── dashboard.html
│   ├── static/
│   │   └── chart.min.js
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env (AWS credentials)
├── .gitignore
└── README.md
```

---

## 📦 How to Run

### 1. Clone the repo

```bash
git clone https://github.com/yourusername/CloudPulse-AI.git
cd cloudpulse-ai/monitor-agent
```

---

### 2. Set up environment variables

Create a `.env` file with your AWS credentials:

```env
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

---

### 3. Build and Run (Docker Compose)

```bash
docker-compose up --build
```

Access dashboard at ➡️ [http://localhost:8080](http://localhost:8080)

---

## 🧪 Running Tests

```bash
go test ./...
```

Unit tests are provided to validate core functionality like drift detection and API responses.

---

## 🛡 Security

- No AWS credentials are hard-coded.
- `.env` is gitignored.
- Minimal Docker images.
- Non-root Docker user.
- Regular system metrics updates.

---

## 🛠 Future Enhancements

- Multi-model endpoint support
- Slack or Email notifications
- Advanced drift detection (KL divergence, etc.)
- Admin dashboard authentication
- Kubernetes deployment option

---

## ✨ Credits

Built by Hamza Ajmal

---