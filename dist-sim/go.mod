module petrisim

go 1.17

replace centralsim => ../centralsim

require (
	centralsim v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

require golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
