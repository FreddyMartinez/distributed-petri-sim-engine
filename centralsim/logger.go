package centralsim

import (
	"log"
	"os"

	"github.com/DistributedClocks/GoVector/govec"
)

/*
Estructura usada para escribir en archivos de log locales
*/
type Logger struct {
	NoFmtLog  *log.Logger
	Tansition *log.Logger
	Event     *log.Logger
	Mark      *log.Logger
	Clock     *log.Logger
	GoVec     *govec.GoLog
}

func CreateLogger(processId string) *Logger {
	file, err := os.OpenFile("./logs/"+processId+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(file, processId+"\t", log.Ltime)
	transition := log.New(file, processId+"\tTRANSITION:\t\t", log.Ltime)
	event := log.New(file, processId+"\tEVENT: \t\t", log.Ltime)
	mark := log.New(file, processId+"\tLOOKAHEAD: \t\t", log.Ltime)
	clock := log.New(file, processId+"\tSIM CLOCK: \t\t", log.Lmicroseconds)

	defaultConfig := govec.GetDefaultConfig()
	defaultConfig.UseTimestamps = true
	goVector := govec.InitGoVector(processId, "/logs/govector/"+processId, defaultConfig)

	return &Logger{Event: event, NoFmtLog: logger, Tansition: transition, Mark: mark, GoVec: goVector, Clock: clock}
}

func (log *Logger) GoVectLog(message string) {
	log.GoVec.LogLocalEvent(message, govec.GetDefaultLogOptions())
}
