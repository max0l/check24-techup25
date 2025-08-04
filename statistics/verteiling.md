# Verteilung der Spalten

Mit der berechnung der Verteilung der Spalteninhalten, kann man eine art baum Erstellen, um so schneller nach eintragen filtern zu können.

Allerdings ist nicht jede Spalte "suchbar". Es sind nur folgende Spalten suchbar:
```go
type OfferQuerry struct {
	DepartureAirports     []string  
	EarliestDepartureDate time.Time 
	LatestReturnDate      time.Time 
	CountAdults           int       
	CountChildren         int
	Duration              int
}
```
Es ist aber nicht schlau nach den Dates zu filtern, da diese zu unterschiedlich sind.
Diese sollten entweder nur nach Wochen getrennt werden oder durch die Dauer getrennt werden.

Das Filtern/Suchen mithilfe der Dauer könnte aber länger dauern, es sei denn man kombiniert das auch noch. Wobei hier
aber auch auf die Art des Filterns geachtet werden muss. Das beste wäre wohl, wenn man nach Woche filtert (also 1.1 bis 
7.1, 8.1 bis 14.1 usw.) und dann da noch die Dauer hat.

Es wäre also Folgende Reihenfolge:
1. countchildren
2. 

## Wichtigste Spalten

### Spalte: outbounddepartureairport
Top 5 Werte und ihr Anteil:
DUS: 21513370 mal (23.19%)
CGN: 9518179 mal (10.26%)
STR: 9501368 mal (10.24%)
FRA: 9410366 mal (10.14%)
HAM: 7689193 mal (8.29%)
Verschiedene Werte insgesamt: 48

### Spalte: countadults
Top 5 Werte und ihr Anteil:
2: 72075942 mal (77.68%)
1: 9555560 mal (10.30%)
3: 5416075 mal (5.84%)
4: 3660763 mal (3.95%)
6: 763949 mal (0.82%)
Verschiedene Werte insgesamt: 20

### Spalte: countchildren
Top 5 Werte und ihr Anteil:
0: 54951599 mal (59.23%)
1: 18925609 mal (20.40%)
2: 15975687 mal (17.22%)
3: 2434714 mal (2.62%)
4: 431635 mal (0.47%)
Verschiedene Werte insgesamt: 12

### Spalte: roomtype
Top 5 Werte und ihr Anteil:
double: 51803015 mal (55.83%)
apartment: 15255130 mal (16.44%)
family: 5394444 mal (5.81%)
juniorsuite: 4861063 mal (5.24%)
suite: 4628103 mal (4.99%)
Verschiedene Werte insgesamt: 19

### Zusatzanalyse für roomtype 'double'
Anzahl Angebote mit roomtype 'double': 51803015
Davon mit countchildren == 0: 38126666 (73.60%)
Davon mit countchildren > 0: 13676349 (26.40%)

## Alle Daten

Spalte: hotelid
Top 5 Werte und ihr Anteil:
1402: 1329479 mal (1.43%)
7536: 1286445 mal (1.39%)
1462: 1154744 mal (1.24%)
1347: 1072598 mal (1.16%)
6718: 886708 mal (0.96%)
Verschiedene Werte insgesamt: 960

Spalte: countadults
Top 5 Werte und ihr Anteil:
2: 72075942 mal (77.68%)
1: 9555560 mal (10.30%)
3: 5416075 mal (5.84%)
4: 3660763 mal (3.95%)
6: 763949 mal (0.82%)
Verschiedene Werte insgesamt: 20

Spalte: countchildren
Top 5 Werte und ihr Anteil:
0: 54951599 mal (59.23%)
1: 18925609 mal (20.40%)
2: 15975687 mal (17.22%)
3: 2434714 mal (2.62%)
4: 431635 mal (0.47%)
Verschiedene Werte insgesamt: 12

Spalte: inbounddepartureairport
Top 5 Werte und ihr Anteil:
PMI: 92783430 mal (100.00%)
Verschiedene Werte insgesamt: 1

Spalte: inboundarrivalairport
Top 5 Werte und ihr Anteil:
DUS: 21513370 mal (23.19%)
CGN: 9518179 mal (10.26%)
STR: 9501368 mal (10.24%)
FRA: 9410366 mal (10.14%)
HAM: 7689193 mal (8.29%)
Verschiedene Werte insgesamt: 48

Spalte: outbounddepartureairport
Top 5 Werte und ihr Anteil:
DUS: 21513370 mal (23.19%)
CGN: 9518179 mal (10.26%)
STR: 9501368 mal (10.24%)
FRA: 9410366 mal (10.14%)
HAM: 7689193 mal (8.29%)
Verschiedene Werte insgesamt: 48

Spalte: outboundarrivalairport
Top 5 Werte und ihr Anteil:
PMI: 92783430 mal (100.00%)
Verschiedene Werte insgesamt: 1

Spalte: mealtype
Top 5 Werte und ihr Anteil:
halfboard: 27196855 mal (29.31%)
allinclusive: 26355028 mal (28.40%)
breakfast: 24691553 mal (26.61%)
none: 9048278 mal (9.75%)
halfboardplus: 2392567 mal (2.58%)
Verschiedene Werte insgesamt: 13

Spalte: oceanview
Top 5 Werte und ihr Anteil:
False: 81124671 mal (87.43%)
True: 11658759 mal (12.57%)
Verschiedene Werte insgesamt: 2

Spalte: roomtype
Top 5 Werte und ihr Anteil:
double: 51803015 mal (55.83%)
apartment: 15255130 mal (16.44%)
family: 5394444 mal (5.81%)
juniorsuite: 4861063 mal (5.24%)
suite: 4628103 mal (4.99%)
Verschiedene Werte insgesamt: 19                           