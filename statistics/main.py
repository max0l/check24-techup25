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
rows_without_real_hotel = 0

double_total = 0
double_children0 = 0


for chunk in pd.read_csv(FILE_PATH, chunksize=CHUNKSIZE):
    total_rows += len(chunk)
    for col in columns:
        counters[col].update(chunk[col].astype(str))

    rows_without_real_hotel += (chunk['hotelid'].astype(int) > 998).sum()

    mask_double = chunk['roomtype'].astype(str) == 'double'
    double_total += mask_double.sum()
    double_children0 += ((chunk['countchildren'] == 0) & mask_double).sum()

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
print(f"Anzahl Zeilen ohne echten Hotel: {rows_without_real_hotel} ({rows_without_real_hotel/total_rows:.2%})")