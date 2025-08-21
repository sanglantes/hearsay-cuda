FROM nvidia/cuda:12.1.0-cudnn8-runtime-ubuntu22.04

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app/api

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      python3 python3-venv python3-pip ca-certificates wget build-essential \
      libssl-dev libffi-dev libblas-dev liblapack-dev && \
    ln -s /usr/bin/python3 /usr/bin/python && \
    python -m pip install --upgrade pip setuptools wheel && \
    rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt .

RUN python -m pip install --no-cache-dir \
      "torch" --index-url https://download.pytorch.org/whl/cu121 && \
    python -m pip install --no-cache-dir -r requirements.txt
COPY . .

ENV MPLBACKEND=Agg