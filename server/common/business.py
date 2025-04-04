import logging
from common.utils import Bet, load_bets, store_bets, has_won

def parse_batch(batchMessage):
    bets = []
    lines = batchMessage.strip().split("\n")
    for line in lines:
        trimmed = line.strip()
        if trimmed == "":
            continue
        parts = trimmed.split(",")
        if len(parts) != 6:
            raise ValueError(f"Invalid bet format: {line}")
        betObj = Bet(parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
        bets.append(betObj)
    return bets

def process_bets(allBets):
    return store_bets(allBets)

def find_winners():
    logging.info("action: sorteo | result: success")
    winners_dic = {}
    bets = load_bets()
    for bet in bets:
        if has_won(bet):
            agency = bet.agency
            if agency in winners_dic:
                winners_dic[agency].append(bet.document)
            else:
                winners_dic[agency] = [bet.document]
    return winners_dic
