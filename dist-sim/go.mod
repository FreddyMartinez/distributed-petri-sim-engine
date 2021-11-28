module petrisim

go 1.17

replace centralsim => ../centralsim

require (
	centralsim v0.0.0-00010101000000-000000000000
	github.com/DistributedClocks/GoVector v0.0.0-20210402100930-db949c81a0af
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

require (
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.1.4 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)
