package storage

import "sync"

var Mu sync.Mutex
var URLStore = make(map[string]string)
