import pandas as pd
from collections import Counter, defaultdict

FILE_PATH = '../offers/offer.csv'
CHUNKSIZE = 500000

columns = [
    'hotelid', 'countadults', 'countchildren', 'inbounddepartureairport',
    'inboundarrivalairport', 'outbounddepartureairport', 'outboundarrivalairport',
    'mealtype', 'oceanview', 'roomtype', 'price'
]

counters = {col: Counter() for col in columns}
total_rows = 0

# NEU: Zähler für die Zusatzanalyse
double_total = 0
double_children0 = 0

for chunk in pd.read_csv(FILE_PATH, chunksize=CHUNKSIZE):
    total_rows += len(chunk)
    for col in columns:
        counters[col].update(chunk[col].astype(str))

    # Zusatzanalyse
    mask_double = chunk['roomtype'].astype(str) == 'double'
    double_total += mask_double.sum()
    double_children0 += ((chunk['countchildren'] == 0) & mask_double).sum()

# Ergebnisse anzeigen
for col in columns:
    print(f"\nSpalte: {col}")
    top5 = counters[col].most_common(5)
    print("Top 5 Werte und ihr Anteil:")
    for value, count in top5:
        print(f"  {value}: {count} mal ({count/total_rows:.2%})")
    print(f"Verschiedene Werte insgesamt: {len(counters[col])}")

print(f"\n---\nZusatzanalyse für roomtype 'double':")
print(f"Anzahl Angebote mit roomtype 'double': {double_total}")
print(f"Davon mit countchildren == 0: {double_children0} ({double_children0/double_total:.2%})")
print(f"Davon mit countchildren > 0: {double_total - double_children0} ({(double_total-double_children0)/double_total:.2%})")