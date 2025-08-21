import database
from joblib import Memory

memory = Memory("./cache")

@memory.cache
def sentiment_over_many_messages(nick: str, expire: int = 0):
    from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
    
    analyzer = SentimentIntensityAnalyzer()

    msgs = database.get_messages_from_nick(nick)
    average_compound_score = sum(analyzer.polarity_scores(v)["compound"] for v in msgs)/len(msgs)

    return average_compound_score