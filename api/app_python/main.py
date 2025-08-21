import importlib
import os
from fastapi import FastAPI
from fastapi.responses import JSONResponse
import requests

from typing import Optional
from pydantic import BaseModel

import s_readability as s_readability
import _sentiment
import joblib
import time, requests

current_period = lambda: int(time.time() // 7200)

app = FastAPI()

@app.get("/ping")
def pong() -> dict[str, str]:
    return {"ping": "pong"}


@app.get(
    "/readability",
    summary="Return the Flesch-Kincaid score of a nick."
)
def readability(nick: str) -> JSONResponse:
    import s_readability as s_readability

    return JSONResponse({
        "score": s_readability.flesch_score(nick, current_period())
    })


@app.get(
    "/retrain",
    summary="Retrain the model. Optionally return a confusion-matrix link."
)
def retrain(
    min_messages: int,
    cm: Optional[int] = 0,
    cf: Optional[int] = 0,
    bert: Optional[int] = 0,
    gpu: Optional[int] = 0
) -> JSONResponse:
    import s_retrain
    cm = bool(cm)

    if importlib.util.find_spec("sentence_transformers") is None:
        bert = 0
    pipeline = s_retrain.create_pipeline(1, bert, gpu)

    X, y = s_retrain.get_X_y(min_messages, cf)
    start = time.time()
    pipeline.fit(X, y)
    elapsed = time.time() - start
    joblib.dump(pipeline, "/app/data/pipeline.joblib")
    joblib.dump(pipeline.named_steps["clf"].classes_.astype(str), "/app/data/labels.joblib")

    url = ""
    accuracy = 0.0
    f1 = 0.0
    if cm:
        cm_table, labels, accuracy, f1 = s_retrain.evaluate_pipeline(pipeline, X, y)
        
        joblib.dump(cm_table, "/app/data/cm_table.joblib") # Used for me command.

        s_retrain.plot_and_save_confusion_matrix(cm_table, labels)
        with open("cm.png", "rb") as f:
            resp = requests.post("https://tmpfiles.org/api/v1/upload", files={"file": f})
            respj = resp.json()
            if respj["status"] == "success":
                url = respj["data"]["url"]
            else:
                url = "failed"

    return JSONResponse(content={
        "time": elapsed,
        "url": url,
        "accuracy": accuracy,
        "f1": f1
    })


class AttributeRequest(BaseModel):
    msg: str
    min_messages: int
    confidence: bool = False
@app.post(
    "/attribute",
    summary="Attribute a message to a chatter."
)
def attribute(req: AttributeRequest) -> JSONResponse:
    import s_retrain

    if not os.path.exists("/app/data/pipeline.joblib"):
        pipeline = s_retrain.create_pipeline()
        X, y = s_retrain.get_X_y(req.min_messages)
        pipeline.fit(X, y)

        joblib.dump(pipeline, "/app/data/pipeline.joblib")
    else:
        pipeline = joblib.load("/app/data/pipeline.joblib")
    
    author = pipeline.predict([req.msg])[0]

    if req.confidence:
        confidence = pipeline.decision_function([req.msg]).tolist()[0]

        labels = map(str, pipeline.named_steps["clf"].classes_)

        conf_map = dict(zip(labels, confidence))
        conf_map = sorted(conf_map.items(), key=lambda x: x[1], reverse=True)[:3]
        conf_str = ', '.join(f"{lc[0]}_ ({lc[1]:.2f})" for lc in conf_map)

    return JSONResponse(content={"author": author, "confidence": conf_str})


@app.post(
    "/profile_attribute",
    summary="Attribute a profile to a chatter."
)
def attribute(req: AttributeRequest) -> JSONResponse:
    import s_retrain

    group_k = len(req.msg.split("/:MSG/"))
    pipeline = s_retrain.create_pipeline(group_k)
    X, y = s_retrain.get_X_y_block(req.min_messages, 0, group_k, current_period())
    pipeline.fit(X, y)
    
    author = pipeline.predict([req.msg.replace("/:MSG/", "   ")])[0]

    if req.confidence:
        confidence = pipeline.decision_function([req.msg.replace("/:MSG/", "   ")]).tolist()[0]

        labels = map(str, pipeline.named_steps["clf"].classes_)

        conf_map = dict(zip(labels, confidence))
        conf_map = sorted(conf_map.items(), key=lambda x: x[1], reverse=True)[:3]
        conf_str = ', '.join(f"{lc[0]}_ ({lc[1]:.2f})" for lc in conf_map)

    return JSONResponse(content={"author": author, "confidence": conf_str})


@app.get(
    "/attribute_list",
    summary="List current nicks in the model."
)
def attribute_list() -> JSONResponse:
    try:
        labels = joblib.load("/app/data/labels.joblib")
        return JSONResponse(content={"authors":'_, '.join(labels)+'_'})
    except FileNotFoundError:
        return JSONResponse(content={"authors": "No nicks were found. Retrain the model."})
    except Exception as e:
        return JSONResponse(content={"authors": str(e)})


class sentimentRequest(BaseModel):
    msg: str
@app.post(
    "/sentiment",
    summary="Extract the sentiment from a message."
)
def sentiment(req: sentimentRequest):
    from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
    
    analyzer = SentimentIntensityAnalyzer()
    vs = analyzer.polarity_scores(req.msg)

    comp = vs["compound"]
    if comp >= 0.05:
        hr = "positive"
    elif -0.05 < comp < 0.05:
        hr = "neutral"
    else:
        hr = "negative"

    return JSONResponse(content={
        "pos": vs["pos"],
        "neu": vs["neu"],
        "neg": vs["neg"],
        "hr": hr,
        "compound": comp
    })


@app.get(
    "/me",
    summary="Information about you."
)
def me(author: str) -> JSONResponse:
    comp = _sentiment.sentiment_over_many_messages(author, current_period())
    if comp >= 0.05:
        hr = "Positive"
    elif -0.05 < comp < 0.05:
        hr = "Neutral"
    else:
        hr = "Negative"
        
    try:
        labels = joblib.load("/app/data/labels.joblib").tolist()
        cm = joblib.load("/app/data/cm_table.joblib")

        you = labels.index(author)
        mconfusion = cm[you, :] + cm[:, you]
        mconfusion[you] = -1

        neighbour_no = int(mconfusion.argmax())
        neighbour =  labels[neighbour_no]+'_'
    except Exception as e:
        print(e)
        neighbour = "Train the attribution model with evaluation for results."

    return JSONResponse(content={
        "readability": s_readability.flesch_score(author, current_period()),
        "sentiment": comp,
        "sentiment_hr": hr,
        "neighbour": neighbour
    })