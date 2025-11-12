FROM python:3.11-slim

WORKDIR /app

# Copy requirements and install
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy code
COPY . .

# Environment variable (path to key inside container)
ENV FIREBASE_CREDENTIALS=/run/secrets/firebase.json

EXPOSE 50061
CMD ["python", "server.py"]

