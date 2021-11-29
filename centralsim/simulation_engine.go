/*Package centralsim with several files to offer a centralized simulation
PROPOSITO: Tipo abstracto para realizar la simulacion de una (sub)RdP.
COMENTARIOS:
	- El resultado de una simulacion local sera un slice dinamico de
	componentes, de forma que cada una de ella sera una structura estatica de
	dos enteros, el primero de ellos sera el codigo de la transiciondisparada y
	el segundo sera el valor del reloj local para el que se disparo.
*/
package centralsim

import (
	"fmt"
	"sync"
	"time"
)

// TypeClock defines integer size for holding time.
type TypeClock int64

// ResultadoTransition holds fired transition id and time of firing
type ResultadoTransition struct {
	CodTransition     IndLocalTrans
	ValorRelojDisparo TypeClock
}

// SimulationEngine is the basic data type for simulation execution
type SimulationEngine struct {
	iiRelojlocal          TypeClock             // Valor de mi reloj local
	ilMislefs             Lefs                  // Estructura de datos del simulador
	IlEventos             EventList             //Lista de eventos a procesar
	ivTransResults        []ResultadoTransition // slice dinamico con los resultados
	EventNumber           float64               // cantidad de eventos ejecutados
	Log                   *Logger
	lookAheads            map[int]TypeClock // usado para saber los lookAheads de los procesos precedentes
	maxLookAhead          TypeClock         // se usa para dar lookAhead cuando no hay transiciones sensibilizadas
	sendEventCh           chan Event
	incomEventsCh         chan IncommingEvent // Recibe eventos de otros procesos
	reqLookAheadCh        chan int            // Canal para solicitar lookAhead al proceso indicado
	receiveLookAheadCh    chan LookAhead      // Canal para recibir LookAhead de proceso precedente
	receiveLookAheadReqCh chan int            // Recibe solicitud de LookAhead de proceso posterior
	sendLookAheadCh       chan LookAhead      // Envía LookAhead propio a proceso posterior
	isWaitingEvent        bool
	waitForEvent          chan bool
	mux                   sync.Mutex
}

// MakeSimulationEngine : inicializar SimulationEngine struct
func MakeSimulationEngine(
	alLaLef Lefs,
	logger *Logger,
	sendEv chan Event,
	incomingEvent chan IncommingEvent,
	requestLookAhead chan int,
	receiveLACh chan LookAhead,
	receiveLAReqCh chan int,
	sendLookAheadCh chan LookAhead,
	lookAheads map[int]TypeClock,
	maxLookAhead TypeClock,
) *SimulationEngine {

	m := SimulationEngine{}

	m.iiRelojlocal = 0
	m.ilMislefs = alLaLef
	m.IlEventos = make(EventList, 0)
	m.ivTransResults = make([]ResultadoTransition, 0)
	m.EventNumber = 0
	m.Log = logger
	m.sendEventCh = sendEv
	m.incomEventsCh = incomingEvent
	m.reqLookAheadCh = requestLookAhead
	m.receiveLookAheadCh = receiveLACh
	m.receiveLookAheadReqCh = receiveLAReqCh
	m.sendLookAheadCh = sendLookAheadCh
	m.lookAheads = lookAheads
	m.maxLookAhead = maxLookAhead
	m.isWaitingEvent = false
	m.waitForEvent = make(chan bool)
	m.mux = sync.Mutex{}

	m.Log.NoFmtLog.Println("Motor de simulación creado")

	go m.manageIncommingEvents()
	go m.CalculateLookAhead()

	return &m
}

// disparar una transicion. Esto es, generar todos los eventos
//	   ocurridos por el disparo de una transicion
//   RECIBE: Indice en el vector de la transicion a disparar
func (se *SimulationEngine) dispararTransicion(ilTr IndLocalTrans) {
	// Prepare 5 local variables
	trList := se.ilMislefs.IaRed              // transition list
	timeTrans := trList[ilTr].IiTiempo        // time to spread to new events
	timeDur := trList[ilTr].IiDuracionDisparo // firing time length
	listIul := trList[ilTr].TransConstIul     // Iul list of pairs Trans, Ctes
	listPul := trList[ilTr].TransConstPul     // Pul list of pairs Trans, Ctes

	// First apply Iul propagations (Inmediate : 0 propagation time)
	for _, trCo := range listIul {
		localIndex := se.ilMislefs.IaRed.findIndex(IndLocalTrans(trCo[0])) // Encontrar id de transición
		trList[localIndex].updateFuncValue(TypeConst(trCo[1]))
	}

	// Generamos eventos ocurridos por disparo de transicion ilTr
	for _, trCo := range listPul {
		// tiempo = tiempo de la transicion + coste disparo
		se.IlEventos.inserta(Event{timeTrans + timeDur,
			IndLocalTrans(trCo[0]),
			TypeConst(trCo[1])})
	}
}

/* fireEnabledTransitions dispara todas las transiciones sensibilizadas
   		PROPOSITO: Accede a lista de transiciones sensibilizadas y procede con
	   	su disparo, lo que generara nuevos eventos y modificara el marcado de
		transicion disparada. Igualmente anotara en resultados el disparo de
		cada transicion para el reloj actual dado
*/
func (se *SimulationEngine) fireEnabledTransitions(aiLocalClock TypeClock) {
	for se.ilMislefs.haySensibilizadas() { //while
		liCodTrans := se.ilMislefs.getSensibilizada()
		se.dispararTransicion(liCodTrans)
		se.Log.GoVectLog(fmt.Sprintf("Dispara transición %v", liCodTrans))

		// Anotar el Resultado que disparo la liCodTrans en tiempoaiLocalClock
		se.ivTransResults = append(se.ivTransResults, ResultadoTransition{se.ilMislefs.IaRed[liCodTrans].IiIndLocal, aiLocalClock})
	}
}

// tratarEventos : Accede a lista eventos y trata todos con tiempo aiTiempo
func (se *SimulationEngine) tratarEventos() {
	var leEvento Event
	aiTiempo := se.iiRelojlocal

	se.Log.Event.Println("Tratar Eventos", se.IlEventos)
	for se.IlEventos.hayEventos(aiTiempo) {
		leEvento = se.IlEventos.popPrimerEvento() // extraer evento más reciente

		idTr := leEvento.IiTransicion // obtener transición del evento
		trList := se.ilMislefs.IaRed  // obtener lista de transiciones de Lefs

		if idTr < 0 { // Enviar evento a la transición correspondiente
			se.sendEventCh <- leEvento
		} else {
			idTr := trList.findIndex(idTr) // Encontrar el índice
			// Establecer nuevo valor de la funcion
			trList[idTr].updateFuncValue(leEvento.IiCte)
			// Establecer nuevo valor del tiempo
			trList[idTr].actualizaTiempo(leEvento.IiTiempo)
			se.Log.GoVectLog(fmt.Sprintf("Evento local %v", se.EventNumber))
		}

		se.EventNumber++
	}
}

// avanzarTiempo : Modifica reloj local con minimo tiempo de entre
//	   recibidos del exterior o del primer evento en lista de eventos
func (se *SimulationEngine) avanzarTiempo() TypeClock {
	nextTime := se.IlEventos.tiempoPrimerEvento()
	// Compara el menor tiempo del siguiente evento con el menor LookAhead
	for _, l := range se.lookAheads {
		if l < nextTime {
			nextTime = l
		}
	}
	se.Log.Clock.Println("NEXT CLOCK...... : ", nextTime)
	return nextTime
}

// devolverResultados : Mostrar los resultados de la simulacion
func (se *SimulationEngine) devolverResultados() {
	resultados := "----------------------------------------\n"
	resultados += "Resultados del simulador local\n"
	resultados += "----------------------------------------\n"
	if len(se.ivTransResults) == 0 {
		resultados += "No esperes ningun resultado...\n"
	}

	for _, liResult := range se.ivTransResults {
		resultados +=
			"TIEMPO: " + fmt.Sprintf("%v", liResult.ValorRelojDisparo) +
				" -> TRANSICION: " + fmt.Sprintf("%v", liResult.CodTransition) + "\n"
	}

	resultados += "\n ========== TOTAL DE TRANSICIONES DISPARADAS = " +
		fmt.Sprintf("%d", len(se.ivTransResults)) + "\n"

	fmt.Println(resultados)
}

// SimularUnpaso de una RdP con duración disparo >= 1
func (se *SimulationEngine) simularUnpaso() {
	se.ilMislefs.actualizaSensibilizadas(se.iiRelojlocal)
	// Si no hay transiciones sensibilizadas ni eventos por procesar, espera evento
	if !se.ilMislefs.haySensibilizadas() && se.IlEventos.ListaEventosVacia() {
		// Espera evento
		se.Log.NoFmtLog.Println("ESPERA EVENTO")
		se.isWaitingEvent = true
		<-se.waitForEvent
	}
	// si los eventos son de tiempo menor a los lookahead, los procesa, si no, pide lookahead
	for i, l := range se.lookAheads {
		se.Log.Mark.Println(fmt.Sprintf("LOOKAHEAD P%v ACTUAL: %v", i, l))
		if l < se.IlEventos.tiempoPrimerEvento() {
			se.getLookAhead(i)
		}
	}
	se.mux.Lock()

	se.Log.NoFmtLog.Println("-----------Stack de transiciones sensibilizadas---------")
	se.ilMislefs.IsTransSensib.ImprimeTransStack(se.Log)
	se.Log.NoFmtLog.Println("-----------Final Stack de transiciones---------")

	// Fire enabled transitions and produce events
	se.fireEnabledTransitions(se.iiRelojlocal)

	se.Log.NoFmtLog.Println("-----------Lista eventos después de disparos---------")
	se.IlEventos.Imprime(se.Log)
	se.Log.NoFmtLog.Println("-----------Final lista eventos---------")

	if !se.IlEventos.ListaEventosVacia() {
		// Cuando hay eventos futuros
		se.iiRelojlocal = se.avanzarTiempo()
		se.Log.Clock.Println("Avanza el tiempo -> ", se.iiRelojlocal)
		se.Log.GoVectLog(fmt.Sprintf("Avanza el tiempo -> %v", se.iiRelojlocal))
	}
	se.tratarEventos()
	se.mux.Unlock()
}

// SimularPeriodo de una RdP
// RECIBE: - Ciclo inicial (por si marcado recibido no se corresponde al
//				inicial sino a uno obtenido tras simular ai_cicloinicial ciclos)
//		   - Ciclo con el que terminamos
func (se *SimulationEngine) SimularPeriodo(CicloInicial, CicloFinal TypeClock) {
	ldIni := time.Now()

	// Inicializamos el reloj local
	// ------------------------------------------------------------------
	se.iiRelojlocal = CicloInicial

	for se.iiRelojlocal < CicloFinal {
		///*		//DEPURACION
		se.Log.NoFmtLog.Println("RELOJ LOCAL !!!  = ", se.iiRelojlocal)
		// se.ilMislefs.ImprimeLefs(se.Log)
		//*/
		se.simularUnpaso()
	}

	elapsedTime := time.Since(ldIni)

	fmt.Printf("Eventos por segundo = %f",
		se.EventNumber/elapsedTime.Seconds())

	/*	// Devolver los resultados de la simulacion
		se.devolverResultados()
		result := "\n---------------------\n"
		result += "TIEMPO SIMULADO en ciclos: " +
			fmt.Sprintf("%d", Nciclos-CicloInicial) + "\n"
		result += "TIEMPO ejecución REAL simulación: " +
			fmt.Sprintf("%v", elapsedTime.String()) + "\n"
		fmt.Println(result)
	*/
}
