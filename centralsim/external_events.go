package centralsim

type LookAhead struct {
	Process       int
	ExpectedEvent Event
}

// Rutina encargada de añadir eventos entrantes a la lista
func (se *SimulationEngine) manageIncommingEvents() {
	for {
		select {
		case event := <-se.incomEventsCh:
			se.mux.Lock()
			se.Log.Event.Println("Aqui debería meter el elemento en la lista!!!")
			se.IlEventos.inserta(event)
			se.Log.Event.Println(se.IlEventos)
			se.mux.Unlock()
		}
	}
}

// Rutina encargada de calcular el valor del LookAhead cuando lo requiere otro proceso
func (se *SimulationEngine) CalculateLookAhead() {
	for {
		select {
		case pid := <-se.receiveLookAheadReqCh:
			// TODO calc LookAhead using current transition and
			se.Log.Mark.Println("LookAhead Solicitado por", pid)
			ev := Event{}
			la := LookAhead{Process: pid, ExpectedEvent: ev}
			se.sendLookAheadCh <- la
		}
	}
}
