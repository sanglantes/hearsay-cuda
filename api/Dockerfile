FROM nvidia/cuda:12.8.0-cudnn-runtime-ubuntu22.04

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app/api

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      python3.11 python3.11-venv python3-pip ca-certificates wget build-essential \
      libssl-dev libffi-dev libblas-dev liblapack-dev libopenblas-dev gfortran && \
    ln -sf /usr/bin/python3.11 /usr/bin/python && \
    rm -rf /var/lib/apt/lists/*

RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

RUN pip install --upgrade pip setuptools wheel

COPY ./requirements.txt .

RUN pip install --no-cache-dir "torch" --index-url https://download.pytorch.org/whl/cu128 && \
    pip install --no-cache-dir --force-reinstall numpy scipy && \
    pip install --no-cache-dir -r requirements.txt


RUN apt-get remove -y build-essential gfortran && apt-get autoremove -y && rm -rf /var/lib/apt/lists/* /root/.cache/pip

ENV MPLBACKEND=Agg

COPY . .