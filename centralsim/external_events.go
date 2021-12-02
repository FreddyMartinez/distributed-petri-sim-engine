package centralsim

import (
	"errors"
)

type LookAhead struct {
	Process int
	Time    TypeClock
}

// Estructura utilitaria, usada para enviar el Evento con el id del proceso que lo generó
type IncommingEvent struct {
	Event     Event
	ProcessId int
}

// Rutina encargada de añadir eventos entrantes a la lista
func (se *SimulationEngine) manageIncommingEvents() {
	for {
		select {
		case event := <-se.incomEventsCh:
			se.mux.Lock()
			se.IlEventos.inserta(event.Event)
			if se.lookAheads[event.ProcessId] < event.Event.IiTiempo {
				se.lookAheads[event.ProcessId] = event.Event.IiTiempo
			}

			if se.isWaitingEvent {
				se.isWaitingEvent = false
				se.waitForEvent <- true
			}
			se.mux.Unlock()
		}
	}
}

// Rutina encargada de calcular el valor del LookAhead cuando lo requiere otro proceso
func (se *SimulationEngine) CalculateLookAhead() {
	for {
		select {
		case lar := <-se.receiveLookAheadReqCh:
			se.mux.Lock() // Bloquea recursos para leer estado local

			la := LookAhead{Process: lar.Process}
			// si el proceso que solicita también puede enviar eventos, actualiza el look
			if current, ok := se.lookAheads[lar.Process]; ok && lar.Time > current {
				se.Log.Clock.Println("Actualiza LookAhead de proceso", lar.Process)
				se.lookAheads[lar.Process] = lar.Time
			}

			// si no hay transiciones sensibilizadas, da el tiempo futuro máximo
			if se.ilMislefs.IsTransSensib.isEmpty() {
				se.updateTimeWithLA() // actualiza el reloj local con el menor lookAhead
				// El máximo es el LookAhead mínimo más el tiempo que le tome a un token atravesar la red
				la.Time = se.iiRelojlocal + se.maxLookAhead
			} else {
				la.Time = se.iiRelojlocal + 1 // asume que el tiempo mínimo en que puede generar un evento externo es 1
			}
			se.mux.Unlock()

			se.sendLookAheadCh <- la
		}
	}
}

// Devuelve la transición de salida de la subred
func (se *SimulationEngine) findOutbound() (Transition, error) {
	for _, t := range se.ilMislefs.IaRed {
		if t.EsSalida {
			return t, nil
		}
	}
	return Transition{}, errors.New("Transición de salida no encontrada")
}

func (se *SimulationEngine) updateTimeWithLA() {
	minLookAhead := TypeClock(100) // se inicializa en 100 para reducirlo
	for i, l := range se.lookAheads {
		if l < minLookAhead {
			if l <= se.iiRelojlocal { // Si LookAhead no permite avanzar, se pide nuevo para ampliar más el tiempo
				se.Log.Clock.Println("LookAhead actual", l)
				go se.getLookAhead(i) // cuando se hace esto, los lookAhead se acumulan, generando problemas de sincronización
			}
			minLookAhead = l
		}
	}
	if minLookAhead > se.iiRelojlocal {
		se.iiRelojlocal = minLookAhead
		se.Log.Clock.Println("Avanza el tiempo -> ", se.iiRelojlocal)
	}
}

func getFutureEvent(constantsList [][2]int, inputTranId int) ([2]int, error) {
	for _, trCo := range constantsList { // Par transición / constante
		transitionId := -1 * (trCo[0] + 1) // Transforma el valor en un id global
		if transitionId == inputTranId {
			return trCo, nil
		}
	}
	return [2]int{}, errors.New("No se encontró la transición buscada")
}

func (se *SimulationEngine) getLookAhead(processId int) {
	// solicita Look Ahead
	la := LookAhead{Process: processId, Time: se.iiRelojlocal + 1}
	se.reqLookAheadCh <- la
	// espera Look Ahead
	ev := <-se.receiveLookAheadCh
	// Si el evento es de un tiempo mayor se inserta, si no, se vuelve a pedir
	se.mux.Lock()
	if ev.Time > se.lookAheads[ev.Process] { // puede ser menor si llega después de un evento
		se.lookAheads[ev.Process] = ev.Time
	}
	se.mux.Unlock()
}
