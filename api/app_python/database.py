import sqlite3
import os
from joblib import Memory
from collections import defaultdict

memory = Memory("./cache")

DP = "/app/data/database.db"
DB_TIMESTAMP = lambda: int(os.path.getmtime(DP)) // 1000

def get_connection() -> sqlite3.Connection:
    return sqlite3.connect(DP)

def get_nicks_with_x_plus_messages(x: int) -> list[str]:
    with sqlite3.connect(DP) as conn:
        res = conn.execute("SELECT nick FROM messages GROUP BY nick HAVING COUNT(*) > ?", (x,))
        return [u[0] for u in res]
    
@memory.cache
def get_messages_with_x_plus_messages(x: int, cf: int = 0, DBT: int = DB_TIMESTAMP()) -> dict[str, list[str]]:
    author_message = defaultdict(list)

    base_query = """
        WITH eligible_authors AS (
            SELECT m.nick
            FROM messages m
            JOIN users u ON m.nick = u.nick
            WHERE u.opt = 1
            GROUP BY m.nick
            HAVING COUNT(*) >= ? 
               AND (? = 0 OR MAX(m.time) > datetime('now', '-' || ? || ' days'))
        ),
        ranked_messages AS (
            SELECT m.nick,
                   m.message,
                   ROW_NUMBER() OVER (PARTITION BY m.nick ORDER BY m.time DESC) AS rn
            FROM messages m
            JOIN eligible_authors ea ON m.nick = ea.nick
        )
        SELECT nick, message
        FROM ranked_messages
        WHERE rn <= 10000
    """
    params = (x, cf, cf)

    with sqlite3.connect(DP) as conn:
        res = conn.execute(base_query, params)
        for nick, message in res:
            author_message[nick].append(message)

    return author_message


@memory.cache
def get_messages_from_nick(nick: str, DBT: int = DB_TIMESTAMP()) -> list[str]:
    with sqlite3.connect(DP) as conn:
        res = conn.execute("SELECT message FROM messages WHERE nick = ? ORDER BY id DESC LIMIT 10000", (nick,))
        return [msg[0] for msg in res]

def is_nick_eligible(count: int, nick: str) -> bool:
    with sqlite3.connect(DP) as conn:
        res = conn.execute("""SELECT
                           COUNT(*)
                           FROM messages m
                           JOIN users u WHERE u.nick = m.nick
                           AND opt = 1
                           AND m.nick = (?)
                           """, (nick,))
        return int(res.fetchone()[0]) >= count