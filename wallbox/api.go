package wallbox

// ChargeStatus is the EV's charging status from A to F
type ChargeStatus string

// Charging states; Status D bis F aktuell nicht verwendet
const (
	StatusNotConfig ChargeStatus = "NC"
	StatusNone      ChargeStatus = ""
	StatusAvailable ChargeStatus = "A" // Fzg. angeschlossen: nein    Laden aktiv: nein    Ladestation betriebsbereit, Fahrzeug getrennt
	StatusOccupied  ChargeStatus = "B" // Fzg. angeschlossen:   ja    Laden aktiv: nein    Fahrzeug verbunden, Netzspannung liegt nicht an
	StatusCharging  ChargeStatus = "C" // Fzg. angeschlossen:   ja    Laden aktiv:   ja    Fahrzeug lädt, Netzspannung liegt an
	StatusD         ChargeStatus = "D" // Fzg. angeschlossen:   ja    Laden aktiv:   ja    Fahrzeug lädt mit externer Belüftungsanforderung (für Blei-Säure-Batterien)
	StatusE         ChargeStatus = "E" // Fzg. angeschlossen:   ja    Laden aktiv: nein    Fehler Fahrzeug / Kabel (CP-Kurzschluss, 0V)
	StatusF         ChargeStatus = "F" // Fzg. angeschlossen:   ja    Laden aktiv: nein    Fehler EVSE oder Abstecken simulieren (CP-Wake-up, -12V)
)

func statusTextWithStatus(status ChargeStatus) string {
	switch status {
	case StatusAvailable:
		return "Verfügbar"
	case StatusOccupied:
		return "Belegt"
	case StatusCharging:
		return "Laden"
	}
	return "Unbekannt"
}
