package centralsim

import (
	"fmt"
)

//--------------------------------------------------------------------------

// IndLocalTrans is a index of a transition in the local lefs list
type IndLocalTrans int

//TypeConst is the constant to propagate in lefs
type TypeConst int

// TransitionConstant is the pair transition and cte to propagate to
//type TransitionConstant struct {
//	ITrans IndLocalTrans
//	Cnstnt TypeConst
//}

//------------------------------------------------------------------------

// Transition : Tipo abstracto  para guardar la informacion de una transicion
type Transition struct {
	// IiIdLocal en la tabla global de transiciones
	IiIndLocal IndLocalTrans `json:"ii_idglobal"`

	// iiValorLef es el valor que tiene la funcion de
	// sensibilizacion en el instante de tiempo que nos da
	// la variable ii_tiempo
	IiValorLef TypeConst `json:"ii_valor"`
	IiTiempo   TypeClock `json:"ii_tiempo"`

	// tiempo que dura el disparo de la transicion
	IiDuracionDisparo TypeClock `json:"ii_duracion_disparo"`

	// vector con parejas :
	//		transicion junto con cte a actualizarle de forma inmediata
	TransConstIul [][2]int `json:"ii_listactes_IUL"`
	// vector con parejas :
	//		de transiciones a las que tengo que propagar cte
	// 		en el tiempo de disparo de esta transicion, junto con la cte que
	// 		tengo que propagar
	TransConstPul [][2]int `json:"ii_listactes_PUL"`

	// Tiempos de LookAhead
	// cada valor corresponde con las transiciones de la subred
	TiempoHastaMarca TiempoHasta `json:"iL_tiemposhastamarca"`

	EsSalida bool `json:"ib_desalida"`
}

type TiempoHasta struct {
	LiTiempos []int `json:"il_tiempos"`
}

// actualizaTiempo modifica el tiempo de la transicion dada
func (t *Transition) actualizaTiempo(aiTi TypeClock) {
	// Modificacion del tiempo
	t.IiTiempo = aiTi
}

// updateFuncValue modifica valor funcion de sensibilizacion de transicion dada
// RECIBE: Codigo de la transicion y valor con el que modificar
//		OJO, no es el valor definitivo, sino la CTE a añadir al valor que tenia
//		antes la funcion
func (t *Transition) updateFuncValue(aiValLef TypeConst) {
	// Modificacion del valor de la funcion lef
	t.IiValorLef += aiValLef
}

// Imprime los atributos de una transicion para depurar errores
func (t *Transition) Imprime(log *Logger) {
	log.Tansition.Println("ID GLOBAL: ", t.IiIndLocal)
	log.NoFmtLog.Println(
		fmt.Sprintf("VALOR LEF: %v TIEMPO: %v DURACION DISPARO: %v", t.IiValorLef, t.IiTiempo, t.IiDuracionDisparo))
	log.NoFmtLog.Println(" LISTA DE CTES IUL: ")
	for _, v := range t.TransConstIul {
		log.NoFmtLog.Println("\tTRANSICION: ", v[0], "\t\tCTE: ", v[1])
	}
	log.NoFmtLog.Println(" LISTA DE CTES PUL: ")
	for _, v := range t.TransConstPul {
		log.NoFmtLog.Println("\tTRANSICION: ", v[0], "\t\tCTE: ", v[1])
	}
	if t.EsSalida {
		log.NoFmtLog.Println("\tTIEMPOS LOOKAHEAD: ", t.TiempoHastaMarca.LiTiempos)
	}
}

// ImprimeValores de la transición
func (t *Transition) ImprimeValores() {
	fmt.Println("Transicion -> ")
	fmt.Println("\tIDLOCALTRANSICION: ", t.IiIndLocal)
	fmt.Println("\t\tVALOR LEF: ", t.IiValorLef)
	fmt.Println("\t\tTIEMPO: ", t.IiTiempo)
}

//--------------------------------------------------------------------------

// TransitionList is a list of transitions themselves
type TransitionList []Transition //Slice de transiciones como Lista

// length return length of ListTransitions with type adapted to IndLocalTrans
func (lt TransitionList) length() IndLocalTrans {
	return IndLocalTrans(len(lt))
}

func (lt TransitionList) findIndex(id IndLocalTrans) IndLocalTrans {
	for i, t := range lt {
		if t.IiIndLocal == id {
			return IndLocalTrans(i)
		}
	}
	return -1
}

//----------------------------------------------------------------------

// TransitionStack is a Stack of transition indices
type TransitionStack []IndLocalTrans

// MakeTransitionStack crea lista de tamaño aiLongitud
func MakeTransitionStack() TransitionStack {
	// cero length and capacidad capacity
	return make(TransitionStack, 0)
}

// push transition id to stack
func (st *TransitionStack) push(iTr IndLocalTrans) {
	*st = append(*st, iTr)
}

// pop transition id from stack
func (st *TransitionStack) pop() IndLocalTrans {
	if (*st).isEmpty() {
		return -1
	}

	iTr := (*st)[len(*st)-1] // obtener dato de lo alto de la pila
	*st = (*st)[:len(*st)-1] //desempilar

	return iTr
}

// isEmpty  the transition stack ?
func (st TransitionStack) isEmpty() bool {
	return len(st) == 0
}

func (st TransitionStack) ImprimeTransStack(log *Logger) {
	if st.isEmpty() {
		log.NoFmtLog.Println("\tStack TRANSICIONES VACIA")
	} else {
		for _, iTr := range st {
			log.NoFmtLog.Println("\t\t\t", iTr)
		}
	}
}
