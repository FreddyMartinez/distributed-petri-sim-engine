package centralsim

import (
	"errors"
	"fmt"
)

type LookAhead struct {
	Process int
	Time    TypeClock
}

// Rutina encargada de añadir eventos entrantes a la lista
func (se *SimulationEngine) manageIncommingEvents() {
	for {
		select {
		case event := <-se.incomEventsCh:
			se.mux.Lock()
			se.IlEventos.inserta(event)
			if se.isWaitingEvent {
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

			la := LookAhead{Process: lar}
			// si no hay transiciones sensibilizadas, da el lookAhead máximo, el cual es el tiempo mínimo que le tome a un token atravesar la red
			if se.ilMislefs.IsTransSensib.isEmpty() {
				la.Time = se.iiRelojlocal + se.maxLookAhead
			} else {
				la.Time = se.iiRelojlocal + 1 // asume que el tiempo mínimo en que puede generar un evento externo es 1
			}
			se.mux.Unlock()
			se.Log.Mark.Println(fmt.Sprintf("Envía LookAhead a P%v, con tiempo %v", la.Process, la.Time))
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
	se.reqLookAheadCh <- processId
	// espera Look Ahead
	ev := <-se.receiveLookAheadCh
	// Si el evento es de un tiempo igual o mayor se inserta, si no, se vuelve a pedir
	if ev.Time == -1 || ev.Time >= se.iiRelojlocal {
		se.mux.Lock()
		se.lookAheads[ev.Process] = ev.Time
		se.mux.Unlock()
	} else {
		se.getLookAhead(processId)
	}
}
